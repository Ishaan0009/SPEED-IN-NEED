//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"runtime"
	"speedinneed/engine"
	"strings"
	"time"
	"unsafe"
)

type remasterGlyph struct{ U, V, W, Advance, DU, DV, Cell, Pad float64 }
type remasterInstance struct {
	Data           [][8]float32
	ObjectIDs      []int
	Key, Kind      string
	X, Z, Radius   float64
	VBO            uint32
	Count          int
	Fade           float64
	TopY034        float64
	GrassSource051 []byte
	Copies051      int
	LODFade052     float64
	LODSide052     int
	Packed052      []float32
	Merged052      bool
}
type remasterTile struct {
	Key          string
	X, Z         float64
	Radius       float64
	RoadCorridor bool
}
type remasterParticle struct {
	Pos, Vel            pos3
	Life, MaxLife, Size float64
	Exhaust             bool
}

var remaster = struct {
	Enabled, Ready, QA               bool
	Mode, Time, Laps, Opponents, Row int
	Definitions                      [2]*engine.WorldDefinition
	Loaded                           [2]bool
	Glyphs                           map[string]remasterGlyph
	Atlas                            uint32
	Instances                        []remasterInstance
	Tiles                            []remasterTile
	Standalone, Lamps                []int
	BufferMode                       int
	RoadMeshes                       []worldBatch
	RoadTiles                        []remasterTile
	Drag, Dragged                    bool
	DownX, DownY, MouseX, MouseY     int
	Orbit, Pitch, Idle               float64
	CameraVelocity                   pos3
	CameraAnchor                     pos3
	CameraReset                      bool
	YawLoc, InstanceLoc              int32
	Particles                        [320]remasterParticle
	ParticleCursor                   int
	SmokeClock                       float64
	Status                           string
	FullMap, SceneReady              bool
	Loading                          bool
	SettingsTab                      int
	HeightTexture                    uint32
}{Enabled: true, BufferMode: -1, CameraReset: true, Laps: 3, Opponents: 4}

func initializeRemasterSettings() {
	remaster.Mode = engine.ClampInt(beta.cfg.GameMode, 0, 1)
	remaster.Time = engine.ClampInt(beta.cfg.TimeOfDay-1, 0, 2)
	remaster.Laps = engine.ClampInt(beta.cfg.RaceLaps, 1, 8)
	remaster.Opponents = engine.ClampInt(beta.cfg.Opponents, 0, 7)
	beta.mapSelect = remaster.Mode
	beta.mode = remaster.Mode
	app.lightingOn = true
	app.boolFPP = beta.cfg.CameraMode == 0
}
func saveRemasterSettings() {
	beta.cfg.GameMode = remaster.Mode
	beta.cfg.TimeOfDay = remaster.Time + 1
	beta.cfg.RaceLaps = remaster.Laps
	beta.cfg.Opponents = remaster.Opponents
	saveCfg()
}
func initRemasterGraphics() {
	initBeta020Sky()
	initCharacters032()
	raw, _ := assets.ReadFile("assets/b037_ui_font_sdf.json")
	json.Unmarshal(raw, &remaster.Glyphs)
	remaster.Atlas = app.tex["b037_font_fallback.png"]
	initFont035()
	useLighting(true)
	sceneInt("uTerrainRock", 8)
	pbrTexture("b035_rock_color.jpg")
	glExt("glActiveTexture", 0x84C8)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(pbrTexture("b035_rock_color.jpg")))
	glExt("glActiveTexture", 0x84C0)
	// Decode replacement road textures before driving, avoiding first-use I/O
	// in the first world frame after leaving the menus.
	pbrTexture("b034_asphalt_lanes.jpg")
	pbrTexture("b034_asphalt_normal.jpg")
	useLighting(false)
	for _, key := range []string{"b051_grass0", "b051_grass1", "b033_garage", "b012_title_road", "b012_avenue_tree", "b012_building00", "rm_clouds", "rm_showroom", "rm_car2", "rm_car3", "rm_dashboard", "rm_steering", "rm_npc0", "rm_npc1", "rm_npc2", "rm_npc3", "rm_flame", "rm_nitro"} {
		loadPBRAsset(key)
	}
	for _, key := range []string{"ferrari", "porsche", "rm_car2", "rm_car3"} {
		loadPBRAsset(key + "_lod")
	}
	for i := 0; i < 4; i++ {
		loadPBRAsset(fmt.Sprintf("b022_cabin_%d", i))
		loadPBRAsset(fmt.Sprintf("b033_cabin_%d", i))
	}
	beta011InstanceScaleLoc = int32(glExt("glGetAttribLocation", uintptr(lightPipeline.program), uintptr(unsafe.Pointer(&[]byte("aInstanceScale\x00")[0]))))
	remaster.InstanceLoc = int32(glExt("glGetAttribLocation", uintptr(lightPipeline.program), uintptr(unsafe.Pointer(&[]byte("aInstance\x00")[0]))))
	remaster.YawLoc = int32(glExt("glGetAttribLocation", uintptr(lightPipeline.program), uintptr(unsafe.Pointer(&[]byte("aInstanceYaw\x00")[0]))))
	gpuWorld.instanceLoc = remaster.InstanceLoc
	gpuWorld.instancing = glExtAvailable("glDrawElementsInstanced") && glExtAvailable("glVertexAttribDivisor")
	remaster.Ready = true
	for _, key := range []string{"b011_sun", "b011_moon", "b011_npc0", "b011_npc1", "b011_npc2", "b011_npc3", "b011_npc4", "b011_npc5", "b011_npc6", "b011_npc7", "b011_npc8", "b011_npc9"} {
		loadPBRAsset(key)
	}
}
func remasterDefinition(mode int) *engine.WorldDefinition {
	mode = engine.ClampInt(mode, 0, 1)
	if remaster.Definitions[mode] != nil {
		return remaster.Definitions[mode]
	}
	raw, e := assets.ReadFile([]string{"assets/b038_world_free.json", "assets/b038_world_race.json"}[mode])
	if e != nil {
		remaster.Status = e.Error()
		return nil
	}
	d, e := engine.DecodeWorldDefinition(raw)
	if e != nil {
		remaster.Status = e.Error()
		return nil
	}
	remaster.Definitions[mode] = d
	return d
}
func activateRemasteredWorld(restart bool) {
	remaster.Loading = true
	defer func() { remaster.Loading = false }()
	drawAokawaLoading(.05, "PREPARING MAP")
	d := remasterDefinition(remaster.Mode)
	if d == nil {
		return
	}
	if !restart && app.world.Remaster != nil && app.mapID == remaster.Mode {
		app.world.Remaster.Time = remaster.Time
		return
	}
	if remaster.Mode == 0 && !workshop051.Active {
		prepareTerrain051(d)
	}
	remaster.SceneReady = false
	autopilot034 = false
	motion034.Ready = false
	remaster.FullMap = false
	app.world = engine.NewRemasteredWorld(d, remaster.Mode, beta.carSelect, remaster.Time, remaster.Laps, remaster.Opponents)
	app.world.SimulationLOD050 = true
	app.world.SetRaceDifficulty051(beta.cfg.RaceDifficulty040)
	initializeNavigation033(d)
	app.world.SetTrafficDensity(beta.cfg.TrafficDensity)
	beta011Skids.Marks = nil
	beta011Skids.Valid = false
	app.mapID = remaster.Mode
	beta.mapSelect = remaster.Mode
	beta.mode = remaster.Mode
	app.prevCar = app.world.Car
	app.renderCar = app.world.Car
	app.renderAlpha = 1
	beta.cameraHeading = app.world.Car.Heading
	app.camYaw = 0
	app.camPitch = 0
	remaster.CameraReset = true
	if !remaster.Loaded[remaster.Mode] {
		keys := map[string]bool{}
		for _, o := range d.Objects {
			if d.Beta026 && (o.Asset == "b022r_lane_paint" || o.Asset == "b021_crosswalks" || o.Asset == "b011_crosswalks") {
				continue
			}
			keys[o.Asset] = true
		}
		for _, k := range d.TerrainTiles {
			keys[k] = true
		}
		loaded := 0
		for k := range keys {
			if pbrAssets[k] == nil {
				loadPBRAsset(k)
			}
			loaded++
			if loaded%8 == 0 {
				drawAokawaLoading(.25+.55*float64(loaded)/float64(len(keys)), "LOADING MAP ASSETS")
			}
		}
		remaster.Loaded[remaster.Mode] = true
	}
	if remaster.Mode == 0 && !workshop051.Active {
		if pbrAssets["b035_grass0_far_lod"] == nil {
			loadPBRAsset("b035_grass0_far_lod")
			loadPBRAsset("b035_grass1_far_lod")
		}
		rootRocks051(d)
	}
	if remaster.BufferMode != remaster.Mode {
		drawAokawaLoading(.85, "BUILDING WORLD BUFFERS")
		buildRemasterWorldBuffers(d)
		uploadAokawaHeightfield(d)
		remaster.BufferMode = remaster.Mode
	}
	runtime.GC() // Free upload/decode buffers on the loading screen, before driving.
}
func startRemasteredDrive() {
	if multiplayer050.Session != nil && !multiplayer050.Session.IsHost() {
		multiplayer050.Message = "Only the host can change the session world"
		multiplayer050.Open = true
		return
	}
	photo041.Active, photo041.Drag = false, false
	walker051 = engine.Walker051{}
	if !validateProgressSelection040() {
		return
	}
	if beta.cfg.RandomTime040 {
		startRandomTime040()
	}
	activateRemasteredWorld(true)
	if app.world.Remaster == nil {
		return
	}
	applySetupSpawnRefine()
	if multiplayer050.Session != nil && multiplayer050.Session.IsHost() {
		multiplayer050.Epoch = uint64(time.Now().UnixMilli())
		multiplayer050.Race = map[string]*racePeer050{}
	}
	app.started = true
	app.paused = false
	selectScreen(screenDrive)
	app.lightingOn = true
	saveRemasterSettings()
}
func remasterWndProc(hwnd uintptr, m uint32, w, l uintptr) bool {
	if remaster.Loading && (m == WM_KEYDOWN || m == WM_LBUTTONDOWN || m == WM_RBUTTONDOWN || m == WM_MOUSEMOVE) {
		return true
	}
	if !remaster.Enabled {
		return false
	}
	if photoPointer041(m, w, l) {
		return true
	}
	if m == 0x215 || m == 0x08 {
		remaster.Drag = false
		app.camHeld = false
	}
	if m == 0x06 && w&65535 == 0 && !qa.enabled && beta.screen == screenDrive {
		remaster.CameraReset = true
		selectScreen(screenPause)
	}
	if m == WM_KEYDOWN {
		if (l>>30)&1 != 0 && !(beta.screen != screenDrive && w >= 0x25 && w <= 0x28) {
			return true
		}
		return remasterKey(w)
	}
	x, y := int(int16(l&65535)), int(int16((l>>16)&65535))
	if m == 0x20A && beta.screen == screenSettings {
		n := len(settingNames[remaster.SettingsTab])
		delta := int(int16(w>>16)) / 120
		beta.selection = engine.ClampInt(beta.selection-delta*2, 0, n-1)
		return true
	}
	if m == 0x20A && garageView037() {
		garageZoom037(-float64(int16(w>>16)) * .0007)
		return true
	}
	if m == 0x20A && beta.screen == screenDrive {
		if !remaster.FullMap {
			zoomCamera036(-float64(int16(w>>16)) * .0035)
		}
		return true
	}
	if beta.screen == screenDrive {
		return false
	}
	return beta014MenuPointer(hwnd, m, x, y)
}

func remasterKey(k uintptr) bool {
	if beta.screen == screenDrive && !garage050.Open && !multiplayer050.Open && k == 'E' {
		if multiplayer050.Session == nil && app.world.ToggleWalker051(&walker051) {
			remaster.CameraReset = true
			app.boolFPP = false
		}
		return true
	}
	if workshop051.Active && k == 0x75 {
		remaster.Time = (remaster.Time + 1) % 3
		app.world.Remaster.Time = remaster.Time
		return true
	}
	if photoKey041(k) {
		return true
	}
	if k != 0x7a && beta020MenuKey(k) {
		return true
	}
	if k == 0x7b && beta.screen == screenDrive {
		photo020Pending = true
		return true
	}
	if k == 0x7A {
		toggleFullscreen()
		saveRemasterSettings()
		return true
	}
	if beta.screen == screenSettings {
		return aokawaSettingsKey(k)
	}
	if false {
		if k == VK_ESCAPE && beta.bindWaiting < 0 {
			remaster.Time = engine.ClampInt(beta.cfg.TimeOfDay-1, 0, 2)
			if app.world.Remaster != nil {
				app.world.Remaster.Time = remaster.Time
			}
			selectScreen(beta.prevScreen)
			saveRemasterSettings()
			return true
		}
		return betaInput(k)
	}
	if beta.screen == screenDrive {
		if remaster.FullMap {
			return navMapKey033(k)
		}
		switch k {
		case 'P':
			toggleAutopilot034()
		case VK_ESCAPE:
			remaster.CameraReset = true
			selectScreen(screenPause)
		case 'F':
			cycleRadio030()
		case 'V':
			beta.cfg.CameraMode = (beta.cfg.CameraMode + 1) % 4
			app.boolFPP = beta.cfg.CameraMode == 0
			app.camYaw = 0
			app.camPitch = 0
			remaster.CameraReset = true
			saveRemasterSettings()
		case 'H':
			beta.cfg.Headlights = !beta.cfg.Headlights
			saveRemasterSettings()
		case 'R':
			app.world.RecoverRemasteredCar()
			app.camYaw = 0
			app.camPitch = 0
			app.prevCar = app.world.Car
			remaster.CameraReset = true
		case 'M':
			remaster.FullMap = !remaster.FullMap
			if remaster.FullMap {
				beta011MapOpened = time.Now()
			}
			app.camHeld = false
			win(user32, "ReleaseCapture")
			saveRemasterSettings()
		case 'C':
			togglePhoto041()
			return true
		case 0x70:
			app.showHelp = !app.showHelp
		case 0x71:
			app.showPerf = !app.showPerf
		case 0x78:
			toggleBenchmark()
		}
		return true
	}
	if beta.screen == screenTitle {
		if k == VK_ESCAPE {
			app.run = false
		} else if k == VK_RETURN || k == VK_SPACE {
			selectScreen(screenMaps)
		}
		return true
	}
	if beta.screen == screenResults {
		if k == VK_RETURN || k == VK_ESCAPE {
			app.started = false
			selectScreen(screenMaps)
		}
		return true
	}
	if beta.screen == screenPause {
		if k == VK_ESCAPE || k == 'P' {
			remaster.CameraReset = true
			selectScreen(screenDrive)
			return true
		}
		if k == 0x26 {
			beta.selection = (beta.selection + 3) % 4
		}
		if k == 0x28 {
			beta.selection = (beta.selection + 1) % 4
		}
		if k == VK_RETURN {
			remasterPauseAction(beta.selection)
		}
		return true
	}
	rows := 6
	if remaster.Mode == 0 {
		rows = 4
	}
	switch k {
	case VK_ESCAPE:
		selectScreen(screenTitle)
	case 0x26:
		remaster.Row = (remaster.Row + rows - 1) % rows
	case 0x28:
		remaster.Row = (remaster.Row + 1) % rows
	case 0x25:
		remasterCycle(remaster.Row, -1)
	case 0x27:
		remasterCycle(remaster.Row, 1)
	case 'Q':
		remasterCycle(3, -1)
	case 'E':
		remasterCycle(3, 1)
	case VK_RETURN:
		aokawaAdvanceSetup()
	}
	return true
}
func remasterPauseAction(row int) {
	switch row {
	case 0:
		remaster.CameraReset = true
		selectScreen(screenDrive)
	case 1:
		beta.prevScreen = screenPause
		selectScreen(screenSettings)
	case 2:
		app.started = false
		remaster.FullMap = false
		remaster.Row = 0
		selectScreen(screenMaps)
	case 3:
		app.run = false
	}
}
func remasterPaintCount() int { return visiblePaintCount040() }

func remasterCycle(row, delta int) {
	switch row {
	case 0:
		remaster.Mode = (remaster.Mode + 1) % 2
		if p := activeProfile040(); p != nil && !p.CanRace040() {
			remaster.Mode = 0
			progressGate040("CIRCUIT RACING / LOCKED UNTIL LEVEL 10")
		}
		beta.mode = remaster.Mode
		beta.mapSelect = remaster.Mode
	case 1:
		beta.carSelect = nextProgressCar040(beta.carSelect, delta)
		if beta.carSelect == 3 {
			beta.paint = 0
		} else {
			beta.paint %= remasterPaintCount()
		}
	case 2:
		beta.cfg.RandomTime040 = false
		remaster.Time = (remaster.Time + 3 + delta) % 3
	case 3:
		n := remasterPaintCount()
		if n > 0 {
			beta.paint = (beta.paint + n + delta) % n
			saveSetup050(setup050())
		}
	case 4:
		remaster.Laps = 1 + (remaster.Laps-1+8+delta)%8
	case 5:
		remaster.Opponents = (remaster.Opponents + 8 + delta) % 8
	}
	saveRemasterSettings()
}
func remasterClick(x, y int) {
	if beta020Click(x, y) {
		return
	}
	fx, fy := float64(x)/float64(app.width), float64(y)/float64(app.height)
	if beta.screen == screenTitle {
		selectScreen(screenMaps)
		return
	}
	if beta.screen == screenSettings {
		aokawaSettingsClick(x, y)
		return
	}
	if fx > .91 && fy < .13 && beta.screen != screenPause {
		beta.prevScreen = beta.screen
		selectScreen(screenSettings)
		return
	}
	if beta.screen == screenPause {
		for i := 0; i < 4; i++ {
			a := .40 + float64(i)*.073
			if fx > .37 && fx < .63 && fy > a && fy < a+.054 {
				remasterPauseAction(i)
				return
			}
		}
		return
	}
	if beta.screen == screenResults {
		app.started = false
		selectScreen(screenMaps)
		return
	}
	if aokawaSetupClick(x, y) {
		return
	}
	if false && fx > .30 && fx < .70 {
		if fy > .815 && fy < .895 {
			startRemasteredDrive()
			return
		}
		for i, top := range []float64{.285, .395, .505, .615, .705, .76} {
			hh := .075
			if i >= 4 {
				hh = .045
			}
			if fy > top && fy < top+hh {
				remaster.Row = i
				delta := 1
				if fx < .5 {
					delta = -1
				}
				remasterCycle(i, delta)
				return
			}
		}
	}
}

func remasterUIBegin() {
	useLighting(false)
	glExt("glActiveTexture", 0x84C0)
	w, h := float64(app.width), float64(app.height)
	gl("glMatrixMode", GL_PROJECTION)
	gl("glLoadIdentity")
	m := [16]float32{float32(2 / w), 0, 0, 0, 0, float32(-2 / h), 0, 0, 0, 0, -1, 0, -1, 1, 0, 1}
	gl("glLoadMatrixf", uintptr(unsafe.Pointer(&m[0])))
	gl("glMatrixMode", GL_MODELVIEW)
	gl("glLoadIdentity")
	dis(GL_DEPTH_TEST)
	dis(GL_CULL_FACE)
	dis(0x0BC0) // alpha test belongs to the 3D pass
	dis(0x8037) // polygon offset must not leak into map/menu quads
	dis(GL_FOG)
	en(GL_BLEND)
}
func remasterTextWidth(s string, h float64) float64 {
	h = math.Max(h, math.Min(14, float64(app.height)*.016))
	if beta.screen != screenDrive {
		h *= beta.cfg.UIScale
	}
	n := 0.
	var prev rune
	for _, c := range strings.ToUpper(s) {
		n += fontKerning035(prev, c) * h
		if g, ok := remaster.Glyphs[string(c)]; ok {
			n += g.Advance * h
		} else {
			n += h * .57
		}
		prev = c
	}
	return n
}
func remasterText(s string, x, y, h float64, c [4]float64) {
	h = math.Max(h, math.Min(14, float64(app.height)*.016))
	if beta.screen != screenDrive {
		h *= beta.cfg.UIScale
	}
	if remaster.Atlas == 0 {
		hudText(s, x, y, h/1.43, c)
		return
	}
	if usingFont035() {
		drawFont035(s, x, y, h, c)
		return
	}
	en(GL_TEXTURE_2D)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(remaster.Atlas))
	col(c[0], c[1], c[2], c[3])
	begin(GL_QUADS)
	for _, r := range strings.ToUpper(s) {
		if r == ' ' {
			x += h * .5
			continue
		}
		g, ok := remaster.Glyphs[string(r)]
		if !ok {
			g = remaster.Glyphs["?"]
		}
		cell := g.Cell
		if cell == 0 {
			cell = 192. / 168
		}
		qw, qh := h*cell, h*cell
		xx, yy := x-h*12/168, y-h*12/168
		dv := g.DV
		if dv == 0 {
			dv = .25
		}
		tex(g.U, g.V)
		vtx(xx, yy, 0)
		tex(g.U+.125, g.V)
		vtx(xx+qw, yy, 0)
		tex(g.U+.125, g.V+dv)
		vtx(xx+qw, yy+qh, 0)
		tex(g.U, g.V+dv)
		vtx(xx, yy+qh, 0)
		x += h * g.Advance
	}
	end()
	dis(GL_TEXTURE_2D)
}
func remasterCenter(s string, y, h float64, c [4]float64) {
	if width := remasterTextWidth(s, h); width > float64(app.width)*.92 {
		h *= float64(app.width) * .92 / width
	}
	remasterText(s, (float64(app.width)-remasterTextWidth(s, h))*.5, y, h, c)
}
func remasterLegacyMenuUI() {
	remasterUIBegin()
	w, h := float64(app.width), float64(app.height)
	// Legibility is independent of the orbiting 3D backdrop.
	hudRect(0, 0, w, h, .005, .012, .022, .28)
	if beta.screen == screenTitle {
		remasterCenter("SPEED IN NEED", h*.13, h*.075, white)
		remasterCenter("BETA v0.5.3", h*.235, h*.019, amber)
		hudRect(w*.22, h*.775, w*.56, h*.09, .008, .016, .025, .75)
		remasterCenter("PRESS ENTER TO DRIVE", h*.805, h*.023, [4]float64{1, 1, 1, .72 + .28*math.Sin(beta.menuTime*2)})

	} else if beta.screen == screenSettings {
		hudRect(w*.025, h*.1, w*.95, h*.86, .009, .018, .028, .96)
		remasterCenter("SETTINGS", h*.18, h*.039, white)
		labels, values := betaSettingLabels()
		for i, l := range labels {
			r := settingRect(i)
			x, y, rw, rh := r.X*w, r.Y*h, r.W*w, r.H*h
			if beta.selection == i {
				hudRect(x, y, rw, rh, .05, .15, .19, 1)
			}
			remasterText(l, x+10, y+rh*.23, h*.014, white)
			remasterText(values[i], x+rw*.68, y+rh*.23, h*.014, cyan)
		}
		remasterCenter("ESC BACK     ARROWS ADJUST     CLICK TO CHANGE", h*.925, h*.013, cyan)
		if beta.bindWaiting >= 0 {
			remasterCenter("PRESS A NEW KEY", h*.29, h*.021, amber)
		}
	} else if beta.screen == screenPause {
		hudRect(w*.28, h*.19, w*.44, h*.67, .008, .018, .028, .96)
		remasterCenter("PAUSED", h*.25, h*.05, white)
		for i, s := range []string{"RESUME", "SETTINGS", "RESTART", "MAIN MENU"} {
			y := h * (.37 + float64(i)*.105)
			hudRect(w*.32, y, w*.36, h*.08, .035, .08, .11, .98)
			if i == beta.selection {
				hudRect(w*.32, y, 4, h*.08, 1, .47, .16, 1)
			}
			remasterCenter(s, y+h*.027, h*.023, white)
		}
	} else if beta.screen == screenResults {
		hudRect(w*.23, h*.22, w*.54, h*.57, .008, .018, .028, .96)
		remasterCenter("RACE COMPLETE", h*.3, h*.039, white)
		s := app.world.Remaster
		if s != nil {
			remasterCenter(fmt.Sprintf("POSITION %d OF %d", s.Position, s.Opponents+1), h*.41, h*.028, amber)
			remasterCenter(fmt.Sprintf("TIME %.2f SEC", s.FinishTime), h*.49, h*.024, white)
			remasterCenter(fmt.Sprintf("BEST LAP %.2f SEC", s.BestLap), h*.56, h*.019, cyan)
		}
		remasterCenter("ENTER TO CONTINUE", h*.69, h*.02, white)
	} else {
		hudRect(w*.27, h*.11, w*.46, h*.81, .008, .018, .028, .95)
		remasterCenter("SPEED IN NEED", h*.155, h*.035, white)
		mapName := "MAP ONE   COAST TO CANYON   9 KM2"
		if remaster.Mode == 1 {
			mapName = "MAP TWO   COASTAL CIRCUIT   3.72 KM"
		}
		remasterCenter(mapName, h*.225, h*.012, cyan)
		labels := []string{"MODE", "CAR", "TIME", "PAINT", "LAPS", "RACERS"}
		paint := "FIXED LIVERY"
		if beta.carSelect != 3 {
			paint = fmt.Sprintf("COLOUR %02d   Q / E", beta.paint+1)
		}
		values := []string{[]string{"FREE ROAM", "CIRCUIT RACING"}[remaster.Mode], engine.RemasteredVehicleSpec(beta.carSelect).Name, []string{"GOLDEN HOUR", "SUNSET", "AFTER DARK"}[remaster.Time], paint, fmt.Sprint(remaster.Laps), fmt.Sprint(remaster.Opponents)}
		for i, top := range []float64{.285, .395, .505, .615, .705, .76} {
			if i >= 4 && remaster.Mode == 0 {
				continue
			}
			hh := .075
			if i >= 4 {
				hh = .045
			}
			y := h * top
			hudRect(w*.30, y, w*.4, h*hh, .025, .065, .088, .98)
			if remaster.Row == i {
				hudRect(w*.30, y, 3, h*hh, 1, .49, .18, 1)
			}
			if i < 4 {
				remasterText(labels[i], w*.32, y+h*.012, h*.011, cyan)
				remasterCenter(values[i], y+h*.036, h*.019, white)
			} else {
				remasterText(labels[i], w*.33, y+h*.015, h*.014, cyan)
				remasterText(values[i], w*.6, y+h*.015, h*.017, white)
			}
			remasterText("<", w*.306, y+h*hh*.5, h*.016, amber)
			remasterText(">", w*.68, y+h*hh*.5, h*.016, amber)
		}
		if remaster.Mode == 0 {
			remasterCenter("TOWN   FOREST   MOUNTAINS   COAST", h*.738, h*.012, cyan)
		}
		hudRect(w*.30, h*.815, w*.4, h*.08, 1, .39, .10, 1)
		remasterCenter("PLAY", h*.84, h*.027, [4]float64{.015, .023, .028, 1})
	}
	if beta.screen != screenTitle && beta.screen != screenSettings {
		// White mechanical gear icon, generous square hit target.
		hudRect(w*.922, h*.025, w*.056, h*.085, .012, .025, .035, .8)
		push()
		move(w*.95, h*.067, 0)
		for i := 0; i < 8; i++ {
			push()
			rotate(float64(i)*45, 0, 0, 1)
			hudRect(-h*.004, -h*.025, h*.008, h*.012, 1, 1, 1, 1)
			pop()
		}
		for i := 0; i < 24; i++ {
			a, b := float64(i)*math.Pi/12, float64(i+1)*math.Pi/12
			col(1, 1, 1, 1)
			begin(GL_QUADS)
			vtx(math.Cos(a)*h*.013, math.Sin(a)*h*.013, 0)
			vtx(math.Cos(a)*h*.022, math.Sin(a)*h*.022, 0)
			vtx(math.Cos(b)*h*.022, math.Sin(b)*h*.022, 0)
			vtx(math.Cos(b)*h*.013, math.Sin(b)*h*.013, 0)
			end()
		}
		pop()
	}
	if remaster.Status != "" {
		remasterCenter(remaster.Status, h*.95, h*.014, amber)
	}
	en(GL_DEPTH_TEST)
}

func drawRemasterHUD014(c engine.Car) {
	remasterUIBegin()
	w, h := float64(app.width), float64(app.height)
	scaleUI := beta.cfg.UIScale
	drawRadio030()
	label := ""
	drawRemasterMinimap(c)
	s := app.world.Remaster
	if s.Mode == 1 {
		label = fmt.Sprintf("LAP %d / %d   POS %d / %d", engine.ClampInt(s.Lap, 1, s.Laps), s.Laps, s.Position, s.Opponents+1)
	}
	if label != "" {
		remasterText(label, w*.04, h*.145, h*.017, white)
	}
	if s.Mode == 1 {
		remasterText(fmt.Sprintf("TIME %.1f", s.RaceTime), w*.04, h*.128, h*.016, cyan)
		if s.Countdown > 0 {
			remasterCenter(fmt.Sprint(int(math.Ceil(s.Countdown))), h*.32, h*.12, amber)
		} else if s.RaceTime < 1 {
			remasterCenter("GO", h*.32, h*.10, cyan)
		}
	}
	hudRect(w*.79, h*.77, w*.185, h*.185, .008, .018, .028, .88)
	remasterText(fmt.Sprintf("%03.0f", math.Abs(c.Speed)*3.6), w*.805, h*.79, h*.052*math.Min(scaleUI, 1.25), white)
	gear := fmt.Sprint(c.Gearbox.Gear)
	if c.Gearbox.Gear < 0 {
		gear = "R"
	}
	remasterText("KM/H   GEAR "+gear, w*.805, h*.865, h*.013, cyan)
	hudRect(w*.805, h*.905, w*.15, h*.009, .09, .16, .2, 1)
	hudRect(w*.805, h*.905, w*.15*engine.Clamp(c.Nitro/100, 0, 1), h*.009, .05, .85, 1, 1)
	remasterText("NITRO", w*.805, h*.93, h*.011, cyan)
	remasterText("ESC PAUSE   V CAMERA / F RADIO   M MAP   H LIGHTS", w*.035, h*.95, h*.012, white)
	if c.DriftActive && math.Abs(c.Lateral) > 1 {
		remasterCenter(fmt.Sprintf("DRIFT %.0f", c.DriftScore), h*.72, h*.026, amber)
	}
	if app.showPerf {
		remasterText(fmt.Sprintf("FPS %.0f   CPU %.1f MS   GPU %.1f MS   DRAWS %d", app.floatFPS, app.renderMS, app.gpuMS, gpuWorld.draws), w*.04, h*.19, h*.014, cyan)
	}
	if profiler.message != "" {
		remasterCenter(profiler.message, h*.22, h*.014, cyan)
	}
	if app.showHelp {
		hudRect(w*.25, h*.22, w*.5, h*.42, .008, .018, .028, .94)
		for i, t := range []string{"W ACCELERATE   S BRAKE / REVERSE", "A D STEER   SPACE HANDBRAKE", "MOUSE LOOK   C RECENTER", "SHIFT NITRO   R ROAD RECOVERY", "H LIGHTS   V CAMERA / F RADIO   M FULL MAP", "F11 FULLSCREEN   F2 PROFILER", "F9 BENCHMARK   F1 CLOSE HELP"} {
			remasterCenter(t, h*(.27+float64(i)*.055), h*.017, white)
		}
	}
	en(GL_DEPTH_TEST)
}

// Hold a monotonically increasing clock for benchmark diagnostics.
var remasterBoot = time.Now()

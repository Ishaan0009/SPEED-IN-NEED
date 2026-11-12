//go:build windows

package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
	"unsafe"
)

func remasterSceneOverlay() bool {
	return app.started && (remaster.FullMap || beta.screen == screenPause || beta.screen == screenSettings && beta.prevScreen == screenPause)
}
func aokawaCameraDistance() float64 {
	// Fixed presets remain bounded even at top speed or after scroll input.
	if beta.cfg.CameraMode == 0 {
		return 0
	}
	return engine.Clamp(app.camDist, 3.5, 14)
}
func aokawaSetupRows() int {
	if remaster.Mode == 1 {
		return 6
	}
	return 4
}
func aokawaAdvanceSetup014() {
	if remaster.Row < aokawaSetupRows()-1 {
		remaster.Row++
		return
	}
	startRemasteredDrive()
}
func aokawaSetupClick(x, y int) bool {
	fx, fy := float64(x)/float64(app.width), float64(y)/float64(app.height)
	if fx > .07 && fx < .45 && fy > .42 && fy < .56 {
		delta := 1
		if fx < .26 {
			delta = -1
		}
		remasterCycle(remaster.Row, delta)
		return true
	}
	if fy > .68 && fy < .76 {
		if fx > .07 && fx < .21 {
			if remaster.Row > 0 {
				remaster.Row--
			} else {
				selectScreen(screenTitle)
			}
			return true
		}
		if fx > .23 && fx < .45 {
			aokawaAdvanceSetup()
			return true
		}
	}
	return false
}
func remasterMenuUI014() {
	remasterUIBegin()
	w, h := float64(app.width), float64(app.height)
	if beta.screen == screenPause {
		hudRect(0, 0, w, h, .005, .012, .02, .18)
		hudRect(w*.35, h*.28, w*.30, h*.47, .008, .018, .027, .93)
		remasterCenter("PAUSED", h*.315, h*.030, white)
		for i, label := range []string{"RESUME", "SETTINGS", "MAIN MENU", "QUIT"} {
			y := h * (.40 + float64(i)*.073)
			hudRect(w*.37, y, w*.26, h*.054, .035, .065, .080, .96)
			if beta.selection == i {
				hudRect(w*.37, y, 3, h*.054, .1, .88, .9, 1)
			}
			remasterCenter(label, y+h*.018, h*.016, white)
		}
		en(GL_DEPTH_TEST)
		return
	}
	if beta.screen == screenSettings {
		drawAokawaSettings()
		en(GL_DEPTH_TEST)
		return
	}
	if beta.screen == screenResults {
		remasterLegacyMenuUI()
		return
	}
	if beta.screen == screenTitle {
		drawBeta011TitleLogo(w*.065, h*.10, w*.44, h*.25)

		remasterText("BETA v0.5.3", w*.07, h*.39, h*.013, amber)
		hudRect(w*.22, h*.82, w*.56, h*.072, .008, .018, .028, .78)
		remasterCenter("PRESS ENTER TO DRIVE", h*.844, h*.021, white)
		en(GL_DEPTH_TEST)
		return
	}
	hudRect(w*.045, h*.10, w*.43, h*.73, .008, .018, .027, .91)
	remasterText("SPEED IN NEED", w*.07, h*.14, h*.031, white)
	remasterText("DRIVE SETUP", w*.07, h*.21, h*.013, cyan)
	row := engine.ClampInt(remaster.Row, 0, aokawaSetupRows()-1)
	remaster.Row = row
	for i := 0; i < aokawaSetupRows(); i++ {
		color := [3]float64{.09, .17, .20}
		if i <= row {
			color = [3]float64{.08, .84, .87}
		}
		hudRect(w*(.07+float64(i)*.059), h*.28, w*.045, 3, color[0], color[1], color[2], 1)
	}
	labels := []string{"GAME MODE", "YOUR CAR", "TIME OF DAY", "BODY COLOUR", "RACE LAPS", "AI RACERS"}
	paint := fmt.Sprintf("COLOUR %02d", beta.paint+1)
	if beta.carSelect == 3 {
		paint = "FIXED TRUENO LIVERY"
	}
	values := []string{[]string{"FREE ROAM", "CIRCUIT RACING"}[remaster.Mode], engine.RemasteredVehicleSpec(beta.carSelect).Name, []string{"GOLDEN HOUR", "SUNSET", "AFTER DARK"}[remaster.Time], paint, fmt.Sprint(remaster.Laps), fmt.Sprint(remaster.Opponents)}
	remasterText(fmt.Sprintf("%02d / %02d", row+1, aokawaSetupRows()), w*.07, h*.335, h*.012, cyan)
	remasterText(labels[row], w*.07, h*.375, h*.023, white)
	hudRect(w*.07, h*.44, w*.38, h*.10, .035, .08, .10, .98)
	remasterText("<", w*.082, h*.475, h*.023, amber)
	remasterText(">", w*.424, h*.475, h*.023, amber)
	size := h * .020
	width := remasterTextWidth(values[row], size)
	if width > w*.30 {
		size *= w * .30 / width
		width = remasterTextWidth(values[row], size)
	}
	remasterText(values[row], w*.26-width/2, h*.478, size, white)
	descriptions := []string{"CITY / FOREST / MOUNTAIN CAVE", "FOUR DISTINCT HANDLING PROFILES", "INDEPENDENT FROM MAP SELECTION", "BODY PAINT ONLY", "ORDERED CHECKPOINTS / FULL LAPS", "RACING AI / NO CIVILIAN TRAFFIC"}
	if remaster.Mode == 1 && row == 0 {
		descriptions[0] = "SEPARATE CLOSED COASTAL CIRCUIT"
	}
	remasterText(descriptions[row], w*.07, h*.59, h*.011, cyan)
	hudRect(w*.07, h*.69, w*.14, h*.06, .035, .065, .08, 1)
	remasterText("BACK", w*.105, h*.709, h*.017, white)
	hudRect(w*.23, h*.69, w*.22, h*.06, .08, .83, .86, 1)
	label := "NEXT"
	if row == aokawaSetupRows()-1 {
		label = "PLAY"
	}
	remasterText(label, w*.30, h*.709, h*.017, [4]float64{.01, .025, .035, 1})
	remasterText("ARROWS ADJUST / ENTER CONTINUES", w*.07, h*.785, h*.010, cyan)
	drawAokawaGear()
	en(GL_DEPTH_TEST)
}
func drawAokawaGear() {
	w, h := float64(app.width), float64(app.height)
	hudRect(w*.935, h*.04, h*.045, h*.045, .01, .02, .03, .72)
	push()
	move(w*.949, h*.0625, 0)
	resize(.65, .65, 1)
	for i := 0; i < 8; i++ {
		push()
		rotate(float64(i)*45, 0, 0, 1)
		hudRect(-h*.004, -h*.024, h*.008, h*.012, 1, 1, 1, 1)
		pop()
	}
	col(1, 1, 1, 1)
	begin(GL_QUADS)
	for i := 0; i < 32; i++ {
		a, b := float64(i)*math.Pi/16, float64(i+1)*math.Pi/16
		vtx(math.Cos(a)*h*.017, math.Sin(a)*h*.017, 0)
		vtx(math.Cos(b)*h*.017, math.Sin(b)*h*.017, 0)
		vtx(math.Cos(b)*h*.010, math.Sin(b)*h*.010, 0)
		vtx(math.Cos(a)*h*.010, math.Sin(a)*h*.010, 0)
	}
	end()
	pop()
}

var settingNames = [][]string{
	{"PRESET", "WINDOW SIZE", "RENDER SCALE", "V-SYNC", "FRAME LIMIT", "FULLSCREEN", "UI SCALE", "TRAFFIC DENSITY", "ADAPTIVE QUALITY", "INTEGRATED MODE / RESTART"},
	{"SUN SHADOWS", "VIEW DISTANCE", "FOLIAGE DISTANCE", "TEXTURE FILTERING", "TEXTURE QUALITY", "ANTI-ALIASING", "BLOOM", "CONTACT AO", "COLOR GRADE", "SPEED MOTION BLUR"},
	{"MESH DETAIL", "GROUND DETAIL", "LIGHTING DETAIL", "SHADOW FILTER", "FOLIAGE DENSITY", "FOLIAGE WIND"},
	{"EXPOSURE", "COLOR SATURATION", "MOTION BLUR AMOUNT", "BLOOM AMOUNT", "TEXTURE BUDGET / RESTART"},
	{"INVERT LOOK X", "CAMERA SENS", "ACCELERATE", "BRAKE / REVERSE", "LEFT", "RIGHT", "NITRO", "HANDBRAKE"},
	{"MASTER VOLUME", "ENGINE / EFFECTS", "RADIO / MUSIC", "AMBIENCE", "MUTE ALL", "RADIO STATION"},
}

func aokawaSettingValues() []string {
	if remaster.SettingsTab >= 6 {
		return extraSettingValues033()
	}
	c := beta.cfg
	switch remaster.SettingsTab {
	case 0:
		return []string{[]string{"LOW / PERFORMANCE", "MEDIUM / BALANCED", "HIGH / QUALITY", "ULTRA / CINEMATIC"}[app.quality], []string{"1280 X 720", "1600 X 900", "1920 X 1080", "2560 X 1440"}[engine.ClampInt(c.Resolution, 0, 3)], fmt.Sprintf("%.0f%%", c.RenderScale*100), onoff(c.VSync), frameCapLabel035(c.FrameCap), onoff(app.fullscreen), fmt.Sprintf("%.0f%%", c.UIScale*100), []string{"OFF", "LIGHT / 256", "NORMAL / 1032", "BUSY / 1056"}[engine.ClampInt(c.TrafficDensity, 0, 3)], onoff(c.AdaptiveQuality), onoff(c.Lightweight)}
	case 1:
		return []string{[]string{"OFF", "2048 / SHORT", "3072 / MEDIUM", "4096 / LONG"}[c.Shadows], fmt.Sprintf("%.0f%%", c.ViewDistance*100), fmt.Sprintf("%.0f%%", c.Vegetation*100), fmt.Sprintf("%dX", c.Anisotropy), []string{"HIGH / FULL MIPS", "MEDIUM / +1 MIP", "LOW / +2 MIPS"}[engine.ClampInt(c.TextureBias, 0, 2)], []string{"OFF", "FXAA"}[boolInt(c.FXAA)], onoff(c.Bloom), onoff(c.SSAO), []string{"NATURAL", "VIVID", "WARM"}[engine.ClampInt(c.ColorGrade, 0, 2)], onoff(c.MotionBlur)}
	case 2:
		return []string{[]string{"LOW / 30 M", "BALANCED / 60 M", "HIGH / 100 M", "ULTRA / 150 M"}[engine.ClampInt(c.MeshDetail, 0, 3)], []string{"LOW / DIFFUSE", "MEDIUM / NORMALS", "HIGH / MICRODETAIL"}[engine.ClampInt(c.GroundDetail, 0, 2)], []string{"LOW / 4 LIGHTS", "BALANCED / 8 LIGHTS", "HIGH / 10 LIGHTS", "ULTRA / 12 LIGHTS"}[engine.ClampInt(c.LightingDetail, 0, 3)], []string{"HARD", "PCF 4 TAP", "PCF 9 TAP"}[engine.ClampInt(c.ShadowFilter, 0, 2)], fmt.Sprintf("%.0f%%", c.FoliageDensity*100), onoff(c.Wind)}
	case 3:
		return []string{fmt.Sprintf("%.0f%%", c.Exposure*100), fmt.Sprintf("%.0f%%", c.Saturation*100), fmt.Sprintf("%.0f%%", c.BlurStrength*100), fmt.Sprintf("%.0f%%", c.BloomStrength*100), []string{"512 / LOW", "1K / BALANCED", "2K / HIGH", "4K / ULTRA", "8K / SOURCE"}[engine.ClampInt(c.TextureBudget, 0, 4)], skySettingName040()}
	case 5:
		return []string{fmt.Sprintf("%.0f%%", c.MasterVolume*100), fmt.Sprintf("%.0f%%", c.EffectsVolume*100), fmt.Sprintf("%.0f%%", c.MusicVolume*100), fmt.Sprintf("%.0f%%", c.AmbienceVolume*100), onoff(c.AudioMuted), radioNames030[engine.ClampInt(c.RadioStation, 0, len(radioNames030)-1)]}
	default:
		return []string{onoff(app.invertMouseX), fmt.Sprintf("%.2f", c.CameraSensitivity), bindingName(c.Bindings[0]), bindingName(c.Bindings[1]), bindingName(c.Bindings[2]), bindingName(c.Bindings[3]), bindingName(c.Bindings[4]), bindingName(c.Bindings[5]), onoff(c.InvertLookYRefine), onoff(c.CameraAutoReturnRefine)}
	}
}
func onoff(v bool) string {
	if v {
		return "ON"
	}
	return "OFF"
}
func drawAokawaSettings() { drawSettings033() }
func aokawaSettingsKey(k uintptr) bool {
	if pad033.BindWaiting >= 0 {
		if k == VK_ESCAPE {
			pad033.BindWaiting = -1
		}
		return true
	}
	if beta.bindWaiting >= 0 {
		if k != VK_ESCAPE {
			beta.cfg.Bindings[beta.bindWaiting] = uint32(k)
		}
		beta.bindWaiting = -1
		saveCfg()
		return true
	}
	if k == VK_ESCAPE {
		if !settingsRefine.Category {
			settingsRefine.Category = true
			return true
		}
		selectScreen(beta.prevScreen)
		saveCfg()
		return true
	}
	if k == 9 {
		stepSettingsGroupRefine(1)
		return true
	}
	if settingsKeyRefine(k) {
		return true
	}
	n := len(settingNames[remaster.SettingsTab])
	if k == 0x26 {
		beta.selection = (beta.selection + n - 1) % n
	}
	if k == 0x28 {
		beta.selection = (beta.selection + 1) % n
	}
	delta := 0
	if k == 0x25 {
		delta = -1
	}
	if k == 0x27 || k == VK_RETURN {
		delta = 1
	}
	if delta != 0 {
		aokawaAdjustSetting(delta)
	}
	return true
}
func aokawaSettingsClick(x, y int) { settingsClick033(x, y) }
func aokawaAdjustSetting(delta int) {
	if remaster.SettingsTab == 4 && beta.selection >= 8 {
		if beta.selection == 8 {
			beta.cfg.InvertLookYRefine = !beta.cfg.InvertLookYRefine
		} else {
			beta.cfg.CameraAutoReturnRefine = !beta.cfg.CameraAutoReturnRefine
		}
		saveCfg()
		return
	}
	if adjustExtraSetting033(delta) {
		return
	}
	playAudio030("ui_move", .35)
	if remaster.SettingsTab == 5 {
		c := &beta.cfg
		switch beta.selection {
		case 0:
			c.MasterVolume = engine.Clamp(c.MasterVolume+float64(delta)*.05, 0, 1)
		case 1:
			c.EffectsVolume = engine.Clamp(c.EffectsVolume+float64(delta)*.05, 0, 1)
		case 2:
			c.MusicVolume = engine.Clamp(c.MusicVolume+float64(delta)*.05, 0, 1)
		case 3:
			c.AmbienceVolume = engine.Clamp(c.AmbienceVolume+float64(delta)*.05, 0, 1)
		case 4:
			c.AudioMuted = !c.AudioMuted
		case 5:
			c.RadioStation = (c.RadioStation + len(radioNames030) + delta) % len(radioNames030)
			playAudio030("radio_tune", .5)
		}
		saveCfg()
		return
	}
	c := &beta.cfg
	row := beta.selection
	if remaster.SettingsTab == 0 && row == 9 {
		c.Lightweight = !c.Lightweight
		adaptive026.Reset()
		saveCfg()
		return
	}
	if remaster.SettingsTab == 0 && row == 8 {
		c.AdaptiveQuality = !c.AdaptiveQuality
		adaptive026.Reset()
		saveCfg()
		return
	}
	if remaster.SettingsTab == 0 && row == 7 {
		c.TrafficDensity = (c.TrafficDensity + 4 + delta) % 4
		if app.world != nil {
			app.world.SetTrafficDensity(c.TrafficDensity)
		}
		saveCfg()
		return
	}
	if remaster.SettingsTab == 1 && row == 8 {
		c.ColorGrade = (c.ColorGrade + 3 + delta) % 3
		saveCfg()
		return
	}
	switch remaster.SettingsTab {
	case 0:
		switch row {
		case 0:
			applyAokawaPreset((app.quality + 4 + delta) % 4)
		case 1:
			c.Resolution = (c.Resolution + 4 + delta) % 4
			applyAokawaWindowSize()
		case 2:
			c.RenderScale = engine.Clamp(c.RenderScale+float64(delta)*.05, .5, 1)
		case 3:
			c.VSync = !c.VSync
			applyAokawaVSync()
		case 4:
			caps := []int{0, 30, 60, 90, 120, 144, 150, 165, 240}
			i := 0
			for j, v := range caps {
				if v == c.FrameCap {
					i = j
				}
			}
			c.FrameCap = caps[(i+len(caps)+delta)%len(caps)]
		case 5:
			toggleFullscreen()
		case 6:
			c.UIScale = engine.Clamp(c.UIScale+float64(delta)*.1, 1, 1.5)
		}
	case 1:
		switch row {
		case 0:
			c.Shadows = (c.Shadows + 4 + delta) % 4
		case 1:
			c.ViewDistance = engine.Clamp(c.ViewDistance+float64(delta)*.1, .7, 1.5)
		case 2:
			c.Vegetation = engine.Clamp(c.Vegetation+float64(delta)*.1, .5, 1.5)
		case 3:
			levels := []int{1, 2, 4, 8, 16}
			i := 0
			for j, v := range levels {
				if v == c.Anisotropy {
					i = j
				}
			}
			c.Anisotropy = levels[(i+5+delta)%5]
			applyAokawaTextureSettings()
		case 4:
			c.TextureBias = (c.TextureBias + 3 + delta) % 3
			applyAokawaTextureSettings()
		case 5:
			c.FXAA = !c.FXAA
		case 6:
			c.Bloom = !c.Bloom
		case 7:
			c.SSAO = !c.SSAO
		case 9:
			c.MotionBlur = !c.MotionBlur
		}
	case 2:
		switch row {
		case 0:
			c.MeshDetail = (c.MeshDetail + 4 + delta) % 4
		case 1:
			c.GroundDetail = (c.GroundDetail + 3 + delta) % 3
		case 2:
			c.LightingDetail = (c.LightingDetail + 4 + delta) % 4
		case 3:
			c.ShadowFilter = (c.ShadowFilter + 3 + delta) % 3
		case 4:
			c.FoliageDensity = engine.Clamp(c.FoliageDensity+float64(delta)*.1, .35, 1)
		case 5:
			c.Wind = !c.Wind
		}
	case 3:
		switch row {
		case 0:
			c.Exposure = engine.Clamp(c.Exposure+float64(delta)*.05, .65, 1.4)
		case 1:
			c.Saturation = engine.Clamp(c.Saturation+float64(delta)*.05, .7, 1.3)
		case 2:
			c.BlurStrength = engine.Clamp(c.BlurStrength+float64(delta)*.1, 0, 1)
		case 3:
			c.BloomStrength = engine.Clamp(c.BloomStrength+float64(delta)*.1, 0, 1.5)
		case 4:
			c.TextureBudget = (c.TextureBudget + 5 + delta) % 5
		case 5:
			c.SkyDetail040 = (c.SkyDetail040 + 5 + delta) % 5
		}
	case 4:
		switch row {
		case 0:
			app.invertMouseX = !app.invertMouseX
		case 1:
			c.CameraSensitivity = engine.Clamp(c.CameraSensitivity+float64(delta)*.1, .25, 3)
		default:
			beta.bindWaiting = row - 2
		}
	}
	saveCfg()
}
func applyAokawaPreset(p int) {
	app.quality = engine.ClampInt(p, 0, 3)
	c := &beta.cfg
	c.Quality = app.quality
	adaptive026.Reset()
	configureQuality050(c, app.quality)
	if app.quality == 0 && c.TrafficDensity > 1 {
		c.TrafficDensity = 1
		if app.world != nil && app.world.Remaster != nil {
			app.world.SetTrafficDensity(1)
		}
	}
	applyAokawaTextureSettings()
}
func applyAokawaVSync() {
	app.vsync = glExtAvailable("wglSwapIntervalEXT") && glExt("wglSwapIntervalEXT", uintptr(boolInt(beta.cfg.VSync))) != 0 && beta.cfg.VSync
}
func applyAokawaWindowSize() {
	if app.fullscreen {
		return
	}
	sizes := [][2]int{{1280, 720}, {1600, 900}, {1920, 1080}, {2560, 1440}}
	v := sizes[engine.ClampInt(beta.cfg.Resolution, 0, 3)]
	r := windowRect{right: int32(v[0]), bottom: int32(v[1])}
	win(user32, "AdjustWindowRect", uintptr(unsafe.Pointer(&r)), WS_OVERLAPPEDWINDOW, 0)
	win(user32, "SetWindowPos", app.hwnd, 0, 0, 0, uintptr(r.right-r.left), uintptr(r.bottom-r.top), 6)
}
func applyAokawaTextureSettings() {
	if app.rc == 0 {
		return
	}
	for _, id := range pbrTextures {
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(id))
		gl("glTexParameterf", GL_TEXTURE_2D, 0x8501, b32(float64(beta.cfg.TextureBias)))
		if supportsAnisotropy() {
			gl("glTexParameterf", GL_TEXTURE_2D, 0x84FE, b32(anisotropy026()))
		}
	}
}

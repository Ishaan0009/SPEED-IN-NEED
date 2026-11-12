package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"math"
	"path/filepath"
	"speedinneed/engine"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

type avatarProfile040 struct {
	ID, Name, Label, Image, File, Thumbnail string
	AtlasUV                                 [4]float64 `json:"atlas_uv"`
}

var profiles040 struct {
	Store                                           *engine.ProfileStore040
	Path, Message, DraftName, DeletePending, Metric string
	Page, Cursor, Avatar, AvatarFirst, BoardFirst   int
	Editing, AvatarFocus, MemoryOnly                bool
	Avatars                                         []avatarProfile040
	Allowed                                         map[string]bool
	Atlas                                           string
	Textures                                        map[string]uint32
}

// Call after readCfg. Saves live beside the portable settings file, and an
// unreadable store enters read-only recovery mode rather than replacing data.
func initializeProfiles040(path string) {
	profiles040.Path = path
	profiles040.MemoryOnly = qa.enabled || path == ""
	profiles040.Metric = "nearpass"
	profiles040.Textures = map[string]uint32{}
	profiles040.Allowed = map[string]bool{}
	var gallery struct {
		Default, Atlas string
		Avatars        []avatarProfile040
	}
	if raw, e := assets.ReadFile("assets/avatars040/manifest.json"); e == nil {
		json.Unmarshal(raw, &gallery)
	}
	profiles040.Avatars = gallery.Avatars
	profiles040.Atlas = gallery.Atlas
	for _, a := range gallery.Avatars {
		profiles040.Allowed[a.ID] = true
	}
	if profiles040.MemoryOnly {
		profiles040.Store = engine.NewProfileStore040()
	} else {
		store, e := engine.LoadProfileStore040(path)
		profiles040.Store = store
		if e != nil {
			profiles040.Store = engine.NewProfileStore040()
			profiles040.MemoryOnly = true
			profiles040.Message = "SAVE COULD NOT BE READ / ORIGINAL FILES PRESERVED"
		}
	}
	if profiles040.Store == nil {
		profiles040.Store = engine.NewProfileStore040()
	}
	for _, p := range profiles040.Store.Profiles {
		if !profiles040.Allowed[p.Avatar] && gallery.Default != "" {
			p.Avatar = gallery.Default
		}
	}
	active := profiles040.Store.Active040()
	if active != nil && len(active.Settings) == 0 {
		raw, _ := json.Marshal(beta.cfg)
		active.SetSettings040(raw)
	}
	applyProfileSettings040()
}

func activeProfile040() *engine.DriverProfile040 {
	if profiles040.Store == nil {
		return nil
	}
	return profiles040.Store.Active040()
}
func flushProfiles040() {
	if profiles040.MemoryOnly || qa.enabled || profiles040.Store == nil || profiles040.Path == "" {
		return
	}
	if e := engine.SaveProfileStore040(profiles040.Path, profiles040.Store); e != nil {
		profiles040.Message = "PROFILE COULD NOT BE SAVED / TRY AGAIN"
	}
}
func saveProfileSettings040() {
	p := activeProfile040()
	if p == nil {
		return
	}
	// Capture the same current selections that saveCfg captures, even when
	// switching profiles directly from the title screen.
	cfg := beta.cfg
	cfg.SelectedCar = beta.carSelect
	cfg.SelectedMap = beta.mapSelect
	cfg.Paint = beta.paint
	cfg.Quality = app.quality
	cfg.CameraDistance = app.camDist
	cfg.InvertMouseX = app.invertMouseX
	cfg.Sensitivity = app.sensitivity
	cfg.FirstPerson = app.boolFPP
	cfg.GameMode = remaster.Mode
	cfg.TimeOfDay = remaster.Time + 1
	cfg.RaceLaps = remaster.Laps
	cfg.Opponents = remaster.Opponents
	if raw, e := json.Marshal(cfg); e == nil {
		p.SetSettings040(raw)
	}
	flushProfiles040()
}
func applyProfileSettings040() {
	p := activeProfile040()
	if p == nil {
		return
	}
	cfg := defaultCfg()
	if len(p.Settings) > 0 {
		json.Unmarshal(p.Settings040(), &cfg)
	} else {
		cfg.SelectedCar = 3
		cfg.Paint = 0
		cfg.GameMode = 0
		cfg.RaceDifficulty040 = 0
	}
	if !p.CanUseCar040(cfg.SelectedCar) {
		cfg.SelectedCar = 3
	}
	if !p.CanRace040() {
		cfg.GameMode = 0
	}
	if !p.CanRaceDifficulty040(cfg.RaceDifficulty040) {
		cfg.RaceDifficulty040 = 0
	}
	cfg.UIScale = engine.Clamp(cfg.UIScale, .75, 1.5)
	cfg.Quality = engine.ClampInt(cfg.Quality, 0, 3)
	cfg.Exposure = engine.Clamp(cfg.Exposure, .65, 1.4)
	cfg.Saturation = engine.Clamp(cfg.Saturation, .7, 1.3)
	cfg.RenderScale = engine.Clamp(cfg.RenderScale, .5, 1)
	cfg.CameraDistance = engine.Clamp(cfg.CameraDistance, 3.2, 17)
	cfg.CameraSensitivity = engine.Clamp(cfg.CameraSensitivity, .2, 3)
	cfg.GamepadDeadzone = engine.Clamp(cfg.GamepadDeadzone, .05, .5)
	cfg.RaceDifficulty040 = engine.ClampInt(cfg.RaceDifficulty040, 0, 2)
	maxPaint := p.PaintCount040(cfg.SelectedCar, len(paintPalette))
	cfg.Paint = engine.ClampInt(cfg.Paint, 0, int(math.Max(0, float64(maxPaint-1))))
	beta.cfg = cfg
	beta.carSelect = cfg.SelectedCar
	beta.paint = cfg.Paint
	beta.mapSelect = engine.ClampInt(cfg.GameMode, 0, 1)
	app.quality = cfg.Quality
	app.camDist = cfg.CameraDistance
	app.invertMouseX = cfg.InvertMouseX
	app.sensitivity = cfg.Sensitivity
	app.boolFPP = cfg.CameraMode == 0
	app.boolMouse = false
	app.lightingOn = cfg.Lighting
	initializeRemasterSettings()
	remaster.SceneReady = false
	remaster.CameraReset = true
	preview014.ready = false
	if app.tex != nil {
		applyAokawaVSync()
		applyAokawaTextureSettings()
	}
	for i, a := range profiles040.Avatars {
		if a.ID == p.Avatar {
			profiles040.Avatar = i
			profiles040.AvatarFirst = i / 24 * 24
			break
		}
	}
}

func profileTexture040(path string) uint32 {
	if profiles040.Textures == nil {
		profiles040.Textures = map[string]uint32{}
	}
	if t, ok := profiles040.Textures[path]; ok {
		return t
	}
	// Gallery content is restricted to the bundled supplied-image directory.
	path = strings.ReplaceAll(path, "\\", "/")
	if !strings.HasPrefix(path, "assets/avatars040/") || strings.Contains(path, "..") {
		return 0
	}
	raw, e := assets.ReadFile(path)
	if e != nil {
		return 0
	}
	im, _, e := image.Decode(bytes.NewReader(raw))
	if e != nil {
		return 0
	}
	b := im.Bounds()
	rgba := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(rgba, rgba.Bounds(), im, b.Min, draw.Src)
	var texture uint32
	gl("glGenTextures", 1, uintptr(unsafe.Pointer(&texture)))
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(texture))
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_LINEAR)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_LINEAR)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, 0x812f)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, 0x812f)
	gl("glTexImage2D", GL_TEXTURE_2D, 0, GL_RGBA, uintptr(b.Dx()), uintptr(b.Dy()), 0, GL_RGBA, GL_UNSIGNED_BYTE, uintptr(unsafe.Pointer(&rgba.Pix[0])))
	profiles040.Textures[path] = texture
	return texture
}
func drawProfileAvatar040(id string, x, y, size float64) {
	var portrait *avatarProfile040
	for i := range profiles040.Avatars {
		if profiles040.Avatars[i].ID == id {
			portrait = &profiles040.Avatars[i]
			break
		}
	}
	if portrait == nil && len(profiles040.Avatars) > 0 {
		portrait = &profiles040.Avatars[0]
	}
	if portrait == nil {
		roundRect032(x, y, size, size, 3, [4]float64{.11, .15, .16, 1})
		return
	}
	texture := profileTexture040(profiles040.Atlas)
	uv := portrait.AtlasUV
	if texture == 0 {
		path := portrait.Thumbnail
		if path == "" {
			path = portrait.Image
		}
		if path == "" {
			path = portrait.File
		}
		texture = profileTexture040(path)
		uv = [4]float64{0, 0, 1, 1}
	}
	if texture == 0 {
		return
	}
	glExt("glActiveTexture", 0x84c0)
	en(GL_TEXTURE_2D)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(texture))
	col(1, 1, 1, 1)
	begin(GL_QUADS)
	tex(uv[0], uv[1])
	vtx(x, y, 0)
	tex(uv[2], uv[1])
	vtx(x+size, y, 0)
	tex(uv[2], uv[3])
	vtx(x+size, y+size, 0)
	tex(uv[0], uv[3])
	vtx(x, y+size, 0)
	end()
	dis(GL_TEXTURE_2D)
}
func profileButton040(id, label string, x, y, w, h float64, selected bool) {
	ui020.Buttons = append(ui020.Buttons, button020{id, x, y, w, h})
	hover := float64(remaster.MouseX) >= x && float64(remaster.MouseX) <= x+w && float64(remaster.MouseY) >= y && float64(remaster.MouseY) <= y+h
	color := [4]float64{.075, .10, .11, .66}
	if hover || selected || focus033.ID == id {
		color = [4]float64{.14, .19, .18, .95}
	}
	roundRect032(x, y, w, h, 3, color)
	if selected {
		hudRect(x, y+h-2, w, 2, cyan[0], cyan[1], cyan[2], 1)
	}
	headline035(label, x+w*.055, y+h*.31, h*.31, w*.89, white)
}
func titleProfile040(w, h float64) {
	p := activeProfile040()
	if p == nil {
		return
	}
	x, y, size := w*.755, h*.055, h*.070
	roundRect032(x-12, y-10, w*.205, size+20, 4, [4]float64{.025, .04, .045, .83})
	drawProfileAvatar040(p.Avatar, x, y, size)
	headline035(p.Name, x+size+14, y+size*.11, h*.018, w*.12, white)
	label := fmt.Sprintf("LEVEL %02d", p.Level040())
	if p.AllUnlocked040() {
		label = fmt.Sprintf("LEVEL %02d / ALL UNLOCKED", p.Level040())
	}
	remasterText(label, x+size+14, y+size*.65, h*.010, cyan)
	ui020.Buttons = append(ui020.Buttons, button020{"profiles040", x - 12, y - 10, w * .205, size + 20})
	profileButton040("leaderboard040", "LEADERBOARD", x-12, y+size+20, w*.205, h*.036, false)
}
func drawProfilesOverlay040(w, h float64) bool {
	if profiles040.Page == 0 {
		return false
	}
	ui020.Buttons = ui020.Buttons[:0]
	hudRect(0, 0, w, h, .008, .014, .018, .88)
	roundRect032(w*.045, h*.055, w*.91, h*.89, 5, [4]float64{.025, .039, .045, .98})
	if profiles040.Page == 2 {
		drawProfileLeaderboard040(w, h)
	} else {
		drawProfileManager040(w, h)
	}
	if profiles040.Message != "" {
		headline035(profiles040.Message, w*.08, h*.885, h*.010, w*.69, cyan)
	}
	profileButton040("profile-close040", "BACK", w*.81, h*.875, w*.10, h*.038, false)
	return true
}
func drawProfileManager040(w, h float64) {
	remasterText("DRIVER PROFILES", w*.08, h*.10, h*.028, white)
	remasterText("SEPARATE SETTINGS. SEPARATE PROGRESS.", w*.08, h*.158, h*.010, mutedRefine)
	s := profiles040.Store
	if s == nil {
		return
	}
	profiles040.Cursor = engine.ClampInt(profiles040.Cursor, 0, len(s.Profiles)-1)
	first := profiles040.Cursor / 7 * 7
	for row := 0; row < 7 && first+row < len(s.Profiles); row++ {
		index := first + row
		p := s.Profiles[index]
		y := h * (.22 + float64(row)*.071)
		selected := p.ID == s.ActiveID
		profileButton040("profile-select040:"+p.ID, "", w*.08, y, w*.275, h*.060, selected)
		drawProfileAvatar040(p.Avatar, w*.085, y+h*.005, h*.05)
		headline035(p.Name, w*.125, y+h*.010, h*.015, w*.18, white)
		remasterText(fmt.Sprintf("LV %02d", p.Level040()), w*.125, y+h*.037, h*.009, mutedRefine)
		if index == profiles040.Cursor && !profiles040.AvatarFocus {
			hudRect(w*.076, y+4, 2, h*.05, cyan[0], cyan[1], cyan[2], 1)
		}
	}
	if len(s.Profiles) > 7 {
		remasterText(fmt.Sprintf("%d / %d  UP / DOWN TO BROWSE", first+1, len(s.Profiles)), w*.08, h*.735, h*.008, mutedRefine)
	}
	remasterText("NEW DRIVER NAME", w*.08, h*.765, h*.009, mutedRefine)
	input := profiles040.DraftName
	if input == "" {
		input = "CLICK TO TYPE"
	}
	if profiles040.Editing && int(app.t*2)%2 == 0 {
		input += "_"
	}
	profileButton040("profile-name040", input, w*.08, h*.790, w*.19, h*.043, profiles040.Editing)
	profileButton040("profile-create040", "CREATE", w*.278, h*.790, w*.077, h*.043, false)
	p := s.Active040()
	if p == nil {
		return
	}
	drawProfileAvatar040(p.Avatar, w*.42, h*.16, h*.10)
	progress := fmt.Sprintf("LEVEL %02d / %d XP", p.Level040(), p.XP)
	if p.AllUnlocked040() {
		progress = "ALL CARS / ALL MODES / ALL DIFFICULTIES"
	}
	headline035(progress, w*.49, h*.18, h*.012, w*.39, cyan)
	if !p.AllUnlocked040() {
		headline035("NEXT LEVEL / "+fmt.Sprint(engine.XPForLevel040(p.Level040()+1))+" XP", w*.49, h*.215, h*.010, w*.37, mutedRefine)
	}
	remasterText("CUSTOM AVATAR GALLERY", w*.42, h*.285, h*.012, white)
	profiles040.Avatar = engine.ClampInt(profiles040.Avatar, 0, int(math.Max(0, float64(len(profiles040.Avatars)-1))))
	for row := 0; row < 24 && profiles040.AvatarFirst+row < len(profiles040.Avatars); row++ {
		index := profiles040.AvatarFirst + row
		a := profiles040.Avatars[index]
		size := h * .089
		x, y := w*.42+float64(row%8)*h*.103, h*.335+float64(row/8)*h*.113
		roundRect032(x-3, y-3, size+6, size+6, 3, [4]float64{.13, .19, .19, .85})
		drawProfileAvatar040(a.ID, x, y, size)
		if a.ID == p.Avatar || profiles040.AvatarFocus && index == profiles040.Avatar {
			hudRect(x, y+size+3, size, 2, cyan[0], cyan[1], cyan[2], 1)
		}
		ui020.Buttons = append(ui020.Buttons, button020{"profile-avatar040:" + strconv.Itoa(index), x, y, size, size})
	}
	profileButton040("profile-avatar-prev040", "<", w*.42, h*.705, w*.05, h*.036, false)
	profileButton040("profile-avatar-next040", ">", w*.48, h*.705, w*.05, h*.036, false)
	remasterText(fmt.Sprintf("%02d / %02d", profiles040.AvatarFirst/24+1, int(math.Max(1, math.Ceil(float64(len(profiles040.Avatars))/24)))), w*.55, h*.717, h*.010, mutedRefine)
	label := "CHOOSE FROM THE SUPPLIED PORTRAITS"
	if profiles040.AvatarFocus {
		label = "ARROWS: AVATAR / ENTER: SELECT / TAB: PROFILES"
	}
	headline035(label, w*.42, h*.774, h*.010, w*.45, mutedRefine)
	if !p.AllUnlocked040() {
		label = "DELETE PROFILE"
		if profiles040.DeletePending == p.ID {
			label = "CONFIRM DELETE"
		}
		profileButton040("profile-delete040", label, w*.42, h*.82, w*.18, h*.035, false)
	}
}
func profileLapTime040(seconds float64) string {
	if seconds <= 0 {
		return "--"
	}
	minutes := int(seconds) / 60
	return fmt.Sprintf("%d:%06.3f", minutes, seconds-float64(minutes*60))
}
func drawProfileLeaderboard040(w, h float64) {
	remasterText("LOCAL LEADERBOARD", w*.08, h*.10, h*.028, white)
	remasterText("BEST RUNS ACROSS YOUR DRIVER PROFILES", w*.08, h*.158, h*.010, mutedRefine)
	for i, item := range [][2]string{{"nearpass", "NEAR PASSES"}, {"drift", "BEST DRIFT"}, {"lap", "FASTEST LAP"}} {
		profileButton040("profile-sort040:"+item[0], item[1], w*(.08+float64(i)*.18), h*.205, w*.165, h*.038, profiles040.Metric == item[0])
	}
	for i, label := range []string{"DRIVER", "NEAR PASSES", "BEST DRIFT RUN", "MINIMUM LAP"} {
		remasterText(label, w*[]float64{.13, .43, .60, .79}[i], h*.283, h*.010, mutedRefine)
	}
	rows := profiles040.Store.Leaderboard040(profiles040.Metric)
	profiles040.BoardFirst = engine.ClampInt(profiles040.BoardFirst, 0, int(math.Max(0, float64(len(rows)-1))))
	for row := 0; row < 7 && profiles040.BoardFirst+row < len(rows); row++ {
		index := profiles040.BoardFirst + row
		p := rows[index]
		y := h * (.325 + float64(row)*.071)
		if p.ID == profiles040.Store.ActiveID {
			roundRect032(w*.078, y-6, w*.835, h*.063, 3, [4]float64{.10, .16, .15, .85})
		}
		remasterText(fmt.Sprintf("%02d", index+1), w*.088, y+h*.013, h*.014, cyan)
		drawProfileAvatar040(p.Avatar, w*.13, y, h*.048)
		headline035(p.Name, w*.166, y+h*.008, h*.014, w*.22, white)
		remasterText(fmt.Sprintf("LEVEL %02d", p.Level), w*.166, y+h*.035, h*.008, mutedRefine)
		remasterText(fmt.Sprint(p.NearPasses), w*.43, y+h*.011, h*.018, white)
		drift := "--"
		if p.BestDriftScore > 0 {
			drift = fmt.Sprintf("%.0f", p.BestDriftScore)
		}
		remasterText(drift, w*.60, y+h*.009, h*.018, white)
		if p.BestDriftSeconds > 0 {
			remasterText(fmt.Sprintf("%.1f SEC", p.BestDriftSeconds), w*.60, y+h*.036, h*.008, mutedRefine)
		}
		remasterText(profileLapTime040(p.BestLapSeconds), w*.79, y+h*.011, h*.016, white)
	}
	profileButton040("profile-board-prev040", "<", w*.08, h*.85, w*.05, h*.034, false)
	profileButton040("profile-board-next040", ">", w*.14, h*.85, w*.05, h*.034, false)
	headline035("NEAR PASS / 100 KM/H+    DRIFT / ONE CONTINUOUS RUN    LAP / ANY DIFFICULTY", w*.24, h*.861, h*.008, w*.54, mutedRefine)
}

func profileAction040(id string) bool {
	if id == "profiles040" || id == "leaderboard040" {
		if profiles040.Store == nil {
			initializeProfiles040(filepath.Join(filepath.Dir(cfgPath()), "profiles.json"))
		}
		profiles040.Page = 1
		if id == "leaderboard040" {
			profiles040.Page = 2
		}
		profiles040.Message = ""
		profiles040.Editing = false
		profiles040.DeletePending = ""
		return true
	}
	if !strings.HasPrefix(id, "profile-") {
		return false
	}
	if profiles040.Page == 0 {
		return false
	}
	s := profiles040.Store
	if s == nil {
		return true
	}
	switch id {
	case "profile-close040":
		profiles040.Page = 0
		profiles040.Editing = false
		saveProfileSettings040()
	case "profile-name040":
		profiles040.Editing = true
		profiles040.Message = "TYPE A NAME / ENTER TO CREATE"
	case "profile-create040":
		if profiles040.MemoryOnly && !qa.enabled && profiles040.Path != "" {
			profiles040.Message = "PROFILES ARE READ ONLY / ORIGINAL SAVES NEED RECOVERY"
			return true
		}
		avatar := "custom_001"
		if len(profiles040.Avatars) > 0 {
			avatar = profiles040.Avatars[profiles040.Avatar].ID
		}
		saveProfileSettings040()
		p, e := s.Create040(profiles040.DraftName, avatar)
		if e != nil {
			profiles040.Message = strings.ToUpper(e.Error())
			return true
		}
		profiles040.Cursor = len(s.Profiles) - 1
		profiles040.DraftName = ""
		profiles040.Editing = false
		profiles040.Message = "NEW DRIVER / TRUENO UNLOCKED / RACING AT LEVEL 10"
		p.Settings = nil
		applyProfileSettings040()
		saveProfileSettings040()
	case "profile-delete040":
		p := s.Active040()
		if p == nil || p.AllUnlocked040() {
			return true
		}
		if profiles040.DeletePending != p.ID {
			profiles040.DeletePending = p.ID
			profiles040.Message = "SELECT CONFIRM DELETE TO REMOVE THIS PROFILE"
			return true
		}
		s.Delete040(p.ID)
		profiles040.DeletePending = ""
		profiles040.Cursor = 0
		applyProfileSettings040()
		flushProfiles040()
		profiles040.Message = "PROFILE REMOVED"
	case "profile-avatar-prev040":
		profiles040.AvatarFirst = int(math.Max(0, float64(profiles040.AvatarFirst-24)))
	case "profile-avatar-next040":
		if profiles040.AvatarFirst+24 < len(profiles040.Avatars) {
			profiles040.AvatarFirst += 24
		}
	case "profile-board-prev040":
		profiles040.BoardFirst = int(math.Max(0, float64(profiles040.BoardFirst-7)))
	case "profile-board-next040":
		if profiles040.BoardFirst+7 < len(s.Profiles) {
			profiles040.BoardFirst += 7
		}
	default:
		if strings.HasPrefix(id, "profile-select040:") {
			key := strings.TrimPrefix(id, "profile-select040:")
			saveProfileSettings040()
			if e := s.Select040(key); e == nil {
				applyProfileSettings040()
				flushProfiles040()
				profiles040.Message = "PROFILE SELECTED"
				profiles040.DeletePending = ""
				for i, p := range s.Profiles {
					if p.ID == key {
						profiles040.Cursor = i
						break
					}
				}
			}
		} else if strings.HasPrefix(id, "profile-avatar040:") {
			index, e := strconv.Atoi(strings.TrimPrefix(id, "profile-avatar040:"))
			if e == nil && index >= 0 && index < len(profiles040.Avatars) {
				profiles040.Avatar = index
				s.Active040().SetAvatar040(profiles040.Avatars[index].ID, profiles040.Allowed)
				flushProfiles040()
				profiles040.Message = "AVATAR SELECTED"
			}
		} else if strings.HasPrefix(id, "profile-sort040:") {
			profiles040.Metric = strings.TrimPrefix(id, "profile-sort040:")
			profiles040.BoardFirst = 0
		}
	}
	return true
}
func profileKey040(key uintptr) bool {
	if profiles040.Page == 0 {
		return false
	}
	if key == 0x1b {
		if profiles040.Editing {
			profiles040.Editing = false
		} else {
			profileAction040("profile-close040")
		}
		return true
	}
	if profiles040.Editing {
		if key == 0x08 && profiles040.DraftName != "" {
			r := []rune(profiles040.DraftName)
			profiles040.DraftName = string(r[:len(r)-1])
		}
		if key == 0x0d {
			profileAction040("profile-create040")
		}
		return true
	}
	if profiles040.Page == 2 {
		switch key {
		case 0x26:
			profileAction040("profile-board-prev040")
		case 0x28:
			profileAction040("profile-board-next040")
		case 0x25:
			profiles040.Metric = map[string]string{"nearpass": "lap", "lap": "drift", "drift": "nearpass"}[profiles040.Metric]
		case 0x27:
			profiles040.Metric = map[string]string{"nearpass": "drift", "drift": "lap", "lap": "nearpass"}[profiles040.Metric]
		}
		return true
	}
	if key == 0x09 {
		profiles040.AvatarFocus = !profiles040.AvatarFocus
		return true
	}
	if profiles040.AvatarFocus {
		delta := 0
		switch key {
		case 0x25:
			delta = -1
		case 0x27:
			delta = 1
		case 0x26:
			delta = -8
		case 0x28:
			delta = 8
		case 0x0d:
			profileAction040("profile-avatar040:" + strconv.Itoa(profiles040.Avatar))
		}
		profiles040.Avatar = engine.ClampInt(profiles040.Avatar+delta, 0, int(math.Max(0, float64(len(profiles040.Avatars)-1))))
		profiles040.AvatarFirst = profiles040.Avatar / 24 * 24
		return true
	}
	switch key {
	case 0x26:
		profiles040.Cursor = int(math.Max(0, float64(profiles040.Cursor-1)))
	case 0x28:
		profiles040.Cursor = engine.ClampInt(profiles040.Cursor+1, 0, len(profiles040.Store.Profiles)-1)
	case 0x0d:
		profileAction040("profile-select040:" + profiles040.Store.Profiles[profiles040.Cursor].ID)
	case 'N':
		profiles040.Editing = true
	}
	return true
}
func profileChar040(r rune) {
	if profiles040.Page == 0 || !profiles040.Editing || unicode.IsControl(r) || r == '/' || r == '\\' || !unicode.IsPrint(r) || utf8.RuneCountInString(profiles040.DraftName) >= 24 {
		return
	}
	profiles040.DraftName += string(r)
}

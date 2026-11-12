package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
	"strings"
)

type button020 struct {
	ID         string
	X, Y, W, H float64
}

var ui020 struct {
	Stage, Hover int
	Buttons      []button020
}

func beta020TexturedPanel(key string, x, y, w, h float64) {
	if app.tex[key] == 0 {
		hudRect(x, y, w, h, .06, .11, .13, 1)
		return
	}
	en(GL_TEXTURE_2D)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(app.tex[key]))
	col(1, 1, 1, 1)
	begin(GL_QUADS)
	tex(0, 0)
	vtx(x, y, 0)
	tex(1, 0)
	vtx(x+w, y, 0)
	tex(1, 1)
	vtx(x+w, y+h, 0)
	tex(0, 1)
	vtx(x, y+h, 0)
	end()
	dis(GL_TEXTURE_2D)
}

func beta020Button(id, label string, x, y, w, h float64, selected bool) {
	if beta.screen == screenTitle {
		titleButton035(id, label, x, y, w, h)
		return
	}
	syncMenuFocus033()
	focused := focus033.ID == id
	selected = selected || focused && !strings.HasPrefix(id, "car")
	index := len(ui020.Buttons)
	ui020.Buttons = append(ui020.Buttons, button020{id, x, y, w, h})
	hover := float64(remaster.MouseX) >= x && float64(remaster.MouseX) <= x+w && float64(remaster.MouseY) >= y && float64(remaster.MouseY) <= y+h
	if hover {
		ui020.Hover = index
	}
	target := 0.
	if hover || selected || focused {
		target = 1
	}
	ui021.Hover[id] += (target - ui021.Hover[id]) * (1 - math.Exp(-ui021.DT*14))
	glow := ui021.Hover[id]
	roundRect032(x, y, w, h, 3, [4]float64{.07 + .04*glow, .09 + .05*glow, .10 + .05*glow, .20 + .50*glow})
	if selected {
		hudRect(x+14, y+h-3, w-28, 2, cyan[0], cyan[1], cyan[2], .95)
	}
	size := math.Min(float64(app.height)*.020, h*.34)
	if width := remasterTextWidth(label, size); width > w*.85 {
		size *= w * .85 / width
	}
	remasterText(label, x+(w-remasterTextWidth(label, size))/2, y+(h-size*beta.cfg.UIScale)/2, size, white)
}

func beta020Advance() {
	if ui020.Stage == 2 && !customize051 {
		customize051 = true
		focus033.Ready = false
		return
	}
	customize051 = false
	if !validateProgressSelection040() {
		return
	}
	if ui020.Stage == 3 && remaster.Mode == 0 || ui020.Stage == 4 {
		startRemasteredDrive()
		return
	}
	ui020.Stage = engine.ClampInt(ui020.Stage+1, 0, 4)
	remaster.Row = []int{0, 0, 1, 2, 4}[ui020.Stage]
	saveRemasterSettings()
}

func beta020Action(id string) {
	if id == "customize051" {
		customize051 = true
		focus033.Ready = false
		return
	}
	var vehicle, row, delta int
	if n, e := fmt.Sscanf(id, "vehicle051:%d", &vehicle); e == nil && n == 1 {
		selectCar051(vehicle)
		return
	}
	if n, e := fmt.Sscanf(id, "modify051:%d:%d", &row, &delta); e == nil && n == 2 {
		garage050.Selected = row
		adjustGarage050(delta)
		return
	}
	if id == "rain051" {
		weather050.Rain = !weather050.Rain
		return
	}
	if id == "back" && ui020.Stage == 2 && customize051 {
		customize051 = false
		focus033.Ready = false
		return
	}
	if (id == "pause-main" || id == "main") && workshop051.Active {
		exitWorkshop051()
		app.started = false
		ui020.Stage = 2
		customize051 = true
		selectScreen(screenMaps)
		return
	}
	if id == "playGarage051" {
		enterWorkshop051()
		return
	}
	if id == "pause-main" || id == "main" {
		exitWorkshop051()
		walker051 = engine.Walker051{}
	}
	if id == "multiplayer050" {
		multiplayer050.Open = true
		return
	}
	if profileAction040(id) {
		return
	}
	if !allowProgressAction040(id) {
		return
	}
	playAudio030("ui_confirm", .45)
	var colour int
	if n, err := fmt.Sscanf(id, "paint:%d", &colour); err == nil && n == 1 && colour >= 0 && colour < len(paintPalette) {
		beta.paint = colour
		saveSetup050(setup050())
		saveRemasterSettings()
		return
	}
	switch id {
	case "difficulty040":
		for i := 0; i < 3; i++ {
			beta.cfg.RaceDifficulty040 = (beta.cfg.RaceDifficulty040 + 1) % 3
			if p := activeProfile040(); p == nil || p.CanRaceDifficulty040(beta.cfg.RaceDifficulty040) {
				break
			}
		}
		saveRemasterSettings()
	case "start":
		ui020.Stage = 0
		remaster.Row = 0
		selectScreen(screenMaps)
	case "next":
		beta020Advance()
	case "back":
		if ui020.Stage == 0 {
			selectScreen(screenTitle)
		} else {
			ui020.Stage--
			remaster.Row = []int{0, 0, 1, 2, 4}[ui020.Stage]
		}
	case "main":
		app.started = false
		ui020.Stage = 0
		selectScreen(screenTitle)
	case "mode0":
		remaster.Mode = 0
		beta.mode = 0
		beta.mapSelect = 0
	case "mode1":
		remaster.Mode = 1
		beta.mode = 1
		beta.mapSelect = 1
	case "car0", "car1", "car2", "car3":
		remasterCycle(1, int(id[3]-'0')-beta.carSelect)
		preview014.ready = false
	case "car-":
		cycleCar051(-1)
		preview014.ready = false
	case "car+":
		cycleCar051(1)
		preview014.ready = false
	case "paint-":
		remasterCycle(3, -1)
	case "paint+":
		remasterCycle(3, 1)
	case "day":
		beta.cfg.RandomTime040 = false
		remaster.Time = 0
	case "afternoon":
		beta.cfg.RandomTime040 = false
		remaster.Time = 1
	case "night":
		beta.cfg.RandomTime040 = false
		remaster.Time = 2
	case "random-time040":
		beta.cfg.RandomTime040 = true
	case "laps-":
		remasterCycle(4, -1)
	case "laps+":
		remasterCycle(4, 1)
	case "racers-":
		remasterCycle(5, -1)
	case "racers+":
		remasterCycle(5, 1)
	case "settings":
		beta.prevScreen = beta.screen
		selectScreen(screenSettings)
	case "resume":
		remasterPauseAction(0)
	case "pause-settings":
		remasterPauseAction(1)
	case "pause-main":
		ui020.Stage = 0
		remasterPauseAction(2)
	case "quit":
		remasterPauseAction(3)
	}
	saveRemasterSettings()
}

func beta020Click(x, y int) bool {
	if pickSetupSpawnRefine(x, y) {
		return true
	}
	if beta.screen != screenTitle && beta.screen != screenMaps && beta.screen != screenPause {
		return false
	}
	for _, b := range ui020.Buttons {
		if float64(x) >= b.X && float64(x) <= b.X+b.W && float64(y) >= b.Y && float64(y) <= b.Y+b.H {
			beta020Action(b.ID)
			return true
		}
	}
	return true
}

func beta020MenuKey(k uintptr) bool {
	if profileKey040(k) {
		return true
	}
	return menuKey033(k)
}

func beta020MenuUI() {
	defer observeMenuRemastered051()
	beta021UIFrame()
	remasterUIBegin()
	w, h := float64(app.width), float64(app.height)
	ui020.Buttons = ui020.Buttons[:0]
	ui020.Hover = -1
	if drawProfilesOverlay040(w, h) {
		en(GL_DEPTH_TEST)
		return
	}
	drawProgressNotice040(w, h)
	if beta.screen == screenSettings {
		drawAokawaSettings()
		en(GL_DEPTH_TEST)
		return
	}
	if beta.screen == screenResults {
		if drawLANResults050(w, h) {
			en(GL_DEPTH_TEST)
			return
		}
		remasterLegacyMenuUI()
		return
	}
	if beta.screen == screenPause {
		// A compact corner panel leaves the exact frozen driving view visible.
		beta021Panel(w*.035, h*.49, w*.25, h*.43, .86)
		remasterText("PAUSED", w*.055, h*.525, h*.029, white)
		for i, item := range [][2]string{{"resume", "RESUME"}, {"pause-settings", "SETTINGS"}, {"pause-main", "MAIN MENU"}, {"quit", "QUIT"}} {
			if item[0] == "pause-main" && workshop051.Active {
				item[1] = "BACK TO CUSTOMIZE"
			}
			beta020Button(item[0], item[1], w*.055, h*(.60+float64(i)*.070), w*.21, h*.053, beta.selection == i)
		}
		en(GL_DEPTH_TEST)
		return
	}
	if beta.screen == screenTitle {
		drawTitle033(w, h)
		en(GL_DEPTH_TEST)
		return
	}
	stage := engine.ClampInt(ui020.Stage, 0, 4)
	if stage != 2 {
		hudRect(0, 0, w, h, .008, .016, .024, .72)
	}
	heading := []string{"SELECT MODE", "SELECT MAP", "SELECT CAR", "LIGHTING", "RACE SETUP"}[stage]
	if stage == 2 && customize051 {
		heading = "CUSTOMIZE"
	}
	remasterText(heading, w*.055, h*.075, h*.032, white)
	remasterText(fmt.Sprintf("%02d / %02d", stage+1, 4+remaster.Mode), w*.055, h*.135, h*.012, cyan)
	for i := 0; i < 4+remaster.Mode; i++ {
		a, b, c := .08, .16, .19
		if i <= stage {
			a, b, c = 1, .60, .20
		}
		hudRect(w*(.055+float64(i)*.038), h*.17, w*.029, 3, a, b, c, 1)
	}
	switch stage {
	case 0:
		drawModeChoice033(w, h)
	case 1:
		drawMapChoice033(w, h)
	case 2:
		drawCarChoice033(w, h)
	case 3:
		beta021LightingChoices()
	case 4:
		for i, item := range []struct {
			Label, Minus, Plus string
			Value              int
		}{{"LAPS", "laps-", "laps+", remaster.Laps}, {"AI RACERS", "racers-", "racers+", remaster.Opponents}} {
			y := h * (.29 + float64(i)*.20)
			remasterText(item.Label, w*.27, y, h*.019, white)
			beta020Button(item.Minus, "<", w*.27, y+h*.055, w*.08, h*.065, false)
			beta020Button(item.Plus, ">", w*.65, y+h*.055, w*.08, h*.065, false)
			remasterCenter(fmt.Sprint(item.Value), y+h*.075, h*.025, cyan)
		}
		beta020Button("difficulty040", "DIFFICULTY / "+[]string{"EASY", "NORMAL", "HARD"}[engine.ClampInt(beta.cfg.RaceDifficulty040, 0, 2)], w*.27, h*.70, w*.46, h*.05, true)
		if p := activeProfile040(); p != nil && !p.CanRaceDifficulty040(1) {
			remasterCenter("WIN EASY IN FIRST PLACE TO UNLOCK NORMAL", h*.78, h*.010, mutedRefine)
		} else if p != nil && !p.CanRaceDifficulty040(2) {
			remasterCenter("WIN NORMAL IN FIRST PLACE TO UNLOCK HARD", h*.78, h*.010, mutedRefine)
		}
	}
	beta020Button("back", "PREVIOUS", w*.055, h*.86, w*.15, h*.067, false)
	next := "NEXT"
	if stage == 2 && !customize051 {
		next = "CUSTOMIZE"
	}
	if stage == 3 && remaster.Mode == 0 || stage == 4 {
		next = "PLAY"
	}
	beta020Button("next", next, w*.225, h*.86, w*.24, h*.067, true)
	beta020Button("main", "MAIN MENU", w*.485, h*.86, w*.18, h*.067, false)
	drawAokawaGear()
	ui020.Buttons = append(ui020.Buttons, button020{"settings", w * .923, h * .03, w * .052, h * .08})
	remasterText("BETA 0.5.3", w*.76, h*.955, h*.010, cyan)
	en(GL_DEPTH_TEST)
}

func beta020HUD(c engine.Car) {
	if workshop051.Active {
		remasterUIBegin()
		drawWorkshopHUD051(float64(app.width), float64(app.height))
		en(GL_DEPTH_TEST)
		return
	}
	remasterUIBegin()
	w, h := float64(app.width), float64(app.height)
	s := app.world.Remaster
	remasterText(radioNames030[engine.ClampInt(beta.cfg.RadioStation, 0, len(radioNames030)-1)], w*.69, h*.045, h*.008, mutedRefine)
	drawRemasterMinimap(c)
	if s.Mode == 1 {
		position, total := racePosition050()
		label := fmt.Sprintf("POS %d / %d    LAP %d / %d", position, total, engine.ClampInt(s.Lap, 1, s.Laps), s.Laps)
		hudRect(w*.025, h*.12, w*.36, h*.092, .01, .02, .028, .82)
		remasterText(label, w*.04, h*.135, h*.022, white)
		remasterText(fmt.Sprintf("TIME %s   BEST %s", raceClock051(s.RaceTime), raceClock051(s.BestLap)), w*.04, h*.175, h*.013, cyan)
	}
	drawSpeed033(c)
	drawProgressHUD040(w, h)
	if s.Mode == 1 && s.Countdown > 0 {
		remasterCenter(fmt.Sprint(int(math.Ceil(s.Countdown))), h*.36, h*.10, white)
	}
	if app.showPerf {
		remasterText(fmt.Sprintf("%.0f FPS  /  CPU %.1f MS  /  GPU %.1f MS  /  %d CARS", app.floatFPS, app.simMS, app.gpuMS, len(app.world.Traffic)), w*.03, h*.915, h*.011, cyan)
	}
	en(GL_DEPTH_TEST)
}

func remasterMenuUI()              { beta020MenuUI() }
func drawRemasterHUD(c engine.Car) { beta020HUD(c) }
func aokawaAdvanceSetup()          { beta020Advance() }
func beta020Hover(x, y int)        { menuHover033(x, y) }

func raceClock051(seconds float64) string {
	if seconds <= 0 {
		return "--:--.--"
	}
	minutes := int(seconds) / 60
	return fmt.Sprintf("%02d:%05.2f", minutes, math.Mod(seconds, 60))
}

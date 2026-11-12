//go:build windows

package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
	"unsafe"
)

var amber = [4]float64{1, .48, .16, 1}

func menuText(s string, x, y, size float64, c [4]float64) {
	// Row layout reserves space for 150% UI. Cap headings independently.
	max := float64(app.height) * .048 / (1.43 * uiFontScale())
	hudText(s, x, y, math.Min(size, max), c)
}
func menuButton(x, y, w, h float64, title, detail string, active bool) {
	if active {
		hudRect(x, y, w, h, .055, .17, .21, .96)
		hudRect(x, y, 4, h, 1, .48, .16, 1)
	} else {
		hudRect(x, y, w, h, .025, .055, .075, .93)
	}
	menuText(title, x+19, y+h*.10, math.Min(23, h*.46/(1.43*uiFontScale())), white)
	if detail != "" {
		menuText(detail, x+20, y+h*.69, math.Min(12, h*.23/(1.43*uiFontScale())), func() [4]float64 {
			if active {
				return amber
			}
			return cyan
		}())
	}
}
func drawBetaMenuUI() {
	w, h := float64(app.width), float64(app.height)
	gl("glMatrixMode", GL_PROJECTION)
	gl("glLoadIdentity")
	m := [16]float32{float32(2 / w), 0, 0, 0, 0, float32(-2 / h), 0, 0, 0, 0, -1, 0, -1, 1, 0, 1}
	gl("glLoadMatrixf", uintptr(unsafe.Pointer(&m[0])))
	gl("glMatrixMode", GL_MODELVIEW)
	gl("glLoadIdentity")
	dis(GL_DEPTH_TEST)
	dis(GL_FOG)
	en(GL_BLEND)
	hudRect(0, 0, w, h, .006, .012, .018, .13)
	panel := .48
	if beta.screen == screenSettings {
		panel = .96
	}
	hudRect(0, 0, w*panel, h, .009, .021, .030, .96)
	hudRect(0, 0, 5, h, 1, .45, .12, 1)
	hudText("SPEED IN", w*.052, h*.062, math.Min(43, h*.073/(1.43*uiFontScale())), white)
	hudText("NEED", w*.052, h*.145, math.Min(43, h*.073/(1.43*uiFontScale())), white)
	hudRect(w*.052, h*.25, w*.055, 3, 1, .47, .13, 1)
	menuText("BETA v0.5.3", w*.12, h*.237, 13, amber)
	if beta.screen != screenSettings && beta.screen != screenMaps && beta.screen != screenPause {
		menuText("BUILT FOR THE DRIVE", w*.61, h*.12, 17, white)
		menuText([]string{"296 CHALLENGE", "911 TURBO CLASSIC"}[beta.carSelect], w*.61, h*.73, 28, white)
		menuText("PBR PAINT  /  INDEPENDENT SUSPENSION", w*.61, h*.80, 12, cyan)
	}
	switch beta.screen {
	case screenTitle:
		menuText("FIND YOUR NEXT ROAD", w*.054, h*.31, 16, cyan)
		for i, s := range []string{"DRIVE", "GARAGE", "SETTINGS", "EXIT"} {
			menuButton(w*.05, h*(.375+float64(i)*.105), w*.38, h*.082, s, []string{"THREE WORLDS. FOUR WAYS TO DRIVE.", "TWO CARS. YOUR COLOUR.", "GRAPHICS, CAMERA AND CONTROLS", "RETURN TO DESKTOP"}[i], beta.selection == i)
		}
		menuText("ARROWS + ENTER  /  MOUSE", w*.055, h*.855, 13, cyan)
	case screenMaps:
		menuText("CHOOSE YOUR WORLD", w*.054, h*.30, 23, white)
		for i, s := range []string{"COASTAL SUNSET", "NEON CITY", "CANYON ROADS"} {
			menuButton(w*.053, h*(.369+float64(i)*.135), w*.39, h*.10, fmt.Sprintf("0%d   %s", i+1, s), []string{"OCEAN ROAD  /  GOLDEN HOUR", "CITY AVENUES  /  AFTER DARK", "SWITCHBACKS  /  LATE AFTERNOON"}[i], beta.mapSelect == i)
		}
		menuButton(w*.053, h*.815, w*.39, h*.085, "CONTINUE TO GARAGE", "ENTER  /  CLICK", true)
		menuText([]string{"COASTAL SUNSET", "NEON CITY", "CANYON ROADS"}[beta.mapSelect], w*.54, h*.13, 29, white)
		menuText("LIVE WORLD PREVIEW", w*.54, h*.21, 13, amber)
		menuText("STUNT RAMPS ON CONNECTING ROADS", w*.54, h*.84, 14, white)
		menuText("CIVILIAN TRAFFIC ON THE MAIN LOOP", w*.54, h*.89, 12, cyan)
	case screenCars:
		menuText("GARAGE / 0"+fmt.Sprint(beta.carSelect+1), w*.054, h*.30, 17, cyan)
		menuText([]string{"296 CHALLENGE", "911 TURBO CLASSIC"}[beta.carSelect], w*.054, h*.37, 29, white)
		p := engine.VehicleSpec(beta.carSelect)
		menuText(fmt.Sprintf("%.0f KG  /  %.0f NM  /  6 SPEED", p.Mass, p.PeakTorque), w*.055, h*.443, 13, cyan)
		menuButton(w*.053, h*.51, w*.182, h*.065, "< PREVIOUS", "", false)
		menuButton(w*.253, h*.51, w*.19, h*.065, "NEXT >", "", false)
		menuText("BODY PAINT   Q / E", w*.055, h*.60, 13, cyan)
		for i, c := range paintPalette {
			x := w * (.055 + float64(i)*.038)
			hudRect(x, h*.65, w*.027, h*.032, c[0], c[1], c[2], 1)
			if i == beta.paint {
				hudRect(x, h*.69, w*.027, 3, 1, 1, 1, 1)
			}
		}
		menuButton(w*.053, h*.724, w*.39, h*.066, []string{"FREE DRIVE", "SPRINT RACE", "DRIFT TRIAL", "TRAFFIC RUSH"}[beta.mode], "N / CLICK TO CHANGE MODE", true)
		menuButton(w*.053, h*.815, w*.39, h*.085, "START DRIVING", "ENTER  /  CLICK", true)
	case screenSettings:
		menuText("SETTINGS", w*.055, h*.295, 24, white)
		menuText("CONTROLS & CAMERA", w*.055, h*.365, 14, amber)
		menuText("DISPLAY & WORLD", w*.55, h*.365, 14, amber)
		labels, values := betaSettingLabels()
		for i, label := range labels {
			r := settingRect(i)
			x, y, rw, rh := r.X*w, r.Y*h, r.W*w, r.H*h
			if i == beta.selection {
				hudRect(x, y, rw, rh, .055, .16, .20, .97)
				hudRect(x, y, 3, rh, 1, .47, .12, 1)
			}
			menuText(label, x+10, y+3, 14, white)
			menuText(values[i], x+rw*.68, y+3, 14, cyan)
		}
		menuText("H  HEADLIGHTS   /   F1  DRIVING HELP", w*.55, h*.82, 13, white)
		if beta.bindWaiting >= 0 {
			menuText("PRESS A NEW KEY. ESC CANCELS.", w*.055, h*.91, 17, amber)
		} else {
			menuText("ARROWS ADJUST  /  CLICK TO CHANGE", w*.055, h*.91, 13, cyan)
		}
	case screenPause:
		menuText("PAUSED", w*.055, h*.30, 33, white)
		for i, s := range []string{"RESUME DRIVE", "SETTINGS", "RESTART DRIVE", "MAIN MENU"} {
			menuButton(w*.053, h*(.405+float64(i)*.115), w*.39, h*.09, s, []string{"ENTER / ESC", "CONTROLS AND DISPLAY", "RESET CAR AND CHALLENGE", "CHOOSE A NEW WORLD"}[i], beta.selection == i)
		}
	case screenResults:
		menuText("RUN COMPLETE", w*.055, h*.34, 32, white)
		menuText(fmt.Sprintf("TIME  %.1f SEC", app.world.ChallengeTime), w*.055, h*.49, 24, cyan)
		menuText(fmt.Sprintf("%d CREDITS", app.world.Credits), w*.055, h*.58, 24, amber)
		menuText("ENTER TO CHOOSE YOUR NEXT ROAD", w*.055, h*.77, 15, white)
	}
	menuText("ESC  BACK", w*.055, h*.95, 12, cyan)
	menuText("SPEED IN NEED  /  BETA 0.1", w*.72, h*.95, 12, white)
	en(GL_DEPTH_TEST)
}
func betaSettingLabels() ([]string, []string) {
	labels := []string{"STEERING MODE", "INVERT MOUSE", "STEERING SENS", "CAMERA SENS", "CAMERA DISTANCE", "GRAPHICS", "FULLSCREEN", "ACCELERATE", "BRAKE / REVERSE", "STEER LEFT", "STEER RIGHT", "NITRO", "HANDBRAKE", "UI SCALE", "TIME OF DAY"}
	values := []string{[]string{"WASD", "MOUSE"}[boolInt(app.boolMouse)], []string{"OFF", "ON"}[boolInt(app.invertMouseX)], fmt.Sprintf("%.2f", app.sensitivity), fmt.Sprintf("%.2f", beta.cfg.CameraSensitivity), fmt.Sprintf("%.1f M", app.camDist), []string{"LOW", "BALANCED", "HIGH"}[app.quality], []string{"OFF", "ON"}[boolInt(app.fullscreen)], bindingName(beta.cfg.Bindings[0]), bindingName(beta.cfg.Bindings[1]), bindingName(beta.cfg.Bindings[2]), bindingName(beta.cfg.Bindings[3]), bindingName(beta.cfg.Bindings[4]), bindingName(beta.cfg.Bindings[5]), fmt.Sprintf("%.0f%%", beta.cfg.UIScale*100), []string{"DAY", "GOLDEN HOUR", "SUNSET", "AFTER DARK"}[beta.cfg.TimeOfDay%4]}
	return labels, values
}
func betaClick(x, y int) {
	w, h := float64(app.width), float64(app.height)
	fx, fy := float64(x)/w, float64(y)/h
	if fx < .22 && fy > .94 {
		betaInput(VK_ESCAPE)
		return
	}
	switch beta.screen {
	case screenTitle:
		if fx >= .05 && fx <= .43 {
			for i := 0; i < 4; i++ {
				top := .375 + float64(i)*.105
				if fy >= top && fy <= top+.082 {
					beta.selection = i
					betaInput(VK_RETURN)
					return
				}
			}
		}
	case screenMaps:
		if i, ok := nativeMenuHit("maps", x, y, app.width, app.height); ok {
			beta.mapSelect = i
			activateSelectedMap(false)
			saveCfg()
			return
		}
		if fx >= .05 && fx <= .45 && fy >= .81 && fy <= .91 {
			betaInput(VK_RETURN)
		}
	case screenCars:
		if fx < .45 && fy >= .81 && fy <= .91 {
			betaInput(VK_RETURN)
			return
		}
		if fy >= .51 && fy <= .575 && fx < .45 {
			if fx < .24 {
				betaInput('A')
			} else {
				betaInput('D')
			}
			return
		}
		if fy >= .64 && fy <= .70 {
			for i := range paintPalette {
				if fx >= .055+float64(i)*.038 && fx <= .082+float64(i)*.038 {
					beta.paint = i
					saveCfg()
					return
				}
			}
		}
		if fx < .45 && fy >= .72 && fy <= .79 {
			betaInput('N')
		}
	case screenPause:
		if fx >= .05 && fx <= .45 {
			for i := 0; i < 4; i++ {
				top := .405 + float64(i)*.115
				if fy >= top && fy <= top+.09 {
					beta.selection = i
					betaInput(VK_RETURN)
					return
				}
			}
		}
	case screenSettings:
		for i := 0; i < 15; i++ {
			if settingRect(i).Contains(fx, fy) {
				beta.selection = i
				if i >= 7 && i <= 12 {
					betaInput(VK_RETURN)
				} else {
					betaInput(0x27)
				}
				return
			}
		}
	}
}

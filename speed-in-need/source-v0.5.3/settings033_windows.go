package main

import (
	"fmt"
	"speedinneed/engine"
)

var tabs033 = []string{"DISPLAY", "GRAPHICS", "DETAIL", "IMAGE", "KEYS", "AUDIO", "MAP", "GAMEPAD", "RTX / API"}

func init() {
	settingNames = append(settingNames, []string{"MINIMAP ORIENTATION", "CAMERA SHAKE", "VEHICLE CRASH EFFECTS"}, []string{"CONTROLLER", "CONNECTED DEVICE", "STICK DEADZONE", "STEERING RESPONSE", "VIBRATION", "STEER AXIS", "ACCELERATOR AXIS", "BRAKE AXIS", "BIND HANDBRAKE", "BIND NITRO", "BIND CAMERA", "BIND HEADLIGHTS", "BIND MAP", "BIND RADIO", "BIND RECOVER", "BIND PAUSE", "RESTORE PAD DEFAULTS", "TRIGGER LAYOUT", "INVERT STEERING AXIS"})
	settingNames = append(settingNames, graphicsLabels038)
}
func extraSettingValues033() []string {
	if remaster.SettingsTab == 8 {
		return graphicsValues038()
	}
	c := beta.cfg
	if remaster.SettingsTab == 6 {
		return []string{[]string{"ROTATING MAP / FIXED PLAYER", "FIXED NORTH / ROTATING PLAYER"}[engine.ClampInt(c.MapOrientation, 0, 1)], fmt.Sprintf("%.0f%%", c.CameraShake*100), onoff(c.CrashEffects041)}
	}
	device := "NOT CONNECTED"
	if pad033.State.Connected {
		device = pad033.State.Name
	}
	axes := []string{"LEFT X", "LEFT Y", "RIGHT X", "RIGHT Y", "LEFT TRIGGER / U", "RIGHT TRIGGER / V"}
	out := []string{onoff(c.GamepadEnabled), device, fmt.Sprintf("%.0f%%", c.GamepadDeadzone*100), fmt.Sprintf("%.0f%%", c.GamepadSensitivity*100), onoff(c.GamepadVibration), axes[engine.ClampInt(c.GamepadSteerAxis, 0, 5)], axes[engine.ClampInt(c.GamepadThrottleAxis, 0, 5)], axes[engine.ClampInt(c.GamepadBrakeAxis, 0, 5)]}
	for i, v := range c.GamepadButtons {
		label := padButtonName033(v)
		if i == 1 {
			label = "L1 / R1 (LB / RB)"
		}
		if i == 6 {
			label = "HOLD D-PAD DOWN / 1 SEC"
		}
		if pad033.BindWaiting == i {
			label = "PRESS A CONTROLLER BUTTON..."
		}
		out = append(out, label)
	}
	return append(out, "RESTORE", []string{"SEPARATE AXES", "SHARED AXIS: + ACCELERATOR", "SHARED AXIS: - ACCELERATOR"}[engine.ClampInt(c.GamepadTriggerMode, 0, 2)], onoff(c.GamepadSteerInvert))
}
func adjustExtraSetting033(delta int) bool {
	if adjustGraphics038(delta) {
		return true
	}
	c := &beta.cfg
	if remaster.SettingsTab == 6 {
		if beta.selection == 0 {
			c.MapOrientation = (c.MapOrientation + 1) % 2
		} else if beta.selection == 2 {
			c.CrashEffects041 = !c.CrashEffects041
		} else {
			c.CameraShake = engine.Clamp(c.CameraShake+float64(delta)*.1, 0, 1)
		}
		saveCfg()
		return true
	}
	if remaster.SettingsTab != 7 {
		return false
	}
	switch beta.selection {
	case 0:
		c.GamepadEnabled = !c.GamepadEnabled
	case 2:
		c.GamepadDeadzone = engine.Clamp(c.GamepadDeadzone+float64(delta)*.02, .04, .35)
	case 3:
		c.GamepadSensitivity = engine.Clamp(c.GamepadSensitivity+float64(delta)*.1, .4, 2)
	case 4:
		c.GamepadVibration = !c.GamepadVibration
	case 5:
		c.GamepadSteerAxis = (c.GamepadSteerAxis + 6 + delta) % 6
	case 6:
		c.GamepadThrottleAxis = (c.GamepadThrottleAxis + 6 + delta) % 6
	case 7:
		c.GamepadBrakeAxis = (c.GamepadBrakeAxis + 6 + delta) % 6
	case 17:
		c.GamepadTriggerMode = (c.GamepadTriggerMode + 3 + delta) % 3
	case 18:
		c.GamepadSteerInvert = !c.GamepadSteerInvert
	case 14: // Reset is deliberately fixed to hold D-pad Down.
		return true
	case 16:
		c.GamepadTriggerMode = 0
		c.GamepadSteerInvert = true
		c.GamepadButtons = defaultPadButtons033
		c.GamepadSteerAxis = 0
		c.GamepadThrottleAxis = 5
		c.GamepadBrakeAxis = 4
		c.GamepadDeadzone = .14
		c.GamepadSensitivity = 1
		pad033.BindWaiting = -1
	default:
		if beta.selection >= 8 && beta.selection < 16 {
			pad033.BindWaiting = beta.selection - 8
		}
	}
	saveCfg()
	return true
}

var viewport033 struct{ Tab, First int }

func settingFirst033() int {
	if viewport033.Tab != remaster.SettingsTab {
		viewport033.Tab = remaster.SettingsTab
		viewport033.First = 0
	}
	if beta.selection < viewport033.First {
		viewport033.First = beta.selection
	}
	if beta.selection >= viewport033.First+9 {
		viewport033.First = beta.selection - 8
	}
	return viewport033.First
}
func drawSettings033Legacy() {
	w, h := float64(app.width), float64(app.height)
	beta021Panel(w*.065, h*.07, w*.87, h*.86, .98)
	remasterText("SETTINGS", w*.10, h*.105, h*.034, white)
	for i, name := range tabs033 {
		x := w * (.10 + float64(i)*.091)
		c := white
		if i == remaster.SettingsTab {
			c = amber
			hudRect(x, h*.216, w*.078, 2, c[0], c[1], c[2], 1)
		}
		remasterText(name, x, h*.18, h*.013, c)
	}
	labels, values := settingNames[remaster.SettingsTab], aokawaSettingValues()
	first := settingFirst033()
	for row := 0; row < 9 && first+row < len(labels); row++ {
		i := first + row
		y := h * (.265 + float64(row)*.055)
		if i == beta.selection {
			roundRect032(w*.095, y-h*.008, w*.81, h*.048, 6, [4]float64{.16, .13, .09, .92})
			hudRect(w*.095, y-h*.008, 3, h*.048, 1, .57, .15, 1)
		}
		labelColor038 := white
		if graphicsRowDisabled038(i) {
			labelColor038 = [4]float64{.45, .5, .55, 1}
		}
		remasterText(labels[i], w*.11, y, h*.013, labelColor038)
		value := values[i]
		size := h * .013
		if width := remasterTextWidth(value, size); width > w*.36 {
			size *= w * .36 / width
		}
		valueColor038 := cyan
		if graphicsRowDisabled038(i) {
			valueColor038 = labelColor038
		}
		remasterText(value, w*.53, y, size, valueColor038)
	}
	if len(labels) > 9 {
		remasterText(fmt.Sprintf("%02d / %02d", beta.selection+1, len(labels)), w*.85, h*.81, h*.011, amber)
	}
	if beta.bindWaiting >= 0 || pad033.BindWaiting >= 0 {
		remasterText("PRESS YOUR CONTROL  /  ESC CANCELS", w*.11, h*.82, h*.013, amber)
	} else {
		remasterText("ARROWS / STICK: SELECT    ENTER / A: CHANGE    TAB / LB / RB: CATEGORY    ESC / B: BACK", w*.10, h*.87, h*.010, white)
	}
}
func settingsClick033Legacy(x, y int) {
	fx, fy := float64(x)/float64(app.width), float64(y)/float64(app.height)
	if fy > .165 && fy < .225 && fx >= .10 && fx < .925 {
		remaster.SettingsTab = engine.ClampInt(int((fx-.10)/.091), 0, len(tabs033)-1)
		beta.selection = 0
		return
	}
	if fy >= .25 && fy < .765 && fx > .095 && fx < .905 {
		row := engine.ClampInt(int((fy-.257)/.055), 0, 8) + settingFirst033()
		if row < len(settingNames[remaster.SettingsTab]) {
			beta.selection = row
			delta := 1
			if fx < .50 {
				delta = -1
			}
			aokawaAdjustSetting(delta)
		}
		return
	}
	if fy > .85 {
		aokawaSettingsKey(VK_ESCAPE)
	}
}

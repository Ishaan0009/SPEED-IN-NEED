package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
	"syscall"
	"time"
	"unsafe"
)

// Windows structures and success codes follow the Microsoft XInput/WinMM APIs.
// https://learn.microsoft.com/windows/win32/api/xinput/nf-xinput-xinputgetstate
type xinputState033 struct {
	Packet         uint32
	Buttons        uint16
	LT, RT         byte
	LX, LY, RX, RY int16
}
type joyState033 struct{ Size, Flags, X, Y, Z, R, U, V, Buttons, Button, POV, Reserved1, Reserved2 uint32 }
type joyCaps033 struct {
	Mid, Pid                                   uint16
	Name                                       [32]uint16
	XYZ                                        [6]uint32
	NumButtons, PeriodMin, PeriodMax           uint32
	RUV                                        [6]uint32
	Capabilities, MaxAxes, NumAxes, MaxButtons uint32
	RegKey                                     [32]uint16
	OEM                                        [260]uint16
}
type padSample033 struct {
	Connected bool
	Name      string
	Axes      [6]float64
	Buttons   uint32
}

var defaultPadButtons033 = [8]uint32{0x2000, 0x0200, 0x8000, 0x0040, 0x0020, 0x0008, 0x0002, 0x0010}
var padActionNames033 = []string{"HANDBRAKE", "NITRO", "CAMERA", "HEADLIGHTS", "MAP", "RADIO", "RECOVER", "PAUSE"}
var pad033 = struct {
	Ready                       bool
	Get, Set                    *syscall.LazyProc
	JoyPos, JoyCaps             *syscall.LazyProc
	State                       padSample033
	Index, Backend, BindWaiting int
	Caps                        joyCaps033
	Rest                        [6]float64
	Previous                    uint32
	Retry, Repeat               time.Time
	Direction                   int
	Rumble                      uint16
}{Index: -1, BindWaiting: -1}

func padDeadzone033(v, d float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	d = engine.Clamp(d, .02, .4)
	a := math.Abs(v)
	if a <= d {
		return 0
	}
	return math.Copysign(engine.Clamp((a-d)/(1-d), 0, 1), v)
}
func padButtonName033(mask uint32) string {
	names := map[uint32]string{0x1000: "A / SOUTH", 0x2000: "B / EAST", 0x4000: "X / WEST", 0x8000: "Y / NORTH", 0x100: "LB / L1", 0x200: "RB / R1", 0x20: "BACK / SELECT", 0x10: "START", 0x40: "LEFT STICK", 0x80: "RIGHT STICK", 1: "DPAD UP", 2: "DPAD DOWN", 4: "DPAD LEFT", 8: "DPAD RIGHT"}
	if s, ok := names[mask]; ok {
		return s
	}
	return fmt.Sprintf("BUTTON 0X%X", mask)
}
func initGamepad033() {
	if pad033.Ready {
		return
	}
	pad033.Ready = true
	for _, name := range []string{"xinput1_4.dll", "xinput9_1_0.dll"} {
		d := syscall.NewLazyDLL(name)
		if d.Load() != nil {
			continue
		}
		get := d.NewProc("XInputGetState")
		if get.Find() != nil {
			continue
		}
		pad033.Get = get
		set := d.NewProc("XInputSetState")
		if set.Find() == nil {
			pad033.Set = set
		}
		break
	}
	d := syscall.NewLazyDLL("winmm.dll")
	pad033.JoyPos = d.NewProc("joyGetPosEx")
	pad033.JoyCaps = d.NewProc("joyGetDevCapsW")
}
func readXInput033(index int) (padSample033, bool) {
	if pad033.Get == nil {
		return padSample033{}, false
	}
	var s xinputState033
	code, _, _ := pad033.Get.Call(uintptr(index), uintptr(unsafe.Pointer(&s)))
	if code != 0 {
		return padSample033{}, false
	}
	return padSample033{Connected: true, Name: fmt.Sprintf("XINPUT CONTROLLER %d", index+1), Axes: [6]float64{float64(s.LX) / 32768, float64(s.LY) / 32768, float64(s.RX) / 32768, float64(s.RY) / 32768, float64(s.LT) / 255, float64(s.RT) / 255}, Buttons: uint32(s.Buttons)}, true
}
func joyAxes033(s joyState033, c joyCaps033) [6]float64 {
	values := [6]uint32{s.X, s.Y, s.Z, s.R, s.U, s.V}
	var out = [6]float64{.5, .5, .5, .5, 0, 0}
	ranges := [12]uint32{}
	copy(ranges[:6], c.XYZ[:])
	copy(ranges[6:], c.RUV[:])
	for i, v := range values {
		lo, hi := ranges[i*2], ranges[i*2+1]
		if hi > lo {
			out[i] = engine.Clamp((float64(v)-float64(lo))/float64(hi-lo), 0, 1)
		}
	}
	return out
}
func readJoystick033(index int, calibrate bool) (padSample033, bool) {
	if pad033.JoyPos == nil {
		return padSample033{}, false
	}
	s := joyState033{Size: uint32(unsafe.Sizeof(joyState033{})), Flags: 0xff}
	code, _, _ := pad033.JoyPos.Call(uintptr(index), uintptr(unsafe.Pointer(&s)))
	if code != 0 {
		return padSample033{}, false
	}
	if calibrate {
		var c joyCaps033
		code, _, _ := pad033.JoyCaps.Call(uintptr(index), uintptr(unsafe.Pointer(&c)), unsafe.Sizeof(c))
		if code != 0 {
			return padSample033{}, false
		}
		pad033.Caps = c
		pad033.Rest = joyAxes033(s, c)
	}
	a := joyAxes033(s, pad033.Caps)
	state := padSample033{Connected: true, Name: "USB / " + syscall.UTF16ToString(pad033.Caps.Name[:])}
	for i := 0; i < 4; i++ {
		state.Axes[i] = centeredJoy034(a[i])
	}
	state.Axes[1] *= -1
	state.Axes[3] *= -1
	for i := 4; i < 6; i++ {
		state.Axes[i] = engine.Clamp(math.Abs(a[i]-pad033.Rest[i])/math.Max(.1, math.Max(pad033.Rest[i], 1-pad033.Rest[i])), 0, 1)
	}
	order := []uint32{0x1000, 0x2000, 0x4000, 0x8000, 0x100, 0x200, 0x20, 0x10, 0x40, 0x80}
	if sonyPad034(state.Name, pad033.Caps.Mid) {
		order = []uint32{0x4000, 0x1000, 0x2000, 0x8000, 0x100, 0x200, 1 << 16, 1 << 17, 0x20, 0x10, 0x40, 0x80}
	}
	for i := 0; i < 32; i++ {
		if s.Buttons&(uint32(1)<<uint(i)) != 0 {
			if i < len(order) {
				state.Buttons |= order[i]
			} else if i < 26 {
				state.Buttons |= uint32(1) << uint(i+6)
			}
		}
	}
	if pad033.Caps.Capabilities&0x10 != 0 && s.POV != 0xffff && s.POV != 0xffffffff {
		angle := int(s.POV / 100)
		if angle >= 315 || angle <= 45 {
			state.Buttons |= 1
		}
		if angle >= 45 && angle <= 135 {
			state.Buttons |= 8
		}
		if angle >= 135 && angle <= 225 {
			state.Buttons |= 2
		}
		if angle >= 225 && angle <= 315 {
			state.Buttons |= 4
		}
	}
	return state, true
}
func pollGamepad033() padSample033 {
	initGamepad033()
	if pad033.Index >= 0 {
		var s padSample033
		var ok bool
		if pad033.Backend == 1 {
			s, ok = readXInput033(pad033.Index)
		} else {
			s, ok = readJoystick033(pad033.Index, false)
		}
		if ok {
			return s
		}
		pad033.Index = -1
		pad033.Rumble = 0
	}
	if time.Now().Before(pad033.Retry) {
		return padSample033{}
	}
	pad033.Retry = time.Now().Add(750 * time.Millisecond)
	for i := 0; i < 4; i++ {
		if s, ok := readXInput033(i); ok {
			pad033.Index, pad033.Backend = i, 1
			return s
		}
	}
	for i := 0; i < 16; i++ {
		if s, ok := readJoystick033(i, true); ok {
			pad033.Index, pad033.Backend = i, 2
			return s
		}
	}
	return padSample033{}
}
func padRumble033(value uint16) {
	if value == pad033.Rumble {
		return
	}
	pad033.Rumble = value
	if pad033.Backend == 1 && pad033.Index >= 0 && pad033.Set != nil {
		v := [2]uint16{value, uint16(float64(value) * .45)}
		pad033.Set.Call(uintptr(pad033.Index), uintptr(unsafe.Pointer(&v[0])))
	}
}
func padAxis033(index int) float64 {
	index = engine.ClampInt(index, 0, 5)
	v := pad033.State.Axes[index]
	if index < 4 {
		return math.Abs(padDeadzone033(v, beta.cfg.GamepadDeadzone))
	}
	return engine.Clamp((v-.035)/.965, 0, 1)
}
func updateGamepad033(dt float64) {
	if qa.enabled {
		return
	}
	if !beta.cfg.GamepadEnabled {
		camera036.LookBehind = false
		resetHold036(false, 0)
		padRumble033(0)
		pad033.State = padSample033{}
		return
	}
	s := pollGamepad033()
	wasConnected := pad033.State.Connected
	if !s.Connected && wasConnected && beta.screen == screenDrive {
		selectScreen(screenPause)
	}
	if !gameForeground040() {
		padRumble033(0)
		s.Buttons = 0
		s.Axes = [6]float64{}
	}
	pad033.State = s
	down := s.Buttons &^ pad033.Previous
	pad033.Previous = s.Buttons
	if pad033.BindWaiting >= 0 {
		if down != 0 {
			bit := down & (^down + 1)
			bindPadAction033(pad033.BindWaiting, bit)
			pad033.BindWaiting = -1
			saveCfg()
		}
		return
	}
	if beta.screen != screenDrive || remaster.FullMap {
		resetHold036(false, 0)
		camera036.LookBehind = false
	}
	if remaster.FullMap {
		step := navMapSpan033() * engine.Clamp(dt, 0, .05) * .4
		x, y := padDeadzone033(s.Axes[0], .2), -padDeadzone033(s.Axes[1], .2)
		if s.Buttons&4 != 0 {
			x = -1
		}
		if s.Buttons&8 != 0 {
			x = 1
		}
		if s.Buttons&1 != 0 {
			y = -1
		}
		if s.Buttons&2 != 0 {
			y = 1
		}
		nav033.Cursor = nav033.Cursor.Add(engine.Vec{X: x * step, Z: y * step})
		if down&0x1000 != 0 {
			setWaypoint033(nav033.Cursor)
		}
		if down&0x4000 != 0 {
			nav033.Active = false
		}
		if down&0x10 != 0 {
			remaster.FullMap = false
			selectScreen(screenPause)
			beta020SyncPointer()
		} else if down&(0x2000|0x20) != 0 {
			remaster.FullMap = false
			beta020SyncPointer()
		}
		if down&0x100 != 0 {
			nav033.Zoom = engine.Clamp(nav033.Zoom/1.25, .65, 14)
		}
		if down&0x200 != 0 {
			nav033.Zoom = engine.Clamp(nav033.Zoom*1.25, .65, 14)
		}
		padRumble033(0)
		return
	}
	if beta.screen == screenDrive {
		if photo041.Active {
			photoPad041(s, down, dt)
			padRumble033(0)
			return
		}
		drivePad036(s, down, dt)
		v := 0.0
		if beta.cfg.GamepadVibration && app.world != nil {
			c := app.world.Car
			v = engine.Clamp(math.Abs(c.Lateral)/18, 0, .20)
			if c.Braking && math.Abs(c.Speed) > 20 {
				v = math.Max(v, .16)
			}
			if c.Boosting {
				v = math.Max(v, .22)
			}
		}
		padRumble033(uint16(v * 65535))
		return
	}
	padRumble033(0)
	if profiles040.Page == 0 {
		garagePadRefine(s, dt)
	} else if down&0x300 != 0 {
		profileKey040(0x09)
		down &^= 0x300
	}
	if beta.screen == screenPause && down&0x10 != 0 {
		remasterKey(VK_ESCAPE)
		return
	}
	if down&0x2000 != 0 {
		remasterKey(VK_ESCAPE)
	}
	if down&0x1000 != 0 {
		remasterKey(VK_RETURN)
	}
	if down&0x100 != 0 {
		if beta.screen == screenSettings {
			stepSettingsGroupRefine(-1)
		} else {
			remasterKey('Q')
		}
	}
	if down&0x200 != 0 {
		if beta.screen == screenSettings {
			stepSettingsGroupRefine(1)
		} else {
			remasterKey('E')
		}
	}
	direction := 0
	lx, ly := s.Axes[0], s.Axes[1]
	if s.Buttons&1 != 0 || ly > .55 {
		direction = 0x26
	}
	if s.Buttons&2 != 0 || ly < -.55 {
		direction = 0x28
	}
	if s.Buttons&4 != 0 || lx < -.55 {
		direction = 0x25
	}
	if s.Buttons&8 != 0 || lx > .55 {
		direction = 0x27
	}
	if direction == 0 {
		pad033.Direction = 0
	} else if direction != pad033.Direction || time.Now().After(pad033.Repeat) {
		wait := 250 * time.Millisecond
		if direction != pad033.Direction {
			wait = 400 * time.Millisecond
		}
		pad033.Direction = direction
		pad033.Repeat = time.Now().Add(wait)
		remasterKey(uintptr(direction))
	}
}
func bindPadAction033(action int, mask uint32) bool {
	if action < 0 || action >= len(beta.cfg.GamepadButtons) || mask == 0 || mask&(mask-1) != 0 {
		return false
	}
	old := beta.cfg.GamepadButtons[action]
	for i, v := range beta.cfg.GamepadButtons {
		if i != action && v == mask {
			beta.cfg.GamepadButtons[i] = old
		}
	}
	beta.cfg.GamepadButtons[action] = mask
	return true
}
func mergeGamepad033(in engine.Input) engine.Input {
	keyboardDriving := in.Throttle > 0 || in.Brake > 0 || in.Boost
	if !pad033.State.Connected || !beta.cfg.GamepadEnabled {
		return in
	}
	i := engine.ClampInt(beta.cfg.GamepadSteerAxis, 0, 5)
	v := padDeadzone033(pad033.State.Axes[i], beta.cfg.GamepadDeadzone) * engine.Clamp(beta.cfg.GamepadSensitivity, .4, 2)
	if math.Abs(in.Steer) < .01 {
		if beta.cfg.GamepadSteerInvert {
			v = -v
		}
		in.Steer = engine.Clamp(v, -1, 1)
	}
	forwardGate, reverseGate := 0., 0.
	if beta.cfg.GamepadTriggerMode == 0 {
		forwardGate = padAxis033(beta.cfg.GamepadThrottleAxis)
		reverseGate = padAxis033(beta.cfg.GamepadBrakeAxis)
	} else {
		shared := padDeadzone033(pad033.State.Axes[engine.ClampInt(beta.cfg.GamepadThrottleAxis, 0, 5)], beta.cfg.GamepadDeadzone)
		if beta.cfg.GamepadTriggerMode == 2 {
			shared = -shared
		}
		forwardGate, reverseGate = math.Max(0, shared), math.Max(0, -shared)
	}
	// Either trigger enables the engine. The left stick chooses direction and
	// drive amount; both triggers together remain a valid ignition request.
	longitudinal := padDeadzone033(pad033.State.Axes[1], beta.cfg.GamepadDeadzone)
	if pad033.State.Buttons&(1<<17) != 0 {
		forwardGate = 1
	}
	if pad033.State.Buttons&(1<<16) != 0 {
		reverseGate = 1
	}
	if (forwardGate > .06 || reverseGate > .06) && longitudinal > 0 {
		in.Throttle = math.Max(in.Throttle, longitudinal)
	}
	if (forwardGate > .06 || reverseGate > .06) && longitudinal < 0 {
		in.Brake = math.Max(in.Brake, -longitudinal)
	}
	in.EngineGated = !keyboardDriving
	in.EngineEnabled = forwardGate > .06 || reverseGate > .06
	in.Handbrake = in.Handbrake || pad033.State.Buttons&beta.cfg.GamepadButtons[0] != 0
	in.Boost = in.Boost || (pad033.State.Buttons&(beta.cfg.GamepadButtons[1]|0x0300) != 0 && in.Throttle > .05)
	return in
}

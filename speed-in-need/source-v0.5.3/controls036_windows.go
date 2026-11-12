package main

import (
	"math"
	"speedinneed/engine"
	"time"
)

var camera036 struct {
	LookBehind, Dirty bool
	Changed           time.Time
	ResetHold         float64
	ResetFired        bool
}

func zoomCamera036(delta float64) {
	if app.boolFPP || beta.cfg.CameraMode == 0 {
		return
	}
	if app.camDist < 3.2 || app.camDist > 17 {
		app.camDist = 5.8
	}
	app.camDist = engine.Clamp(app.camDist+delta, 3.2, 17)
	beta.cfg.CameraDistance = app.camDist
	camera036.Dirty = true
	camera036.Changed = time.Now()
}
func persistCamera036() {
	if camera036.Dirty && time.Since(camera036.Changed) > 700*time.Millisecond && !qa.enabled {
		camera036.Dirty = false
		saveCfg()
	}
}
func lookBackYaw036() float64 {
	if camera036.LookBehind {
		return math.Pi
	}
	return 0
}

// Pure hold state: a reset needs a continuous deliberate one-second hold.
func resetHold036(held bool, dt float64) bool {
	if !held {
		camera036.ResetHold = 0
		camera036.ResetFired = false
		return false
	}
	camera036.ResetHold += engine.Clamp(dt, 0, .1)
	if camera036.ResetHold >= 1 && !camera036.ResetFired {
		camera036.ResetFired = true
		return true
	}
	return false
}
func drivePad036(s padSample033, down uint32, dt float64) {
	for i, k := range []uintptr{0, 0, 'V', 'H', 'M', 0, 0, VK_ESCAPE} {
		if k != 0 && down&beta.cfg.GamepadButtons[i] != 0 {
			remasterKey(k)
		}
	}
	if down&0x1000 != 0 {
		togglePhoto041()
	}
	if down&4 != 0 {
		beta.cfg.RadioStation = (beta.cfg.RadioStation + len(radioNames030) - 1) % len(radioNames030)
		playAudio030("radio_tune", .7)
		saveCfg()
	}
	if down&beta.cfg.GamepadButtons[5] != 0 {
		cycleRadio030()
	}
	if resetHold036(s.Buttons&2 != 0, dt) {
		recoverRoad038()
	}
	if s.Buttons&1 != 0 && time.Since(horn036Last) > 300*time.Millisecond {
		playAudio030("horn", .65)
		horn036Last = time.Now()
	}
	behind := s.Buttons&0x4000 != 0
	if behind != camera036.LookBehind {
		remaster.CameraReset = true
	}
	camera036.LookBehind = behind
	x, y := padDeadzone033(s.Axes[2], .18), padDeadzone033(s.Axes[3], .18)
	if s.Buttons&0x80 != 0 {
		if math.Abs(y) > .01 {
			zoomCamera036(-y * dt * 3.2)
		}
	} else if math.Abs(x)+math.Abs(y) > .01 {
		if beta.cfg.InvertLookYRefine {
			y = -y
		}
		if app.invertMouseX {
			x = -x
		}
		app.camYaw = math.Remainder(app.camYaw-x*dt*2.1*beta.cfg.CameraSensitivity, engine.Tau)
		app.camPitch = engine.Clamp(app.camPitch-y*dt*.8*beta.cfg.CameraSensitivity, -.45, .65)
		beta.idleMouse = 0
	}
}

var horn036Last time.Time

func migratePad036() {
	if beta.cfg.PadMappingVersion < 36 {
		beta.cfg.GamepadButtons = defaultPadButtons033
		beta.cfg.PadMappingVersion = 36
	}
}

func hornSamples036() []float32 {
	samples := make([]float32, audioRate030*28/100)
	for i := range samples {
		t := float64(i) / audioRate030
		envelope := math.Min(1, t/.008) * math.Min(1, (.28-t)/.025)
		samples[i] = float32((math.Sin(t*engine.Tau*440) + .65*math.Sin(t*engine.Tau*554.37)) * .20 * envelope)
	}
	return samples
}

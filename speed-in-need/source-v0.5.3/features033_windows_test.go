package main

import (
	"math"
	"speedinneed/engine"
	"testing"
	"time"
	"unsafe"
)

func TestController033NativeABIAndInput(t *testing.T) {
	if unsafe.Sizeof(xinputState033{}) != 16 || unsafe.Sizeof(joyState033{}) != 52 || unsafe.Sizeof(joyCaps033{}) != 728 {
		t.Fatal("Windows controller structure mismatch")
	}
	if unsafe.Offsetof(joyCaps033{}.RUV) != 104 || unsafe.Offsetof(joyCaps033{}.Capabilities) != 128 {
		t.Fatal("WinMM capabilities offsets")
	}
	for _, v := range []float64{0, .08, -.08, math.NaN(), math.Inf(1)} {
		if padDeadzone033(v, .14) != 0 {
			t.Fatal("neutral drift", v)
		}
	}
	if padDeadzone033(-1, .14) != -1 || padDeadzone033(1, .14) != 1 {
		t.Fatal("stick range")
	}
	savedCfg, savedPad := beta.cfg, pad033
	defer func() { beta.cfg = savedCfg; pad033 = savedPad }()
	beta.cfg = defaultCfg()
	pad033.State = padSample033{Connected: true, Axes: [6]float64{.8, 0, 0, 0, 0, 0}}
	got := mergeGamepad033(engine.Input{Steer: -1, Brake: .7})
	if got.Steer != -1 || got.Brake != .7 || got.Throttle != 0 {
		t.Fatal("keyboard priority or phantom trigger", got)
	}
	pad033.State.Axes[5] = 1
	pad033.State.Axes[1] = 1
	got = mergeGamepad033(engine.Input{})
	if got.Throttle != 1 || got.Brake != 0 || got.Steer >= -.5 {
		t.Fatal("analog driving", got)
	}
	if !bindPadAction033(0, defaultPadButtons033[1]) || beta.cfg.GamepadButtons[1] != defaultPadButtons033[0] {
		t.Fatal("conflicting binding must swap")
	}
	if bindPadAction033(2, 3) || bindPadAction033(-1, 1) {
		t.Fatal("invalid binding accepted")
	}
	beta.cfg.GamepadTriggerMode = 1
	beta.cfg.GamepadThrottleAxis = 0
	pad033.State.Axes[0] = -.9
	pad033.State.Axes[1] = -1
	got = mergeGamepad033(engine.Input{})
	if got.Brake < .8 || got.Throttle != 0 {
		t.Fatal("shared trigger brake", got)
	}
	beta.cfg.GamepadTriggerMode = 2
	pad033.State.Axes[1] = 1
	got = mergeGamepad033(engine.Input{})
	if got.Throttle < .8 || got.Brake != 0 {
		t.Fatal("inverted shared trigger", got)
	}
	pad033.State.Connected = false
	in := engine.Input{Steer: .4, Throttle: .6}
	if mergeGamepad033(in) != in {
		t.Fatal("disconnected input changed")
	}
}

func TestNavigation033ShippedRoads(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	start := time.Now()
	g := engine.NewRoadNavigator(d)
	t.Logf("graph: %d nodes, %d edges, %d components; built in %s", len(g.Nodes), len(g.Segments), g.Components, time.Since(start))
	if len(g.Segments) < 5000 {
		t.Fatal("authored roads lost")
	}
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	for _, p := range []engine.Vec{{X: 75, Z: -675}, {X: -1000, Z: 0}, {X: 300, Z: 500}, {X: 0, Z: -1200}, {X: 1200, Z: 1000}} {
		begin := time.Now()
		path := g.FindPath(w.Car.Pos, p, w.Car.Body.Height-.74)
		t.Logf("target %+v: road %.1fm, last mile %.1fm, %d points, %s", p, path.Metres, path.OffRoadMetres, len(path.Points), time.Since(begin))
		if !path.Reachable || len(path.Points) < 2 || math.IsNaN(path.Metres) {
			t.Fatal("unreachable road navigation", p)
		}
		if path.RoadEnd.Dist(p) > 350 {
			t.Fatal("unexpectedly disconnected destination", p, path.RoadEnd)
		}
	}
}

func TestController033HardwareProbe(t *testing.T) {
	saved := pad033
	defer func() { padRumble033(0); pad033 = saved }()
	initGamepad033()
	state := pollGamepad033()
	t.Logf("XInput available=%v; WinMM available=%v; connected=%v; device=%s", pad033.Get != nil, pad033.JoyPos != nil, state.Connected, state.Name)
}

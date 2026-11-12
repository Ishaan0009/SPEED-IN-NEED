package main

import (
	"encoding/json"
	"math"
	"speedinneed/engine"
	"testing"
)

func worldDefinition034(t *testing.T, mode string) *engine.WorldDefinition {
	t.Helper()
	raw, err := assets.ReadFile("assets/b034_world_" + mode + ".json")
	if err != nil {
		t.Fatal(err)
	}
	d, err := engine.DecodeWorldDefinition(raw)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
func TestStick034DrivesPhysicalWheels(t *testing.T) {
	cfg, pad := beta.cfg, pad033
	defer func() { beta.cfg = cfg; pad033 = pad }()
	beta.cfg = defaultCfg()
	pad033.State = padSample033{Connected: true}
	d := worldDefinition034(t, "free")
	for _, direction := range []float64{1, -1} {
		w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
		w.SetTrafficDensity(0)
		start := w.Car.Pos
		pad033.State.Axes = [6]float64{0, direction, 0, 0, 0, 0}
		if direction > 0 {
			pad033.State.Axes[5] = 1
		} else {
			pad033.State.Axes[4] = 1
		}
		peak := 0.0
		for i := 0; i < 240; i++ {
			w.Step(mergeGamepad033(engine.Input{}), 1./120)
			peak = math.Max(peak, direction*w.Car.Speed)
		}
		if w.Car.Pos.Dist(start) < 2 || peak < 1 {
			t.Fatalf("left stick failed to move: direction=%v speed=%v distance=%v", direction, w.Car.Speed, w.Car.Pos.Dist(start))
		}
	}
	pad033.State.Axes = [6]float64{1, 0, 0, 0, 0, 0}
	in := mergeGamepad033(engine.Input{})
	if in.Steer >= 0 {
		t.Fatal("stick right must match keyboard D steering")
	}
	pad033.State.Axes = [6]float64{}
	pad033.State.Buttons = beta.cfg.GamepadButtons[0]
	if !mergeGamepad033(engine.Input{}).Handbrake {
		t.Fatal("the configured handbrake button must engage the rear brake")
	}
	pad033.State.Buttons = 0
	pad033.State.Axes[5] = 1
	if mergeGamepad033(engine.Input{}).Throttle != 0 {
		t.Fatal("R2 alone must only enable the motor")
	}
	pad033.State.Axes[1] = -1
	in = mergeGamepad033(engine.Input{})
	if in.Throttle != 0 || in.Brake != 1 {
		t.Fatal("R2+south must request reverse", in)
	}
	pad033.State.Axes[1] = 1
	pad033.State.Axes[4] = 1
	pad033.State.Axes[5] = 0
	in = mergeGamepad033(engine.Input{})
	if in.Throttle != 1 || in.Brake != 0 {
		t.Fatal("LT+north must request forward", in)
	}
}
func TestUSB034NeutralAndHeldReconnect(t *testing.T) {
	if centeredJoy034(.5) != 0 || centeredJoy034(1) != 1 || centeredJoy034(0) != -1 {
		t.Fatal("calibrated endpoints")
	}
	axes := joyAxes033(joyState033{X: 32767, Y: 32767}, joyCaps033{XYZ: [6]uint32{0, 65535, 0, 65535, 0, 0}})
	if math.Abs(centeredJoy034(axes[0])) > .001 || centeredJoy034(axes[2]) != 0 || axes[4] != 0 {
		t.Fatal("absent axes caused phantom input", axes)
	}
	if !sonyPad034("Wireless Controller", 0) || !sonyPad034("USB Gamepad", 0x54c) {
		t.Fatal("PlayStation mapping")
	}
}
func TestAutopilot034RouteAndManualOverride(t *testing.T) {
	oldWorld, oldNav, oldAP := app.world, nav033, autopilot034
	defer func() { app.world = oldWorld; nav033 = oldNav; autopilot034 = oldAP }()
	d := worldDefinition034(t, "free")
	app.world = engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	c := app.world.Car
	f := engine.HeadingVec(c.Heading)
	nav033.Active = true
	nav033.Path = engine.NavPath033{Reachable: true, Points: []engine.Vec{c.Pos, c.Pos.Add(f.Mul(50)), c.Pos.Add(f.Mul(100))}}
	autopilot034 = true
	in := mergeAutopilot034(engine.Input{})
	if in.Throttle <= 0 || math.Abs(in.Steer) > .001 {
		t.Fatal("straight route", in)
	}
	in = mergeAutopilot034(engine.Input{Brake: .8})
	if autopilot034 || in.Brake != .8 {
		t.Fatal("brake must override AP immediately")
	}
	autopilot034 = true
	nav033.Active = false
	mergeAutopilot034(engine.Input{})
	if autopilot034 {
		t.Fatal("cleared waypoint must release AP")
	}
}
func TestRelease034WorldRoutesAndCircuit(t *testing.T) {
	for _, mode := range []string{"free", "race"} {
		updated := worldDefinition034(t, mode)
		raw, err := assets.ReadFile("assets/remaster_world_" + mode + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var old engine.WorldDefinition
		if json.Unmarshal(raw, &old) != nil {
			t.Fatal("baseline")
		}
		a, _ := json.Marshal(updated.Roads)
		b, _ := json.Marshal(old.Roads)
		if string(a) != string(b) {
			t.Fatal("v033 roads changed", mode)
		}
		if mode == "race" {
			a, _ = json.Marshal(updated.Grid)
			b, _ = json.Marshal(old.Grid)
			if string(a) != string(b) {
				t.Fatal("circuit ground changed")
			}
		}
		graph := engine.NewRoadNavigator(updated)
		path := graph.FindPath(updated.Spawn, updated.Roads[0].Points[len(updated.Roads[0].Points)/2], updated.Grid.Height(updated.Spawn))
		if !path.Reachable {
			t.Fatal("new runtime route disconnected", mode)
		}
	}
}

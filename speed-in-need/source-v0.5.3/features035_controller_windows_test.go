package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
	"testing"
)

func TestController035IgnitionDirectionMatrix(t *testing.T) {
	cfg, pad := beta.cfg, pad033
	defer func() { beta.cfg = cfg; pad033 = pad }()
	beta.cfg = defaultCfg()
	for _, stick := range []float64{-1, 0, 1} {
		for triggers := 0; triggers < 4; triggers++ {
			t.Run(fmt.Sprintf("stick%g-triggers%d", stick, triggers), func(t *testing.T) {
				pad033.State = padSample033{Connected: true, Axes: [6]float64{.7, stick, 0, 0, float64(triggers & 1), float64((triggers >> 1) & 1)}}
				in := mergeGamepad033(engine.Input{})
				forward, reverse := 0., 0.
				if triggers != 0 && stick > 0 {
					forward = 1
				}
				if triggers != 0 && stick < 0 {
					reverse = 1
				}
				if in.Throttle != forward || in.Brake != reverse || !in.EngineGated || in.EngineEnabled != (triggers != 0) || in.Steer >= 0 {
					t.Fatal("incorrect ignition/direction/steering", in)
				}
			})
		}
	}
	pad033.State = padSample033{Connected: true, Axes: [6]float64{0, .55, 0, 0, 0, .4}}
	in := mergeGamepad033(engine.Input{})
	if math.Abs(in.Throttle-padDeadzone033(.55, beta.cfg.GamepadDeadzone)) > 1e-9 {
		t.Fatal("trigger pressure changed the accelerator; stick must set drive", in)
	}
	if in := mergeGamepad033(engine.Input{Throttle: .6}); in.EngineGated || in.Throttle < .6 {
		t.Fatal("controller disabled keyboard motor", in)
	}
}

func TestController035StickWithoutIgnitionOnlySteersTires(t *testing.T) {
	cfg, pad := beta.cfg, pad033
	defer func() { beta.cfg = cfg; pad033 = pad }()
	beta.cfg = defaultCfg()
	pad033.State = padSample033{Connected: true, Axes: [6]float64{1, 1}}
	d := &engine.WorldDefinition{Grid: engine.HeightGrid{W: 3, H: 3, MinX: -500, MinZ: -500, Step: 500, Values: make([]float32, 9)}, Roads: []engine.RoadDefinition{{Points: []engine.Vec{{X: -400}, {X: 400}}, Width: 800}}}
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	w.SetTrafficDensity(0)
	start := w.Car.Pos
	for i := 0; i < 120; i++ {
		w.Step(mergeGamepad033(engine.Input{}), 1./120)
	}
	if w.Car.EngineRunning || math.Abs(w.Car.Speed) > .02 || w.Car.Pos.Dist(start) > .02 || math.Abs(w.Car.Rig.Wheels[0].SteerAngle) < .2 {
		t.Fatalf("engine-off stick moved car or failed to steer: %+v", w.Car)
	}
	pad033.State.Axes[5] = 1
	for i := 0; i < 240; i++ {
		w.Step(mergeGamepad033(engine.Input{}), 1./120)
	}
	if !w.Car.EngineRunning || w.Car.Speed < 1 || w.Car.Pos.Dist(start) < 1 {
		t.Fatal("matching trigger+stick did not drive")
	}
	pad033.State.Axes[5] = 0
	w.Step(mergeGamepad033(engine.Input{}), 1./120)
	if w.Car.EngineRunning || w.Car.Gearbox.RPM != 0 || math.Abs(w.Car.Rig.Wheels[0].SteerAngle) < .2 {
		t.Fatal("released trigger did not switch engine off while retaining tire steering")
	}
}

func TestTraffic035ShippedCivilianNetwork(t *testing.T) {
	d := worldDefinition034(t, "free")
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	core := w.BuildTrafficCore035()
	eligible := 0
	for _, ok := range core {
		if ok {
			eligible++
		}
	}
	t.Logf("civilian through routes=%d/%d traffic pool=%d", eligible, len(core), len(w.Remaster.TrafficPool))
	if eligible < 30 || eligible >= len(core) || len(w.Remaster.TrafficPool) != 1056 {
		t.Fatal("civilian network pruning removed through roads or retained all dead ends")
	}
	for _, npc := range w.Remaster.TrafficPool {
		if !core[npc.Path] {
			t.Fatal("civilian spawned in excluded dead-end road")
		}
	}
}

func TestController035EngineOffMutesMotorOnly(t *testing.T) {
	clip := make([]float32, 2205)
	for i := range clip {
		clip[i] = float32(math.Sin(float64(i)*.1) * .3)
	}
	frame := audioFrame030{Master: 1, Effects: 1, Driving: true, EngineOff: true, RPM: 900}
	mixer := audioMixer030{Clips: map[string][]float32{"engine0": clip, "idle0": clip}, Gain: 1}
	out := make([]int16, 1024)
	mixer.render(out, frame)
	for _, sample := range out {
		if sample != 0 {
			t.Fatal("engine-off controller still played motor recording")
		}
	}
	frame.EngineOff = false
	mixer.render(out, frame)
	nonzero := false
	for _, sample := range out {
		nonzero = nonzero || sample != 0
	}
	if !nonzero {
		t.Fatal("trigger ignition failed to restore motor sound")
	}
}

func TestAutopilot035OverridesNeutralControllerIgnition(t *testing.T) {
	oldWorld, oldNav, oldAP, cfg, pad := app.world, nav033, autopilot034, beta.cfg, pad033
	defer func() { app.world = oldWorld; nav033 = oldNav; autopilot034 = oldAP; beta.cfg = cfg; pad033 = pad }()
	beta.cfg = defaultCfg()
	pad033.State = padSample033{Connected: true}
	d := &engine.WorldDefinition{Grid: engine.HeightGrid{W: 3, H: 3, MinX: -500, MinZ: -500, Step: 500, Values: make([]float32, 9)}, Roads: []engine.RoadDefinition{{Points: []engine.Vec{{Z: -400}, {Z: 400}}, Width: 20}}}
	app.world = engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	c := app.world.Car
	f := engine.HeadingVec(c.Heading)
	nav033.Active = true
	nav033.Path = engine.NavPath033{Reachable: true, Points: []engine.Vec{c.Pos, c.Pos.Add(f.Mul(80))}}
	autopilot034 = true
	in := mergeAutopilot034(mergeGamepad033(engine.Input{}))
	if in.EngineGated || in.Throttle <= 0 {
		t.Fatal("neutral controller disabled autopilot", in)
	}
}

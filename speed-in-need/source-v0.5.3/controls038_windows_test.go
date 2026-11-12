package main

import (
	"speedinneed/engine"
	"testing"
)

func TestController038BothBumpersNitroAndEitherIgnition(t *testing.T) {
	cfg, pad := beta.cfg, pad033
	defer func() { beta.cfg = cfg; pad033 = pad }()
	beta.cfg = defaultCfg()
	for _, button := range []uint32{0x100, 0x200, 0x300} {
		for _, trigger := range []int{4, 5} {
			pad033.State = padSample033{Connected: true, Buttons: button}
			pad033.State.Axes[1] = 1
			pad033.State.Axes[trigger] = 1
			in := mergeGamepad033(engine.Input{})
			if !in.Boost || in.Throttle != 1 || !in.EngineEnabled {
				t.Fatal(button, trigger, in)
			}
			pad033.State.Axes[1] = -1
			in = mergeGamepad033(engine.Input{})
			if in.Boost || in.Brake != 1 || in.Throttle != 0 {
				t.Fatal("reverse must not request nitro", in)
			}
		}
	}
}

func TestController038RoadRecoveryHoldExecutesOnce(t *testing.T) {
	world, camera, notice, prev, ap := app.world, camera036, roadRecoveryNotice038, app.prevCar, autopilot034
	defer func() {
		app.world = world
		camera036 = camera
		roadRecoveryNotice038 = notice
		app.prevCar = prev
		autopilot034 = ap
	}()
	d := &engine.WorldDefinition{Grid: engine.HeightGrid{W: 3, H: 3, MinX: -100, MinZ: -100, Step: 100, Values: make([]float32, 9)}, Roads: []engine.RoadDefinition{{Width: 12, Points: []engine.Vec{{Z: -90}, {Z: 90}}}}}
	app.world = engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	app.world.SetTrafficDensity(0)
	app.world.Car.Pos = engine.Vec{X: 50}
	resetHold036(false, 0)
	for i := 0; i < 61; i++ {
		if resetHold036(true, 1./60) {
			recoverRoad038()
		}
	}
	if app.world.Car.Pos.X > 7 || !camera036.ResetFired {
		t.Fatal("road recovery did not execute", app.world.Car.Pos)
	}
	app.world.Car.Pos.X = 50
	for i := 0; i < 120; i++ {
		if resetHold036(true, 1./60) {
			t.Fatal("held key repeated")
		}
	}
	resetHold036(false, 0)
	for i := 0; i < 61; i++ {
		if resetHold036(true, 1./60) {
			recoverRoad038()
		}
	}
	if app.world.Car.Pos.X > 7 {
		t.Fatal("release failed to rearm")
	}
}

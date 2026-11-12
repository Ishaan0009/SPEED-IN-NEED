package engine

import (
	"math"
	"testing"
)

func isolated052(mapID int) *World {
	w := NewWorld(mapID)
	w.Traffic = nil
	w.Trees = nil
	w.VisibleTrees = 0
	w.Poles = nil
	w.Rocks = nil
	w.Buildings = nil
	w.Pickups = nil
	return w
}

func TestAlpha052DryWorldPlacement(t *testing.T) {
	for mapID := 0; mapID < 3; mapID++ {
		w := NewWorld(mapID)
		for i, tree := range w.Trees {
			if w.IsWater(tree.Pos, 5) || w.RampClearance(tree.Pos, 10) {
				t.Fatalf("map %d tree %d overlaps water or ramp corridor", mapID, i)
			}
		}
		for i, rock := range w.Rocks {
			if w.IsWater(rock.Pos, 5) || w.RampClearance(rock.Pos, 10) {
				t.Fatalf("map %d rock %d overlaps water or ramp corridor", mapID, i)
			}
		}
		if len(w.Ramps) != 2 {
			t.Fatalf("map %d missing ramps", mapID)
		}
	}
}

func TestAlpha052RampSurfaceAndFlight(t *testing.T) {
	w := isolated052(2)
	r := w.Ramps[0]
	f := HeadingVec(r.Heading)
	last := math.Inf(-1)
	for i := 0; i <= 100; i++ {
		p, _ := r.Sample(r.Length * float64(i) / 100)
		h, ok := r.Surface(p)
		if !ok || h < last-.05 {
			t.Fatalf("discontinuous ramp deck at %d: %g", i, h)
		}
		last = h
	}
	w.Car.Pos = r.Start.Sub(f.Mul(5))
	w.Car.Heading = r.Heading
	w.Car.Speed = 45
	w.InitializeChassis()
	airborne, landed := false, false
	for i := 0; i < 1800; i++ {
		w.Car.Pos, f = r.Sample(-5 + float64(i)*45./120)
		w.Car.Heading = math.Atan2(f.X, f.Z)
		w.StepChassis(Input{Throttle: 1}, 0, 1./120)
		if !w.Car.Body.Grounded && w.Car.Body.Height-w.GroundHeight(w.Car.Pos) > 2 {
			airborne = true
		}
		if airborne && w.Car.Body.Grounded {
			landed = true
		}
		for _, v := range []float64{w.Car.Body.Height, w.Car.Body.VerticalSpeed, w.Car.Rig.Roll, w.Car.Rig.Pitch} {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				t.Fatalf("ramp produced nonfinite state at tick %d", i)
			}
		}
	}
	if !airborne || !landed {
		t.Fatalf("ramp must launch and land: airborne=%t landed=%t", airborne, landed)
	}
}

func TestAlpha052AirNitroAddsForwardAndVerticalImpulse(t *testing.T) {
	a, b := isolated052(2), isolated052(2)
	for _, w := range []*World{a, b} {
		w.Car.Pos = Vec{}
		w.Car.Heading = 0
		w.Car.Speed = 30
		w.Car.Nitro = 100
		w.Car.Rig.Pitch = -.3
		w.Car.Body = ChassisState{Height: 30, VerticalSpeed: 4, Initialized: true, Grounded: false, AirTime: 1, AirVelocity: Vec{0, 30}}
	}
	for i := 0; i < 20; i++ {
		a.Step(Input{}, 1./120)
		b.Step(Input{Boost: true}, 1./120)
	}
	if b.Car.Body.VerticalSpeed < a.Car.Body.VerticalSpeed+.5 || b.Car.Pos.Z < a.Car.Pos.Z+.1 || b.Car.Nitro >= 100 {
		t.Fatalf("air nitro failed: ordinary=%+v boosted=%+v", a.Car.Body, b.Car.Body)
	}
}

func TestAlpha052AirBrakesCannotStopWorldMomentum(t *testing.T) {
	a, b := isolated052(2), isolated052(2)
	for _, w := range []*World{a, b} {
		w.Car.Pos = Vec{}
		w.Car.Speed = 30
		w.Car.Heading = 0
		w.Car.Body = ChassisState{Height: 40, Initialized: true, AirTime: 1, AirVelocity: Vec{0, 30}}
	}
	for i := 0; i < 30; i++ {
		a.Step(Input{}, 1./120)
		b.Step(Input{Brake: 1, Handbrake: true, Steer: 1}, 1./120)
	}
	if a.Car.Pos.Dist(b.Car.Pos) > .001 {
		t.Fatalf("air steering/braking changed ballistic momentum: %v vs %v", a.Car.Pos, b.Car.Pos)
	}
}

func TestAlpha052ParkedChassisSettles(t *testing.T) {
	w := isolated052(2)
	w.Car.Pos = Vec{0, 0}
	w.InitializeChassis()
	for i := 0; i < 1200; i++ {
		w.StepChassis(Input{}, 0, 1./120)
	}
	if math.Abs(w.Car.Rig.Roll) > .12 || math.Abs(w.Car.Body.RollRate) > .04 || math.Abs(w.Car.Body.VerticalSpeed) > .08 {
		t.Fatalf("parked chassis rocks: %+v %+v", w.Car.Body, w.Car.Rig)
	}
}

func TestAlpha052VerticalInterpolation(t *testing.T) {
	a, b := Car{}, Car{}
	a.Body.Height = 2
	b.Body.Height = 12
	if h := InterpolateCar(a, b, .25).Body.Height; h != 4.5 {
		t.Fatalf("jump interpolation: %g", h)
	}
}

func TestAlpha052InvalidMapAndZeroTick(t *testing.T) {
	w := NewWorld(99)
	if w.Map != 2 {
		t.Fatal("invalid map not clamped")
	}
	before := w.Car
	w.Step(Input{Throttle: 1, Boost: true}, 0)
	if before != w.Car {
		t.Fatal("zero-length tick changed state")
	}
}

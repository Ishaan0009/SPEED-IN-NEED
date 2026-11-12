package engine

import (
	"math"
	"testing"
)

func TestBeta012CombinedTireGripBudget(t *testing.T) {
	for class := 0; class < 4; class++ {
		w := remasteredFixture(0, class)
		w.Car.Speed = 24
		for tick := 0; tick < 900; tick++ {
			w.Step(Input{Throttle: .7, Steer: .65, Handbrake: tick > 200 && tick < 340}, 1./120)
			for _, f := range w.Car.TireForces {
				if f.NormalLoad < .1 {
					continue
				}
				q := math.Hypot(f.Longitudinal/(f.MuX*f.NormalLoad), f.Lateral/(f.MuY*f.NormalLoad))
				if math.IsNaN(q) || q > 1.00001 {
					t.Fatalf("class %d exceeded combined friction: %.3f", class, q)
				}
			}
		}
	}
}

func TestBeta012SpaceInitiatesAndReleaseRecoversDrift(t *testing.T) {
	for class := 0; class < 4; class++ {
		w := remasteredFixture(0, class)
		w.Car.Speed = 27
		peak := 0.0
		for tick := 0; tick < 110; tick++ {
			w.Step(Input{Throttle: .65, Steer: .65, Handbrake: true}, 1./120)
			peak = math.Max(peak, math.Abs(w.Car.DriftAngle))
		}
		if peak < 7 {
			t.Fatalf("class %d handbrake failed to initiate a slide: %.2f deg", class, peak)
		}
		for tick := 0; tick < 480; tick++ {
			w.Step(Input{Throttle: .5}, 1./120)
		}
		if w.Car.DriftActive || w.Car.DriftRecovery > 0 || math.Abs(w.Car.DriftAngle) > 8 {
			t.Fatalf("class %d failed drift recovery: %.2f", class, w.Car.DriftAngle)
		}
	}
}

func TestBeta012BrakeAndReverseDoNotRetainImpactLockout(t *testing.T) {
	for class := 0; class < 4; class++ {
		w := remasteredFixture(0, class)
		w.Car.Speed = 25
		for tick := 0; tick < 720 && w.Car.Speed > .2; tick++ {
			w.Step(Input{Brake: 1}, 1./120)
		}
		if w.Car.Speed > .2 {
			t.Fatal("service brake failed to stop")
		}
		w.Car.ContactImpulse = 30
		for tick := 0; tick < 60; tick++ {
			w.Step(Input{Brake: 1, Steer: .5}, 1./120)
		}
		if w.Car.Speed > -.3 {
			t.Fatalf("class %d reverse recovery too slow: %.2f", class, w.Car.Speed)
		}
	}
}

func TestBeta012TrafficStopsBeforeBusBodyAndTakesCorners(t *testing.T) {
	w := remasteredFixture(0, 0)
	w.Car.Pos = Vec{700, 700}
	p, f := w.Routes[0].AtDistance(130)
	w.Traffic = []Traffic{{Path: 0, T: 100, Direction: 1, Lane: 2.7, Speed: 12, CruiseSpeed: 15, Variant: 0}, {Path: 0, T: 130, Direction: 1, Lane: 2.7, Speed: 0, CruiseSpeed: 0, Variant: 0, Pos: p.Add(Vec{-f.Z, f.X}.Mul(2.7)), Heading: math.Atan2(f.X, f.Z)}}
	for tick := 0; tick < 1200; tick++ {
		w.stepRemasteredTraffic(1. / 120)
		a, b := w.Traffic[0], w.Traffic[1]
		if a.Pos.Dist(b.Pos) < 10.7 {
			t.Fatalf("bus body overlap: %.2f", a.Pos.Dist(b.Pos))
		}
	}
	if w.Traffic[0].Speed > .5 {
		t.Fatal("traffic did not queue behind the stopped bus")
	}
}

func TestBeta012WheelInterpolationCrossesPhaseWrap(t *testing.T) {
	a, b := Car{}, Car{}
	a.Rig.Wheels[0].RollAngle = Tau - .03
	b.Rig.Wheels[0].RollAngle = .03
	c := InterpolateCar(a, b, .5)
	if math.Abs(c.Rig.Wheels[0].RollAngle-Tau) > .001 {
		t.Fatal("wheel spins backward at phase wrap")
	}
}

func TestBeta012StaticAxleLoadsMatchVehicleBalance(t *testing.T) {
	for class := 0; class < 4; class++ {
		w := remasteredFixture(0, class)
		for tick := 0; tick < 360; tick++ {
			w.Step(Input{Handbrake: true}, 1./120)
		}
		front := w.Car.Rig.Wheels[0].NormalLoad + w.Car.Rig.Wheels[1].NormalLoad
		rear := w.Car.Rig.Wheels[2].NormalLoad + w.Car.Rig.Wheels[3].NormalLoad
		if math.Abs(front/(front+rear)-RemasteredVehicleSpec(class).FrontWeight) > .04 {
			t.Fatalf("class %d axle balance: %.3f", class, front/(front+rear))
		}
	}
}

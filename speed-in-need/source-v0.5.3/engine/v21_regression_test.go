package engine

import (
	"math"
	"testing"
)

func isolatedV21World() *World {
	w := NewWorld(0)
	w.Traffic = nil
	w.Trees = nil
	w.VisibleTrees = 0
	w.Poles = nil
	w.Buildings = nil
	w.Pickups = nil
	return w
}

func TestV21ServiceBrakeIsStrongerThanRearHandbrake(t *testing.T) {
	foot := isolatedV21World()
	rear := isolatedV21World()
	foot.Car.Speed, rear.Car.Speed = 22, 22 // ~79 km/h
	for n := 0; n < 60; n++ {
		foot.Step(Input{Brake: 1}, 1.0/120)
		rear.Step(Input{Handbrake: true}, 1.0/120)
	}
	if foot.Car.Speed >= rear.Car.Speed-2 {
		t.Fatalf("foot brake must be stronger: foot=%.2f rear=%.2f", foot.Car.Speed, rear.Car.Speed)
	}
	if foot.Car.Speed < -0.001 {
		t.Fatal("foot brake crossed into reverse without releasing the pedal")
	}
}

func TestV21KeyboardSteeringLowLock(t *testing.T) {
	w := isolatedV21World()
	w.Car.Speed = 22
	for n := 0; n < 45; n++ {
		w.Step(Input{Steer: 1}, 1.0/120)
	}
	if math.Abs(w.Car.VisualSteer) > .20 {
		t.Fatalf("steering too sharp at 80 km/h: %v", w.Car.VisualSteer)
	}
	if w.Car.DriftScore != 0 {
		t.Fatalf("ordinary corner scored handbrake drift: %v", w.Car.DriftScore)
	}
}

func TestV21DriftRequiresUsableSpeed(t *testing.T) {
	high := isolatedV21World()
	high.Car.Speed = 33 // 119 km/h; no full rear slide until below 100 km/h
	high.Step(Input{Steer: .9, Handbrake: true}, 1.0/120)
	if math.Abs(high.Car.Lateral) > .3 {
		t.Fatalf("high speed e-brake caused excessive immediate slide: %.2f", high.Car.Lateral)
	}
	low := isolatedV21World()
	low.Car.Speed = 20
	for n := 0; n < 90; n++ {
		low.Step(Input{Steer: .75, Handbrake: true, Throttle: .6}, 1.0/120)
	}
	if math.Abs(low.Car.DriftAngle) < 5 {
		t.Fatalf("handbrake failed to initiate controlled drift: %.2f", low.Car.DriftAngle)
	}
	for n := 0; n < 160; n++ {
		low.Step(Input{}, 1.0/120)
	}
	if math.Abs(low.Car.DriftAngle) > 4 {
		t.Fatalf("rear tires did not regain grip: %.2f", low.Car.DriftAngle)
	}
}

func TestV21RoadNormalsJoinWithoutCracks(t *testing.T) {
	for _, route := range BuildRoutes() {
		n := len(route.Points)
		for i := 0; i < n; i++ {
			normal := RoadNormalAt(route.Points, i)
			if math.Abs(normal.Length()-1) > .001 {
				t.Fatalf("invalid road edge normal at %s point %d", route.Name, i)
			}
		}
		if route.Closed && RoadNormalAt(route.Points, 0).Dist(RoadNormalAt(route.Points, n-1)) > .001 {
			t.Fatalf("road loop has visible seam at %s", route.Name)
		}
	}
}

func TestV21FootBrakeAtRestDoesNotReverse(t *testing.T) {
	w := isolatedV21World()
	w.Car.Speed = 0
	// Reverse does not engage immediately on first brake contact.
	for n := 0; n < 25; n++ {
		w.Step(Input{Brake: 1}, 1.0/120)
	}
	if math.Abs(w.Car.Speed) > .0001 {
		t.Fatalf("brake selected reverse too early: %.3f", w.Car.Speed)
	}
	// Holding S at rest is now deliberately defined to select reverse.
	for n := 0; n < 140; n++ {
		w.Step(Input{Brake: 1}, 1.0/120)
	}
	if w.Car.Speed >= -.05 {
		t.Fatalf("brake hold did not select reverse: %.3f", w.Car.Speed)
	}
}

package engine

import (
	"math"
	"testing"
)

func TestBeta4StraightDrivingDoesNotDrift(t *testing.T) {
	w := NewWorld(0)
	w.Pickups = nil
	w.Traffic = nil
	w.Trees = nil
	w.Poles = nil
	for i := 0; i < 450; i++ {
		w.Step(Input{Throttle: 1}, 1./120)
	}
	if math.Abs(w.Car.Lateral) > 0.25 || math.Abs(w.Car.DriftAngle) > 1 {
		t.Fatalf("phantom drift: lateral=%v angle=%v", w.Car.Lateral, w.Car.DriftAngle)
	}
	if w.Car.DriftScore != 0 {
		t.Fatalf("unintentional drift scoring %v", w.Car.DriftScore)
	}
	if w.Car.WheelSpin == 0 {
		t.Fatal("wheel animation not linked to distance")
	}
}
func TestBeta4HandbrakeDriftAndRecovery(t *testing.T) {
	w := NewWorld(0)
	w.Pickups = nil
	w.Traffic = nil
	w.Trees = nil
	w.Poles = nil
	for i := 0; i < 230; i++ {
		w.Step(Input{Throttle: 1}, 1./120)
	}
	for i := 0; i < 110; i++ {
		w.Step(Input{Throttle: 1, Steer: -.7, Handbrake: true}, 1./120)
	}
	angle := math.Abs(w.Car.DriftAngle)
	if angle < 7 {
		t.Fatalf("handbrake did not create rear slip: angle=%v lateral=%v", angle, w.Car.Lateral)
	}
	for i := 0; i < 180; i++ {
		w.Step(Input{Throttle: 0, Steer: 0}, 1./120)
	}
	if math.Abs(w.Car.DriftAngle) > 4 {
		t.Fatalf("failed to regain grip: angle=%v", w.Car.DriftAngle)
	}
}
func TestBeta4WheelsTurnWithSteering(t *testing.T) {
	w := NewWorld(0)
	for i := 0; i < 160; i++ {
		w.Step(Input{Throttle: 1, Steer: -.7}, 1./120)
	}
	if w.Car.VisualSteer >= 0 {
		t.Fatalf("visual wheel direction wrong %v", w.Car.VisualSteer)
	}
}

package engine

import "testing"

func TestAlphaCrashReducesSpeedWithoutReverse(t *testing.T) {
	w := NewWorld(0)
	w.Trees = []Tree{{Pos: w.Car.Pos.Add(Vec{0, 2}), Size: 3}}
	w.Traffic = nil
	w.Buildings = nil
	w.Car.Heading = 0
	w.Car.Speed = 28
	w.Step(Input{}, 1.0/120.0)
	if w.Car.Speed < 0 {
		t.Fatalf("collision reversed vehicle: %v", w.Car.Speed)
	}
}
func TestAlphaMapSelectionAllOpen(t *testing.T) {
	for id := 0; id < 3; id++ {
		w := NewWorld(id)
		if w.Map != id {
			t.Fatal("map selection mismatch")
		}
	}
}
func TestAlphaNoSpontaneousSlip(t *testing.T) {
	w := NewWorld(0)
	w.Traffic = nil
	w.Trees = nil
	w.Poles = nil
	w.Car.Speed = 30
	for i := 0; i < 100; i++ {
		w.Step(Input{}, 1.0/120.0)
	}
	if w.Car.DriftScore != 0 {
		t.Fatalf("spontaneous drift: %v", w.Car.DriftScore)
	}
}

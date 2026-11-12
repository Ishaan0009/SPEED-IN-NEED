package engine

import (
	"math"
	"testing"
)

func TestInputs(t *testing.T) {
	w := NewWorld(0)
	for i := 0; i < 130; i++ {
		w.Step(Input{Throttle: 1}, 1./120)
	}
	prev := w.Car.Heading
	for i := 0; i < 110; i++ {
		w.Step(Input{Throttle: 1, Steer: -1}, 1./120)
	}
	if w.Car.Heading >= prev {
		t.Fatal("A should steer left")
	}
	w.Reset()
	for i := 0; i < 130; i++ {
		w.Step(Input{Throttle: 1}, 1./120)
	}
	prev = w.Car.Heading
	for i := 0; i < 110; i++ {
		w.Step(Input{Throttle: 1, Steer: 1}, 1./120)
	}
	if w.Car.Heading <= prev {
		t.Fatal("D should steer right")
	}
}
func TestBoost(t *testing.T) {
	w := NewWorld(0)
	w.Car.Speed = 20
	w.Car.Nitro = 5
	w.Pickups = nil
	for i := 0; i < 200; i++ {
		w.Step(Input{Boost: true, Throttle: 1}, 1./120)
	}
	if w.Car.Nitro > 0 {
		t.Fatalf("unlimited nitro: %v", w.Car.Nitro)
	}
}
func TestStability(t *testing.T) {
	w := NewWorld(0)
	for i := 0; i < 1700; i++ {
		w.Step(Input{Throttle: 1, Steer: math.Sin(float64(i) * .03), Handbrake: i%300 > 180}, 1./120)
	}
	if math.IsNaN(w.Car.Pos.X) || math.IsNaN(w.Car.YawRate) {
		t.Fatal("unstable handling")
	}
}
func TestCheckpoint(t *testing.T) {
	w := NewWorld(0)
	w.StartChallenge(1)
	if len(w.Checkpoints) == 0 {
		t.Fatal("no checkpoint")
	}
	w.Car.Pos = w.Checkpoints[0]
	w.Step(Input{}, 1./120)
	if w.Checkpoint != 1 {
		t.Fatal("checkpoint not advanced")
	}
}

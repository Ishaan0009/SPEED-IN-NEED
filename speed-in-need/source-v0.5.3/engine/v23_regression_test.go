package engine

import (
	"math"
	"testing"
)

func TestV23ServiceBrakeThenReverse(t *testing.T) {
	w := isolatedV21World()
	w.Car.Speed = 18 // starts moving forward
	for i := 0; i < 150; i++ {
		w.Step(Input{Brake: 1}, 1.0/120)
	}
	if w.Car.Speed >= -0.01 {
		t.Fatalf("S was held long enough to stop then reverse: speed %.3f", w.Car.Speed)
	}
	if w.Car.Speed < -8.01 {
		t.Fatalf("reverse exceeded safety speed: %.3f", w.Car.Speed)
	}
	w.Step(Input{}, 1.0/120)
	if w.Car.ReverseHold > 0 {
		t.Fatalf("release must reset reverse selector hold: %.4f", w.Car.ReverseHold)
	}
}
func TestV23NoInstantNitroSpeedCollapse(t *testing.T) {
	w := isolatedV21World()
	w.Car.Speed = 43
	w.Car.Nitro = 100
	for i := 0; i < 80; i++ {
		w.Step(Input{Boost: true, Throttle: 1}, 1.0/120)
	}
	if w.Car.Speed <= 44 {
		t.Fatalf("boost did not increase top speed: %.2f", w.Car.Speed)
	}
	prev := w.Car.Speed
	w.Step(Input{Throttle: 1}, 1.0/120)
	if prev-w.Car.Speed > .7 {
		t.Fatalf("nitro released caused unrealistic instantaneous speed loss: %.2f -> %.2f", prev, w.Car.Speed)
	}
	if math.IsNaN(w.Car.Speed) {
		t.Fatal("speed invalid")
	}
}
func TestV23TreeColliderSmallerThanCanopy(t *testing.T) {
	w := isolatedV21World()
	w.Trees = []Tree{{Pos: Vec{100, 100}, Size: 2.0}}
	w.VisibleTrees = 1
	w.Car.Pos = Vec{103.0, 100}
	w.Car.Speed = 0
	w.Step(Input{}, 1.0/120)
	if w.Car.Pos.X != 103 {
		t.Fatalf("tree visually distant but caused invisible contact at %.2f", w.Car.Pos.X)
	}
}

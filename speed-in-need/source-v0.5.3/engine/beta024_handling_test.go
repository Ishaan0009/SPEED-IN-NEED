package engine

import (
	"math"
	"testing"
)

func TestBeta024ThrottleDoesNotTurnLightBrakingIntoFullBraking(t *testing.T) {
	for car := 0; car < 4; car++ {
		light, full := flat22(car, true), flat22(car, true)
		for _, w := range []*World{light, full} {
			w.Car.Speed = 20
			w.Car.Gearbox.Gear = 3
		}
		for tick := 0; tick < 90; tick++ {
			light.Step(Input{Throttle: 1, Brake: .2}, 1./120)
			full.Step(Input{Throttle: 1, Brake: 1}, 1./120)
		}
		if light.Car.Speed < full.Car.Speed+1 {
			t.Fatalf("car %d: light %.3f full %.3f", car, light.Car.Speed, full.Car.Speed)
		}
	}
}
func TestBeta024VergeDoesNotAbruptlyDropGrip(t *testing.T) {
	w := flat22(0, false)
	w.Car.Pos = Vec{0, 4.8}
	w.Car.Heading = math.Pi / 2
	w.Car.Speed = 12
	w.Step(Input{Throttle: .2}, 1./120)
	before := w.Car.TireForces[0].MuX
	w.Car.Pos.Z = 5.5
	w.Step(Input{Throttle: .2}, 1./120)
	after := w.Car.TireForces[0].MuX
	if before-after > .10 {
		t.Fatalf("single-tick verge grip step: %.3f -> %.3f", before, after)
	}
	if w.Car.DriftActive {
		t.Fatal("crossing verge enabled handbrake drift")
	}
}

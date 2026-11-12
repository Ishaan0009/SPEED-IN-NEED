package engine

import (
	"math"
	"testing"
)

func TestBeta023GrassDragIsGentleAndCannotReverseRest(t *testing.T) {
	for _, v := range []float64{-.001, 0, .001, 1, 15, 30, 60} {
		f := OffroadResistance(1400, 1400*9.81, v)
		if f*v < 0 || math.Abs(f)/1400 > 1.2 {
			t.Fatalf("unreasonable turf drag at %f: %f", v, f)
		}
	}
	for class := 0; class < 4; class++ {
		road, turf := flat22(class, true), flat22(class, false)
		for tick := 0; tick < 1200; tick++ {
			road.Step(Input{Throttle: 1}, 1./120)
			turf.Step(Input{Throttle: 1}, 1./120)
		}
		if turf.Car.Speed < road.Car.Speed*.60 {
			t.Fatalf("class %d bogged down: road %.1f grass %.1f", class, road.Car.Speed, turf.Car.Speed)
		}
	}
}
func TestBeta023SlowSignalImpactReleasesWithoutStopping(t *testing.T) {
	w := flat22(0, true)
	w.Remaster.Def.Objects = []PlacedObject{{Kind: "signal", Scale: [3]float64{1, 1, 1}}}
	w.Remaster.Def.Colliders = []ColliderShape{{Object: 0, Kind: "signal", Center: Vec{0, 4}, Radius: .14, MinY: 0, MaxY: 7}}
	w.Remaster.staticGrid = map[gridKey][]int{gridOf(Vec{0, 4}): {0}}
	w.Car.Pos = Vec{0, 1.8}
	w.Car.Heading = 0
	w.Car.Speed = .6
	w.Car.Body.Height = .74
	w.remasteredStaticContacts(&w.Car, 0)
	if w.Remaster.FallenPoles[0] == nil || w.Car.Speed < .4 {
		t.Fatalf("slow signal remains a rigid wall: speed %f", w.Car.Speed)
	}
}

func TestBeta023AcceleratingAgainstSignalFromRestDoesNotStick(t *testing.T) {
	w := flat22(0, true)
	w.Remaster.Def.Objects = []PlacedObject{{Kind: "signal", Scale: [3]float64{1, 1, 1}}}
	w.Remaster.Def.Colliders = []ColliderShape{{Object: 0, Kind: "signal", Center: Vec{0, 4}, Radius: .14, MinY: 0, MaxY: 7}}
	w.Remaster.staticGrid = map[gridKey][]int{gridOf(Vec{0, 4}): {0}}
	w.Car.Pos = Vec{0, 1.8}
	w.Car.Body.Height = .74
	for i := 0; i < 240; i++ {
		w.Step(Input{Throttle: 1}, 1./120)
	}
	if w.Remaster.FallenPoles[0] == nil || w.Car.Speed < 1 {
		t.Fatalf("car remains stuck pushing a signal: speed %.3f", w.Car.Speed)
	}
}

package engine

import (
	"math"
	"testing"
)

func TestV2TrafficCruiseAndHeading(t *testing.T) {
	for mapID := 0; mapID < 3; mapID++ {
		w := NewWorld(mapID)
		w.Trees = nil
		w.VisibleTrees = 0
		w.Buildings = nil
		w.Poles = nil
		for i, car := range w.Traffic {
			if car.Speed < 0 || car.Speed > car.CruiseSpeed || car.CruiseSpeed > 12 {
				t.Fatalf("bad NPC speed %v car %d", car, mapID)
			}
			if car.Variant%4 == 3 && car.CruiseSpeed > 6.7 {
				t.Fatalf("truck too fast %v", car.CruiseSpeed)
			}
			_, dir := w.Routes[car.Path].AtDistance(car.T)
			dir = dir.Mul(car.Direction)
			if math.Abs(math.Atan2(dir.X, dir.Z)-car.Heading) > 0.01 {
				t.Fatalf("wrong NPC heading %v", i)
			}
		}
		for frame := 0; frame < 200; frame++ {
			w.Step(Input{}, 1.0/120)
		}
		for i, car := range w.Traffic {
			if car.Speed < 0 || car.Speed > car.CruiseSpeed+.001 {
				t.Fatalf("NPC speed drift %d:%v", i, car.Speed)
			}
			if math.IsNaN(car.Heading) || math.IsInf(car.Heading, 0) {
				t.Fatalf("nonfinite heading")
			}
		}
	}
}
func TestV2NoWorldClamp(t *testing.T) {
	w := NewWorld(0)
	w.Car.Pos = Vec{X: 481, Z: 0}
	w.Traffic = nil
	w.Trees = nil
	w.VisibleTrees = 0
	w.Poles = nil
	w.Step(Input{}, 1.0/120)
	if w.Car.Pos.X < 480 {
		t.Fatalf("invisible world wall: %+v", w.Car.Pos)
	}
}
func TestV2InvisibleTreeCollisionMatchesDrawBudget(t *testing.T) {
	w := NewWorld(0)
	w.Traffic = nil
	w.Buildings = nil
	w.Poles = nil
	w.Trees = []Tree{{Pos: Vec{X: 40, Z: 40}, Size: 3}, {Pos: w.Car.Pos, Size: 4}}
	w.VisibleTrees = 1
	pos := w.Car.Pos
	w.Step(Input{}, 1.0/120)
	if w.Car.Pos.Dist(pos) > .01 {
		t.Fatalf("collided with an invisible decorative tree: %v", w.Car.Pos.Dist(pos))
	}
}
func TestV2SpeedLimits(t *testing.T) {
	w := NewWorld(0)
	w.Traffic = nil
	w.Trees = nil
	w.VisibleTrees = 0
	w.Poles = nil
	for i := 0; i < 120*50; i++ {
		w.Step(Input{Throttle: 1}, 1.0/120)
	}
	if w.Car.Speed > 44.01 {
		t.Fatalf("player speed over cap %v", w.Car.Speed)
	}
}

func TestV2TrafficTurnsAroundAtOpenRouteEnds(t *testing.T) {
	w := NewWorld(0)
	w.Traffic = nil
	w.Trees = nil
	w.Buildings = nil
	w.Poles = nil
	r := w.Routes[1]
	if r.Closed {
		t.Fatal("connector should be open")
	}
	w.Traffic = []Traffic{{Path: 1, T: r.Len - .02, Speed: 8, CruiseSpeed: 8, Direction: 1, Lane: 3.1}}
	w.Step(Input{}, 1.0/120)
	if w.Traffic[0].Direction != -1 {
		t.Fatalf("car teleported at connector end instead of turning: %+v", w.Traffic[0])
	}
	if w.Traffic[0].T > r.Len || w.Traffic[0].T < 0 {
		t.Fatal("wrong endpoint distance")
	}
}

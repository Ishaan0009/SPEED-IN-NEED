package engine

import (
	"math"
	"testing"
)

func TestAlpha04DistinctRoadMaps(t *testing.T) {
	signatures := make(map[string]bool)
	for id := 0; id < 3; id++ {
		world := NewWorld(id)
		if world.Map != id || len(world.Routes) != 3 {
			t.Fatalf("bad map %d", id)
		}
		// Roads, traffic and the GPS all share this one per-map route set.
		first := world.Routes[0].Points[0]
		signature := world.Routes[0].Name
		if signatures[signature] {
			t.Fatalf("map %d reuses road %s", id, signature)
		}
		signatures[signature] = true
		if len(world.Routes[0].Points) < 100 || world.Routes[0].Len < 450 {
			t.Fatalf("short route %d", id)
		}
		if id != 0 && DistRoad(world.Car.Pos, world.Routes) > 3 {
			t.Fatalf("map %d player off route: %v first %v", id, world.Car.Pos, first)
		}
	}
}

func TestAlpha04StationaryBurnoutAndRelease(t *testing.T) {
	world := NewWorld(0)
	world.Traffic = nil
	world.Trees = nil
	world.VisibleTrees = 0
	world.Poles = nil
	world.Rocks = nil
	for i := 0; i < 200; i++ {
		world.Step(Input{Throttle: 1, Handbrake: true}, 1./120)
	}
	if !world.Car.Burnout || math.Abs(world.Car.Speed) > .05 {
		t.Fatalf("handbrake engine hold: %+v", world.Car)
	}
	if world.Car.WheelSpin != 0 {
		t.Fatalf("rear tires rolling during brake lock: %v", world.Car.WheelSpin)
	}
	for i := 0; i < 50; i++ {
		world.Step(Input{Throttle: 1}, 1./120)
	}
	if world.Car.Burnout || world.Car.Speed < 1 {
		t.Fatalf("cannot accelerate after releasing handbrake: %+v", world.Car)
	}
}

func TestAlpha04RocksCollideAndTreesRemainSolid(t *testing.T) {
	w := NewWorld(2)
	w.Traffic = nil
	w.Poles = nil
	w.Trees = nil
	w.VisibleTrees = 0
	if len(w.Rocks) == 0 {
		t.Fatal("no roadside rocks")
	}
	rock := w.Rocks[0]
	w.Car.Pos = rock.Pos.Add(Vec{.4, .3})
	w.Car.Speed = 20
	w.Step(Input{}, 1./120)
	if w.Car.Pos.Dist(rock.Pos) < rock.Radius+.98 {
		t.Fatalf("rock ignored: %+v", w.Car)
	}
	if w.Car.Speed >= 20 {
		t.Fatal("rock impact did not reduce speed")
	}
	w.Rocks = nil
	w.Trees = []Tree{{Pos: Vec{X: 70, Z: 40}, Size: 1.6}}
	w.VisibleTrees = 1
	w.Car.Pos = Vec{70.2, 40.3}
	w.Car.Speed = 10
	w.Step(Input{}, 1./120)
	if w.Car.Speed >= 10 {
		t.Fatal("tree trunk not solid")
	}
}

func TestAlpha04BreakawayPoleHasLightImpact(t *testing.T) {
	w := NewWorld(0)
	w.Traffic = nil
	w.Trees = nil
	w.VisibleTrees = 0
	w.Rocks = nil
	w.Poles = []Pole{{Pos: Vec{X: 66, Z: 40}}}
	w.Car.Pos = Vec{X: 66.3, Z: 40}
	w.Car.Speed = 15
	w.Step(Input{}, 1./120)
	if !w.Poles[0].Falling {
		t.Fatal("streetlight not falling")
	}
	if w.Car.Speed < 11.5 {
		t.Fatalf("lamp impact too strong: %v", w.Car.Speed)
	}
}

func TestAlpha04CoastalWaterBlocked(t *testing.T) {
	w := NewWorld(0)
	w.Traffic = nil
	w.Trees = nil
	w.VisibleTrees = 0
	w.Poles = nil
	w.Rocks = nil
	w.Car.Pos = Vec{X: -260, Z: -285}
	w.Car.Speed = 20
	w.Step(Input{}, 1./120)
	if w.Car.Pos.X < -203.01 || w.Car.Speed > 6 {
		t.Fatalf("car drives on water: %+v", w.Car)
	}
}

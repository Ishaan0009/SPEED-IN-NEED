package engine

import (
	"math"
	"testing"
)

func TestAokawaBarrierTransfersMomentumAndSettles(t *testing.T) {
	w := remasteredFixture(0, 0)
	w.Car.Pos = Vec{}
	w.Car.Heading = 0
	w.Car.Speed = 20
	w.Remaster.Def.Objects = []PlacedObject{{Asset: "visible barrier", Kind: "barrier", Solid: true, Collider: 0}}
	w.Remaster.Def.Colliders = []ColliderShape{{Center: Vec{0, 2.4}, Half: Vec{1, .3}, MinY: 0, MaxY: .8, Object: 0, Kind: "barrier"}}
	w.Remaster.buildSpatial(w)
	w.remasteredStaticContacts(&w.Car, 0)
	p := w.Remaster.DynamicProps[0]
	if p == nil || p.Velocity.Z < 5 || w.Car.Speed < 10 {
		t.Fatal("barrier behaves like a fixed wall")
	}
	for i := 0; i < 120*12; i++ {
		w.stepAokawaProps(1. / 120)
	}
	if p.Pos.Z < 5 || p.Velocity.Length() > .01 || p.Height < -.001 || math.IsNaN(p.Roll) {
		t.Fatalf("barrier did not settle %+v", p)
	}
	// The old placement is empty after the barrier moves.
	w.Car.Pos = Vec{}
	w.Car.Speed = 5
	w.remasteredStaticContacts(&w.Car, 0)
	if w.Car.Speed != 5 {
		t.Fatal("ghost barrier remained at its old location")
	}
}
func TestAokawaCoastingBrakeLamps(t *testing.T) {
	for class := 0; class < 4; class++ {
		w := remasteredFixture(0, class)
		w.Car.Speed = 20
		w.Step(Input{}, 1./120)
		if !w.Car.CoastingLights {
			t.Fatal("release-throttle lamps missing")
		}
		w.Step(Input{Throttle: 1}, 1./120)
		if w.Car.CoastingLights {
			t.Fatal("coasting lamps stuck on under throttle")
		}
	}
}
func TestAokawaAirNitroRemainsHeavy(t *testing.T) {
	for class := 0; class < 4; class++ {
		a, b := remasteredFixture(0, class), remasteredFixture(0, class)
		for _, w := range []*World{a, b} {
			w.Car.Body.Height = 30
			w.Car.Body.Grounded = false
			w.Car.Speed = 30
			w.Car.Body.AirVelocity = HeadingVec(w.Car.Heading).Mul(30)
		}
		a.Step(Input{}, 1./120)
		b.Step(Input{Boost: true}, 1./120)
		extra := (b.Car.Speed - a.Car.Speed) * 120
		if extra < .2 || extra > 1 {
			t.Fatalf("air nitro extra acceleration %.2f", extra)
		}
		if math.Abs(b.Car.Body.VerticalSpeed-a.Car.Body.VerticalSpeed) > .001 {
			t.Fatal("level nitro lifts the car")
		}
	}
}
func TestAokawaNPCsKeepTheRightLane(t *testing.T) {
	w := remasteredFixture(0, 0)
	w.Traffic = []Traffic{{Path: 0, T: 100, Direction: 1, Lane: 2.7, Speed: 8, CruiseSpeed: 8}, {Path: 0, T: 400, Direction: -1, Lane: 2.7, Speed: 8, CruiseSpeed: 8}}
	w.stepRemasteredTraffic(1. / 120)
	for _, npc := range w.Traffic {
		p, f := w.Routes[0].AtDistance(npc.T)
		f = f.Mul(npc.Direction)
		// Horizontal screen-right from the renderer's forward-cross-up basis.
		right := Vec{-f.Z, f.X}
		delta := npc.Pos.Sub(p)
		if delta.X*right.X+delta.Z*right.Z < 2.5 {
			t.Fatal("NPC on left lane")
		}
	}
}

func TestAokawaNPCContinuesThroughConnectedRoad(t *testing.T) {
	w := remasteredFixture(0, 0)
	w.Routes = []Route{newRoute("approach", []Vec{{0, 0}, {0, 100}}), newRoute("cross street", []Vec{{-100, 100}, {100, 100}})}
	w.Remaster.Def.Roads = []RoadDefinition{{Kind: "town"}, {Kind: "town"}}
	npc := Traffic{Path: 0, T: 100.2, Direction: 1, Lane: 2.7, Speed: 10, Pos: Vec{2.7, 100}, Heading: 0}
	if !w.connectAokawaTraffic(&npc, .2) || npc.Path != 1 {
		t.Fatal("NPC failed to join connected road")
	}
	if math.Abs(npc.YawKick) > math.Pi*.6 {
		t.Fatal("NPC chose an illegal U-turn")
	}
}

func TestAokawaDeadEndStopsWithoutRepeatedAcceleration(t *testing.T) {
	w := remasteredFixture(0, 0)
	w.Routes = []Route{newRoute("dead end", []Vec{{0, 0}, {0, 100}})}
	w.Remaster.Def.Roads = []RoadDefinition{{Kind: "town"}}
	w.Car.Pos = Vec{400, 400}
	w.Traffic = []Traffic{{Path: 0, T: 99.99, Speed: 10, CruiseSpeed: 10, Direction: 1, Lane: 2.7, Pos: Vec{2.7, 99.99}}}
	for i := 0; i < 120; i++ {
		w.stepRemasteredTraffic(1. / 120)
	}
	npc := w.Traffic[0]
	if npc.Speed != 0 || npc.CruiseSpeed != 0 || npc.Direction != 1 {
		t.Fatal("dead-end traffic keeps accelerating or makes a U-turn")
	}
}
func TestAokawaNPCStopsBeforeRedAndContinuesOnGreen(t *testing.T) {
	w := remasteredFixture(0, 0)
	w.Routes = []Route{newRoute("straight", []Vec{{0, 0}, {0, 300}})}
	w.Car.Pos = Vec{300, 300}
	w.Traffic = []Traffic{{Path: 0, T: 20, Direction: 1, Lane: 2.7, Speed: 12, CruiseSpeed: 12, Pos: Vec{2.7, 20}}}
	w.Remaster.Def.Signals = []SignalDefinition{{Route: 0, Distance: 80, Direction: 1, Phase: 20}}
	for i := 0; i < 120*12; i++ {
		w.stepRemasteredTraffic(1. / 120)
	}
	if w.Traffic[0].T > 76.3 || w.Traffic[0].Speed > .1 {
		t.Fatalf("NPC crossed stop line at %.2f / %.2f", w.Traffic[0].T, w.Traffic[0].Speed)
	}
	w.Remaster.Def.Signals[0].Phase = 0
	for i := 0; i < 120*8; i++ {
		w.stepRemasteredTraffic(1. / 120)
	}
	if w.Traffic[0].T < 90 {
		t.Fatal("NPC failed to leave on green")
	}
}

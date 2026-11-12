package engine

import (
	"math"
	"testing"
)

func TestBeta013MarkingsStopAtCrossingCarriageways(t *testing.T) {
	d := remasteredFixture(0, 0).Remaster.Def
	d.Roads = []RoadDefinition{{Points: []Vec{{-100, 0}, {100, 0}}, Width: 12}, {Points: []Vec{{0, -100}, {0, 100}}, Width: 12}}
	w := NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	for _, p := range []Vec{{0, 0}, {4, 0}, {-4, 5}} {
		if w.RoadMarkingAllowed(p, Vec{1, 0}, 0) {
			t.Fatalf("lane paint crosses junction at %v", p)
		}
	}
	if !w.RoadMarkingAllowed(Vec{-30, 0}, Vec{1, 0}, 0) {
		t.Fatal("ordinary road markings removed")
	}
}

// This tests the transient that caused snap-back after a handbrake release,
// including both steering directions and all four independently tuned cars.
func TestBeta013GripSteeringReturnsWithoutSnap(t *testing.T) {
	for class := 0; class < 4; class++ {
		for _, sign := range []float64{-1, 1} {
			w := remasteredFixture(0, class)
			w.Car.Speed = 28
			for tick := 0; tick < 80; tick++ {
				w.Step(Input{Throttle: .5, Steer: sign * .7, Handbrake: true}, 1./120)
			}
			for tick := 0; tick < 360; tick++ {
				w.Step(Input{Throttle: .3}, 1./120)
				if math.IsNaN(w.Car.YawRate) || math.Abs(w.Car.YawRate) > 2.4 || math.Abs(w.Car.Rig.Roll) > .191 {
					t.Fatal("unstable chassis")
				}
			}
			if math.Abs(w.Car.VisualSteer) > .001 || math.Abs(w.Car.YawRate) > .15 || math.Abs(w.Car.DriftAngle) > 8 {
				t.Fatalf("class %d did not settle: steer %.3f yaw %.3f slip %.2f", class, w.Car.VisualSteer, w.Car.YawRate, w.Car.DriftAngle)
			}
		}
	}
}

func TestBeta013TrafficFacesItsVelocityInBothLanes(t *testing.T) {
	w := remasteredFixture(0, 0)
	w.Car.Pos = Vec{700, 700}
	for variant := 0; variant < 14; variant++ {
		for _, direction := range []float64{-1, 1} {
			p, f := w.Routes[0].AtDistance(130)
			f = f.Mul(direction)
			w.Traffic = []Traffic{{Path: 0, T: 130, Direction: direction, Lane: 2.7, Speed: 10, CruiseSpeed: 12, Variant: variant, Pos: p.Add(Vec{-f.Z, f.X}.Mul(2.7)), Heading: math.Atan2(f.X, f.Z)}}
			before := w.Traffic[0].Pos
			w.stepRemasteredTraffic(1. / 120)
			actor := w.Traffic[0]
			motion := actor.Pos.Sub(before)
			forward := HeadingVec(actor.Heading)
			if motion.X*forward.X+motion.Z*forward.Z <= 0 {
				t.Fatalf("variant %d direction %.0f travels backwards", variant, direction)
			}
		}
	}
}

func TestBeta013SignalAndSignBreakAwayAtTheirOwnHeight(t *testing.T) {
	for _, kind := range []string{"signal", "sign", "road-tree"} {
		w := remasteredFixture(0, 0)
		d := w.Remaster.Def
		d.Objects = []PlacedObject{{Kind: kind, Pos: [3]float64{0, 0, 4}, Scale: [3]float64{1, 1, 1}, Solid: true, Collider: 0}}
		d.Colliders = []ColliderShape{{Center: Vec{0, 4}, Radius: .16, MinY: 0, MaxY: 3, Kind: kind, Object: 0}}
		w.Remaster.staticGrid = map[gridKey][]int{gridOf(Vec{0, 4}): {0}}
		w.Car.Pos = Vec{0, 1.8}
		w.Car.Heading = 0
		w.Car.Speed = 12
		w.Car.Body.Height = .74
		w.remasteredStaticContacts(&w.Car, 0)
		p := w.Remaster.FallenPoles[0]
		if p == nil || !p.Falling || p.Length != 3 {
			t.Fatalf("%s stays rigid", kind)
		}
		for i := 0; i < 1200; i++ {
			w.stepRemasteredPoles(1. / 120)
		}
		if !p.Settled || p.Angle < 1.3 || p.Angle > math.Pi/2 {
			t.Fatalf("%s does not settle: %+v", kind, p)
		}
	}
}

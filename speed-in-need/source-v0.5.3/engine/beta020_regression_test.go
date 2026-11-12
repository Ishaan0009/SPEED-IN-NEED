package engine

import (
	"math"
	"testing"
)

func TestBeta020NormalDrivingRecoversLateralDisturbance(t *testing.T) {
	for class := 0; class < 4; class++ {
		w := remasteredFixture(0, class)
		w.Car.Speed = 20
		w.Car.Lateral = 3
		initial := math.Hypot(w.Car.Speed, w.Car.Lateral)
		for step := 0; step < 300; step++ {
			w.Step(Input{}, 1./120)
		}
		if math.Abs(w.Car.Lateral) > .8 || math.Abs(w.Car.YawRate) > .20 || w.Car.DriftActive || w.Car.DriftScore != 0 {
			t.Fatalf("class %d normal grip failed: lateral %.3f yaw %.3f drift %v", class, w.Car.Lateral, w.Car.YawRate, w.Car.DriftActive)
		}
		if math.Hypot(w.Car.Speed, w.Car.Lateral) > initial*1.02 {
			t.Fatalf("class %d grip created kinetic energy", class)
		}
	}
}

func TestBeta020HandbrakeRetainsLockedRearAndRollingFront(t *testing.T) {
	for class := 0; class < 4; class++ {
		w := remasteredFixture(0, class)
		w.Car.Speed = 23
		for i := 0; i < 30; i++ {
			w.Step(Input{Handbrake: true, Steer: .6}, 1./120)
		}
		if !w.Car.DriftActive {
			t.Fatalf("class %d handbrake drift unavailable", class)
		}
		for i := 2; i < 4; i++ {
			if w.Car.Rig.Wheels[i].AngularVelocity != 0 {
				t.Fatal("rear handbrake did not lock")
			}
		}
		if math.Abs(w.Car.Rig.Wheels[0].AngularVelocity) < 1 {
			t.Fatal("front wheel incorrectly locked")
		}
	}
}

func TestBeta020TrafficSpatialQueryKeepsEveryBrakingCandidate(t *testing.T) {
	s := &RemasteredState{}
	var actors []Traffic
	for x := -110.; x <= 110; x += 11 {
		for z := -110.; z <= 110; z += 11 {
			actors = append(actors, Traffic{Pos: Vec{x, z}, Speed: 10})
		}
	}
	s.beta020TrafficGrid(actors)
	for _, p := range []Vec{{0, 0}, {39.99, 39.99}, {-39.99, -.01}} {
		found := map[int]bool{}
		for _, i := range s.beta020TrafficNearby(p) {
			found[i] = true
		}
		for i, actor := range actors {
			if actor.Pos.Dist(p) <= 65 && !found[i] {
				t.Fatalf("spatial query missed actor %d at %v from %v", i, actor.Pos, p)
			}
		}
	}
}

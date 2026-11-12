package engine

import (
	"math"
	"testing"
)

func TestHandling035ServiceBrakeStopsAllCars(t *testing.T) {
	for class := 0; class < 4; class++ {
		for _, speed := range []float64{60 / 3.6, 120 / 3.6, 200 / 3.6, 80} {
			w := flat22(class, true)
			w.Car.Speed = speed
			elapsed := 0.
			for math.Abs(w.Car.Speed) > .25 && elapsed < 10 {
				w.Step(Input{Brake: 1}, 1./120)
				elapsed += 1. / 120
			}
			t.Logf("car=%d %.0fkm/h stop=%.2fs", class, speed*3.6, elapsed)
			if elapsed > 9 || w.Car.Speed < -.1 {
				t.Fatalf("weak/unstable service brakes: car%d %.0fm/s %.2fs", class, speed, elapsed)
			}
		}
	}
}

func TestHandling035BothDirectionsFormCompactCircles(t *testing.T) {
	for class := 0; class < 4; class++ {
		for _, sign := range []float64{-1, 1} {
			w := flat22(class, true)
			w.Car.Speed = 9
			distance, turn := 0., 0.
			for tick := 0; tick < 1800; tick++ {
				before, heading := w.Car.Pos, w.Car.Heading
				w.Step(Input{Throttle: Clamp(.11+(9-w.Car.Speed)*.14, 0, .65), Steer: sign}, 1./120)
				if tick > 240 {
					distance += before.Dist(w.Car.Pos)
					turn += math.Abs(w.Car.Heading - heading)
				}
			}
			radius := distance / math.Max(.001, turn)
			t.Logf("car=%d turn=%g radius=%.2fm full rotations=%.2f", class, sign, radius, turn/Tau)
			if turn < Tau || radius > 12 || radius < 2 || w.Car.DriftActive {
				t.Fatal("ordinary steering cannot form a controlled circle")
			}
		}
	}
}

func TestHandling035HandbrakeBurnoutAndControlledDonuts(t *testing.T) {
	for class := 0; class < 4; class++ {
		for _, steer := range []float64{-1, 0, 1} {
			w := flat22(class, true)
			start := w.Car.Pos
			for tick := 0; tick < 600; tick++ {
				w.Step(Input{Throttle: 1, Handbrake: true, Steer: steer}, 1./120)
			}
			if !w.Car.Burnout || !w.Car.RearLocked || !w.Car.Body.Grounded {
				t.Fatalf("burnout state missing class%d steer%.0f", class, steer)
			}
			for _, wheel := range w.Car.Rig.Wheels[2:] {
				if wheel.AngularVelocity != 0 {
					t.Fatal("handbrake rear tires not locked")
				}
			}
			if steer == 0 && w.Car.Pos.Dist(start) > .1 {
				t.Fatal("straight burnout moved vehicle")
			}
			if steer != 0 && (w.Car.Heading*steer < 1.5 || w.Car.Pos.Dist(start) > 20 || math.Abs(w.Car.Speed) > 4) {
				t.Fatalf("uncontrolled/missing donut class%d steer%.0f heading%.2f distance%.2f speed%.2f", class, steer, w.Car.Heading, w.Car.Pos.Dist(start), w.Car.Speed)
			}
		}
	}
}

func TestHandling035TinyRoadSeamDoesNotLaunchCar(t *testing.T) {
	w := flat22(0, true)
	// A 4cm expansion joint in a road support deck must not fire the suspension.
	w.Remaster.Def.Surfaces = []SurfaceTriangle{{A: [3]float64{-300, .04, 0}, B: [3]float64{300, .04, 0}, C: [3]float64{300, .04, 300}, Road: true}, {A: [3]float64{-300, .04, 0}, B: [3]float64{300, .04, 300}, C: [3]float64{-300, .04, 300}, Road: true}}
	w.Remaster.buildSpatial(w)
	w.Car.Pos = Vec{0, -30}
	w.Car.Speed = 35
	w.Car.Body.Initialized = false
	peak := 0.
	for tick := 0; tick < 240; tick++ {
		w.Step(Input{Throttle: .3}, 1./120)
		peak = math.Max(peak, w.Car.Body.Height-w.GroundHeight(w.Car.Pos))
	}
	if peak > 1.1 || w.Car.Body.VerticalSpeed > 2 {
		t.Fatalf("road seam launched car: clearance%.2f vertical%.2f", peak, w.Car.Body.VerticalSpeed)
	}
}

func TestTraffic035PrunesDeadEndBranchesWithoutEditingRoads(t *testing.T) {
	w := remasteredFixture(0, 0)
	w.Remaster.Def.Beta020 = true
	w.Routes = []Route{newRoute("loop", []Vec{{0, 0}, {200, 0}, {200, 200}, {0, 200}, {0, 0}}), newRoute("spur", []Vec{{200, 0}, {350, 0}}), newRoute("branch", []Vec{{350, 0}, {500, 0}}), newRoute("through", []Vec{{0, 0}, {100, 100}, {200, 200}})}
	w.Remaster.Def.Roads = []RoadDefinition{{Width: 12}, {Width: 12}, {Width: 12}, {Width: 12}}
	w.Remaster.TrafficCore035 = nil
	core := w.BuildTrafficCore035()
	if !core[0] || core[1] || core[2] || !core[3] {
		t.Fatal("wrong civilian road network", core)
	}
	pool := w.beta020TrafficPool()
	if len(pool) != 1056 {
		t.Fatalf("through roads lost population: %d", len(pool))
	}
	for _, npc := range pool {
		if npc.Path == 1 || npc.Path == 2 {
			t.Fatal("NPC spawned on dead-end branch")
		}
	}
	if len(w.Routes[1].Points) != 2 || w.Routes[1].Len != 150 {
		t.Fatal("road structure was edited")
	}
	npc := Traffic{Path: 3, T: w.Routes[3].Len, Pos: Vec{200, 200}, Heading: math.Pi / 4, Direction: 1, Lane: 2.7, Speed: 8}
	if !w.connectAokawaTraffic(&npc, .1) || npc.Path != 0 {
		t.Fatal("NPC failed to choose continuous route", npc)
	}
}

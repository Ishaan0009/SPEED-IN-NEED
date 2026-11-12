package engine

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

func TestBeta021DeadEndDoesNotPermanentlyParkTraffic(t *testing.T) {
	w := remasteredFixture(0, 0)
	w.Remaster.Def.Beta020 = true
	w.Routes = []Route{newRoute("dead end", []Vec{{0, 0}, {0, 180}})}
	w.Remaster.Def.Roads = []RoadDefinition{{Kind: "town", Width: 12}}
	w.Remaster.Def.Signals = nil
	w.Car.Pos = Vec{80, 80}
	w.Traffic = []Traffic{{Path: 0, T: 179.9, Pos: Vec{-2.7, 179.9}, Speed: 4, CruiseSpeed: 10, Direction: 1, Lane: 2.7, Variant: 4}}
	for i := 0; i < 600; i++ {
		w.stepRemasteredTraffic(1. / 120)
	}
	if w.Traffic[0].CruiseSpeed <= 0 || w.Traffic[0].Direction != -1 || w.Traffic[0].T > 175 || w.Traffic[0].Speed < 1 {
		t.Fatalf("dead-end actor failed to rejoin traffic: %+v", w.Traffic[0])
	}
}

func TestBeta021CrossRoadPriorityDoesNotBrakeParallelTraffic(t *testing.T) {
	a := Traffic{Path: 2, Pos: Vec{0, 0}, Heading: 0, Speed: 8}
	b := Traffic{Path: 1, Pos: Vec{4, 12}, Heading: 0, Speed: 8}
	if trafficCrossingYield021(a, b) {
		t.Fatal("parallel street caused a crossing yield")
	}
	b.Pos = Vec{-8, 8}
	b.Heading = math.Pi / 2
	if !trafficCrossingYield021(a, b) {
		t.Fatal("real crossing conflict did not yield to road priority")
	}
	b.Speed = 0
	if trafficCrossingYield021(a, b) {
		t.Fatal("stationary cross traffic created circular yielding")
	}
}

func TestBeta021PebblesCauseBumpWithoutWallRecoil(t *testing.T) {
	w := remasteredFixture(0, 0)
	w.Car.Pos = Vec{}
	w.Car.Speed = 10
	w.Car.Heading = 0
	w.Car.Body.Height = .74
	w.Remaster.Def.Objects = []PlacedObject{{Asset: "pebble", Kind: "rock", Pos: [3]float64{0, 0, 0}, Solid: true}}
	w.Remaster.Def.Colliders = []ColliderShape{{Object: 0, Kind: "rock", Center: Vec{}, Half: Vec{.18, .18}, MinY: 0, MaxY: .25}}
	w.Remaster.staticGrid = map[gridKey][]int{gridOf(Vec{}): {0}}
	w.remasteredStaticContacts(&w.Car, 0)
	if w.Car.Speed < 9.8 || w.Car.Body.VerticalSpeed <= 0 || w.Car.Pos.Length() > .01 {
		t.Fatalf("small stone behaved like a wall: %+v", w.Car)
	}
}

// Measure the actual shipped network, including queues and signal phases.
func TestBeta021ShippedTrafficFlow(t *testing.T) {
	raw, err := os.ReadFile("../assets/remaster_world_free.json")
	if err != nil {
		t.Fatal(err)
	}
	d, err := DecodeWorldDefinition(raw)
	if err != nil {
		t.Fatal(err)
	}
	type sample struct {
		Version                 string
		Actors, Stopped         int
		MeanSpeed, MeanProgress float64
	}
	result := []sample{}
	for _, old := range []bool{true, false} {
		w := NewRemasteredWorld(d, 0, 0, 0, 1, 0)
		w.Car.Pos = Vec{1490, 1490}
		r := sample{Version: "0.2.1", Actors: len(w.Traffic)}
		if old {
			r.Version = "0.2 traffic logic"
		}
		for step := 0; step < 60*120; step++ {
			w.Elapsed += 1. / 120
			if old {
				w.stepRemasteredTraffic020(1. / 120)
			} else {
				w.stepRemasteredTraffic(1. / 120)
			}
			for _, car := range w.Traffic {
				r.MeanProgress += car.Speed / 120 / float64(len(w.Traffic))
			}
		}
		for _, car := range w.Traffic {
			r.MeanSpeed += car.Speed / float64(len(w.Traffic))
			if car.Speed < .3 {
				r.Stopped++
			}
			if !old && car.CruiseSpeed == 0 {
				t.Fatal("traffic was permanently disabled")
			}
		}
		result = append(result, r)
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	if err := os.WriteFile("../TRAFFIC_FLOW_BETA_021.json", data, 0644); err != nil {
		t.Fatal(err)
	}
	if result[1].MeanProgress < result[0].MeanProgress*.9 {
		t.Fatalf("fleet flow regressed: %+v", result)
	}
	t.Logf("flow: %+v", result)
}

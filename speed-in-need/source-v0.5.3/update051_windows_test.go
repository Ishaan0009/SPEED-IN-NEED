package main

import (
	"encoding/json"
	"math"
	"speedinneed/engine"
	"testing"
)

func Test051GarageModelsAndWorkshopData(t *testing.T) {
	cars := garageCars051()
	if len(cars) != 19 {
		t.Fatalf("garage catalogue: %d", len(cars))
	}
	for _, c := range cars[1:] {
		raw, e := assets.ReadFile("assets/pbr/" + c.Key + ".pbr")
		if e != nil || len(raw) < 8 {
			t.Fatalf("missing drivable model %s", c.Key)
		}
	}
	raw, e := assets.ReadFile("assets/b051_basement.json")
	if e != nil {
		t.Fatal(e)
	}
	d, e := engine.DecodeWorldDefinition(raw)
	if e != nil {
		t.Fatal(e)
	}
	if len(d.CollisionTriangles) < 100 || len(d.Surfaces) < 2 || d.Name != "BASEMENT / VEHICLE TEST" {
		t.Fatal("garage lacks supplied geometry and collision")
	}
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	p := engine.Walker051{}
	if !w.ToggleWalker051(&p) {
		t.Fatal("no clear exit beside workshop spawn")
	}
}
func Test051MeadowReliefPreservesRoadsAndIsIdempotent(t *testing.T) {
	raw, e := assets.ReadFile("assets/b038_world_free.json")
	if e != nil {
		t.Fatal(e)
	}
	d, e := engine.DecodeWorldDefinition(raw)
	if e != nil {
		t.Fatal(e)
	}
	d.ApplyMeadowRelief051()
	changed := 0
	for i, h := range d.Grid.Values {
		if h != d.ReliefOriginal051.Values[i] {
			changed++
		}
	}
	if changed < 100 {
		t.Fatal("meadow stayed flat")
	}
	for _, road := range d.Roads {
		for _, p := range road.Points {
			if delta := math.Abs(d.MeadowDelta051(p)); delta > .12 {
				t.Fatalf("road elevation changed %.3f", delta)
			}
		}
	}
	before, _ := json.Marshal(d.Grid.Values)
	d.ApplyMeadowRelief051()
	after, _ := json.Marshal(d.Grid.Values)
	if string(before) != string(after) {
		t.Fatal("relief accumulated on reload")
	}
}

func Test051SuppliedWalkerFitsRuntimeBudget(t *testing.T) {
	raw, e := assets.ReadFile("assets/characters/walker.chr")
	if e != nil {
		t.Fatal(e)
	}
	a, e := decodeCharacter032(raw)
	if e != nil {
		t.Fatal(e)
	}
	if a.Name != "walker" || len(a.LODs[0].Indices) < 10000 {
		t.Fatal("supplied walking skin missing")
	}
}

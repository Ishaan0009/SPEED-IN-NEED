package main

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"speedinneed/engine"
	"testing"
)

func TestRoads038PreserveNetworkAndWidenSelectedCorridor(t *testing.T) {
	load := func(name string) *engine.WorldDefinition {
		raw, e := assets.ReadFile("assets/" + name + ".json")
		if e != nil {
			t.Fatal(e)
		}
		d, e := engine.DecodeWorldDefinition(raw)
		if e != nil {
			t.Fatal(e)
		}
		return d
	}
	before, after := load("b037_world_free"), load("b038_world_free")
	var selection struct {
		IDs []int `json:"main_road_ids"`
	}
	raw, e := os.ReadFile("MAIN_ROADS_038.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(raw, &selection); e != nil {
		t.Fatal(e)
	}
	selected := map[int]bool{}
	for _, i := range selection.IDs {
		selected[i] = true
	}
	if len(after.Roads) < len(before.Roads) || len(selected) != 39 {
		t.Fatal("road network changed")
	}
	for i, r := range after.Roads[:len(before.Roads)] {
		expected := before.Roads[i].Width
		if selected[i] {
			expected += 3
		}
		if r.Width != expected || !reflect.DeepEqual(r.Points, before.Roads[i].Points) {
			t.Fatalf("road %d changed outside requested width", i)
		}
	}
	if !reflect.DeepEqual(before.Signals, after.Signals) || before.Spawn != after.Spawn {
		t.Fatal("navigation anchors changed")
	}
	for i, bore := range after.Tunnels {
		if bore.Width != before.Tunnels[i].Width+3 {
			t.Fatal("tunnel did not widen with road")
		}
		w := engine.NewRemasteredWorld(after, 0, 0, 0, 1, 0)
		f := engine.HeadingVec(bore.Heading)
		r := engine.Vec{X: f.Z, Z: -f.X}
		for _, sign := range []float64{-1, 1} {
			for _, side := range []float64{-6.8, 0, 6.8} {
				previous := math.NaN()
				for a := -3.; a <= 36; a += .25 {
					p := bore.Center.Add(f.Mul(sign * (bore.Length*.5 + a))).Add(r.Mul(side))
					h := w.GroundHeight(p)
					if math.IsNaN(h) || math.IsInf(h, 0) || !math.IsNaN(previous) && math.Abs(h-previous) > .22 {
						t.Fatalf("widened portal support step sign %g side %g along %g: %g", sign, side, a, h-previous)
					}
					previous = h
				}
			}
		}
	}
}

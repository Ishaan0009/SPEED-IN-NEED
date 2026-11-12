package main

import (
	"encoding/json"
	"math"
	"speedinneed/engine"
	"testing"
)

func Test050MapCannotPanIntoEmptySpace(t *testing.T) {
	old := nav033
	defer func() { nav033 = old }()
	nav033.Definition = &engine.WorldDefinition{Bounds: [2]engine.Vec{{X: -1500, Z: -1500}, {X: 1500, Z: 1500}}}
	for _, zoom := range []float64{1, 2, 14} {
		nav033.Zoom = zoom
		nav033.Center = engine.Vec{X: 1e8, Z: -1e8}
		clampMap050()
		if math.Abs(nav033.Center.X) > 1500 || math.Abs(nav033.Center.Z) > 1500 {
			t.Fatalf("map escaped at zoom %g: %+v", zoom, nav033.Center)
		}
		if zoom == 1 && nav033.Center != (engine.Vec{}) {
			t.Fatal("whole-island view was not centred")
		}
	}
}
func Test050RoadConnectionsHaveVisibleAndPhysicalSupport(t *testing.T) {
	b, e := assets.ReadFile("assets/b038_world_free.json")
	if e != nil {
		t.Fatal(e)
	}
	var d engine.WorldDefinition
	if e = json.Unmarshal(b, &d); e != nil {
		t.Fatal(e)
	}
	connections := 0
	for _, road := range d.Roads {
		if road.Name != "CONNECTED ROAD" {
			continue
		}
		connections++
		if len(road.Elevations) != len(road.Points) {
			t.Fatal("connector elevation missing")
		}
		for _, y := range road.Elevations {
			if math.IsNaN(y) || math.IsInf(y, 0) {
				t.Fatal("nonfinite connector")
			}
		}
	}
	visible := false
	for _, o := range d.Objects {
		if o.Asset == "b050_road_connections" {
			visible = true
		}
	}
	if connections == 0 || !visible {
		t.Fatal("connectors require both graph edges and asphalt mesh")
	}
}

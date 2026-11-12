package main

import (
	"math"
	"speedinneed/engine"
	"testing"
)

func TestBeta031ShippedSidewalkPopulation(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	t.Logf("pedestrians=%d paths=%d traffic=%d", len(w.Pedestrians), len(w.SidewalkPaths), len(w.Traffic))
	if len(w.Pedestrians) < 50 || len(w.Pedestrians) > 640 {
		t.Fatal("shipped city lacks a bounded sidewalk population")
	}
	for _, p := range w.Pedestrians {
		if math.Abs(p.Height-w.GroundHeight(p.Pos)) > .03 {
			t.Fatalf("walker floats above pavement at %v", p.Pos)
		}
	}
}

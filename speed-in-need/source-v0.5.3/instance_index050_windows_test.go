package main

import (
	"math"
	"sort"
	"testing"
)

func TestInstanceIndex050PreservesVisibleCellsAndDrawOrder(t *testing.T) {
	var groups []remasterInstance
	for _, key := range []string{"grass-a", "grass-b", "tree"} {
		for x := -40; x <= 40; x++ {
			kind := "grass"
			if key == "tree" {
				kind = "tree"
			}
			groups = append(groups, remasterInstance{Key: key, Kind: kind, X: float64(x) * 12, Radius: 10, Count: 1})
		}
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Key != groups[j].Key {
			return groups[i].Key < groups[j].Key
		}
		return groups[i].X < groups[j].X
	})
	index := indexInstances050(groups)
	for _, eye := range []float64{-999, -50, 0, 15, 499, 999} {
		for _, shadow := range []bool{false, true} {
			spans := nearbyInstanceSpans050(groups, index, eye, 55, shadow, nil)
			included := map[int]bool{}
			last := -1
			for _, span := range spans {
				for i := span.Start; i < span.End; i++ {
					if i <= last {
						t.Fatal("draw order changed")
					}
					last = i
					included[i] = true
				}
			}
			for i, g := range groups {
				needed := g.Kind != "grass" || !shadow && math.Abs(g.X-eye) <= 55+g.Radius
				if included[i] != needed {
					t.Fatalf("cell %d / camera %.1f / shadow %v dropped or inserted", i, eye, shadow)
				}
			}
		}
	}
}

package main

import (
	"math"
	"testing"
)

func TestBeta022WideRoadTileRemainsVisibleAtItsForwardCorner(t *testing.T) {
	// A diagonal corner remains in front of the camera even when the tile's
	// centre is well behind it. The old fixed 90 m sphere discarded that road.
	a := &pbrAsset{Bounds: [2][3]float64{{-250, 0, -250}, {0, 1, 0}}}
	tile := remasterTileFromAsset("road", a)
	eye, target := pos3{-12, 2, -12}, pos3{12, 2, 12}
	if !remasterVisible(tile.X, tile.Z, tile.Radius, 800, eye, target, false) {
		t.Fatal("visible road corner was culled")
	}
	for _, x := range []float64{-250, 0} {
		for _, z := range []float64{-250, 0} {
			if math.Hypot(x-tile.X, z-tile.Z) > tile.Radius {
				t.Fatal("road bounds exceed culling sphere")
			}
		}
	}
}

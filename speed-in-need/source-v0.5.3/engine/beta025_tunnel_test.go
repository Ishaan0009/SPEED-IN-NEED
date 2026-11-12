package engine

import (
	"math"
	"testing"
)

func TestBeta025TunnelApronsJoinTheBoreAndOuterRoad(t *testing.T) {
	tunnel := TunnelDefinition{Center: Vec{}, Heading: 0, Length: 100, Width: 13, StartY: 10, EndY: 8, PortalApron: 14, ApronStartY: 11, ApronEndY: 7.5}
	for _, sign := range []float64{-1, 1} {
		last := 0.0
		for j := 0; j <= 128; j++ {
			a := sign * (49 + float64(j)*15/129)
			h, ok := tunnel.Floor(Vec{0, a})
			if !ok || math.IsNaN(h) {
				t.Fatal("missing physical apron support")
			}
			if j > 0 && math.Abs(h-last) > .03 {
				t.Fatalf("step at mouth/apron: %.5f -> %.5f", last, h)
			}
			last = h
		}
		if _, ok := tunnel.Floor(Vec{7, sign * 55}); ok {
			t.Fatal("apron collider extends beyond visible road width")
		}
	}
	if _, ok := tunnel.Floor(Vec{0, 64.01}); ok {
		t.Fatal("apron support extends past its road join")
	}
}

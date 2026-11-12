package main

import (
	"math"
	"speedinneed/engine"
	"testing"
)

func TestTunnel037ApproachSupportIsContinuous(t *testing.T) {
	raw, err := assets.ReadFile("assets/b037_world_free.json")
	if err != nil {
		t.Fatal(err)
	}
	d, err := engine.DecodeWorldDefinition(raw)
	if err != nil {
		t.Fatal(err)
	}
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	bore := d.Tunnels[0]
	f := engine.HeadingVec(bore.Heading)
	r := engine.Vec{X: f.Z, Z: -f.X}
	for _, sign := range []float64{-1, 1} {
		for _, s := range []float64{-4.5, 0, 4.5} {
			previous := math.NaN()
			for a := -3.; a <= 36; a += .25 {
				p := bore.Center.Add(f.Mul(sign * (bore.Length*.5 + a))).Add(r.Mul(s))
				h := w.GroundHeight(p)
				if !math.IsNaN(previous) && math.Abs(h-previous) > .22 {
					t.Fatalf("portal support step at side %.1f / along %.2f: %.3f m", s, a, h-previous)
				}
				if math.IsNaN(h) || math.IsInf(h, 0) {
					t.Fatal("invalid approach support")
				}
				previous = h
			}
		}
	}
}

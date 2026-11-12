package main

import (
	"math"
	"speedinneed/engine"
	"testing"
)

func TestBeta021SkidsClearAsphaltWithoutFollowingTunnelRoof(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 0, 0)
	for _, route := range w.Routes {
		for distance := 0.; distance < route.Len; distance += 25 {
			p, _ := route.AtDistance(distance)
			if !w.InsideTunnel(p) && beta021SkidHeight(w, p) <= d.Grid.Height(p)+.045 {
				t.Fatal("skidmark is below the baked asphalt")
			}
		}
	}
	for _, tunnel := range d.Tunnels {
		p := tunnel.Center
		if y := beta021SkidHeight(w, p); math.Abs(y-w.GroundHeight(p)-.008) > 1e-6 || y > d.Grid.Height(p)-1 {
			t.Fatal("tunnel skidmark projected onto the mountain roof")
		}
	}
	for _, bridge := range d.BridgeSpans {
		p := bridge.Start.Add(bridge.End).Mul(.5)
		if y := beta021SkidHeight(w, p); y < bridge.Y || y > bridge.Y+.08 {
			t.Fatalf("bridge skidmark escaped the deck layer: %.4f, deck %.4f", y, bridge.Y)
		}
	}
}

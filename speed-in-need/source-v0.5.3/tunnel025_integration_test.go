package main

import (
	"math"
	"speedinneed/engine"
	"testing"
)

func TestBeta025ShippedHillCoversTunnelAndHasBroadShoulders(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	if len(d.Tunnels) == 0 {
		t.Fatal("no tunnel")
	}
	tunnel := d.Tunnels[0]
	if tunnel.PortalApron < 10 {
		t.Fatal("v0.2.6 tunnel authoring has not been applied")
	}
	f := engine.HeadingVec(tunnel.Heading)
	r := engine.Vec{X: f.Z, Z: -f.X}
	for _, along := range []float64{-tunnel.Length * .3, 0, tunnel.Length * .3} {
		p := tunnel.Center.Add(f.Mul(along))
		floor, _ := tunnel.Floor(p)
		if d.Grid.Height(p) < floor+8.5 {
			t.Fatal("bore roof is not inside the hill")
		}
		for _, side := range []float64{-35, 35} {
			q := p.Add(r.Mul(side))
			deck := tunnel.StartY + (tunnel.EndY-tunnel.StartY)*(along/tunnel.Length+.5)
			if d.Grid.Height(q) < deck+5 {
				t.Fatal("hill is still only a narrow strip beside the tunnel")
			}
		}
	}
}
func TestBeta025PortalRoadRemainsOpenAndContinuous(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	tunnel := d.Tunnels[0]
	f := engine.HeadingVec(tunnel.Heading)
	for _, sign := range []float64{-1, 1} {
		previous := 0.0
		for j := 0; j <= 240; j++ {
			a := sign * (tunnel.Length*.5 - 8 + float64(j)*.15)
			p := tunnel.Center.Add(f.Mul(a))
			h := w.GroundHeight(p)
			if j > 0 && math.Abs(h-previous) > .15*.35 {
				t.Fatalf("physical step near portal at %.3f: %.4f -> %.4f", a, previous, h)
			}
			previous = h
		}
	}
}

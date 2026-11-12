package main

import (
	"math"
	"testing"
)

func TestBeta012BridgeRailsAreOpenAndPurposeBuilt(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	count := 0
	for _, shape := range d.Colliders {
		if shape.Kind != "bridge_rail" {
			continue
		}
		count++
		if shape.Object < 0 || shape.Object >= len(d.Objects) {
			t.Fatal("bridge rail has no visible structure")
		}
		o := d.Objects[shape.Object]
		if o.Asset == "rm_barrier" || !o.Solid {
			t.Fatal("bridge reused a stretched roadside prop")
		}
		if shape.Half.X > .2 || shape.Half.X <= 0 || shape.MaxY-shape.MinY > 1.4 {
			t.Fatal("bridge rail blocks the lane")
		}
	}
	if count != len(d.BridgeSpans)*2 {
		t.Fatal("bridge edge protection incomplete")
	}
	for _, b := range d.BridgeSpans {
		dx, dz := b.End.X-b.Start.X, b.End.Z-b.Start.Z
		length := math.Hypot(dx, dz)
		fx, fz := dx/length, dz/length
		covered := [2]bool{}
		for _, c := range d.Colliders {
			if c.Kind != "bridge_rail" {
				continue
			}
			x, z := c.Center.X-(b.Start.X+b.End.X)*.5, c.Center.Z-(b.Start.Z+b.End.Z)*.5
			along, side := x*fx+z*fz, x*fz-z*fx
			if math.Abs(along) > .2 || math.Abs(math.Abs(side)-(b.Width*.5+2.7)) > .2 {
				continue
			}
			if c.Half.Z < length*.5-.1 || math.Abs(math.Sin(c.Heading)*fx+math.Cos(c.Heading)*fz) < .99 {
				t.Fatal("bridge barrier leaves an uncovered span")
			}
			if side < 0 {
				covered[0] = true
			} else {
				covered[1] = true
			}
		}
		if !covered[0] || !covered[1] {
			t.Fatal("bridge needs continuous protection on both footpath edges")
		}
	}
}

func TestOneIslandBuildingsUseVisibleFootprintShapes(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	count := 0
	for _, s := range d.Colliders {
		if s.Kind != "building" {
			continue
		}
		count++
		if s.Half.X > 15 || s.Half.Z < .15 || s.MaxY-s.MinY < 5 {
			t.Fatal("lot-sized or degenerate building collision")
		}
	}
	if count < 900 {
		t.Fatal("source wall cross sections missing")
	}
}

func TestBeta012ShoreAndHillSlopesAreFinite(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	for _, v := range d.Grid.Values {
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			t.Fatal("invalid terrain")
		}
	}
	if d.Grid.H != d.Grid.W {
		t.Fatal("map expanded")
	}
}

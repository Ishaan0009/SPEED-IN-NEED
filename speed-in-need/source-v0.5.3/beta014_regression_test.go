package main

import (
	"speedinneed/engine"
	"testing"
)

func TestBeta014EveryVisibleCircuitBarrierHasACollisionHull(t *testing.T) {
	d := shippedRemasterDefinition(t, 1)
	barriers := 0
	for i, o := range d.Objects {
		if o.Kind != "barrier" {
			continue
		}
		barriers++
		if o.Collider < 0 || o.Collider >= len(d.Colliders) {
			t.Fatalf("visible barrier %d has no hull", i)
		}
		s := d.Colliders[o.Collider]
		if s.Object != i || s.Kind != "race_barrier" || s.Half.X <= 0 || s.Half.Z <= 0 || s.MaxY <= s.MinY {
			t.Fatalf("bad visible barrier hull %d: %+v", i, s)
		}
	}
	if barriers == 0 {
		t.Fatal("circuit fixtures were lost")
	}
}

func TestBeta014MenuDraggingCannotMoveThePreview(t *testing.T) {
	oldRemaster, oldBeta := remaster, beta
	defer func() { remaster, beta = oldRemaster, oldBeta }()
	beta.screen = screenTitle
	remaster.Orbit = .42
	remaster.Pitch = .14
	beta014MenuPointer(0, WM_LBUTTONDOWN, 200, 200)
	beta014MenuPointer(0, WM_MOUSEMOVE, 1200, 700)
	beta014MenuPointer(0, 0x202, 1200, 700)
	if remaster.Orbit != .42 || remaster.Pitch != .14 || remaster.Drag {
		t.Fatal("menu drag altered preview or kept input captured")
	}
}

func TestBeta014BridgeFootpathsAreSupportedAboveTheRiver(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	for _, b := range d.BridgeSpans {
		f := b.End.Sub(b.Start).Unit()
		r := engine.Vec{X: f.Z, Z: -f.X}
		for _, side := range []float64{-1, 1} {
			p := b.Start.Add(b.End.Sub(b.Start).Mul(.5)).Add(r.Mul(side * (b.Width*.5 + 1.35)))
			y := w.GroundHeight(p)
			if y < b.Y+.16 || y > b.Y+.20 {
				t.Fatalf("bridge sidewalk has no matching support: y=%g deck=%g", y, b.Y)
			}
		}
	}
}

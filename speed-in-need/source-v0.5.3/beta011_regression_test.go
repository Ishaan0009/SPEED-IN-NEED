package main

import (
	"math"
	"speedinneed/engine"
	"testing"
)

func TestBeta011BeachShelfSupportsTheWholeCar(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	p := d.BeachSpawn
	if p.X > -950 {
		t.Fatal("west beach spawn missing")
	}
	lo, hi := 1e9, -1e9
	for _, dx := range []float64{-2.5, 0, 2.5} {
		for _, dz := range []float64{-2.5, 0, 2.5} {
			h := w.GroundHeight(p.Add(engine.Vec{X: dx, Z: dz}))
			lo = math.Min(lo, h)
			hi = math.Max(hi, h)
		}
	}
	if lo < 15.5 || hi-lo > .5 {
		t.Fatalf("beach is an abrupt bank instead of a driving shelf: %.2f..%.2f", lo, hi)
	}
}

func TestBeta012AssignedMapFootprintAndNoExpansion(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	if d.Grid.H != 513 || d.Grid.W != 513 || d.Grid.MinZ != -1500 || d.Atlas.Size != [2]int{1251, 1252} || d.Atlas.OriginPixel != [2]float64{625.5, 626} {
		t.Fatal("map does not match the assigned square image")
	}
	if d.Bounds != [2]engine.Vec{{X: -1500, Z: -1500}, {X: 1500, Z: 1500}} {
		t.Fatal("map extent changed")
	}
	for _, o := range d.Objects {
		if o.Asset == "b011_rocky_pass" || o.Asset == "b011_pass_causeway" {
			t.Fatal("unrequested northern expansion still active")
		}
	}
	for _, b := range d.BridgeSpans {
		if b.Width < 12 {
			t.Fatal("bridge narrower than city streets")
		}
	}
}

func TestBeta011BuildingsHaveFullHeightCollision(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	count := 0
	for _, c := range d.Colliders {
		if c.Kind != "building" {
			continue
		}
		count++
		if c.MaxY-c.MinY < 5 {
			t.Fatalf("pavement-only building collider: object=%d height=%.3f", c.Object, c.MaxY-c.MinY)
		}
	}
	if count < 900 {
		t.Fatal("dense city buildings missing")
	}
}
func TestBeta011BridgeDecksSupportTheRoadWidth(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	if len(d.BridgeSpans) < 3 {
		t.Fatal("bridge spans missing")
	}
	for _, b := range d.BridgeSpans {
		delta := b.End.Sub(b.Start)
		length := delta.Length()
		f := delta.Mul(1 / length)
		r := engine.Vec{X: f.Z, Z: -f.X}
		for distance := 6.; distance < length-6; distance += 3 {
			for _, side := range []float64{-2, 0, 2} {
				p := b.Start.Add(f.Mul(distance)).Add(r.Mul(side))
				if h := w.GroundHeight(p); math.Abs(h-b.Y) > .15 {
					t.Fatalf("bridge has unsupported road: route %d at %v h=%.2f", b.Route, p, h)
				}
			}
		}
	}
}
func TestBeta011TrafficDensityAndCollisionClasses(t *testing.T) {
	w := engine.NewRemasteredWorld(shippedRemasterDefinition(t, 0), 0, 0, 0, 1, 0)
	for level, count := range []int{0, 256, 1032, 1056} {
		w.SetTrafficDensity(level)
		if len(w.Traffic) != count {
			t.Fatalf("density %d count %d", level, len(w.Traffic))
		}
	}
	for variant := 4; variant < 14; variant++ {
		f := engine.TrafficBodyFootprint(variant)
		if f.Z > 2.3 || f.X > 1.1 || engine.TrafficBodyMass(variant) > 1800 {
			t.Fatal("new passenger car reused truck collision")
		}
	}
	if engine.TrafficBodyMass(3) < 6000 {
		t.Fatal("existing truck mass lost")
	}
}
func TestBeta011DeepWaterRecoversToNearbyRoad(t *testing.T) {
	w := engine.NewRemasteredWorld(shippedRemasterDefinition(t, 0), 0, 0, 0, 1, 0)
	w.Traffic = nil
	old := engine.Vec{X: -1490, Z: 650}
	w.Car.Pos = old
	w.Car.Body = engine.ChassisState{}
	w.InitializeChassis()
	w.Step(engine.Input{}, 1./120)
	if w.Car.Pos.Dist(old) > 650 || w.Car.Pos.Dist(old) < 20 {
		t.Fatalf("water recovery did not use a nearby road: %v", w.Car.Pos)
	}
	if w.GroundHeight(w.Car.Pos) < 15 {
		t.Fatal("recovery remained underwater")
	}
}

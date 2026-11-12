package main

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"speedinneed/engine"
	"testing"
)

func Test053PlayerCatalogAndIndependentFinishes(t *testing.T) {
	expected := []string{"b050_bmw", "b050_eb110", "b050_lp670", "b051_r34", "b051_jeep", "ferrari", "porsche", "rm_car2", "rm_car3"}
	if len(playerVariants052) != len(expected) || len(paintPalette) != 36 || len(liveryNames052) != 13 {
		t.Fatal("incomplete 0.5.2 customization catalogue")
	}
	for i, id := range playerVariants052 {
		if garageCars051()[id].Key != expected[i] {
			t.Fatal("non-supplied vehicle exposed in selector", id)
		}
	}
}
func Test052RoadSupportsDoNotWarpWithMeadowRelief(t *testing.T) {
	raw, e := assets.ReadFile("assets/b038_world_free.json")
	if e != nil {
		t.Fatal(e)
	}
	d, e := engine.DecodeWorldDefinition(raw)
	if e != nil {
		t.Fatal(e)
	}
	original := append([]engine.SurfaceTriangle(nil), d.Surfaces...)
	d.ApplyMeadowRelief051()
	roads := 0
	for i, s := range d.Surfaces {
		if !s.Road {
			continue
		}
		roads++
		if s.A != original[i].A || s.B != original[i].B || s.C != original[i].C {
			t.Fatal("meadow relief moved an authored road or bridge", i)
		}
	}
	if roads < 100 {
		t.Fatal("road support data missing")
	}
	for i, c := range d.Colliders {
		if c.Object < 0 || c.Object >= len(d.Objects) {
			t.Fatal("stale object index after geometry repair", i, c.Object)
		}
	}
}
func Test052AllBridgeEdgesHaveContinuousTopRails(t *testing.T) {
	raw, _ := assets.ReadFile("assets/b038_world_free.json")
	d, e := engine.DecodeWorldDefinition(raw)
	if e != nil {
		t.Fatal(e)
	}
	raw, e = assets.ReadFile("assets/pbr/b052_bridge_rails.pbr")
	if e != nil {
		t.Fatal(e)
	}
	n := int(binary.LittleEndian.Uint32(raw[4:]))
	var m pbrAsset
	if json.Unmarshal(raw[8:8+n], &m) != nil {
		t.Fatal("bad railing metadata")
	}
	at := 8 + n
	var triangles [][3][3]float64
	for _, b := range m.Batches {
		v := make([][3]float64, b.Vertices)
		for i := range v {
			for j := range v[i] {
				v[i][j] = float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[at+i*48+j*4:])))
			}
		}
		at += b.Vertices * 48
		for i := 0; i < b.Indices; i += 3 {
			var tri [3][3]float64
			for j := 0; j < 3; j++ {
				id := int(binary.LittleEndian.Uint32(raw[at+(i+j)*4:]))
				tri[j] = v[id]
			}
			triangles = append(triangles, tri)
		}
		at += b.Indices * 4
	}
	for id, b := range d.BridgeSpans {
		f := b.End.Sub(b.Start).Unit()
		r := engine.Vec{X: f.Z, Z: -f.X}
		for _, sign := range []float64{-1, 1} {
			for _, u := range []float64{.02, .20, .50, .80, .98} {
				p := b.Start.Add(b.End.Sub(b.Start).Mul(u)).Add(r.Mul(sign * (b.Width*.5 + 2.15)))
				hit := false
				for _, tri := range triangles {
					a, b, c := tri[0], tri[1], tri[2]
					den := (b[2]-c[2])*(a[0]-c[0]) + (c[0]-b[0])*(a[2]-c[2])
					if math.Abs(den) < 1e-10 {
						continue
					}
					v := ((b[2]-c[2])*(p.X-c[0]) + (c[0]-b[0])*(p.Z-c[2])) / den
					w := ((c[2]-a[2])*(p.X-c[0]) + (a[0]-c[0])*(p.Z-c[2])) / den
					if v >= -.0001 && w >= -.0001 && v+w <= 1.0001 && math.Abs(a[1]-b[1]) < .01 && a[1] > d.BridgeSpans[id].Y+1.25 {
						hit = true
						break
					}
				}
				if !hit {
					t.Fatalf("bridge %d edge %.0f has a gap at %.2f", id, sign, u)
				}
			}
		}
	}
}
func Test052WalkingClipHasGroundedFiniteFrames(t *testing.T) {
	raw, e := assets.ReadFile("assets/characters/walker052.walk")
	if e != nil {
		t.Fatal(e)
	}
	if len(raw) < 8 || string(raw[:4]) != "WAL3" {
		t.Fatal("walking animation missing")
	}
	n := int(binary.LittleEndian.Uint32(raw[4:]))
	var m struct {
		Vertices, Indices, Frames int
		Stride, BindMatrixError   float64
	}
	if e = json.Unmarshal(raw[8:8+n], &m); e != nil {
		t.Fatal(e)
	}
	if m.Frames != 33 || m.Vertices < 1000 {
		t.Fatal("source clip not baked")
	}
	at := 8 + n + m.Vertices*8 + m.Indices*4
	for frame := 0; frame < m.Frames; frame++ {
		low, high := math.Inf(1), math.Inf(-1)
		for v := 0; v < m.Vertices; v++ {
			for k := 0; k < 6; k++ {
				value := float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[at:])))
				at += 4
				if math.IsNaN(value) || math.IsInf(value, 0) {
					t.Fatal("invalid animation vertex")
				}
				if k == 1 {
					low = math.Min(low, value)
					high = math.Max(high, value)
				}
			}
		}
		if math.Abs(low) > .0001 || high < 1.6 || high > 2.0 {
			t.Fatalf("frame %d has ungrounded or stretched body: %g..%g", frame, low, high)
		}
	}
	if at != len(raw) {
		t.Fatal("animation payload length mismatch")
	}
}
func Test053HybridGrassUsesMatchingEnvelopeAndStableRoots(t *testing.T) {
	for _, key := range []string{"b035_grass0", "b035_grass1"} {
		for _, suffix := range []string{"", "_lod", "_far_lod"} {
			raw, e := assets.ReadFile("assets/pbr/" + key + suffix + ".pbr")
			if e != nil {
				t.Fatal(e)
			}
			size := int(binary.LittleEndian.Uint32(raw[4:]))
			var m pbrAsset
			if json.Unmarshal(raw[8:8+size], &m) != nil {
				t.Fatal("bad grass")
			}
			if m.Bounds[0][1] < -.001 || m.Bounds[1][1] < .60 || m.Bounds[1][1] > .71 {
				t.Fatal("grass envelope drift", key, suffix, m.Bounds)
			}
			triangles := m.Batches[0].Indices / 3
			if suffix == "" && triangles < 60 {
				t.Fatal("near grass is still broad cards")
			}
			if suffix != "" && triangles > 12 {
				t.Fatal("distant silhouette exceeds budget")
			}
		}
	}
}

func Test052WalkInterpolationCannotExtrapolateOrReadIdleFrame(t *testing.T) {
	for _, frame := range []float64{-1.2, 0, .1, .5, 1.9, 15.7, 31.999, 32, 64.2} {
		a, b, w := walkFrameIndices052(frame, 32)
		if a < 1 || a > 32 || b < 1 || b > 32 || w < 0 || w >= 1 {
			t.Fatalf("invalid cycle interpolation %g: %d %d %g", frame, a, b, w)
		}
	}
}

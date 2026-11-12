package main

import (
	"math"
	"speedinneed/engine"
	"testing"
)

func TestBeta023EveryRoadHasContinuousPhysicalGrade(t *testing.T) {
	d := shippedRemasterDefinition(t, 0)
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	samples := 0
	for ri, road := range d.Roads {
		if road.Kind == "stunt" {
			continue
		}
		for i := 0; i+1 < len(road.Points); i++ {
			a, b := road.Points[i], road.Points[i+1]
			delta := b.Sub(a)
			length := delta.Length()
			if length < .05 {
				continue
			}
			n := int(math.Ceil(length / 2))
			previous := w.GroundHeight(a)
			for j := 1; j <= n; j++ {
				p := a.Add(delta.Mul(float64(j) / float64(n)))
				height := w.GroundHeight(p)
				grade := math.Abs(height-previous) / (length / float64(n))
				oldHeight := previous
				previous = height
				samples++
				if grade > .35 {
					t.Fatalf("road %d %s at (%.4f,%.4f) has a %.1f%% physical step: %.4f -> %.4f; grid %.4f; surfaces %d", ri, road.Name, p.X, p.Z, grade*100, oldHeight, height, d.Grid.Height(p), len(d.Surfaces))
				}
			}
		}
	}
	t.Logf("Audited %d physical road samples at <=2m spacing", samples)
}

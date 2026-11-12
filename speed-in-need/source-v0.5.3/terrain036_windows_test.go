package main

import (
	"math"
	"math/rand"
	"speedinneed/engine"
	"testing"
)

func TestTerrain036CoefficientsPreserveBothTriangles(t *testing.T) {
	rng := rand.New(rand.NewSource(360036))
	g := engine.HeightGrid{W: 17, H: 17, Step: 6, MinX: -48, MinZ: -48, Values: make([]float32, 289)}
	for i := range g.Values {
		g.Values[i] = float32(rng.Float64() * 220)
	}
	packed := heightCoefficients036(g)
	for i := 0; i < 50000; i++ {
		x, z := rng.Float64()*15.999, rng.Float64()*15.999
		ix, iz := int(x), int(z)
		u, v := x-float64(ix), z-float64(iz)
		at := (iz*g.W + ix) * 4
		c := packed[at : at+4]
		y := float64(c[0]) + float64(c[1])*u + float64(c[2])*v + float64(c[3])*math.Max(u+v-1, 0)
		want := g.Height(engine.Vec{X: g.MinX + x*g.Step, Z: g.MinZ + z*g.Step})
		if math.Abs(y-want) > .00007 {
			t.Fatalf("terrain roots moved by %.8f m", y-want)
		}
	}
}

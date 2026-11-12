package main

import "speedinneed/engine"

// One nearest-filtered RGBA sample reconstructs the exact two terrain triangles.
// a + (b-a)x + (c-a)y + (a+d-b-c)max(x+y-1,0)
// preserves the diagonal and rooted grass while replacing four vertex fetches.
func heightCoefficients036(g engine.HeightGrid) []float32 {
	out := make([]float32, g.W*g.H*4)
	for z := 0; z < g.H-1; z++ {
		for x := 0; x < g.W-1; x++ {
			i := z*g.W + x
			a, b, c, d := g.Values[i], g.Values[i+1], g.Values[i+g.W], g.Values[i+g.W+1]
			out[i*4], out[i*4+1], out[i*4+2], out[i*4+3] = a, b-a, c-a, a+d-b-c
		}
	}
	return out
}

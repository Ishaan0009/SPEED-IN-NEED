package main

import (
	"encoding/binary"
	"math"
)

// s3dWheelPivot measures the actual local wheel axis centre, rather than
// assuming imported GLB wheel vertices are already origin-centred.
// Using extrema rather than the vertex-weighted centroid avoids asymmetry
// from dense hubcap/spoke triangulation.
func s3dWheelPivot(data []byte) ([3]float64, bool) {
	var out [3]float64
	if len(data) < 8 || string(data[:4]) != "S3D1" {
		return out, false
	}
	n := int(binary.LittleEndian.Uint32(data[4:8]))
	if n < 3 || n > (len(data)-8)/16 {
		return out, false
	}
	min, max := [3]float64{math.Inf(1), math.Inf(1), math.Inf(1)}, [3]float64{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for i := 0; i < n; i++ {
		p := 8 + i*16
		for a := 0; a < 3; a++ {
			x := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[p+a*4:])))
			if math.IsNaN(x) || math.IsInf(x, 0) || math.Abs(x) > 1e5 {
				return out, false
			}
			min[a] = math.Min(min[a], x)
			max[a] = math.Max(max[a], x)
		}
	}
	for a := 0; a < 3; a++ {
		out[a] = (min[a] + max[a]) * .5
	}
	return out, true
}

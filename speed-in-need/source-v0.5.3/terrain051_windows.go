package main

import (
	"encoding/binary"
	"math"
	"speedinneed/engine"
)

var terrainKeys051 = map[string]bool{}

func prepareTerrain051(d *engine.WorldDefinition) {
	d.ApplyMeadowRelief051()
	decorateMeadow051(d)
	for _, k := range d.TerrainTiles {
		terrainKeys051[k] = true
		terrainKeys051[k+"_lod"] = true
	}
}

// Apply the same grid delta to the original detailed terrain at upload time.
// Existing microdetail is retained, with no extra vertex-shader work per frame.
func reliefVertices051(key string, raw []byte) {
	if !terrainKeys051[key] || app.world == nil || app.world.Remaster == nil {
		return
	}
	d := app.world.Remaster.Def
	get := func(p int) float64 { return float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[p:]))) }
	put := func(p int, v float64) { binary.LittleEndian.PutUint32(raw[p:], math.Float32bits(float32(v))) }
	for i := 0; i+48 <= len(raw); i += 48 {
		at := engine.Vec{X: get(i), Z: get(i + 8)}
		put(i+4, get(i+4)+d.MeadowDelta051(at))
		dx := (d.MeadowDelta051(at.Add(engine.Vec{X: .5})) - d.MeadowDelta051(at.Sub(engine.Vec{X: .5})))
		dz := (d.MeadowDelta051(at.Add(engine.Vec{Z: .5})) - d.MeadowDelta051(at.Sub(engine.Vec{Z: .5})))
		nx, ny, nz := get(i+12)-dx*get(i+16), get(i+16), get(i+20)-dz*get(i+16)
		length := math.Sqrt(nx*nx + ny*ny + nz*nz)
		if length > .001 {
			put(i+12, nx/length)
			put(i+16, ny/length)
			put(i+20, nz/length)
		}
	}
}

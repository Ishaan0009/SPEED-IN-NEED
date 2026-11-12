package main

import (
	"math"
	"speedinneed/engine"
)

func beta021SkidHeight(w *engine.World, p engine.Vec) float64 {
	y := w.GroundHeight(p) + .008
	// Free-drive asphalt meshes sit 45 mm above the terrain raster. Mesh
	// support can be at raster level; a decal must clear the visible asphalt.
	// The tunnel is a separate layer beneath the mountain raster.
	if w.Remaster != nil && w.Remaster.Mode == 0 && !w.InsideTunnel(p) {
		y = math.Max(y, w.Remaster.Def.Grid.Height(p)+.052)
	}
	return y
}

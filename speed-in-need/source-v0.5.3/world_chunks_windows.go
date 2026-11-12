//go:build windows

package main

import (
	"math"
	"speedinneed/engine"
)

// A graphics-independent 180 m world cell. Each display list can be cheaply
// rejected based on camera position/heading instead of submitting the entire
// world at once. Future VAO/VBO storage uses this same spatial organisation.
const renderCellSize = 180.0
const worldMin = -990.0

type renderChunk struct {
	terrain, vegetation uint32
	cx, cz              float64
}

var worldChunks []renderChunk

func releaseRenderChunks() {
	for _, c := range worldChunks {
		if c.terrain != 0 {
			gl("glDeleteLists", uintptr(c.terrain), 1)
		}
		if c.vegetation != 0 {
			gl("glDeleteLists", uintptr(c.vegetation), 1)
		}
	}
	worldChunks = nil
}

func renderGroundCell(x0, z0 float64) {
	en(GL_TEXTURE_2D)
	bind("grass.jpg")
	for z := z0; z < z0+renderCellSize; z += 18 {
		begin(GL_QUADS)
		for x := x0; x < x0+renderCellSize; x += 18 {
			sh := math.Sin(x*.03 + z*.07)
			v := .70 + .20*sh
			if app.mapID == 1 {
				col(.70*v, .83*v, .82*v, 1)
			} else if app.mapID == 2 {
				col(.88*v, .87*v, .75*v, 1)
			} else {
				col(.88*v, .96*v, .86*v, 1)
			}
			tq(pos3{x, app.world.TerrainHeight(engine.Vec{X: x, Z: z}) - .065, z}, pos3{x + 18, app.world.TerrainHeight(engine.Vec{X: x + 18, Z: z}) - .065, z}, pos3{x + 18, app.world.TerrainHeight(engine.Vec{X: x + 18, Z: z + 18}) - .065, z + 18}, pos3{x, app.world.TerrainHeight(engine.Vec{X: x, Z: z + 18}) - .065, z + 18}, x/28, z/28, (x+18)/28, (z+18)/28)
		}
		end()
	}
}

func renderGrassCell(x0, z0 float64) {
	dis(GL_TEXTURE_2D)
	count := []int{700, 1550, 2500}[engine.ClampInt(app.quality, 0, 2)]
	begin(GL_TRIANGLES)
	for i := 0; i < count; i++ {
		x := math.Mod(float64(i)*31.137+37, 870) - 435
		z := math.Mod(float64(i)*97.713+91, 840) - 420
		if x < x0 || x >= x0+renderCellSize || z < z0 || z >= z0+renderCellSize {
			continue
		}
		if engine.DistRoad(engine.Vec{X: x, Z: z}, app.world.Routes) < 12.5 || app.world.IsWater(engine.Vec{X: x, Z: z}, 2) || app.world.RampClearance(engine.Vec{X: x, Z: z}, 2) {
			continue
		}
		y := engine.Terrain(x, z) - .65
		height := .15 + math.Mod(float64(i)*.0713, .34)
		radius := .09 + math.Mod(float64(i)*.047, .11)
		shade := .55 + math.Mod(float64(i)*.013, .25)
		col(.18*shade, .48*shade, .16*shade, 1)
		vtx(x-radius, y, z)
		vtx(x+radius, y, z)
		vtx(x, y+height, z+radius*.2)
		vtx(x, y, z-radius)
		vtx(x, y, z+radius)
		vtx(x+radius*.2, y+height*.82, z)
	}
	end()
}

func buildWorldChunks() {
	releaseRenderChunks()
	for z0 := worldMin; z0 < 1000; z0 += renderCellSize {
		for x0 := worldMin; x0 < 1000; x0 += renderCellSize {
			tile := renderChunk{cx: x0 + renderCellSize*.5, cz: z0 + renderCellSize*.5}
			tile.terrain = uint32(gl("glGenLists", 1))
			gl("glNewList", uintptr(tile.terrain), GL_COMPILE)
			renderGroundCell(x0, z0)
			gl("glEndList")
			// Vegetation only where grass and trees are actually populated.
			populated := math.Abs(tile.cx) < 550 && math.Abs(tile.cz) < 560
			if populated {
				tile.vegetation = uint32(gl("glGenLists", 1))
				gl("glNewList", uintptr(tile.vegetation), GL_COMPILE)
				budget := []int{120, 230, 360}[engine.ClampInt(app.quality, 0, 2)]
				for i, t := range app.world.Trees {
					if i >= budget {
						break
					}
					if t.Pos.X >= x0 && t.Pos.X < x0+renderCellSize && t.Pos.Z >= z0 && t.Pos.Z < z0+renderCellSize {
						renderTree(t)
					}
				}
				renderGrassCell(x0, z0)
				gl("glEndList")
			}
			worldChunks = append(worldChunks, tile)
		}
	}
}

func renderVisibleWorldChunks(eye, target pos3) {
	fwd := engine.Vec{X: target.x - eye.x, Z: target.z - eye.z}.Unit()
	for _, c := range worldChunks {
		to := engine.Vec{X: c.cx - eye.x, Z: c.cz - eye.z}
		dist := to.Length()
		if dist > 805 {
			continue
		}
		// Preserve a generous margin behind the camera for view rotations.
		if dist > 235 && (to.X*fwd.X+to.Z*fwd.Z) < -130 {
			continue
		}
		if c.terrain != 0 {
			gl("glCallList", uintptr(c.terrain))
		}
		if c.vegetation != 0 && dist < 475 {
			gl("glCallList", uintptr(c.vegetation))
		}
	}
}

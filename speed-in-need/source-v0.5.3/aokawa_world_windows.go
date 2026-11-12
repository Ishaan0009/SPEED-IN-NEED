//go:build windows

package main

import (
	"math"
	"runtime"
	"speedinneed/engine"
	"unsafe"
)

func uploadAokawaHeightfield(d *engine.WorldDefinition) {
	for _, key := range d.TerrainTiles {
		loadPBRAsset(key + "_lod")
	}
	if remaster.HeightTexture != 0 {
		gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&remaster.HeightTexture)))
	}
	g := d.Grid
	glExt("glActiveTexture", 0x84C9)
	gl("glGenTextures", 1, uintptr(unsafe.Pointer(&remaster.HeightTexture)))
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(remaster.HeightTexture))
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, 0x2600)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, 0x2600)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, 0x812F)
	gl("glTexParameteri", GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, 0x812F)
	coefficients := heightCoefficients036(g)
	gl("glTexImage2D", GL_TEXTURE_2D, 0, 0x8814, uintptr(g.W), uintptr(g.H), 0, GL_RGBA, glFloat, uintptr(unsafe.Pointer(&coefficients[0])))
	runtime.KeepAlive(coefficients)
	glExt("glActiveTexture", 0x84C0)
	useLighting(true)
	sceneInt("uBayRoads", boolInt(app.mapID == 0))
	sceneInt("uCoastWorld", boolInt(app.mapID == 0))
	sceneVector("uCoastDomain", pos3{g.MinX, g.MinZ, g.Step * float64(g.W-1)})
	sceneFloat("uCoastDepth", g.Step*float64(g.H-1))
	glExt("glActiveTexture", 0x84CC)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(app.tex["b011_landcover.png"]))
	sceneInt("uLandcover", 12)
	glExt("glActiveTexture", 0x84CA)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(app.tex["b041_road_clearance.png"]))
	sceneInt("uRoadMask", 10)
	glExt("glActiveTexture", 0x84CB)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(pbrTexture("road_surface_color.jpg")))
	sceneInt("uRoadAlbedo", 11)
	glExt("glActiveTexture", 0x84C0)
	sceneInt("uHeightField", 9)
	for _, entry := range []struct {
		name, file string
		unit       int
	}{
		{"uJunctionMask", "b038_junction_mask.png", 13},
		{"uTerrainSand", "b035_surface_atlas.jpg", 14},
		{"uWaterData", "b031_water_data.png", 15},
	} {
		glExt("glActiveTexture", uintptr(0x84C0+entry.unit))
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(pbrTexture(entry.file)))
		sceneInt(entry.name, entry.unit)
	}
	sceneFloat("uWaterHeight", d.WaterHeight)
	glExt("glActiveTexture", 0x84C0)
	sceneVector("uHeightGrid", pos3{g.MinX, g.MinZ, g.Step})
	sceneVector("uHeightSize", pos3{float64(g.W), float64(g.H), 0})
	useLighting(false)
	runtime.KeepAlive(g.Values)
}

func aokawaRoadHidden(p engine.Vec) bool {
	for _, t := range app.world.Remaster.Def.Tunnels {
		if _, ok := t.Floor(p); ok {
			return true
		}
	}
	return false
}

func aokawaRoadSurface(p engine.Vec) float64 { return app.world.GroundHeight(p) }

func aokawaLampLocal(id int, front bool, side float64) pos3 {
	// Per-model anchors in normalized mesh coordinates.
	if front {
		return []pos3{{side * .79, .55, 2.04}, {side * .69, .64, 2.02}, {side * .76, .56, 2.04}, {side * .54, .66, 2.10}}[engine.ClampInt(id, 0, 3)]
	}
	return []pos3{{side * .81, .73, -2.15}, {side * .60, .74, -2.06}, {side * .79, .59, -2.21}, {side * .66, .69, -1.88}}[engine.ClampInt(id, 0, 3)]
}
func aokawaToWorld(c engine.Car, p pos3) pos3 {
	// Match the chassis yaw -> pitch -> roll rendering hierarchy.
	cr, sr := math.Cos(c.Rig.Roll), math.Sin(c.Rig.Roll)
	p = pos3{cr*p.x - sr*p.y, sr*p.x + cr*p.y, p.z}
	cp, sp := math.Cos(c.Rig.Pitch), math.Sin(c.Rig.Pitch)
	p = pos3{p.x, cp*p.y - sp*p.z, sp*p.y + cp*p.z}
	ch, sh := math.Cos(c.Heading), math.Sin(c.Heading)
	return pos3{c.Pos.X + ch*p.x + sh*p.z, c.Body.Height - .74 + p.y, c.Pos.Z - sh*p.x + ch*p.z}
}
func aokawaExhausts(id int) []pos3 {
	switch id {
	case 1:
		return []pos3{{-.56, .27, -2.12}, {.56, .27, -2.12}}
	case 2:
		return []pos3{{-.085, .50, -2.285}, {.085, .50, -2.285}}
	case 3:
		return []pos3{{-.63, .22, -1.99}}
	default:
		return []pos3{{-.29, .64, -2.20}, {.29, .64, -2.20}}
	}
}
func drawAokawaNitro(c engine.Car) {
	if !c.Boosting {
		return
	}
	for _, local := range aokawaExhausts(app.world.CarClass) {
		push()
		move(c.Pos.X, c.Body.Height-.74, c.Pos.Z)
		rotate(c.Heading*180/math.Pi, 0, 1, 0)
		rotate(c.Rig.Pitch*180/math.Pi, 1, 0, 0)
		rotate(c.Rig.Roll*180/math.Pi, 0, 0, 1)
		move(local.x, local.y, local.z)
		// The supplied flame is X-long. Rotate its long axis out of the tailpipe.
		rotate(90, 0, 1, 0)
		move(.25, 0, 0)
		resize(1.3+math.Sin(app.t*52)*.16, 1.9, 1.9)
		sceneVector("uEmission", pos3{.15, 1.3, 5.5})
		drawPBRAsset("rm_flame", 0)
		sceneVector("uEmission", pos3{})
		pop()
	}
}

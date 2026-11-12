//go:build windows

package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"sort"
	"speedinneed/engine"
	"strings"
	"time"
	"unsafe"
)

func nrm(x, y, z float64) {
	if capture != nil {
		capture.normal = pos3{x, y, z}
		return
	}
	gl("glNormal3f", b32(x), b32(y), b32(z))
}
func resize(x, y, z float64) {
	if capture != nil {
		capture.multiply([16]float32{float32(x), 0, 0, 0, 0, float32(y), 0, 0, 0, 0, float32(z), 0, 0, 0, 0, 1})
		return
	}
	gl("glScalef", b32(x), b32(y), b32(z))
}

func releaseRemasterBuffers() {
	grassEpoch053++
	releaseRemasterSky()
	releaseBuildingCache026()
	if remasterMapVBO != 0 {
		glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&remasterMapVBO)))
		remasterMapVBO = 0
	}
	for _, g := range remaster.Instances {
		b := g.VBO
		glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&b)))
	}
	for _, b := range remaster.RoadMeshes {
		glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&b.vbo)))
		glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&b.ebo)))
		if b.vao != 0 {
			glExt("glDeleteVertexArrays", 1, uintptr(unsafe.Pointer(&b.vao)))
		}
	}
	remaster.Instances = nil
	instanceIndex050 = nil
	instanceSpansScratch050 = nil
	remaster.Tiles = nil
	remaster.RoadTiles = nil
	remaster.Standalone = nil
	remaster.Lamps = nil
	remaster.RoadMeshes = nil
}
func buildRemasterWorldBuffers(d *engine.WorldDefinition) {
	releaseRemasterBuffers()
	type groupKey struct {
		Key, Kind string
		X, Z      int
	}
	groups := map[groupKey][][8]float32{}
	groupIDs := map[groupKey][]int{}
	barrier020LastCount = -1
	checkedLOD052 := map[string]bool{}
	for i, o := range d.Objects {
		if (o.Kind == "tree" || o.Kind == "road-tree" || o.Kind == "bush" || o.Kind == "lamp") && !checkedLOD052[o.Asset] {
			loadPBRAsset(o.Asset + "_lod")
			checkedLOD052[o.Asset] = true
		}

		if d.Beta026 && (o.Asset == "b022r_lane_paint" || o.Asset == "b021_crosswalks" || o.Asset == "b011_crosswalks") {
			continue
		}
		if o.Kind == "road" {
			if a := pbrAssets[o.Asset]; a != nil {
				remaster.RoadTiles = append(remaster.RoadTiles, remasterTileFromAsset(o.Asset, a))
			}
			continue
		}
		if o.Kind == "lamp" {
			remaster.Lamps = append(remaster.Lamps, i)
		}
		if o.Kind == "signal" || o.Kind == "road" || o.Kind == "road-marking" || o.Kind == "pickup" || o.Kind == "tunnel" || o.Kind == "bridge" || o.Kind == "bridge_rail" || o.Kind == "bridge-support" || o.Kind == "landmark" || o.Kind == "backdrop" || o.Kind == "circuit" || o.Kind == "forest" || o.Kind == "ground-rock" || o.Kind == "water" || o.Radius > 80 {
			remaster.Standalone = append(remaster.Standalone, i)
			continue
		}
		k := groupKey{o.Asset, o.Kind, int(math.Floor(o.Pos[0] / 100)), int(math.Floor(o.Pos[2] / 100))}
		if o.Kind == "barrier" || o.Kind == "prop" || o.Kind == "lamp" || o.Kind == "signal" || o.Kind == "sign" || o.Kind == "road-tree" {
			groupIDs[k] = append(groupIDs[k], i)
		}
		groups[k] = append(groups[k], [8]float32{float32(o.Pos[0]), float32(o.Pos[1]), float32(o.Pos[2]), 1, float32(o.Heading), float32(o.Scale[0]), float32(o.Scale[1]), float32(o.Scale[2])})
	}
	for k, v := range groups {
		g := remasterInstance{Key: k.Key, Kind: k.Kind, X: float64(k.X*100 + 50), Z: float64(k.Z*100 + 50), Radius: 90, Count: len(v), Data: v, ObjectIDs: groupIDs[k]}
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&g.VBO)))
		glExt("glBindBuffer", glArrayBuffer, uintptr(g.VBO))
		glExt("glBufferData", glArrayBuffer, uintptr(len(v)*32), uintptr(unsafe.Pointer(&v[0])), glStaticDraw)
		remaster.Instances = append(remaster.Instances, g)
		runtime.KeepAlive(v)
	}
	loadBeta020Vegetation(d)
	for i := range remaster.Instances {
		g := &remaster.Instances[i]
		g.TopY034 = math.Inf(-1)
		if a := pbrAssets[g.Key]; a != nil {
			for _, p := range g.Data {
				g.TopY034 = math.Max(g.TopY034, float64(p[1])+a.Bounds[1][1]*float64(p[6])+3)
			}
		}
	}
	// Stable ordering avoids small floating point changes in transparent vegetation blending.
	sort.Slice(remaster.Instances, func(i, j int) bool {
		a, b := remaster.Instances[i], remaster.Instances[j]
		if a.Key != b.Key {
			return a.Key < b.Key
		}
		if a.X != b.X {
			return a.X < b.X
		}
		return a.Z < b.Z
	})
	glExt("glBindBuffer", glArrayBuffer, 0)
	for _, key := range d.TerrainTiles {
		a := pbrAssets[key]
		if a == nil {
			continue
		}
		tile := remasterTileFromAsset(key, a)
		for _, road := range d.Roads {
			for _, point := range road.Points {
				if math.Hypot(point.X-tile.X, point.Z-tile.Z) < tile.Radius+road.Width*.5 {
					tile.RoadCorridor = true
					break
				}
			}
			if tile.RoadCorridor {
				break
			}
		}
		remaster.Tiles = append(remaster.Tiles, tile)
	}
	if app.mapID == 0 {
		remaster.RoadMeshes = uploadWorld(captureWorld(drawRemasterRoadGeometry))
	} else {
		remaster.RoadMeshes = uploadWorld(captureWorld(drawRemasterCircuitMarkings))
	}
	gpuWorld.ready = true
	buildRemasterMinimap()
}
func drawRemasterRoadGeometry() {
	w := app.world
	d := w.Remaster.Def
	if d.Beta020 {
		return
	} // Continuous authored lane and crossing paint owns this map.
	for ri, road := range d.Roads {
		for i := 0; i < len(road.Points)-1; i++ {
			a, b := road.Points[i], road.Points[i+1]
			delta := b.Sub(a)
			length := delta.Length()
			if length < .001 {
				continue
			}
			f := delta.Mul(1 / length)
			r := engine.Vec{X: f.Z, Z: -f.X}
			alongSteps := int(math.Ceil(length / 1.5))
			acrossSteps := int(math.Ceil(road.Width / 1.5))
			bind("asphalt.jpg")
			col(.82, .84, .85, 1)
			for j := 0; j < alongSteps && app.mapID != 0; j++ {
				start, endT := float64(j)/float64(alongSteps), float64(j+1)/float64(alongSteps)
				mid := a.Add(delta.Mul((start + endT) * .5))
				if aokawaRoadHidden(mid) {
					continue
				}
				if false {
					continue
				}
				for k := 0; k < acrossSteps; k++ {
					left, right := (float64(k)/float64(acrossSteps)-.5)*road.Width, (float64(k+1)/float64(acrossSteps)-.5)*road.Width
					begin(GL_QUADS)
					nrm(0, 1, 0)
					for _, uv := range [][2]float64{{start, left}, {endT, left}, {endT, right}, {start, right}} {
						q := a.Add(delta.Mul(uv[0])).Add(r.Mul(uv[1]))
						tex(uv[1]/4, (float64(i)*length+uv[0]*length)/4)
						vtx(q.X, w.GroundHeight(q)+.025, q.Z)
					}
					end()
				}
			}
			mid := a.Add(b).Mul(.5)
			if aokawaRoadHidden(mid) {
				continue
			}
			dis(GL_TEXTURE_2D)
			for _, side := range []float64{-1, 1} {
				x := a.Add(r.Mul(side * (road.Width*.5 - .32)))
				y := b.Add(r.Mul(side * (road.Width*.5 - .32)))
				remasterRoadStripe(x, y, .10, [3]float64{.88, .87, .76}, ri)
			}
			if i%3 != 0 && ri != 6 {
				remasterRoadStripe(a, b, .10, [3]float64{.96, .77, .25}, ri)
			}
		}
	}
	// The remastered map already has authored crossing paint. Drawing another
	// crossing for every approach signal duplicated and cut the same zebra stripes.
	if d.Beta020 {
		return
	}
	for _, sig := range d.Signals {
		f := engine.HeadingVec(sig.Heading)
		r := engine.Vec{X: f.Z, Z: -f.X}
		_, width, _ := w.RemasteredRoadAt(sig.Pos)
		for x := -width*.5 + .7; x < width*.5-.7; x += 1.15 {
			p := sig.Pos.Add(r.Mul(x))
			a := p.Sub(f.Mul(2.1))
			b := p.Add(f.Mul(2.1))
			remasterRoadStripe(a, b, .72, [3]float64{.94, .94, .87})
		}
		a := sig.Pos.Sub(f.Mul(4.1)).Sub(r.Mul(width*.5 - .3))
		b := a.Add(r.Mul(width - .6))
		remasterRoadStripe(a, b, .34, [3]float64{.94, .94, .87})
	}
}

func remasterTileFromAsset(key string, a *pbrAsset) remasterTile {
	return remasterTile{Key: key, X: (a.Bounds[0][0] + a.Bounds[1][0]) * .5, Z: (a.Bounds[0][2] + a.Bounds[1][2]) * .5,
		Radius: math.Hypot(a.Bounds[1][0]-a.Bounds[0][0], a.Bounds[1][2]-a.Bounds[0][2])*.5 + .05}
}
func remasterRoadStripe(a, b engine.Vec, width float64, c [3]float64, route ...int) {
	if app.mapID == 0 {
		mid := a.Add(b).Mul(.5)
		for _, t := range app.world.Remaster.Def.Tunnels {
			side, along := t.Coordinates(mid)
			if math.Abs(side) < 25 && math.Abs(along) < t.Length/2+44 {
				return
			}
		}
	}
	d := b.Sub(a)
	l := d.Length()
	if l < .001 {
		return
	}
	r := engine.Vec{X: d.Z / l * width * .5, Z: -d.X / l * width * .5}
	dis(GL_TEXTURE_2D)
	col(c[0], c[1], c[2], 1)
	nrm(0, 1, 0)
	offset := .080
	if app.mapID == 1 {
		offset = .025
	}
	steps := int(math.Ceil(l / 1.2))
	for i := 0; i < steps; i++ {
		p := a.Add(d.Mul(float64(i) / float64(steps)))
		q := a.Add(d.Mul(float64(i+1) / float64(steps)))
		if app.mapID == 0 && len(route) > 0 && !app.world.RoadMarkingAllowed(p.Add(q).Mul(.5), d.Mul(1/l), route[0]) {
			continue
		}
		begin(GL_QUADS)
		for _, v := range []engine.Vec{p.Sub(r), q.Sub(r), q.Add(r), p.Add(r)} {
			vtx(v.X, app.world.GroundHeight(v)+offset, v.Z)
		}
		end()
	}
}
func drawRemasterCircuitMarkings() {
	r := app.world.Remaster.Def.Roads[0]
	for i := 0; i < len(r.Points)-1; i++ {
		a, b := r.Points[i], r.Points[i+1]
		d := b.Sub(a)
		length := d.Length()
		if length < .001 {
			continue
		}
		side := engine.Vec{X: d.Z / length, Z: -d.X / length}
		for _, sign := range []float64{-1, 1} {
			remasterRoadStripe(a.Add(side.Mul(sign*8.55)), b.Add(side.Mul(sign*8.55)), .12, [3]float64{.94, .94, .88})
		}
		if i%3 != 0 {
			remasterRoadStripe(a, b, .10, [3]float64{.97, .76, .20})
		}
	}
	p, f := app.world.Routes[0].AtDistance(0)
	side := engine.Vec{X: f.Z, Z: -f.X}
	for row := 0; row < 2; row++ {
		for col := 0; col < 18; col++ {
			c := [3]float64{.025, .03, .035}
			if (row+col)%2 == 0 {
				c = [3]float64{.96, .96, .94}
			}
			a := p.Add(side.Mul(float64(col) - 8.5)).Add(f.Mul(float64(row)))
			remasterRoadStripe(a, a.Add(f), .98, c)
		}
	}
}
func remasterVisible(x, z, radius, limit float64, eye, target pos3, shadow bool) bool {
	dx, dz := x-eye.x, z-eye.z
	d := math.Hypot(dx, dz)
	if shadow {
		return math.Hypot(x-app.renderCar.Pos.X, z-app.renderCar.Pos.Z) < limit+radius
	}
	if d > limit+radius {
		return false
	}
	if d < radius+40 {
		return true
	}
	f := norm(sub(target, eye))
	return (dx*f.x+dz*f.z) > -radius-30 && visibleCone034(x, z, radius, eye, target)
}

var remasterVisibleGroups []remasterInstance
var remasterVisibleLODGroups025 []remasterInstance

// Bind a material once across all visible cells of the same asset.
func drawRemasterInstanceGroups(groups []remasterInstance) {
	groups = beta012MergeBuildings(groups)
	if len(groups) == 0 {
		return
	}
	if groups[0].Kind == "grass" && !indirectGrassAvailable053() {
		// These blades are opaque/cutout. Draw nearby cells first so the depth
		// buffer rejects covered blades before their lighting shader runs.
		sort.SliceStable(groups, func(i, j int) bool {
			a, b := groups[i], groups[j]
			ax, az := a.X-app.cam.x, a.Z-app.cam.z
			bx, bz := b.X-app.cam.x, b.Z-app.cam.z
			return ax*ax+az*az < bx*bx+bz*bz
		})
	}
	if groups[0].Kind == "grass" {
		groups = batchGrass052(groups)
		if len(groups) == 0 {
			return
		}
	}
	a := pbrAssets[groups[0].Key]
	if groups[0].Kind == "grass" && app.quality == 3 {
		if groups[0].Key == "b035_grass0" {
			a = pbrAssets["b051_grass0"]
		}
		if groups[0].Key == "b035_grass1" {
			a = pbrAssets["b051_grass1"]
		}
	}
	if a == nil {
		return
	}
	if !gpuWorld.instancing || remaster.InstanceLoc < 0 || remaster.YawLoc < 0 {
		return
	}
	setInstancedWorld(true)
	sceneInt("uNonuniformInstances", 1)
	for _, p := range a.Batches {
		if p.Slot != 0 {
			continue
		}
		m := a.Materials[p.Material]
		if sunShadow.drawing && m.Glass {
			continue
		}
		bindPBRMaterial(m)
		packedSlope := groups[0].Kind == "grass" && (app.world.Remaster.Def.VegetationFile == "beta032_vegetation.bin" || app.world.Remaster.Def.VegetationFile == "b034_vegetation.bin" || app.world.Remaster.Def.VegetationFile == "b035_vegetation.bin")
		sceneInt("uGrassSlope032", boolInt(packedSlope))
		conform035 := 0
		if groups[0].Kind == "grass" && strings.HasPrefix(groups[0].Key, "b035_grass") {
			conform035 = 1
			if strings.HasPrefix(groups[0].Key, "b035_grass_import") {
				conform035 = 2
			}
		}
		sceneInt("uGrassConform035", conform035)
		sceneInt("uConformGround", boolInt(!packedSlope && groups[0].Kind == "grass" || groups[0].Kind == "beach" || groups[0].Kind == "bush"))
		sceneInt("uGrassWind", boolInt(groups[0].Kind == "grass"))
		sceneInt("uTreeWindRefine", boolInt(groups[0].Kind == "tree" || groups[0].Kind == "road-tree"))
		sceneFloat("uGrassFadeEnd", []float64{110, 240, 420, 650}[app.quality]*beta.cfg.Vegetation*foliageDistance026())
		grassPass053 := 0
		if groups[0].Kind == "grass" {
			grassPass053 = 1
			if strings.HasSuffix(groups[0].Key, "_lod") {
				grassPass053 = 2
			}
		}
		sceneInt("uGrassDetail053", grassPass053)
		sceneFloat("uGrassNear053", []float64{18, 26, 34, 42}[app.quality])
		pbrPointers(p)
		loc, yaw := uintptr(remaster.InstanceLoc), uintptr(remaster.YawLoc)
		scaleLoc := uintptr(beta011InstanceScaleLoc)
		glExt("glEnableVertexAttribArray", scaleLoc)
		glExt("glVertexAttribDivisor", scaleLoc, 1)
		glExt("glEnableVertexAttribArray", loc)
		glExt("glVertexAttribDivisor", loc, 1)
		glExt("glEnableVertexAttribArray", yaw)
		glExt("glVertexAttribDivisor", yaw, 1)
		if !drawGrassIndirect053(groups, p, loc, yaw, scaleLoc) {
			for _, g := range groups {
				setWorldFade(g.Fade)
				sceneFloat("uLODFade052", g.LODFade052)
				sceneInt("uLODSide052", g.LODSide052)
				glExt("glBindBuffer", glArrayBuffer, uintptr(g.VBO))
				glExt("glVertexAttribPointer", loc, 4, glFloat, 0, 32, 0)
				glExt("glVertexAttribPointer", yaw, 1, glFloat, 0, 32, 16)
				glExt("glVertexAttribPointer", scaleLoc, 3, glFloat, 0, 32, 20)
				glExt("glDrawElementsInstanced", GL_TRIANGLES, uintptr(p.Indices), 0x1405, 0, uintptr(beta22InstanceCount(g)))
				gpuWorld.draws++
			}
		}
		glExt("glVertexAttribDivisor", scaleLoc, 0)
		glExt("glDisableVertexAttribArray", scaleLoc)
		glExt("glVertexAttribDivisor", loc, 0)
		glExt("glDisableVertexAttribArray", loc)
		glExt("glVertexAttribDivisor", yaw, 0)
		glExt("glDisableVertexAttribArray", yaw)
	}

	unbindWorld()
	resetPBRMaterial()
	setInstancedWorld(false)
	sceneInt("uNonuniformInstances", 0)
	sceneInt("uConformGround", 0)
	sceneInt("uGrassWind", 0)
	sceneInt("uGrassSlope032", 0)
	sceneInt("uGrassConform035", 0)
	sceneInt("uLODSide052", 0)
	sceneInt("uGrassDetail053", 0)
}
func drawRemasterWorld(eye, target pos3, shadow bool, radius float64) {
	refreshGrass051()
	if app.world != nil {
		c := app.world.Car
		sceneVec4("uGrassVehicle051", c.Pos.X, c.Pos.Z, c.Heading, c.Speed)
		spec := engine.VehicleVariantSpec050(app.world.CarClass, c.Variant050)
		sceneVec4("uGrassFootprint052", spec.HalfWidth, spec.HalfLength, 0, 0)
	}
	buildingPass026 = 0
	if shadow {
		buildingPass026 = radius
	}
	syncBeta020Barriers()
	d := app.world.Remaster.Def
	terrainBound := false
	terrainDistance := 5400.0
	if shadow {
		terrainDistance = radius
	}
	for _, tile := range remaster.Tiles {
		if remasterVisible(tile.X, tile.Z, tile.Radius, terrainDistance, eye, target, shadow) {
			a := pbrAssets[tile.Key]
			lodDistance := 700.
			if tile.RoadCorridor {
				lodDistance = 1350 * viewDistance026()
			}
			if !shadow && math.Hypot(tile.X-eye.x, tile.Z-eye.z) > lodDistance {
				if lod := pbrAssets[tile.Key+"_lod"]; lod != nil {
					a = lod
				}
			}
			if a == nil || len(a.Materials) == 0 {
				continue
			}
			if !terrainBound {
				bindPBRMaterial(a.Materials[0])
				terrainBound = true
			}
			for _, p := range a.Batches {
				pbrPointers(p)
				gl("glDrawElements", GL_TRIANGLES, uintptr(p.Indices), 0x1405, 0)
				gpuWorld.draws++
			}
		}
	}
	if terrainBound {
		unbindWorld()
		resetPBRMaterial()
	}
	// Asphalt tiles share a material: bind it once, then submit only nearby tiles.
	if !shadow {
		en(0x8037)
		gl("glPolygonOffset", b32(-1), b32(-1))
		bound := false
		for _, tile := range remaster.RoadTiles {
			if !remasterVisible(tile.X, tile.Z, tile.Radius, 1250*viewDistance026(), eye, target, false) {
				continue
			}
			asset := pbrAssets[tile.Key]
			if asset == nil || len(asset.Materials) == 0 {
				continue
			}
			if !bound {
				bindPBRMaterial(asset.Materials[0])
				bound = true
			}
			for _, part := range asset.Batches {
				pbrPointers(part)
				gl("glDrawElements", GL_TRIANGLES, uintptr(part.Indices), 0x1405, 0)
				gpuWorld.draws++
			}
		}
		if bound {
			unbindWorld()
			resetPBRMaterial()
		}
		dis(0x8037)
	}

	hideAsset025 := os.Getenv("SPEED_DIAGNOSTIC_HIDE_ASSET")
	for _, i := range remaster.Standalone {
		o := d.Objects[i]
		if hideAsset025 != "" && o.Asset == hideAsset025 {
			continue
		}
		if (o.Kind == "water" || o.Kind == "road" || o.Kind == "road-marking") && shadow {
			continue
		}
		if o.Kind == "pickup" && app.world.Remaster.PickupCooldown[i] > 0 {
			continue
		}
		limit := 1000. * viewDistance026()
		switch o.Kind {
		case "road-marking":
			limit = 650 * viewDistance026()
		case "prop":
			limit = 130 * viewDistance026()
		case "sign":
			limit = 220 * viewDistance026()
		case "lamp", "signal":
			limit = 1800 * viewDistance026()
		case "road-tree":
			limit = 2200 * viewDistance026()
		}
		if o.Kind == "backdrop" {
			limit = 4200
		}
		if shadow {
			limit = radius
		}
		cx, cz := o.Pos[0], o.Pos[2]
		if a := pbrAssets[o.Asset]; a != nil {
			lx := (a.Bounds[0][0] + a.Bounds[1][0]) * .5 * o.Scale[0]
			lz := (a.Bounds[0][2] + a.Bounds[1][2]) * .5 * o.Scale[2]
			cx += math.Cos(o.Heading)*lx + math.Sin(o.Heading)*lz
			cz += -math.Sin(o.Heading)*lx + math.Cos(o.Heading)*lz
		}
		if !remasterVisible(cx, cz, o.Radius, limit, eye, target, shadow) {
			continue
		}
		sceneInt("uSignalState", signal023State(o))
		sceneInt("uFlowWater", boolInt(o.Kind == "water"))
		sceneFloat("uFlowTime", app.world.Elapsed)
		push()
		if p := app.world.Remaster.DynamicProps[i]; p != nil {
			move(p.Pos.X, p.Height, p.Pos.Z)
			rotate(p.Heading*180/math.Pi, 0, 1, 0)
			rotate(p.Roll*180/math.Pi, 0, 0, 1)
		} else {
			move(o.Pos[0], o.Pos[1], o.Pos[2])
		}
		if pole := app.world.Remaster.FallenPoles[i]; pole != nil {
			rotate(pole.Angle*180/math.Pi, pole.FallDirection.Z, 0, -pole.FallDirection.X)
		}
		if app.world.Remaster.DynamicProps[i] == nil {
			rotate(o.Heading*180/math.Pi, 0, 1, 0)
		}
		if o.Kind == "pickup" {
			move(0, .3+math.Sin(app.t*2)*.15, 0)
			rotate(app.t*50, 0, 1, 0)
		}
		resize(o.Scale[0], o.Scale[1], o.Scale[2])
		portalApron := !shadow && (o.Asset == "ab_cave_aprons" || o.Asset == "b031_tunnel_aprons" || o.Kind == "road-marking")
		if portalApron {
			en(0x8037)
			gl("glPolygonOffset", b32(-1.5), b32(-3))
		}
		drawPBRAsset(o.Asset, 0)
		if portalApron {
			dis(0x8037)
		}
		pop()
	}
	remasterVisibleGroups = remasterVisibleGroups[:0]
	remasterVisibleLODGroups025 = remasterVisibleLODGroups025[:0]
	sceneInt("uSignalState", 0)
	sceneInt("uFlowWater", 0)
	// Windows environment reads cross the OS boundary and allocate UTF-16 text.
	// Read this optional diagnostic switch once per pass, not once per cell.
	hideKind := os.Getenv("SPEED_DIAGNOSTIC_HIDE_KIND")
	for _, span050 := range visibleInstanceSpans050(eye, shadow) {
		for _, g := range remaster.Instances[span050.Start:span050.End] {
			if hideKind != "" && (g.Kind == hideKind || hideKind == "vegetation" && (g.Kind == "grass" || g.Kind == "bush" || g.Kind == "tree" || g.Kind == "road-tree")) {
				continue
			}
			limit := 650. * viewDistance026()
			switch g.Kind {
			case "grass":
				limit = []float64{110, 240, 420, 650}[app.quality] * beta.cfg.Vegetation * foliageDistance026()
			case "bush":
				limit = []float64{180, 320, 500, 650}[app.quality] * viewDistance026() * foliageDistance026()
			case "tree":
				limit = []float64{700, 1200, 1800, 2600}[app.quality] * viewDistance026()
			case "prop":
				limit = 130 * viewDistance026()
			case "lamp", "signal", "sign":
				limit = 1800 * viewDistance026()
			case "road-tree":
				limit = 2200 * viewDistance026()
			case "barrier":
				limit = 340 * viewDistance026()
			case "pickup":
				limit = 260 * viewDistance026()
			case "building":
				limit = 900 * viewDistance026()
			}
			if shadow {
				if g.Kind == "grass" || g.Kind == "bush" || g.Kind == "pickup" {
					continue
				}
				limit = math.Min(limit, radius)
			}
			if !remasterVisible(g.X, g.Z, g.Radius, limit, eye, target, shadow) {
				continue
			}
			if !shadow && g.Kind == "building" && terrainOccluded034(g, eye) {
				continue
			}
			fade := 1.
			if !shadow && (g.Kind == "grass" || g.Kind == "tree" || g.Kind == "road-tree" || g.Kind == "bush") {
				fade = engine.Clamp((limit+g.Radius-math.Hypot(g.X-eye.x, g.Z-eye.z))/40, 0, 1)
			}
			g.Fade = fade
			lodDistance := []float64{38, 75, 110, 150}[engine.ClampInt(app.quality, 0, 3)]
			if g.Kind == "lamp" {
				lodDistance = 110
			}
			if g.Kind == "bush" {
				lodDistance = 52
			}
			if (g.Kind == "tree" || g.Kind == "road-tree") && pbrAssets[g.Key+"_lod"] != nil {
				lodDistance = []float64{22, 30, 38, 54}[engine.ClampInt(app.quality, 0, 3)] + g.Radius
				if shadow {
					lodDistance = 0
				}
			}
			if g.Kind == "grass" && strings.HasPrefix(g.Key, "b035_") {
				lodDistance = []float64{6, 9, 12, 15}[engine.ClampInt(app.quality, 0, 3)]
			}
			lodRange := math.Hypot(g.X-eye.x, g.Z-eye.z)
			if g.Kind == "grass" {
				lodRange = math.Max(0, lodRange-g.Radius*.30)
			}
			if g.Kind == "grass" && (g.Key == "b035_grass0" || g.Key == "b035_grass1") {
				nearEnd := []float64{18, 26, 34, 42}[app.quality] + 10
				if lodRange < nearEnd+g.Radius {
					remasterVisibleGroups = append(remasterVisibleGroups, g)
				}
				if lodRange > nearEnd-20-g.Radius {
					far := g
					far.Key += "_lod"
					remasterVisibleLODGroups025 = append(remasterVisibleLODGroups025, far)
				}
				continue
			}
			stableGrass := g.Kind == "grass" && (g.Key == "b035_grass0" || g.Key == "b035_grass1")
			plant := g.Kind == "grass" || g.Kind == "bush" || g.Kind == "tree" || g.Kind == "road-tree" || g.Kind == "lamp"
			if !stableGrass && plant && pbrAssets[g.Key+"_lod"] != nil {
				if !shadow && lodRange > lodDistance-8 && lodRange < lodDistance+8 {
					t := engine.Clamp((lodRange-lodDistance+8)/16, 0, 1)
					near, far := g, g
					near.LODFade052 = t
					near.LODSide052 = 1
					far.LODFade052 = t
					far.LODSide052 = 2
					far.Key += "_lod"
					remasterVisibleGroups = append(remasterVisibleGroups, near)
					remasterVisibleLODGroups025 = append(remasterVisibleLODGroups025, far)
					continue
				}
				if shadow || lodRange >= lodDistance+8 {
					g.Key += "_lod"
					remasterVisibleLODGroups025 = append(remasterVisibleLODGroups025, g)
					continue
				}
			}
			remasterVisibleGroups = append(remasterVisibleGroups, g)
		}
	}
	// Each list inherits the source asset ordering. Keep far grass contiguous
	// instead of repeatedly rebinding near/far materials between adjacent cells.
	remasterVisibleGroups = append(remasterVisibleGroups, remasterVisibleLODGroups025...)
	sceneInt("uFlowWater", 0)
	for begin := 0; begin < len(remasterVisibleGroups); {
		end := begin + 1
		for end < len(remasterVisibleGroups) && remasterVisibleGroups[end].Key == remasterVisibleGroups[begin].Key {
			end++
		}
		drawRemasterInstanceGroups(remasterVisibleGroups[begin:end])
		begin = end
	}
	setWorldFade(1)
	drawBeta020MovedBarriers(eye, target, shadow, radius)
	if !shadow {
		en(0x8037)
		gl("glPolygonOffset", b32(-2), b32(-2))
		drawWorldBatches(remaster.RoadMeshes)
		dis(0x8037)
		resetPBRMaterial()
		unbindWorld()
		drawRemasterWater()
		drawRemasterSignals()
	}
}
func drawRemasterWater() {
	drawOcean023()
	if app.world.Remaster.Def.WaterGrid.W > 1 {
		return
	}
	w := app.world
	setPBRMaterial(false, .6, .17)
	dis(GL_TEXTURE_2D)
	for i, water := range w.Remaster.Def.Water {
		if len(water.Polygon) < 3 {
			continue
		}
		y := w.Remaster.Def.WaterLevels[i]
		col(.025, .16, .21, .94)
		nrm(0, 1, 0)
		begin(GL_TRIANGLES)
		for j := 1; j < len(water.Polygon)-1; j++ {
			for _, p := range []engine.Vec{water.Polygon[0], water.Polygon[j], water.Polygon[j+1]} {
				vtx(p.X, y, p.Z)
			}
		}
		end()
	}
	resetPBRMaterial()
}
func drawRemasterSignals() {
	// Signal heads are solid authored meshes. A separate emissive box previously
	// floated away from the fixture and looked like a hologram.
}

func drawRemasterCar(c engine.Car, id int, preview bool) {
	if preview {
		c.Variant050 = setup050().Model
	}
	oldMotionEntity038 := temporal038.Entity
	if !preview && id == app.world.CarClass && c.Pos == app.renderCar.Pos {
		temporal038.Entity = 1
	}
	defer func() { temporal038.Entity = oldMotionEntity038 }()
	oldLamps := remasterVehicleLamps
	defer func() { remasterVehicleLamps = oldLamps }()
	oldCockpit := remasterPlayerCockpit
	remasterPlayerCockpit = app.boolFPP && !sunShadow.drawing && beta.screen == screenDrive && id == app.world.CarClass && c.Pos == app.renderCar.Pos
	defer func() { remasterPlayerCockpit = oldCockpit }()
	remasterVehicleLamps.Active = true
	remasterVehicleLamps.Front = engine.AutomaticLights(sceneHour()) || app.world.InsideTunnel(c.Pos)
	remasterVehicleLamps.Tail = remasterVehicleLamps.Front
	remasterVehicleLamps.Brake = c.Braking || c.CoastingLights
	remasterVehicleLamps.Reverse = c.Gearbox.Gear < 0
	if preview || (id == app.world.CarClass && c.Pos == app.renderCar.Pos) {
		remasterVehicleLamps.Front = beta.cfg.Headlights
		remasterVehicleLamps.Tail = beta.cfg.Headlights || engine.AutomaticLights(sceneHour()) || app.world.InsideTunnel(c.Pos)
	}
	key := modelKey050(c, id, preview)
	if pbrAssets[key] == nil {
		loadPBRAsset(key)
	}
	custom050 := c.Variant050 > 0 || preview && setup050().Model > 0
	tune050 := setup050()
	if !preview && !(id == app.world.CarClass && c.Pos == app.renderCar.Pos) {
		tune050.Wheel = c.Rim050
	}
	livery050 := c.Livery050
	if preview || id == app.world.CarClass && c.Pos == app.renderCar.Pos {
		livery050 = tune050.Livery
	}
	sceneInt("uLivery050", livery050)
	defer sceneInt("uLivery050", 0)
	player := id == app.world.CarClass && c.Pos == app.renderCar.Pos
	if !preview && (sunShadow.drawing || (app.quality == 0 && !app.boolFPP) || (!player && c.Pos.Dist(app.renderCar.Pos) > []float64{30, 60, 100, 150}[engine.ClampInt(beta.cfg.MeshDetail, 0, 3)])) && pbrAssets[key+"_lod"] != nil {
		key += "_lod"
	}
	a := pbrAssets[key]
	if a == nil {
		return
	}
	push()
	move(c.Pos.X, c.Body.Height-.74, c.Pos.Z)
	rotate(c.Heading*180/math.Pi, 0, 1, 0)
	rotate(engine.Clamp(c.Rig.Pitch, -1.05, 1.05)*180/math.Pi, 1, 0, 0)
	rotate(engine.Clamp(c.Rig.Roll, -.75, .75)*180/math.Pi, 0, 0, 1)
	if preview || player {
		drawDriver032(c, engine.ClampInt(id, 0, 3), remasterPlayerCockpit)
	}
	damage041 := 0.
	if beta.cfg.CrashEffects041 && !preview {
		damage041 = engine.Clamp(c.Damage/65, 0, 1)
	}
	sceneFloat("uCrashDamage041", damage041)
	if !custom050 && (preview || player) && pbrAssets["b033_cabin_"+fmt.Sprint(id)] != nil {
		drawPBRAsset("b033_cabin_"+fmt.Sprint(id), 0)
	} else {
		drawPBRAsset(key, 0)
	}
	sceneFloat("uCrashDamage041", 0)
	if preview || player {
		drawSteering037(c, id)
	}
	for i := 0; i < 4; i++ {
		anchor, ok := a.Anchors[fmt.Sprint(i+1)]
		if !ok {
			continue
		}
		wheel := c.Rig.Wheels[i]
		hubY := anchor[1]
		if !preview {
			p := engine.VehicleVariantSpec050(id, c.Variant050).P
			radius := a.WheelRadii[fmt.Sprint(i+1)]
			if radius <= 0 {
				radius = p.WheelRadius
			}
			widths := [4][4]float64{{.144, .144, .157, .157}, {.143, .143, .143, .143}, {.175, .175, .178, .178}, {.129, .129, .129, .129}}
			hubY = app.world.VisualWheelHub026(c, anchor, radius, widths[engine.ClampInt(id, 0, 3)][i], wheel.SteerAngle, wheel.Touching)
		}
		push()
		move(anchor[0], hubY, anchor[2])
		rotate(wheel.SteerAngle*180/math.Pi, 0, 1, 0)
		drawPBRAsset(key, i+5)
		rotate(wheel.RollAngle*180/math.Pi, 1, 0, 0)
		if tune050.Wheel == 0 {
			drawPBRAsset(key, i+1)
		}
		drawRim050(a.WheelRadii[fmt.Sprint(i+1)], anchor[0], tune050.Wheel)
		pop()
	}
	pop()
}
func drawRemasterActors(shadow bool, radius float64) {
	drawWalker051()
	drawLANPlayers050(shadow, radius)
	defer drawPedestrians(shadow, radius)
	w := app.world
	c := app.renderCar
	if true {
		drawRemasterCar(c, w.CarClass, false)
		if !shadow {
			drawAokawaReverseLenses(c, w.CarClass)
		}
		if false && !shadow {
			spec := engine.RemasteredVehicleSpec(w.CarClass)
			drawRemasterLampLenses(c, spec.HalfLength, spec.HalfWidth, .62, beta.cfg.Headlights, c.Braking || c.CoastingLights, c.Gearbox.Gear < 0 || c.Speed < -.1)
		}
	}
	for racerIndex038, r := range w.Remaster.Racers {
		if r.Car.Pos.Dist(c.Pos) < radius {
			old := beta.paint
			beta.paint = (r.Class*3 + 2) % len(paintPalette)
			temporal038.Entity = 2 + racerIndex038
			drawRemasterCar(r.Car, r.Class, false)
			temporal038.Entity = 0
			if !shadow {
				spec := engine.RemasteredVehicleSpec(r.Class)
				drawRemasterLampLenses(r.Car, spec.HalfLength, spec.HalfWidth, .62, engine.AutomaticLights(sceneHour()) || w.InsideTunnel(r.Car.Pos), r.Car.Braking, r.Car.Gearbox.Gear < 0)
			}
			beta.paint = old
		}
	}
	beginTrafficVisual040(shadow)
	for _, index := range beta020TrafficSelection(shadow, radius) {
		t := w.Traffic[index]
		if index < len(w.Remaster.TrafficSnapshot) {
			t = engine.InterpolateTraffic026(w.Remaster.TrafficSnapshot[index], t, app.renderAlpha)
		}
		if !shadow {
			t = smoothTrafficVisual040(index, t, navFrameDT033)
		}
		if t.Pos.Dist(c.Pos) > radius {
			continue
		}
		push()
		move(t.Pos.X, w.GroundHeight(t.Pos)+.03, t.Pos.Z)
		rotate(t.Heading*180/math.Pi, 0, 1, 0)
		rotate(t.Pitch*180/math.Pi, 1, 0, 0)
		rotate(t.Roll*180/math.Pi, 0, 0, 1)
		temporal038.Entity = 1024 + index
		drawTrafficDriver040(t, t.Pos.Dist(app.renderCar.Pos), shadow)
		drawBeta026NPC(t)
		temporal038.Entity = 0
		pop()
		if !shadow {
			length, width, y := 2.12, .78, .65
			if t.Variant == 0 {
				length = 5.35
				width = 1.23
				y = 1.3
			}
			if t.Variant == 2 {
				length = 2.25
				width = 1.10
			}
			if t.Variant == 3 {
				length = 3.65
				width = 1.23
				y = 1.12
			}
			other := engine.Car{Pos: t.Pos, Heading: t.Heading}
			other.Body.Height = w.GroundHeight(t.Pos) + .77
			drawRemasterLampLenses(other, length, width, y, engine.AutomaticLights(sceneHour()) || w.InsideTunnel(t.Pos), t.Braking || t.Speed < .5, false)
		}
	}
}
func drawRemasterLampLenses(c engine.Car, length, width, y float64, front, brake, reverse bool) {
	push()
	move(c.Pos.X, c.Body.Height-.74, c.Pos.Z)
	rotate(c.Heading*180/math.Pi, 0, 1, 0)
	rotate(c.Rig.Pitch*180/math.Pi, 1, 0, 0)
	dis(GL_TEXTURE_2D)
	lens := func(x, z float64, color pos3) {
		col(color.x, color.y, color.z, 1)
		sceneVector("uEmission", scale(color, 2.4))
		nrm(0, 0, math.Copysign(1, z))
		begin(GL_QUADS)
		vtx(x-.07, y-.033, z)
		vtx(x+.07, y-.033, z)
		vtx(x+.07, y+.033, z)
		vtx(x-.07, y+.033, z)
		end()
	}
	for _, side := range []float64{-1, 1} {
		f := engine.HeadingVec(c.Heading)
		viewSide := (app.cam.x-c.Pos.X)*f.X + (app.cam.z-c.Pos.Z)*f.Z
		if front && viewSide > length {
			lens(side*width*.78, length*1.035, pos3{.89, .95, 1})
		}
		if viewSide > -length {
			continue
		}
		if reverse {
			lens(side*width*.55, -length*1.035, pos3{1, .98, .90})
		} else if brake || engine.AutomaticLights(sceneHour()) {
			brightness := .5
			if brake {
				brightness = 1
			}
			lens(side*width*.78, -length*1.035, pos3{brightness, .005, .002})
		}
	}
	sceneVector("uEmission", pos3{})
	pop()
}
func drawRemasterCockpit020(c engine.Car) {
	f := engine.HeadingVec(c.Heading)
	r := engine.Vec{X: f.Z, Z: -f.X}
	floor := c.Body.Height - .74
	// The eye and imported controls occupy the driver seat, below the roof line.
	seat := c.Pos.Add(r.Mul([]float64{.34, .34, .34, -.36}[beta.carSelect])).Add(f.Mul([]float64{.34, .34, -.14, -.22}[beta.carSelect]))
	eyeY := floor + []float64{.90, 1.06, .96, .98}[beta.carSelect]
	// A separate cockpit depth range keeps the supplied controls in front of
	// coarse imported dashboard shells, while retaining the actual car interior.
	gl("glDepthRange", b64(0), b64(.04))
	defer gl("glDepthRange", b64(0), b64(1))
	push()
	move(seat.X, eyeY-.28, seat.Z)
	rotate(c.Heading*180/math.Pi, 0, 1, 0)
	move(0, -.10, .66)
	push()
	rotate(72, 1, 0, 0)
	rotate(-c.VisualSteer*430, 0, 1, 0)
	drawPBRAsset("rm_steering", 0)
	pop()
	move(0, -.10, .34)
	resize(.65, .65, .65)
	rotate(180, 0, 1, 0)
	drawPBRAsset("rm_dashboard", 0)
	pop()

}
func drawRemasteredFrame(dt float64) {
	if !remaster.Ready || app.width <= 0 || app.height <= 0 {
		return
	}
	if qa.enabled && qa.stage == 112 {
		drawCharacterAudit032()
		presentFrame()
		return
	}
	persistCamera036()
	start := time.Now()
	if qa.enabled && qa.stage == 132 {
		drawGarage033(dt)
		drawLoading036(.64, "LOADING ENVIRONMENT")
		presentFrame()
		return
	}
	gpuWorld.draws = 0
	if beta.screen != screenDrive && !remasterSceneOverlay() {
		if beta.screen == screenMaps && ui020.Stage == 1 {
			drawMenuBackdrop036()
		} else if beta.screen == screenTitle {
			drawGarage033(dt)
		} else {
			drawRemasterShowroom(dt)
		}
		remasterMenuUI()
		presentFrame()
		return
	}
	if remasterSceneOverlay() && remaster.SceneReady && hdrScene.fbo != 0 {
		hdrScene.active = true
		injectRasterQA037()
		finishHDRScene()
		if remaster.FullMap {
			drawAokawaFullMap()
		} else {
			remasterMenuUI()
		}
		presentFrame()
		return
	}
	c := engine.InterpolateCar(app.prevCar, app.world.Car, app.renderAlpha)
	app.renderCar = c
	velocityYaw := math.Atan2(math.Sin(c.Heading)*c.Speed+math.Cos(c.Heading)*c.Lateral, math.Cos(c.Heading)*c.Speed-math.Sin(c.Heading)*c.Lateral)
	delta := math.Atan2(math.Sin(velocityYaw-c.Heading), math.Cos(velocityYaw-c.Heading))
	follow := c.Heading + delta*engine.Clamp(math.Abs(c.Lateral)/30, 0, .38)
	beta.cameraHeading += math.Atan2(math.Sin(follow-beta.cameraHeading), math.Cos(follow-beta.cameraHeading)) * (1 - math.Exp(-dt*7))
	floor := c.Body.Height - .74
	var want, target pos3
	if app.boolFPP {
		want, target = beta021CockpitEye(c)
	} else {
		ahead := engine.Clamp(math.Abs(c.Speed)*.065, 0, 2.8)
		target = pos3{c.Pos.X + math.Sin(follow)*ahead, floor + 1.12, c.Pos.Z + math.Cos(follow)*ahead}
		yaw := beta.cameraHeading + app.camYaw + lookBackYaw036()
		distance := aokawaCameraDistance()
		want = pos3{c.Pos.X - math.Sin(yaw)*distance, floor + 2.2 + app.camPitch*4, c.Pos.Z - math.Cos(yaw)*distance}
		fraction := app.world.RemasteredCameraFraction([3]float64{target.x, target.y, target.z}, [3]float64{want.x, want.y, want.z})
		if fraction < 1 {
			want = add(target, scale(sub(want, target), fraction))
		}
		// Terrain clearance raises the boom; shortening it into the chassis made
		// steep shorelines show the car's underside and cockpit polygons.
		for t := .15; t <= 1; t += .08 {
			p := add(target, scale(sub(want, target), t))
			q := engine.Vec{X: p.x, Z: p.z}
			if !app.world.InsideTunnel(q) && p.y < app.world.GroundHeight(q)+.4 {
				want.y += (app.world.GroundHeight(q) + .55 - p.y) / t
			}
		}
		if math.Hypot(want.x-c.Pos.X, want.z-c.Pos.Z) < 2.7 {
			want = pos3{c.Pos.X - math.Sin(yaw)*2.7, math.Max(want.y, floor+3.0), c.Pos.Z - math.Cos(yaw)*2.7}
		}
		want.y = math.Min(want.y, floor+7.0)
	}
	if walker051.Active {
		p := walker051.Ped
		yaw := app.world.Car.Heading + app.camYaw
		target = pos3{p.Pos.X, p.Height + 1.3, p.Pos.Z}
		want = pos3{p.Pos.X - math.Sin(yaw)*3.5, p.Height + 2.0 + app.camPitch*2, p.Pos.Z - math.Cos(yaw)*3.5}
		fraction := app.world.RemasteredCameraFraction([3]float64{target.x, target.y, target.z}, [3]float64{want.x, want.y, want.z})
		if fraction < 1 {
			want = add(target, scale(sub(want, target), fraction))
		}
	}
	if remaster.CameraReset || app.boolFPP {
		app.cam = want
		remaster.CameraVelocity = pos3{}
		beta.cameraHeading = c.Heading
		remaster.CameraReset = false
	} else {
		// Analytic critically damped spring; bounded timestep and no overshoot.
		anchor := pos3{c.Pos.X, floor, c.Pos.Z}
		app.cam = add(app.cam, sub(anchor, remaster.CameraAnchor))
		omega := 11.
		if beta.idleMouse < .25 {
			omega = 24
		}
		e := math.Exp(-omega * dt)
		offset := sub(app.cam, want)
		temp := scale(add(remaster.CameraVelocity, scale(offset, omega)), dt)
		remaster.CameraVelocity = scale(sub(remaster.CameraVelocity, scale(temp, omega)), e)
		app.cam = add(want, scale(add(offset, temp), e))
	}
	remaster.CameraAnchor = pos3{c.Pos.X, floor, c.Pos.Z}
	viewQA036(c, &target)
	viewQA037(c, &target)
	viewQAWorldFinal040(c, &target)
	viewQA052(c, &target)
	photoView041(c, &target)
	beginGPUTiming()
	renderSunShadows(c)
	beginHDRScene()
	fov := 59 + engine.Clamp(math.Abs(c.Speed)*.045, 0, 4)
	if qa.enabled && qa.stage >= 124 && qa.stage <= 129 {
		fov = 46
	}
	if app.boolFPP {
		fov = 68
	}
	if photo041.Active {
		fov = photo041.FOV
	}
	perspective(fov)
	gl("glViewport", 0, 0, uintptr(hdrRenderWidth()), uintptr(hdrRenderHeight()))
	gl("glClear", GL_COLOR_BUFFER_BIT|GL_DEPTH_BUFFER_BIT)
	clearMotion038()
	gl("glLoadIdentity")
	shake := engine.SpeedCameraShake(c.Speed, app.world.Elapsed, app.boolFPP)
	for i := range shake {
		shake[i] *= engine.Clamp(beta.cfg.CameraShake, 0, 1)
	}
	forward := engine.HeadingVec(c.Heading)
	offset := pos3{forward.Z*shake[0] + forward.X*shake[2], shake[1], -forward.X*shake[0] + forward.Z*shake[2]}
	applyCamera(add(app.cam, offset), add(target, scale(offset, .25)))
	beginTemporal038()
	temporalOutputs038(false)
	drawRemasterSky(app.cam)
	temporalOutputs038(true)
	gl("glFogi", GL_FOG_MODE, GL_LINEAR_FOG)
	a, b, fog := fog034()
	if beta.cfg.Lightweight || sceneFactor026() < .999 {
		// Fade distant geometry before the reduced terrain draw radius ends.
		b[0] = min(b[0], float32(900*viewDistance026()))
		a[0] = min(a[0], b[0]*.35)
	}
	gl("glFogfv", GL_FOG_START, uintptr(unsafe.Pointer(&a[0])))
	gl("glFogfv", GL_FOG_END, uintptr(unsafe.Pointer(&b[0])))
	gl("glFogfv", GL_FOG_COLOR, uintptr(unsafe.Pointer(&fog[0])))
	en(GL_FOG)
	useLighting(true)
	updateSceneLights(c)
	setupTemporal038()
	drawClouds051()
	drawRemasterWorld(app.cam, target, false, 1000)
	drawRemasterActors(false, 260*viewDistance026())

	effectsDT := dt
	if remasterSceneOverlay() || photo041.Active {
		effectsDT = 0
	}
	snapshotBeforeParticles038()
	temporalOutputs038(false)
	drawRemasterEffects(c, effectsDT, target)
	drawFireflies034()
	drawWeather050(effectsDT)
	temporalOutputs038(true)
	useLighting(false)
	updateRayScene038()
	runTemporalTransport038()
	injectRasterQA037()
	finishHDRScene()
	remaster.SceneReady = true
	snapshotHUDless038()
	if remaster.FullMap {
		drawAokawaFullMap()
	} else if remasterSceneOverlay() {
		remasterMenuUI()
	} else if photo041.Active {
		photoHUD041()
	} else {
		drawRemasterHUD(c)
		drawControllerHint038()
	}
	endGPUTiming()
	app.renderMS = float64(time.Since(start).Microseconds()) / 1000
	presentFrame()
}
func drawRemasterShowroom(dt float64) { drawGarage033(dt) }

func drawRemasterEffects(c engine.Car, dt float64, target pos3) {
	drawBeta011Skids(c)
	remaster.SmokeClock += dt
	if remaster.SmokeClock > .035 {
		remaster.SmokeClock = math.Mod(remaster.SmokeClock, .035)
		f := engine.HeadingVec(c.Heading)
		r := engine.Vec{X: f.Z, Z: -f.X}
		addParticle := func(side float64, exhaust bool) {
			spec := engine.VehicleVariantSpec050(app.world.CarClass, c.Variant050)
			p := c.Pos.Sub(f.Mul(spec.P.Wheelbase * spec.FrontWeight)).Add(r.Mul(side))
			height := beta021SkidHeight(app.world, p) + .16
			if exhaust {
				for _, outlet := range aokawaExhausts(app.world.CarClass) {
					if math.Abs(outlet.x-side) < .01 {
						q := aokawaToWorld(c, outlet)
						p = engine.Vec{X: q.x, Z: q.z}
						height = q.y
					}
				}
			}
			q := &remaster.Particles[remaster.ParticleCursor%len(remaster.Particles)]
			remaster.ParticleCursor++
			*q = remasterParticle{Pos: pos3{p.X, height, p.Z}, Vel: pos3{f.X*c.Speed*.10 + r.X*side*.12, .45, f.Z*c.Speed*.10 + r.Z*side*.12}, Life: 1.8, MaxLife: 1.8, Size: .20, Exhaust: exhaust}
			if exhaust {
				q.Size = .08
				q.Life = .55
				q.MaxLife = .55
				q.Vel = pos3{-f.X * .8, .25, -f.Z * .8}
			}
		}
		if c.Body.Grounded && (c.Burnout || c.DriftActive && (math.Abs(c.Lateral) > .7 || math.Abs(c.Speed) > 3)) {
			track := engine.RemasteredVehicleSpec(app.world.CarClass).P.Track * .5
			addParticle(-track, false)
			addParticle(track, false)
		} else if c.EngineRunning && (math.Abs(c.Speed) < 12 || c.Boosting) {
			for _, outlet := range aokawaExhausts(app.world.CarClass) {
				addParticle(outlet.x, true)
			}
		}
	}
	useLighting(false)
	dis(GL_FOG)
	dis(GL_TEXTURE_2D)
	gl("glDepthMask", 0)
	right := norm(cross(norm(sub(target, app.cam)), pos3{0, 1, 0}))
	up := pos3{0, 1, 0}
	for i := range remaster.Particles {
		p := &remaster.Particles[i]
		if p.Life <= 0 {
			continue
		}
		p.Life -= dt
		p.Pos = add(p.Pos, scale(p.Vel, dt))
		p.Size += dt * .48
		p.Vel.x *= math.Exp(-dt * .8)
		p.Vel.z *= math.Exp(-dt * .8)
		alpha := engine.Clamp(p.Life/p.MaxLife, 0, 1) * .55
		if p.Exhaust {
			alpha *= .24
		}
		en(GL_TEXTURE_2D)
		bind("b035_smoke.png")
		col(.78, .81, .84, alpha)
		size := p.Size
		begin(GL_QUADS)
		for _, q := range [][4]float64{{-1, -1, 0, 1}, {1, -1, 1, 1}, {1, 1, 1, 0}, {-1, 1, 0, 0}} {
			point := add(p.Pos, add(scale(right, q[0]*size), scale(up, q[1]*size)))
			tex(q[2], q[3])
			vtx(point.x, point.y, point.z)
		}
		end()

	}
	gl("glDepthMask", 1)
	useLighting(true)
	en(GL_FOG)
	// Persistent material lightfield replaces additive fake road pools.
	drawLampLenses051()
	drawAokawaNitro(c)
	if false {
		f := engine.HeadingVec(c.Heading)
		p := c.Pos.Sub(f.Mul(2.1))
		push()
		move(p.X, c.Body.Height-.44, p.Z)
		rotate(c.Heading*180/math.Pi+90, 0, 1, 0)
		sceneVector("uEmission", pos3{.2, 1.5, 4})
		drawPBRAsset("rm_flame", 0)
		sceneVector("uEmission", pos3{})
		pop()
	}
}

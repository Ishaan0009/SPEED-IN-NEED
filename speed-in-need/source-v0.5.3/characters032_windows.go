package main

import (
	"fmt"
	"math"
	"runtime"
	"speedinneed/engine"
	"unsafe"
)

type characterBatch032 struct {
	Vertices                [][12]float32
	Indices                 []uint32
	VBO, EBO                uint32
	MotionVBO               uint32
	Previous, MotionScratch [][3]float32
}
type characterBasis032 struct{ A, U, Axis, V, NormalAxis engine.PedVec }

var characters032 struct {
	Models           [5]*characterAsset032
	Batches          [5]characterBatch032
	DriverDrawn      bool
	PedestriansDrawn int
}

func initCharacters032() {
	for i, name := range []string{"manuel", "nathan", "sophia", "driver", "walker"} {
		raw, err := assets.ReadFile("assets/characters/" + name + ".chr")
		if err != nil {
			pbrErrors = append(pbrErrors, "character: "+name)
			continue
		}
		a, err := decodeCharacter032(raw)
		if err != nil {
			pbrErrors = append(pbrErrors, name+": "+err.Error())
			continue
		}
		characters032.Models[i] = a
		pbrTexture(a.Albedo)
	}
}
func characterPoseBasis032(pose [engine.PedJointCount]engine.PedVec) [15]characterBasis032 {
	var result [15]characterBasis032
	up := pose[1].Sub(pose[0]).Unit()
	front := pedCross(pose[6].Sub(pose[3]), up).Unit()
	if front.Length() < .5 {
		front = engine.PedVec{Z: 1}
	}
	for i := range result {
		var a, b engine.PedVec
		if i == 14 {
			a = pose[2]
			b = a.Add(pose[2].Sub(pose[1]).Unit().Mul(.25))
		} else {
			bone := engine.PedBones[i]
			a, b = pose[bone.A], pose[bone.B]
		}
		axis := b.Sub(a)
		unit := axis.Unit()
		seed := front
		if math.Abs(unit.Dot(seed)) > .94 {
			seed = pose[6].Sub(pose[3]).Unit()
		}
		u := pedCross(unit, seed).Unit()
		v := pedCross(unit, u)
		result[i] = characterBasis032{a, u, axis, v, unit}
	}
	return result
}
func skinPoint032(p [3]float32, b characterBasis032) engine.PedVec {
	return b.A.Add(b.U.Mul(float64(p[0]))).Add(b.Axis.Mul(float64(p[1]))).Add(b.V.Mul(float64(p[2])))
}
func skinNormal032(n [3]float32, b characterBasis032) engine.PedVec {
	return b.U.Mul(float64(n[0])).Add(b.NormalAxis.Mul(float64(n[1]))).Add(b.V.Mul(float64(n[2])))
}
func appendCharacter032(index, lod int, pose [engine.PedJointCount]engine.PedVec, hideHead bool) {
	if index == 3 {
		driverCache035.Valid = false
		driverCache035.Dirty = true
	}
	a := characters032.Models[index]
	if a == nil {
		return
	}
	mesh := a.LODs[lod]
	batch := &characters032.Batches[index]
	base := uint32(len(batch.Vertices))
	matrices := characterPoseBasis032(pose)
	if index == 3 {
		matrices = driverSkinBasis035(pose)
	}
	for _, v := range mesh.Vertices {
		w := float64(v.Weight)
		p := skinPoint032(v.A, matrices[v.BoneA]).Mul(w).Add(skinPoint032(v.B, matrices[v.BoneB]).Mul(1 - w))
		n := skinNormal032(v.NA, matrices[v.BoneA]).Mul(w).Add(skinNormal032(v.NB, matrices[v.BoneB]).Mul(1 - w)).Unit()
		batch.Vertices = append(batch.Vertices, [12]float32{float32(p.X), float32(p.Y), float32(p.Z), float32(n.X), float32(n.Y), float32(n.Z), v.UV[0], v.UV[1], 1, 0, 0, 1})
	}
	if index != 3 && skinEntity038 != "" && hdrScene.motion != 0 && hdrScene.active && !sunShadow.drawing {
		batch.MotionScratch = batch.MotionScratch[:0]
		for _, v := range batch.Vertices[base:] {
			batch.MotionScratch = append(batch.MotionScratch, [3]float32{v[0], v[1], v[2]})
		}
		batch.Previous = append(batch.Previous, previousSkin038(fmt.Sprintf("%s/%d/%d", skinEntity038, index, lod), batch.MotionScratch)...)
	}
	for at := 0; at < len(mesh.Indices); at += 3 {
		if index == 3 && driverHandTriangle035(mesh, at) {
			continue
		}
		if hideHead {
			// First person renders only the driver's arms. Skull/neck weights
			// straddle several FBX bones, so hiding head-bone triangles alone
			// left a visible face and shoulders in front of the camera.
			keep := true
			for _, id := range mesh.Indices[at : at+3] {
				v := mesh.Vertices[id]
				bone := v.BoneA
				if v.Weight < .5 {
					bone = v.BoneB
				}
				keep = keep && (bone == 3 || bone == 4 || bone == 6 || bone == 7)
				// Open-palmed source fingers extend past the wheel and cross
				// the eye camera; the cabin view keeps arms up to the wrists.
				local := v.A
				if v.Weight < .5 {
					local = v.B
				}
				if (bone == 4 || bone == 7) && local[1] > 1.04 {
					keep = false
				}
			}
			if !keep {
				continue
			}
		}

		batch.Indices = append(batch.Indices, base+mesh.Indices[at], base+mesh.Indices[at+1], base+mesh.Indices[at+2])
	}
}
func flushCharacter032(index int) {
	a := characters032.Models[index]
	b := &characters032.Batches[index]
	if a == nil || len(b.Indices) == 0 {
		return
	}
	if b.VBO == 0 {
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&b.VBO)))
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&b.EBO)))
	}
	sceneInt("uInstanced", 0)
	sceneInt("uGrassWind", 0)
	sceneInt("uConformGround", 0)
	glExt("glBindVertexArray", 0)
	for _, id := range []uintptr{4, 5, 6} {
		glExt("glDisableVertexAttribArray", id)
	}
	glExt("glBindBuffer", glArrayBuffer, uintptr(b.VBO))
	if index != 3 || driverCache035.Dirty {
		glExt("glBufferData", glArrayBuffer, uintptr(len(b.Vertices))*48, uintptr(unsafe.Pointer(&b.Vertices[0])), 0x88E0)
	}
	glExt("glBindBuffer", 0x8893, uintptr(b.EBO))
	if index != 3 || driverCache035.Dirty {
		glExt("glBufferData", 0x8893, uintptr(len(b.Indices))*4, uintptr(unsafe.Pointer(&b.Indices[0])), 0x88E0)
	}
	bindPBRMaterial(pbrMaterial{Name: "Character " + a.Name, Albedo: a.Albedo, Base: [4]float64{1, 1, 1, 1}, Rough: .88, Double: true, NormalScale: 1})
	pbrPointers(pbrPart{vbo: b.VBO, ebo: b.EBO})
	if index == 3 && hdrScene.motion != 0 && hdrScene.active && !sunShadow.drawing {
		b.MotionScratch = b.MotionScratch[:0]
		for _, v := range b.Vertices {
			b.MotionScratch = append(b.MotionScratch, [3]float32{v[0], v[1], v[2]})
		}
		b.Previous = append(b.Previous[:0], previousSkin038(fmt.Sprintf("driver/%d", app.world.CarClass), b.MotionScratch)...)
		objectMotion038("driver", 0)
	}
	bindPreviousSkin038(&b.MotionVBO, b.Previous)
	defer unbindPreviousSkin038()
	col(1, 1, 1, 1)
	gl("glDrawElements", GL_TRIANGLES, uintptr(len(b.Indices)), 0x1405, 0)
	gpuWorld.draws++
	runtime.KeepAlive(b.Vertices)
	runtime.KeepAlive(b.Indices)
	unbindWorld()
	resetPBRMaterial()
}
func drawImportedPedestrians032(shadow bool) bool {
	if characters032.Models[0] == nil || characters032.Models[1] == nil || characters032.Models[2] == nil {
		return false
	}
	for i := 0; i < 3; i++ {
		b := &characters032.Batches[i]
		b.Vertices = b.Vertices[:0]
		b.Indices = b.Indices[:0]
		b.Previous = b.Previous[:0]
	}
	characters032.PedestriansDrawn = len(pedestrianMesh.visible)
	for _, candidate := range pedestrianMesh.visible {
		p := &app.world.Pedestrians[candidate.index]
		pose := p.Pose
		if candidate.distance < 100 && !p.Sleeping {
			for i := range pose {
				pose[i] = p.PreviousPose[i].Mul(1 - app.renderAlpha).Add(p.Pose[i].Mul(app.renderAlpha))
			}
		}
		lod := 2
		if candidate.distance < 45 && !shadow {
			lod = 1
		}
		if candidate.distance < 16 && !shadow && app.quality > 0 {
			lod = 0
		}
		skinEntity038 = fmt.Sprintf("pedestrian/%d", p.ID)
		appendCharacter032(p.ID%3, lod, pose, false)
		skinEntity038 = ""
	}
	for i := 0; i < 3; i++ {
		flushCharacter032(i)
	}
	pedestrianMesh.Drawn = len(pedestrianMesh.visible)
	return true
}
func seatedDriverPose032(id int, steer float64) [engine.PedJointCount]engine.PedVec {
	// Calibrate against the actual exterior roof and steering-wheel geometry,
	// rather than the separate first-person camera. The imported head extends
	// above its joint, so leave physical headroom in each closed cabin.
	x := []float64{.305, .34, .34, -.36}[id]
	z := []float64{-.14, -.10, -.28, -.14}[id]
	head := []float64{.81, 1.04, .84, 1.01}[id]
	hip := head - .60
	chest := head - .24
	wheel := []float64{.678, .81, .76, .764}[id]
	pose := [engine.PedJointCount]engine.PedVec{
		{X: x, Y: hip, Z: z}, {X: x, Y: chest, Z: z - .035}, {X: x, Y: head, Z: z + .045},
		{X: x - .17, Y: chest, Z: z - .01}, {X: x - .23, Y: chest - .10, Z: z + .22}, {X: x - .16, Y: wheel, Z: z + .50},
		{X: x + .17, Y: chest, Z: z - .01}, {X: x + .23, Y: chest - .10, Z: z + .22}, {X: x + .16, Y: wheel, Z: z + .50},
		{X: x - .10, Y: hip - .02, Z: z}, {X: x - .12, Y: hip - .10, Z: z + .38}, {X: x - .12, Y: .20, Z: z + .69},
		{X: x + .10, Y: hip - .02, Z: z}, {X: x + .12, Y: hip - .10, Z: z + .38}, {X: x + .12, Y: .20, Z: z + .69},
	}
	pose[3].Y += .10
	pose[6].Y += .10
	return fitDriverArms035(pose, steer)
}
func drawDriver032(c engine.Car, id int, cockpit bool) {
	if walker051.Active && id == app.world.CarClass && c.Pos == app.renderCar.Pos {
		characters032.DriverDrawn = false
		return
	}
	characters032.DriverDrawn = false
	// First person keeps the arm triangles and the replacement gripping hands;
	// head/torso geometry is removed before upload, outside the eye near plane.
	if characters032.Models[3] == nil || sunShadow.drawing {
		return
	}
	// Local coordinates share the car's chassis transform, so the driver follows
	// suspension pitch/roll. Hands turn around the wheel; feet reach the pedals.
	pose := seatedDriverPose053(id, c.Variant050, c.VisualSteer)
	b := &characters032.Batches[3]
	lod := 0
	if beta.cfg.Lightweight {
		lod = 1
	}
	if !driverCache035.Valid || b.VBO == 0 || len(b.Indices) == 0 || driverCache035.Pose != pose || driverCache035.LOD != lod || driverCache035.Cockpit != cockpit {
		b.Vertices = b.Vertices[:0]
		b.Indices = b.Indices[:0]
		appendCharacter032(3, lod, pose, cockpit)
		driverCache035.Pose = pose
		driverCache035.LOD = lod
		driverCache035.Cockpit = cockpit
		driverCache035.Valid = true
		driverCache035.Builds++
	} else {
		driverCache035.Reuses++
	}
	flushCharacter032(3)
	driverCache035.Dirty = false
	drawDriverHands035(pose)
	characters032.DriverDrawn = true
}

// Internal asset review uses the real skinning/material path at close range.
func drawCharacterAudit032() {
	beginHDRScene()
	perspective(38)
	gl("glClearColor", b32(.018), b32(.027), b32(.040), b32(1))
	gl("glClear", GL_COLOR_BUFFER_BIT|GL_DEPTH_BUFFER_BIT)
	gl("glLoadIdentity")
	eye, target := pos3{2.8, 1.55, 5.4}, pos3{0, .9, 0}
	applyCamera(eye, target)
	app.cam = eye
	useLighting(true)
	updateSceneLights(engine.Car{})
	sceneFloat("uNight", .02)
	sceneFloat("uSunset", 0)
	sceneInt("uFogEnabled", 0)
	sceneInt("uShadowEnabled", 0)
	setPBRMaterial(false, 0, .8)
	dis(GL_TEXTURE_2D)
	col(.10, .13, .16, 1)
	nrm(0, 1, 0)
	begin(GL_QUADS)
	vtx(-8, -.02, -8)
	vtx(-8, -.02, 8)
	vtx(8, -.02, 8)
	vtx(8, -.02, -8)
	end()
	resetPBRMaterial()
	for model := 0; model < 3; model++ {
		for _, p := range app.world.Pedestrians {
			if p.ID%3 == model {
				pose := p.Pose
				f := engine.HeadingVec(p.Heading)
				r := engine.Vec{X: f.Z, Z: -f.X}
				for j, q := range pose {
					delta := q.Sub(engine.PedVec{X: p.Pos.X, Y: p.Height, Z: p.Pos.Z})
					pose[j] = engine.PedVec{X: float64(model) - 1.5 + delta.X*r.X + delta.Z*r.Z, Y: delta.Y, Z: delta.X*f.X + delta.Z*f.Z}
				}
				b := &characters032.Batches[model]
				b.Vertices = b.Vertices[:0]
				b.Indices = b.Indices[:0]
				appendCharacter032(model, 0, pose, false)
				flushCharacter032(model)
				break
			}
		}
	}
	push()
	move(1.5, 0, 0)
	drawDriver032(engine.Car{}, 0, false)
	pop()
	useLighting(false)
	finishHDRScene()
}

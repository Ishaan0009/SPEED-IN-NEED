package main

import (
	"math"
	"runtime"
	"sort"
	"speedinneed/engine"
	"unsafe"
)

type pedestrianCandidate struct {
	index    int
	distance float64
}

var pedestrianMesh struct {
	buffer   uint32
	vertices []bufferVertex
	visible  []pedestrianCandidate
	Drawn    int
}

func pedCross(a, b engine.PedVec) engine.PedVec {
	return engine.PedVec{X: a.Y*b.Z - a.Z*b.Y, Y: a.Z*b.X - a.X*b.Z, Z: a.X*b.Y - a.Y*b.X}
}
func pedMeshVertex(p, n engine.PedVec, c [4]uint8) bufferVertex {
	return bufferVertex{P: [3]float32{float32(p.X), float32(p.Y), float32(p.Z)}, N: [3]float32{float32(n.X), float32(n.Y), float32(n.Z)}, C: c}
}
func pedTriangle(a, b, c, na, nb, nc engine.PedVec, color [4]uint8) {
	if pedCross(b.Sub(a), c.Sub(a)).Dot(na.Add(nb).Add(nc)) < 0 {
		b, c = c, b
		nb, nc = nc, nb
	}
	pedestrianMesh.vertices = append(pedestrianMesh.vertices, pedMeshVertex(a, na, color), pedMeshVertex(b, nb, color), pedMeshVertex(c, nc, color))
}
func pedLimb(a, b engine.PedVec, radius float64, color [4]uint8, sides int) {
	axis := b.Sub(a).Unit()
	seed := engine.PedVec{Y: 1}
	if math.Abs(axis.Y) > .9 {
		seed = engine.PedVec{X: 1}
	}
	u := pedCross(axis, seed).Unit()
	v := pedCross(axis, u).Unit()
	for i := 0; i < sides; i++ {
		t0, t1 := float64(i)*engine.Tau/float64(sides), float64(i+1)*engine.Tau/float64(sides)
		n0, n1 := u.Mul(math.Cos(t0)).Add(v.Mul(math.Sin(t0))), u.Mul(math.Cos(t1)).Add(v.Mul(math.Sin(t1)))
		p0, p1, q0, q1 := a.Add(n0.Mul(radius)), a.Add(n1.Mul(radius)), b.Add(n0.Mul(radius*.88)), b.Add(n1.Mul(radius*.88))
		pedTriangle(p0, p1, q1, n0, n1, n1, color)
		pedTriangle(p0, q1, q0, n0, n1, n0, color)
		pedTriangle(a, p1, p0, axis.Mul(-1), axis.Mul(-1), axis.Mul(-1), color)
		pedTriangle(b, q0, q1, axis, axis, axis, color)
	}
}
func pedEllipsoid(center, size engine.PedVec, color [4]uint8, sides, rings int) {
	point := func(a, b float64) (engine.PedVec, engine.PedVec) {
		n := engine.PedVec{X: math.Cos(a) * math.Sin(b), Y: math.Cos(b), Z: math.Sin(a) * math.Sin(b)}
		return center.Add(engine.PedVec{X: n.X * size.X, Y: n.Y * size.Y, Z: n.Z * size.Z}), engine.PedVec{X: n.X / size.X, Y: n.Y / size.Y, Z: n.Z / size.Z}.Unit()
	}
	for j := 0; j < rings; j++ {
		for i := 0; i < sides; i++ {
			a, b := float64(i)*engine.Tau/float64(sides), float64(i+1)*engine.Tau/float64(sides)
			y, z := float64(j)*math.Pi/float64(rings), float64(j+1)*math.Pi/float64(rings)
			p0, n0 := point(a, y)
			p1, n1 := point(b, y)
			p2, n2 := point(b, z)
			p3, n3 := point(a, z)
			if j > 0 {
				pedTriangle(p0, p1, p2, n0, n1, n2, color)
			}
			if j < rings-1 {
				pedTriangle(p0, p2, p3, n0, n2, n3, color)
			}
		}
	}
}
func appendPedestrianMesh(p *engine.Pedestrian, distance float64, shadow bool) {
	pose := p.Pose
	if distance < 100 && !p.Sleeping {
		for i := range pose {
			pose[i] = p.PreviousPose[i].Mul(1 - app.renderAlpha).Add(p.Pose[i].Mul(app.renderAlpha))
		}
	}
	skin := [...][4]uint8{{210, 160, 119, 255}, {164, 110, 77, 255}, {106, 71, 53, 255}, {231, 187, 151, 255}}
	shirts := [...][4]uint8{{38, 73, 104, 255}, {137, 52, 43, 255}, {58, 96, 73, 255}, {185, 166, 126, 255}, {93, 73, 115, 255}, {182, 185, 176, 255}}
	pants := [...][4]uint8{{38, 43, 53, 255}, {45, 63, 82, 255}, {80, 72, 59, 255}, {36, 41, 38, 255}}
	colors := [3][4]uint8{skin[p.ID%len(skin)], shirts[p.ID%len(shirts)], pants[(p.ID/3)%len(pants)]}
	sides := 8
	if distance > 45 || shadow {
		sides = 5
	}
	for _, bone := range engine.PedBones {
		pedLimb(pose[bone.A], pose[bone.B], bone.Radius, colors[bone.Material], sides)
	}
	pedEllipsoid(pose[2], engine.PedVec{X: .135, Y: .17, Z: .14}, colors[0], sides, 4)
	if distance < 65 && !shadow {
		hair := [4]uint8{35, 29, 25, 255}
		if p.ID%5 == 0 {
			hair = [4]uint8{92, 64, 36, 255}
		}
		pedEllipsoid(pose[2].Add(engine.PedVec{Y: .105}), engine.PedVec{X: .137, Y: .075, Z: .142}, hair, sides, 3)
		for _, joint := range []int{3, 4, 6, 7, 10, 13} {
			material, radius := 1, .072
			if joint == 10 || joint == 13 {
				material = 2
				radius = .09
			}
			pedEllipsoid(pose[joint], engine.PedVec{X: radius, Y: radius, Z: radius}, colors[material], 5, 3)
		}
		for _, joint := range []int{5, 8} {
			pedEllipsoid(pose[joint], engine.PedVec{X: .065, Y: .08, Z: .065}, colors[0], 5, 3)
		}
	}
	for _, joint := range []int{11, 14} {
		pedEllipsoid(pose[joint], engine.PedVec{X: .095, Y: .07, Z: .15}, [4]uint8{28, 29, 30, 255}, sides, 3)
	}
}
func drawPedestrians(shadow bool, radius float64) {
	pedestrianMesh.Drawn = 0
	if app.world.Remaster == nil || len(app.world.Pedestrians) == 0 {
		return
	}
	limit := math.Min(radius, 160*viewDistance026())
	maximum := 64
	if app.quality == 0 {
		limit = math.Min(limit, 100)
		maximum = 32
	}
	if shadow {
		limit = math.Min(limit, 65)
		maximum = 20
	}
	pedestrianMesh.visible = pedestrianMesh.visible[:0]
	for i := range app.world.Pedestrians {
		p := &app.world.Pedestrians[i]
		if !shadow && !visibleCone034(p.Pos.X, p.Pos.Z, 2, app.cam, remaster.CameraAnchor) {
			continue
		}
		d := p.Pos.Dist(app.renderCar.Pos)
		if d <= limit {
			pedestrianMesh.visible = append(pedestrianMesh.visible, pedestrianCandidate{i, d})
		}
	}
	sort.Slice(pedestrianMesh.visible, func(i, j int) bool { return pedestrianMesh.visible[i].distance < pedestrianMesh.visible[j].distance })
	if len(pedestrianMesh.visible) > maximum {
		pedestrianMesh.visible = pedestrianMesh.visible[:maximum]
	}
	if drawImportedPedestrians032(shadow) {
		return
	}
	pedestrianMesh.vertices = pedestrianMesh.vertices[:0]
	for _, c := range pedestrianMesh.visible {
		appendPedestrianMesh(&app.world.Pedestrians[c.index], c.distance, shadow)
	}
	if len(pedestrianMesh.vertices) == 0 {
		return
	}
	pedestrianMesh.Drawn = len(pedestrianMesh.visible)
	if pedestrianMesh.buffer == 0 {
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&pedestrianMesh.buffer)))
	}
	resetPBRMaterial()
	setPBRMaterial(true, 0, .9)
	sceneInt("uInstanced", 0)
	sceneInt("uGrassWind", 0)
	sceneVector("uSurfaceResponse", pos3{.05, .15, .85})
	glExt("glBindVertexArray", 0)
	for _, index := range []uintptr{4, 5, 6, 7} {
		glExt("glDisableVertexAttribArray", index)
	}
	gl("glDisableClientState", 0x8078)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(app.whiteTexture))
	en(GL_TEXTURE_2D)
	glExt("glBindBuffer", glArrayBuffer, uintptr(pedestrianMesh.buffer))
	glExt("glBufferData", glArrayBuffer, uintptr(len(pedestrianMesh.vertices))*unsafe.Sizeof(bufferVertex{}), uintptr(unsafe.Pointer(&pedestrianMesh.vertices[0])), 0x88E0)
	for _, state := range []uintptr{glVertexArray, glNormalArray, glColorArray} {
		gl("glEnableClientState", state)
	}
	gl("glVertexPointer", 3, glFloat, unsafe.Sizeof(bufferVertex{}), 0)
	gl("glNormalPointer", glFloat, unsafe.Sizeof(bufferVertex{}), 12)
	gl("glColorPointer", 4, GL_UNSIGNED_BYTE, unsafe.Sizeof(bufferVertex{}), 24)
	gl("glDrawArrays", GL_TRIANGLES, 0, uintptr(len(pedestrianMesh.vertices)))
	gpuWorld.draws++
	runtime.KeepAlive(pedestrianMesh.vertices)
	unbindWorld()
	resetPBRMaterial()
}

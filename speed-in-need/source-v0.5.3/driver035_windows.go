package main

import (
	"math"
	"speedinneed/engine"
	"unsafe"
)

var driverCache035 struct {
	Pose                  [engine.PedJointCount]engine.PedVec
	LOD                   int
	Cockpit, Valid, Dirty bool
	Builds, Reuses        uint64
}

// Match the tilted steering-wheel axis. The old XY-only rotation had the
// opposite sign, and elbows stayed fixed while the open palms stretched.
func rotateDriver035(v, axis engine.PedVec, angle float64) engine.PedVec {
	c, s := math.Cos(angle), math.Sin(angle)
	return v.Mul(c).Add(pedCross(axis, v).Mul(s)).Add(axis.Mul(axis.Dot(v) * (1 - c)))
}
func fitDriverArms035(pose [engine.PedJointCount]engine.PedVec, steer float64) [engine.PedJointCount]engine.PedVec {
	center := pose[5].Add(pose[8]).Mul(.5)
	axis := engine.PedVec{Y: .25, Z: 1}.Unit()
	angle := steeringWheelAngle035(steer)
	for _, arm := range [][3]int{{3, 4, 5}, {6, 7, 8}} {
		shoulder := pose[arm[0]]
		wrist := center.Add(rotateDriver035(pose[arm[2]].Sub(center), axis, angle))
		delta := wrist.Sub(shoulder)
		d := delta.Length()
		unit := delta.Unit()
		upper, lower := .305, .335
		if d > upper+lower-.005 {
			wrist = shoulder.Add(unit.Mul(upper + lower - .005))
			d = upper + lower - .005
		}
		a := (upper*upper - lower*lower + d*d) / (2 * math.Max(d, .001))
		side := -1.
		if arm[0] == 6 {
			side = 1
		}
		pole := engine.PedVec{X: side * .7, Y: -.8, Z: -.2}
		pole = pole.Sub(unit.Mul(pole.Dot(unit))).Unit()
		pose[arm[1]] = shoulder.Add(unit.Mul(a)).Add(pole.Mul(math.Sqrt(math.Max(0, upper*upper-a*a))))
		pose[arm[2]] = wrist
	}
	return pose
}
func steeringWheelAngle035(steer float64) float64 { return engine.Clamp(-steer*3, -.90, .90) }
func driverHandTriangle035(mesh characterLOD032, at int) bool {
	for _, id := range mesh.Indices[at : at+3] {
		v := mesh.Vertices[id]
		bone, p := v.BoneA, v.A
		if v.Weight < .5 {
			bone, p = v.BoneB, v.B
		}
		if (bone == 4 || bone == 7) && p[1] > .94 {
			return true
		}
	}
	return false
}

// Keep a continuous anatomical side reference when arms point forward.
func driverSkinBasis035(pose [engine.PedJointCount]engine.PedVec) [15]characterBasis032 {
	basis := characterPoseBasis032(pose)
	side := pose[6].Sub(pose[3]).Unit()
	for _, i := range []int{3, 4, 6, 7} {
		bone := engine.PedBones[i]
		a := pose[bone.A]
		axis := pose[bone.B].Sub(a)
		unit := axis.Unit()
		u := pedCross(unit, side).Unit()
		v := pedCross(unit, u)
		basis[i] = characterBasis032{a, u, axis, v, unit}
	}
	// Head orientation is independent of the sloped chest-to-head neck bone.
	// Following that bone pitched the face down into the steering wheel.
	axis := engine.PedVec{Y: .25}
	basis[14] = characterBasis032{pose[2], engine.PedVec{X: 1}, axis, engine.PedVec{Z: -1}, engine.PedVec{Y: 1}}
	return basis
}

var driverHands035 struct {
	Buffer        uint32
	MotionBuffer  uint32
	MotionScratch [][3]float32
	Vertices      []bufferVertex
}

func drawDriverHands035(pose [engine.PedJointCount]engine.PedVec) {
	previous := pedestrianMesh.vertices
	pedestrianMesh.vertices = driverHands035.Vertices[:0]
	center := pose[5].Add(pose[8]).Mul(.5)
	axis := engine.PedVec{Y: .25, Z: 1}.Unit()
	glove := [4]uint8{38, 42, 48, 255}
	for _, j := range []int{5, 8} {
		w := pose[j]
		radial := w.Sub(center).Unit()
		tangent := pedCross(axis, radial).Unit()
		pedEllipsoid(w.Add(axis.Mul(.005)), engine.PedVec{X: .042, Y: .040, Z: .030}, glove, 8, 5)
		for _, f := range []float64{-.023, -.008, .008, .023} {
			a := w.Add(tangent.Mul(f)).Add(radial.Mul(.018)).Add(axis.Mul(.012))
			b := a.Sub(axis.Mul(.042)).Sub(radial.Mul(.008))
			end := b.Sub(radial.Mul(.028)).Add(axis.Mul(.009))
			pedLimb(a, b, .008, glove, 5)
			pedLimb(b, end, .007, glove, 5)
		}
		pedLimb(w.Sub(radial.Mul(.025)).Add(tangent.Mul(.025)), w.Sub(radial.Mul(.035)).Sub(axis.Mul(.023)), .014, glove, 6)
	}
	driverHands035.Vertices = pedestrianMesh.vertices
	pedestrianMesh.vertices = previous
	if driverHands035.Buffer == 0 {
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&driverHands035.Buffer)))
	}
	resetPBRMaterial()
	setPBRMaterial(true, 0, .82)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(app.whiteTexture))
	en(GL_TEXTURE_2D)
	glExt("glBindVertexArray", 0)
	for _, a := range []uintptr{4, 5, 6, 7} {
		glExt("glDisableVertexAttribArray", a)
	}
	glExt("glBindBuffer", glArrayBuffer, uintptr(driverHands035.Buffer))
	glExt("glBufferData", glArrayBuffer, uintptr(len(driverHands035.Vertices))*28, uintptr(unsafe.Pointer(&driverHands035.Vertices[0])), 0x88E0)
	for _, state := range []uintptr{glVertexArray, glNormalArray, glColorArray} {
		gl("glEnableClientState", state)
	}
	gl("glDisableClientState", 0x8078)
	gl("glVertexPointer", 3, glFloat, 28, 0)
	gl("glNormalPointer", glFloat, 28, 12)
	gl("glColorPointer", 4, GL_UNSIGNED_BYTE, 28, 24)
	if hdrScene.motion != 0 && hdrScene.active && !sunShadow.drawing {
		driverHands035.MotionScratch = driverHands035.MotionScratch[:0]
		for _, v := range driverHands035.Vertices {
			driverHands035.MotionScratch = append(driverHands035.MotionScratch, v.P)
		}
		previous := previousSkin038("driver/hands", driverHands035.MotionScratch)
		objectMotion038("driver/hands", 0)
		bindPreviousSkin038(&driverHands035.MotionBuffer, previous)
		defer unbindPreviousSkin038()
	}
	gl("glDrawArrays", GL_TRIANGLES, 0, uintptr(len(driverHands035.Vertices)))
	gpuWorld.draws++
	unbindWorld()
	resetPBRMaterial()
}

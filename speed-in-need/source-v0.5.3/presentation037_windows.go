package main

import (
	"math"
	"speedinneed/engine"
	"unsafe"
)

// Every off-screen pass owns its viewport and raster state. A HUD scissor or
// vegetation alpha/cull state must never clip a later full-screen resolve.
func fullFrame037(width, height int) {
	for _, state := range []uintptr{0x0C11, 0x0B90, 0x0BC0, GL_CULL_FACE, 0x8037} {
		dis(state)
	}
	gl("glColorMask", 1, 1, 1, 1)
	gl("glViewport", 0, 0, uintptr(width), uintptr(height))
}

var garage037 struct {
	Yaw, Pitch, Zoom float64
	Drag, Touched    bool
	X, Y             int
}

func garageView037() bool { return beta.screen == screenMaps && ui020.Stage == 2 }
func garageZoom037(delta float64) {
	if garage037.Zoom == 0 {
		garage037.Zoom = 1
	}
	garage037.Zoom = engine.Clamp(garage037.Zoom+delta, .72, 1.42)
}
func garagePointer037(message uint32, x, y int) bool {
	if !garageView037() {
		garage037.Drag = false
		return false
	}
	switch message {
	case WM_LBUTTONDOWN:
		if buttonAt033(x, y) == "" && float64(x) > float64(app.width)*.34 && float64(y) > float64(app.height)*.20 && float64(y) < float64(app.height)*.74 {
			garage037.Drag = true
			garage037.X = x
			garage037.Y = y
			garage037.Touched = true
			remaster.Drag = false
			return true
		}
	case WM_MOUSEMOVE:
		if garage037.Drag {
			garage037.Yaw = math.Remainder(garage037.Yaw-float64(x-garage037.X)*.005, engine.Tau)
			garage037.Pitch = engine.Clamp(garage037.Pitch+float64(y-garage037.Y)*.002, -.24, .48)
			garage037.X = x
			garage037.Y = y
			return true
		}
	case 0x202:
		if garage037.Drag {
			garage037.Drag = false
			return true
		}
	}
	return false
}
func garageCamera037(eye, target pos3) (pos3, pos3) {
	if !garageView037() {
		return eye, target
	}
	if garage037.Zoom == 0 {
		garage037.Zoom = 1
	}
	return add(target, scale(sub(eye, target), garage037.Zoom)), target
}
func titleWord037(text string, x, y, size, width float64) {
	push()
	move(x, y, 0)
	shear := [16]float32{.88, 0, 0, 0, -.18, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
	gl("glMultMatrixf", uintptr(unsafe.Pointer(&shear[0])))
	if actual := remasterTextWidth(text, size); actual > width/.88 {
		size *= width / .88 / actual
	}
	remasterText(text, size*.012, size*.018, size, [4]float64{0, 0, 0, .50})
	remasterText(text, 0, 0, size, [4]float64{.97, .985, 1, 1})
	pop()
}
func drawTitle037(w, h float64) {
	dis(GL_TEXTURE_2D)
	begin(GL_QUADS)
	col(.005, .012, .023, .99)
	vtx(0, 0, 0)
	vtx(0, h, 0)
	col(.01, .025, .04, .03)
	vtx(w*.70, h, 0)
	vtx(w*.70, 0, 0)
	end()
	hudRect(w*.068, h*.14, w*.056, 3, .16, .82, .92, 1)
	remasterText("OPEN ROAD SERIES  /  BETA 037", w*.068, h*.17, h*.013, cyan)
	titleWord037("SPEED", w*.075, h*.25, h*.095, w*.36)
	titleWord037("IN NEED", w*.075, h*.36, h*.095, w*.36)
	hudRect(w*.068, h*.485, w*.275, 2, .95, .52, .22, .9)
	remasterText("CITY STREETS. COAST ROADS. HIGHER GROUND.", w*.068, h*.52, h*.012, white)
	for i, item := range [][2]string{{"start", "START DRIVING"}, {"settings", "SETTINGS"}, {"quit", "EXIT GAME"}} {
		beta020Button(item[0], item[1], w*.068, h*(.61+float64(i)*.092), w*.275, h*.065, i == 0)
	}
	remasterText("V0.5.3  /  NATIVE EDITION", w*.068, h*.93, h*.012, cyan)
}
func miniFrame037(cx, cy, r float64) {
	dis(GL_TEXTURE_2D)
	gl("glLineWidth", b32(2))
	for _, ring := range [][5]float64{{r + 2, .025, .035, .04, .98}, {r + 4, .79, .94, .57, .65}} {
		col(ring[1], ring[2], ring[3], ring[4])
		begin(0x0002)
		for i := 0; i < 96; i++ {
			a := float64(i) * engine.Tau / 96
			vtx(cx+math.Cos(a)*ring[0], cy+math.Sin(a)*ring[0], 0)
		}
		end()
	}
	gl("glLineWidth", b32(1))
}

// Called while already inside the vehicle's suspension transform, in both
// exterior and cockpit views. Slot 9 is the actual imported steering mesh.
func drawSteering037(c engine.Car, id int) {
	if c.Variant050 > 0 {
		return
	}
	pose := seatedDriverPose053(id, c.Variant050, 0)
	p := pose[5].Add(pose[8]).Mul(.5)
	push()
	move(p.X, p.Y, p.Z)
	rotate(steeringWheelAngle035(c.VisualSteer)*180/math.Pi, 0, .25, 1)
	move(-p.X, -p.Y, -p.Z)
	drawPBRAsset("b033_cabin_"+fmtSpeed021(float64(id)), 9)
	pop()
}

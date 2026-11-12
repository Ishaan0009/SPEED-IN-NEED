//go:build windows

package main

import (
	"math"
	"speedinneed/engine"
	"unsafe"
)

var remasterMapVBO uint32
var remasterMapCount int

func buildRemasterMinimap() {
	if remasterMapVBO != 0 {
		glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&remasterMapVBO)))
	}
	var points []float32
	for _, r := range app.world.Routes {
		for i := 0; i < len(r.Points)-1; i++ {
			a, b := r.Points[i], r.Points[i+1]
			points = append(points, float32(a.X), float32(-a.Z), float32(b.X), float32(-b.Z))
		}
	}
	if len(points) == 0 {
		return
	}
	glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&remasterMapVBO)))
	glExt("glBindBuffer", glArrayBuffer, uintptr(remasterMapVBO))
	glExt("glBufferData", glArrayBuffer, uintptr(len(points)*4), uintptr(unsafe.Pointer(&points[0])), glStaticDraw)
	glExt("glBindBuffer", glArrayBuffer, 0)
	remasterMapCount = len(points) / 2
}
func drawLegacyLineMinimap(c engine.Car) {
	if remasterMapVBO == 0 {
		return
	}
	w, h := float64(app.width), float64(app.height)
	size := math.Min(190, h*.22) * math.Min(beta.cfg.UIScale, 1.25)
	x, y := w-size-w*.025, h*.035
	cx, cy := x+size*.5, y+size*.5
	hudRect(x, y, size, size, .008, .018, .027, .87)
	dis(GL_TEXTURE_2D)
	en(GL_SCISSOR_TEST)
	gl("glScissor", uintptr(int(x+3)), uintptr(int(h-y-size+3)), uintptr(int(size-6)), uintptr(int(size-6)))
	push()
	move(cx, cy, 0)
	rotate(-c.Heading*180/math.Pi, 0, 0, 1)
	resize(size/750, size/750, 1)
	move(-c.Pos.X, c.Pos.Z, 0)
	glExt("glBindVertexArray", 0)
	glExt("glBindBuffer", glArrayBuffer, uintptr(remasterMapVBO))
	gl("glEnableClientState", glVertexArray)
	gl("glVertexPointer", 2, glFloat, 8, 0)
	col(.49, .75, .78, 1)
	gl("glLineWidth", b32(1.7))
	gl("glDrawArrays", 0x0001, 0, uintptr(remasterMapCount))
	gl("glDisableClientState", glVertexArray)
	glExt("glBindBuffer", glArrayBuffer, 0)
	pop()
	project := func(p engine.Vec) (float64, float64) {
		r, f := engine.MapRelative(c.Pos, p, c.Heading)
		return cx + r*size/750, cy - f*size/750
	}
	pin := func(p engine.Vec, color [3]float64) {
		xx, yy := project(p)
		hudRect(xx-2, yy-2, 4, 4, color[0], color[1], color[2], 1)
	}
	for _, t := range app.world.Traffic {
		if t.Pos.Dist(c.Pos) < 400 {
			pin(t.Pos, [3]float64{1, .54, .15})
		}
	}
	for _, r := range app.world.Remaster.Racers {
		if r.Car.Pos.Dist(c.Pos) < 400 {
			pin(r.Car.Pos, [3]float64{1, .28, .18})
		}
	}
	if s := app.world.Remaster; s.Mode == 1 && len(s.Gates) > 0 {
		pin(s.Gates[s.NextGate%len(s.Gates)], [3]float64{.1, 1, .8})
	}
	col(.12, .96, 1, 1)
	begin(GL_TRIANGLES)
	vtx(cx, cy-9, 0)
	vtx(cx-5, cy+6, 0)
	vtx(cx+5, cy+6, 0)
	end()
	dis(GL_SCISSOR_TEST)
	remasterText("GPS", x+10, y+9, h*.011, cyan)
	remasterText("N", cx-math.Sin(c.Heading)*size*.39, cy-math.Cos(c.Heading)*size*.39, h*.011, white)
}

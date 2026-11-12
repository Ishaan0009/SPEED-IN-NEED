package main

import (
	"math"
	"runtime"
	"speedinneed/engine"
	"unsafe"
)

type rainDrop053 struct {
	X, Z, Floor float64
	Sheltered   bool
}

var rain053 struct {
	World    *engine.World
	X, Z     float64
	Drops    []rainDrop053
	Vertices []float32
	VBO      uint32
}

// Rain columns have fixed world positions within a 24 m cell. Resolve their
// terrain/tunnel coverage only when that cell changes, then stream one draw.
// Previously every streak queried terrain and crossed into OpenGL nine times.
func drawRainStreaks053(count int, time float64, right pos3) {
	x0, z0 := math.Floor(app.cam.x/24)*24, math.Floor(app.cam.z/24)*24
	if rain053.World != app.world || rain053.X != x0 || rain053.Z != z0 || len(rain053.Drops) != count {
		rain053.World, rain053.X, rain053.Z = app.world, x0, z0
		rain053.Drops = make([]rainDrop053, count)
		for i := range rain053.Drops {
			a := float64(i)
			p := engine.Vec{X: x0 + math.Mod(a*17.371, 72) - 24, Z: z0 + math.Mod(a*31.173, 72) - 24}
			rain053.Drops[i] = rainDrop053{p.X, p.Z, app.world.GroundHeight(p) + .03, app.world.InsideTunnel(p)}
		}
	}
	rain053.Vertices = rain053.Vertices[:0]
	for i, p := range rain053.Drops {
		a := float64(i)
		y := app.cam.y - 8 + math.Mod(a*7.73-time*(17+math.Mod(a, 7))+1000000, 30)
		if p.Sheltered || y < p.Floor {
			continue
		}
		width := .012 + math.Hypot(p.X-app.cam.x, p.Z-app.cam.z)*.00060
		length := .24 + math.Mod(a*1.713, .48)
		alpha := float32(weather050.Amount * (.28 + .20*math.Mod(a*.371, 1)))
		for _, q := range [4][3]float64{{-1, 0, 0}, {1, 0, 1}, {1, 1, 1}, {-1, 1, 0}} {
			rain053.Vertices = append(rain053.Vertices, float32(p.X+right.x*width*q[0]+q[1]*.08), float32(y-q[1]*length+right.y*width*q[0]), float32(p.Z+right.z*width*q[0]+q[1]*.035), .68, .77, .85, alpha, float32(q[2]), float32(q[1]))
		}
	}
	if len(rain053.Vertices) == 0 {
		return
	}
	if rain053.VBO == 0 {
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&rain053.VBO)))
	}
	glExt("glBindVertexArray", 0)
	glExt("glBindBuffer", glArrayBuffer, uintptr(rain053.VBO))
	glExt("glBufferData", glArrayBuffer, uintptr(len(rain053.Vertices)*4), uintptr(unsafe.Pointer(&rain053.Vertices[0])), 0x88E0)
	for _, state := range []uintptr{glVertexArray, glColorArray, 0x8078} {
		gl("glEnableClientState", state)
	}
	gl("glDisableClientState", glNormalArray)
	gl("glVertexPointer", 3, glFloat, 36, 0)
	gl("glColorPointer", 4, glFloat, 36, 12)
	gl("glTexCoordPointer", 2, glFloat, 36, 28)
	gl("glDrawArrays", GL_QUADS, 0, uintptr(len(rain053.Vertices)/9))
	gpuWorld.draws++
	for _, state := range []uintptr{glVertexArray, glColorArray, 0x8078} {
		gl("glDisableClientState", state)
	}
	glExt("glBindBuffer", glArrayBuffer, 0)
	runtime.KeepAlive(rain053.Vertices)
}

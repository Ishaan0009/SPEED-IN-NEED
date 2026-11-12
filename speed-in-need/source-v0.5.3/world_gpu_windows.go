//go:build windows

package main

import (
	"math"
	"runtime"
	"speedinneed/engine"
	"unsafe"
)

// The same authored world geometry feeds indexed GPU batches or the legacy
// fallback. Capture runs only at map/preset load, never in the driving loop.
type worldVertex struct {
	P, N [3]float32
	UV   [2]float32
	C    [4]uint8
}
type cpuMesh struct {
	vertices []worldVertex
	indices  []uint32
}
type worldBatch struct {
	vbo, ebo, vao, texture uint32
	count                  int
}
type worldTile struct {
	cx, cz        float64
	ground, grass []worldBatch
	instances     uint32
	trees         [][4]float32
}
type worldCapture struct {
	matrix  [16]float64
	stack   [][16]float64
	normal  pos3
	uv      [2]float32
	color   [4]uint8
	texture uint32
	mode    uintptr
	pending []worldVertex
	meshes  map[uint32]*cpuMesh
}

var capture *worldCapture
var gpuWorld struct {
	ready, instancing bool
	tiles             []worldTile
	details, tree     []worldBatch
	instanceLoc       int32
	draws             int
}

func identityMatrix() [16]float64 { return [16]float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1} }
func (c *worldCapture) multiply(m [16]float32) {
	var out [16]float64
	for col := 0; col < 4; col++ {
		for row := 0; row < 4; row++ {
			for k := 0; k < 4; k++ {
				out[col*4+row] += c.matrix[k*4+row] * float64(m[col*4+k])
			}
		}
	}
	c.matrix = out
}
func (c *worldCapture) vertex(x, y, z float64) {
	m := c.matrix
	n := norm(pos3{m[0]*c.normal.x + m[4]*c.normal.y + m[8]*c.normal.z, m[1]*c.normal.x + m[5]*c.normal.y + m[9]*c.normal.z, m[2]*c.normal.x + m[6]*c.normal.y + m[10]*c.normal.z})
	c.pending = append(c.pending, worldVertex{P: [3]float32{float32(m[0]*x + m[4]*y + m[8]*z + m[12]), float32(m[1]*x + m[5]*y + m[9]*z + m[13]), float32(m[2]*x + m[6]*y + m[10]*z + m[14])}, N: [3]float32{float32(n.x), float32(n.y), float32(n.z)}, UV: c.uv, C: c.color})
}
func (c *worldCapture) finish() {
	if len(c.pending) == 0 {
		return
	}
	m := c.meshes[c.texture]
	if m == nil {
		m = &cpuMesh{}
		c.meshes[c.texture] = m
	}
	base := uint32(len(m.vertices))
	m.vertices = append(m.vertices, c.pending...)
	stride := 3
	if c.mode == GL_QUADS {
		stride = 4
	}
	for i := 0; i+stride <= len(c.pending); i += stride {
		a := base + uint32(i)
		m.indices = append(m.indices, a, a+1, a+2)
		if stride == 4 {
			m.indices = append(m.indices, a, a+2, a+3)
		}
	}
	c.pending = c.pending[:0]
}
func captureWorld(draw func()) map[uint32]*cpuMesh {
	c := &worldCapture{matrix: identityMatrix(), normal: pos3{0, 1, 0}, color: [4]uint8{255, 255, 255, 255}, texture: app.whiteTexture, meshes: map[uint32]*cpuMesh{}}
	capture = c
	draw()
	c.finish()
	capture = nil
	return c.meshes
}
func uploadWorld(meshes map[uint32]*cpuMesh) []worldBatch {
	var batches []worldBatch
	for texture, m := range meshes {
		if len(m.vertices) == 0 || len(m.indices) == 0 {
			continue
		}
		b := worldBatch{texture: texture, count: len(m.indices)}
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&b.vbo)))
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&b.ebo)))
		if glExtAvailable("glGenVertexArrays") {
			glExt("glGenVertexArrays", 1, uintptr(unsafe.Pointer(&b.vao)))
			glExt("glBindVertexArray", uintptr(b.vao))
		}
		glExt("glBindBuffer", glArrayBuffer, uintptr(b.vbo))
		glExt("glBufferData", glArrayBuffer, uintptr(len(m.vertices))*unsafe.Sizeof(worldVertex{}), uintptr(unsafe.Pointer(&m.vertices[0])), glStaticDraw)
		glExt("glBindBuffer", 0x8893, uintptr(b.ebo))
		glExt("glBufferData", 0x8893, uintptr(len(m.indices)*4), uintptr(unsafe.Pointer(&m.indices[0])), glStaticDraw)
		worldPointers()
		if b.vao != 0 {
			glExt("glBindVertexArray", 0)
		}
		glExt("glBindBuffer", glArrayBuffer, 0)
		glExt("glBindBuffer", 0x8893, 0)
		batches = append(batches, b)
		runtime.KeepAlive(m)
	}
	return batches
}
func worldPointers() {
	stride := unsafe.Sizeof(worldVertex{})
	gl("glEnableClientState", glVertexArray)
	gl("glEnableClientState", glNormalArray)
	gl("glEnableClientState", glColorArray)
	gl("glEnableClientState", 0x8078)
	gl("glVertexPointer", 3, glFloat, stride, 0)
	gl("glNormalPointer", glFloat, stride, 12)
	gl("glTexCoordPointer", 2, glFloat, stride, 24)
	gl("glColorPointer", 4, GL_UNSIGNED_BYTE, stride, 32)
}
func bindWorldBatch(b worldBatch) {
	bindWorldPBR(b.texture)
	if b.vao != 0 {
		glExt("glBindVertexArray", uintptr(b.vao))
	} else {
		glExt("glBindBuffer", glArrayBuffer, uintptr(b.vbo))
		glExt("glBindBuffer", 0x8893, uintptr(b.ebo))
		worldPointers()
	}
	en(GL_TEXTURE_2D)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(b.texture))
}
func unbindWorld() {
	if glExtAvailable("glBindVertexArray") {
		glExt("glBindVertexArray", 0)
	}
	for _, state := range []uintptr{glVertexArray, glNormalArray, glColorArray, 0x8078} {
		gl("glDisableClientState", state)
	}
	glExt("glBindBuffer", glArrayBuffer, 0)
	glExt("glBindBuffer", 0x8893, 0)
}
func drawWorldBatches(batches []worldBatch) {
	for _, b := range batches {
		bindWorldBatch(b)
		gl("glDrawElements", GL_TRIANGLES, uintptr(b.count), 0x1405, 0)
		gpuWorld.draws++
	}
	unbindWorld()
	resetPBRMaterial()
}
func buildBufferedWorld() bool {
	for _, f := range []string{"glGenBuffers", "glBindBuffer", "glBufferData", "glDeleteBuffers"} {
		if !glExtAvailable(f) {
			return false
		}
	}
	releaseBufferedWorld()
	gpuWorld.instancing = glExtAvailable("glDrawElementsInstanced") && glExtAvailable("glVertexAttribDivisor") && glExtAvailable("glVertexAttribPointer") && lightPipeline.program != 0
	if gpuWorld.instancing {
		n := []byte("aInstance\x00")
		gpuWorld.instanceLoc = int32(glExt("glGetAttribLocation", uintptr(lightPipeline.program), uintptr(unsafe.Pointer(&n[0]))))
		gpuWorld.instancing = gpuWorld.instanceLoc >= 0
	}
	gpuWorld.details = uploadWorld(captureWorld(renderStaticDetails))
	gpuWorld.tree = uploadWorld(captureWorld(func() { renderTree(engine.Tree{Size: 2}) }))
	for z := worldMin; z < 1000; z += renderCellSize {
		for x := worldMin; x < 1000; x += renderCellSize {
			tile := worldTile{cx: x + renderCellSize*.5, cz: z + renderCellSize*.5}
			x0, z0 := x, z
			tile.ground = uploadWorld(captureWorld(func() { renderGroundCell(x0, z0) }))
			if math.Abs(tile.cx) < 550 && math.Abs(tile.cz) < 560 {
				tile.grass = uploadWorld(captureWorld(func() { renderGrassCell(x0, z0) }))
				budget := []int{120, 230, 360}[engine.ClampInt(app.quality, 0, 2)]
				for i, t := range app.world.Trees {
					if i >= budget {
						break
					}
					if t.Pos.X >= x && t.Pos.X < x+renderCellSize && t.Pos.Z >= z && t.Pos.Z < z+renderCellSize {
						tile.trees = append(tile.trees, [4]float32{float32(t.Pos.X), float32(engine.Terrain(t.Pos.X, t.Pos.Z)), float32(t.Pos.Z), float32(t.Size)})
					}
				}
				if gpuWorld.instancing && len(tile.trees) > 0 {
					glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&tile.instances)))
					glExt("glBindBuffer", glArrayBuffer, uintptr(tile.instances))
					glExt("glBufferData", glArrayBuffer, uintptr(len(tile.trees)*16), uintptr(unsafe.Pointer(&tile.trees[0])), glStaticDraw)
					glExt("glBindBuffer", glArrayBuffer, 0)
				}
			}
			gpuWorld.tiles = append(gpuWorld.tiles, tile)
		}
	}
	gpuWorld.ready = len(gpuWorld.tiles) > 0
	return gpuWorld.ready
}
func drawBufferedWorld(eye, target pos3) bool {
	if !gpuWorld.ready {
		return false
	}
	gpuWorld.draws = 0
	drawWorldBatches(gpuWorld.details)
	fwd := engine.Vec{X: target.x - eye.x, Z: target.z - eye.z}.Unit()
	for _, t := range gpuWorld.tiles {
		to := engine.Vec{X: t.cx - eye.x, Z: t.cz - eye.z}
		dist := to.Length()
		if dist > 805 || (dist > 235 && to.X*fwd.X+to.Z*fwd.Z < -130) {
			continue
		}
		drawWorldBatches(t.ground)
		if dist > 560 {
			continue
		}
		fade := engine.Clamp((560-dist)/100, 0, 1)
		setWorldFade(fade)
		drawWorldBatches(t.grass)
		if gpuWorld.instancing && lightPipeline.active && t.instances != 0 && dist < 220 && drawPBRTreeInstances(t.instances, len(t.trees)) {
		} else if gpuWorld.instancing && lightPipeline.active && t.instances != 0 {
			setInstancedWorld(true)
			for _, b := range gpuWorld.tree {
				bindWorldBatch(b)
				glExt("glBindBuffer", glArrayBuffer, uintptr(t.instances))
				loc := uintptr(gpuWorld.instanceLoc)
				glExt("glEnableVertexAttribArray", loc)
				glExt("glVertexAttribPointer", loc, 4, glFloat, 0, 16, 0)
				glExt("glVertexAttribDivisor", loc, 1)
				glExt("glDrawElementsInstanced", GL_TRIANGLES, uintptr(b.count), 0x1405, 0, uintptr(len(t.trees)))
				gpuWorld.draws++
				glExt("glVertexAttribDivisor", loc, 0)
				glExt("glDisableVertexAttribArray", loc)
			}
			setInstancedWorld(false)
			unbindWorld()
		} else {
			for _, p := range t.trees {
				push()
				move(float64(p[0]), float64(p[1]), float64(p[2]))
				s := p[3]
				m := [16]float32{s, 0, 0, 0, 0, s, 0, 0, 0, 0, s, 0, 0, 0, 0, 1}
				gl("glMultMatrixf", uintptr(unsafe.Pointer(&m[0])))
				drawWorldBatches(gpuWorld.tree)
				pop()
			}
		}
		setWorldFade(1)
	}
	return true
}
func releaseWorldBatches(batches []worldBatch) {
	for _, b := range batches {
		if b.vao != 0 {
			glExt("glDeleteVertexArrays", 1, uintptr(unsafe.Pointer(&b.vao)))
		}
		if b.vbo != 0 {
			glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&b.vbo)))
		}
		if b.ebo != 0 {
			glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&b.ebo)))
		}
	}
}
func releaseBufferedWorld() {
	releaseWorldBatches(gpuWorld.details)
	releaseWorldBatches(gpuWorld.tree)
	for _, t := range gpuWorld.tiles {
		releaseWorldBatches(t.ground)
		releaseWorldBatches(t.grass)
		if t.instances != 0 {
			glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&t.instances)))
		}
	}
	gpuWorld.tiles = nil
	gpuWorld.details = nil
	gpuWorld.tree = nil
	gpuWorld.ready = false
}

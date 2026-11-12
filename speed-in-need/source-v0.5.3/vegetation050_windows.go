package main

import (
	"encoding/binary"
	"math"
	"runtime"
	"speedinneed/engine"
	"unsafe"
)

// Amplification is spatially instanced, not independent GameObjects. All copies
// stay within an existing validated patch; the shader conforms roots to ground.
func denseGrass050(raw []byte, count int) []float32 {
	return denseGrass051(raw, count, 8)
}
func denseGrass051(raw []byte, count, copies int) []float32 {
	out := make([]float32, count*copies*8)
	for i := 0; i < count; i++ {
		var p [8]float32
		for j := 0; j < 8; j++ {
			p[j] = math.Float32frombits(binary.LittleEndian.Uint32(raw[(i*8+j)*4:]))
		}
		for k := 0; k < copies; k++ {
			q := p
			seed := float64(i*17 + k*113)
			if k > 0 {
				q[0] += float32(math.Sin(seed*1.71) * .62)
				q[2] += float32(math.Cos(seed*2.13) * .62)
				q[4] += float32(seed)
			}
			q[5] *= 1.22
			q[7] *= 1.22
			q[6] *= .82 // Half the 0.5.0 height, keeping the same root coverage.
			copy(out[(k*count+i)*8:], q[:])
		}
	}
	return out
}

// Smaller draw cells keep the expensive dense mesh local to the camera.
// Far cells use the sparse mesh and one copy per original patch.
func splitGrass050(raw []byte, count int, parent remasterInstance) []remasterInstance {
	var cells [16][]byte
	for i := 0; i < count; i++ {
		x := float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[i*32:])))
		z := float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[i*32+8:])))
		col := int(math.Floor((x - (parent.X - 24)) / 12))
		row := int(math.Floor((z - (parent.Z - 24)) / 12))
		if col < 0 {
			col = 0
		}
		if col > 3 {
			col = 3
		}
		if row < 0 {
			row = 0
		}
		if row > 3 {
			row = 3
		}
		cells[row*4+col] = append(cells[row*4+col], raw[i*32:(i+1)*32]...)
	}
	out := make([]remasterInstance, 0, 16)
	for i, data := range cells {
		if len(data) == 0 {
			continue
		}
		g := parent
		g.X = parent.X - 18 + float64(i%4)*12
		g.Z = parent.Z - 18 + float64(i/4)*12
		g.Radius = 10
		g.GrassSource051 = data
		g.Copies051 = grassCopies051()
		dense := denseGrass051(data, len(data)/32, g.Copies051)
		g.Count = len(dense) / 8
		g.Packed052 = dense
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&g.VBO)))
		glExt("glBindBuffer", glArrayBuffer, uintptr(g.VBO))
		glExt("glBufferData", glArrayBuffer, uintptr(len(dense)*4), uintptr(unsafe.Pointer(&dense[0])), glStaticDraw)
		runtime.KeepAlive(dense)
		out = append(out, g)
	}
	return out
}

func grassCopies051() int {
	return []int{1, 2, 3, 3}[engine.ClampInt(app.quality, 0, 3)]
}

// The low preset no longer allocates eight dense copies and draws only 1/32
// of them. Rebuild only on a preset change, retaining compact source instances.
var grassUploadedCopies051 int

func refreshGrass051() {
	copies := grassCopies051()
	if copies == grassUploadedCopies051 {
		return
	}
	grassUploadedCopies051 = copies
	for i := range remaster.Instances {
		g := &remaster.Instances[i]
		if len(g.GrassSource051) == 0 || g.Copies051 == copies {
			continue
		}
		dense := denseGrass051(g.GrassSource051, len(g.GrassSource051)/32, copies)
		g.Copies051, g.Count = copies, len(dense)/8
		g.Packed052 = dense
		glExt("glBindBuffer", glArrayBuffer, uintptr(g.VBO))
		glExt("glBufferData", glArrayBuffer, uintptr(len(dense)*4), uintptr(unsafe.Pointer(&dense[0])), glStaticDraw)
		runtime.KeepAlive(dense)
	}
	glExt("glBindBuffer", glArrayBuffer, 0)
}

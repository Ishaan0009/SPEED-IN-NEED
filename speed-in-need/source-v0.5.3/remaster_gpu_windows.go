//go:build windows

package main

import (
	"encoding/binary"
	"math"
	"runtime"
	"unsafe"
)

// Existing S3D models are uploaded once as packed interleaved GPU arrays.
// This replaces ~1.6M immediate-mode driver calls for the two player bodies
// and enables material-independent garage paint in a single extra draw.
// Real glTF image/normal maps require the separate asset migration pass.
const (
	glArrayBuffer = 0x8892
	glStaticDraw  = 0x88E4
	glVertexArray = 0x8074
	glNormalArray = 0x8075
	glColorArray  = 0x8076
	glFloat       = 0x1406
)

type bufferVertex struct {
	P [3]float32
	N [3]float32
	C [4]uint8
}
type bufferBatch struct {
	handle uint32
	count  int
}
type bufferCar struct {
	solid, paint bufferBatch
	available    bool
}

var bufferedCars [2]bufferCar
var bufferedWheels [2][4]bufferBatch

func initBufferedCars() {
	for _, name := range []string{"glGenBuffers", "glBindBuffer", "glBufferData", "glDeleteBuffers"} {
		if !glExtAvailable(name) {
			return
		}
	}
	for i := 0; i < 2; i++ {
		raw, err := assets.ReadFile("models/" + modelFile(i))
		if err != nil || len(raw) < 8 || string(raw[:4]) != "S3D1" {
			continue
		}
		count := int(binary.LittleEndian.Uint32(raw[4:8]))
		if count <= 0 || count > (len(raw)-8)/16 {
			continue
		}
		opaque := make([]bufferVertex, 0, count*2/3)
		paint := make([]bufferVertex, 0, count/3)
		for v := 0; v+2 < count; v += 3 {
			var xyz [3][3]float64
			var colors [3][4]uint8
			for j := 0; j < 3; j++ {
				off := 8 + (v+j)*16
				for axis := 0; axis < 3; axis++ {
					xyz[j][axis] = float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[off+axis*4:])))
				}
				copy(colors[j][:], raw[off+12:off+16])
				colors[j][3] = 255
			}
			a, b, c := xyz[0], xyz[1], xyz[2]
			ux, uy, uz := b[0]-a[0], b[1]-a[1], b[2]-a[2]
			vx, vy, vz := c[0]-a[0], c[1]-a[1], c[2]-a[2]
			nx, ny, nz := uy*vz-uz*vy, uz*vx-ux*vz, ux*vy-uy*vx
			norm := math.Sqrt(nx*nx + ny*ny + nz*nz)
			if norm < 1e-9 || math.IsNaN(norm) || math.IsInf(norm, 0) {
				continue
			}
			// Additional asset validation prevents malformed oversized faces.
			edgeAB := math.Sqrt(ux*ux + uy*uy + uz*uz)
			edgeAC := math.Sqrt(vx*vx + vy*vy + vz*vz)
			if edgeAB > 17 || edgeAC > 17 {
				continue
			}
			n := [3]float32{float32(nx / norm), float32(ny / norm), float32(nz / norm)}
			body := raw[8+v*16+15] == 255
			for j := 0; j < 3; j++ {
				vertex := bufferVertex{P: [3]float32{float32(xyz[j][0]), float32(xyz[j][1]), float32(xyz[j][2])}, N: n, C: colors[j]}
				if body {
					vertex.C = [4]uint8{255, 255, 255, 255}
					paint = append(paint, vertex)
				} else {
					opaque = append(opaque, vertex)
				}
			}
		}
		bufferedCars[i] = bufferCar{solid: uploadBufferedBatch(opaque), paint: uploadBufferedBatch(paint)}
		bufferedCars[i].available = bufferedCars[i].solid.handle != 0 && bufferedCars[i].paint.handle != 0
	}
}

// initBufferedWheels uploads centred tire submeshes once. Their vertices and
// normals remain resident on the GPU, avoiding repeated glBegin/glVertex driver
// submissions while steering and rotating the wheel with the modelview matrix.
func initBufferedWheels() {
	for _, name := range []string{"glGenBuffers", "glBindBuffer", "glBufferData", "glDeleteBuffers"} {
		if !glExtAvailable(name) {
			return
		}
	}
	for car, model := range []string{"challenge296", "porsche911"} {
		for slot, key := range []string{"FL", "FR", "RL", "RR"} {
			raw, err := assets.ReadFile("models/" + model + "_wheel_" + key + ".s3d")
			if err != nil {
				continue
			}
			pivot, ok := s3dWheelPivot(raw)
			if !ok {
				continue
			}
			count := int(binary.LittleEndian.Uint32(raw[4:8]))
			vertices := make([]bufferVertex, 0, count)
			for i := 0; i+2 < count; i += 3 {
				var xyz [3][3]float64
				for j := 0; j < 3; j++ {
					off := 8 + (i+j)*16
					for axis := 0; axis < 3; axis++ {
						xyz[j][axis] = float64(math.Float32frombits(binary.LittleEndian.Uint32(raw[off+4*axis:]))) - pivot[axis]
					}
				}
				a, b, c := xyz[0], xyz[1], xyz[2]
				ux, uy, uz := b[0]-a[0], b[1]-a[1], b[2]-a[2]
				vx, vy, vz := c[0]-a[0], c[1]-a[1], c[2]-a[2]
				nx, ny, nz := uy*vz-uz*vy, uz*vx-ux*vz, ux*vy-uy*vx
				norm := math.Sqrt(nx*nx + ny*ny + nz*nz)
				if norm < 1e-8 || math.IsNaN(norm) {
					continue
				}
				n := [3]float32{float32(nx / norm), float32(ny / norm), float32(nz / norm)}
				for j := 0; j < 3; j++ {
					off := 8 + (i+j)*16
					vertices = append(vertices, bufferVertex{
						P: [3]float32{float32(xyz[j][0]), float32(xyz[j][1]), float32(xyz[j][2])}, N: n,
						C: [4]uint8{raw[off+12], raw[off+13], raw[off+14], 255},
					})
				}
			}
			bufferedWheels[car][slot] = uploadBufferedBatch(vertices)
		}
	}
}

func drawBufferedWheel(car, slot int) bool {
	if car < 0 || car >= len(bufferedWheels) || slot < 0 || slot >= 4 {
		return false
	}
	batch := bufferedWheels[car][slot]
	if batch.handle == 0 {
		return false
	}
	drawBufferedBatch(batch, false)
	return true
}

func uploadBufferedBatch(vertices []bufferVertex) bufferBatch {
	if len(vertices) == 0 {
		return bufferBatch{}
	}
	var buffer uint32
	glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&buffer)))
	if buffer == 0 {
		return bufferBatch{}
	}
	glExt("glBindBuffer", glArrayBuffer, uintptr(buffer))
	glExt("glBufferData", glArrayBuffer, uintptr(len(vertices)*int(unsafe.Sizeof(vertices[0]))), uintptr(unsafe.Pointer(&vertices[0])), glStaticDraw)
	glExt("glBindBuffer", glArrayBuffer, 0)
	runtime.KeepAlive(vertices)
	return bufferBatch{handle: buffer, count: len(vertices)}
}

func drawBufferedBatch(batch bufferBatch, paint bool) {
	if batch.handle == 0 || batch.count == 0 {
		return
	}
	if paint {
		setPBRMaterial(true, .72, .24)
	} else {
		setPBRMaterial(true, .16, .57)
	}
	glExt("glBindBuffer", glArrayBuffer, uintptr(batch.handle))
	gl("glEnableClientState", glVertexArray)
	gl("glEnableClientState", glNormalArray)
	gl("glVertexPointer", 3, glFloat, uintptr(unsafe.Sizeof(bufferVertex{})), 0)
	gl("glNormalPointer", glFloat, uintptr(unsafe.Sizeof(bufferVertex{})), 12)
	if paint {
		gl("glDisableClientState", glColorArray)
		color := paintPalette[beta.paint]
		col(color[0], color[1], color[2], 1)
	} else {
		gl("glEnableClientState", glColorArray)
		gl("glColorPointer", 4, GL_UNSIGNED_BYTE, uintptr(unsafe.Sizeof(bufferVertex{})), 24)
	}
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(app.whiteTexture))
	gl("glDrawArrays", GL_TRIANGLES, 0, uintptr(batch.count))
	gl("glDisableClientState", glColorArray)
	gl("glDisableClientState", glNormalArray)
	gl("glDisableClientState", glVertexArray)
	glExt("glBindBuffer", glArrayBuffer, 0)
	col(1, 1, 1, 1)
}

func drawBufferedCar(id int) bool {
	if id < 0 || id >= len(bufferedCars) || !bufferedCars[id].available {
		return false
	}
	c := bufferedCars[id]
	drawBufferedBatch(c.solid, false)
	drawBufferedBatch(c.paint, true)
	return true
}
func releaseBufferedCars() {
	for i := range bufferedCars {
		for _, b := range []bufferBatch{bufferedCars[i].solid, bufferedCars[i].paint} {
			if b.handle != 0 {
				glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&b.handle)))
			}
		}
		bufferedCars[i] = bufferCar{}
	}
	for i := range bufferedWheels {
		for j := range bufferedWheels[i] {
			b := bufferedWheels[i][j]
			if b.handle != 0 {
				glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&b.handle)))
			}
			bufferedWheels[i][j] = bufferBatch{}
		}
	}
}

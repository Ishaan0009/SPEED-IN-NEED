package main

import (
	"strings"
	"unsafe"
)

type grassCell053 struct{ X, Z float64 }
type grassRange053 struct{ Start, Count uint32 }
type grassIndirect053 struct {
	VBO, Commands uint32
	Copies, Epoch int
	Ranges        map[grassCell053]grassRange053
	Draws         [][5]uint32
}

var grassEpoch053 int
var indirectGrass053 = map[string]*grassIndirect053{}

func indirectGrassAvailable053() bool { return glExtAvailable("glMultiDrawElementsIndirect") }
func drawGrassIndirect053(groups []remasterInstance, p pbrPart, loc, yaw, scaleLoc uintptr) bool {
	if len(groups) == 0 || groups[0].Kind != "grass" || !indirectGrassAvailable053() {
		return false
	}
	key := strings.TrimSuffix(strings.TrimSuffix(groups[0].Key, "_far_lod"), "_lod")
	if key != "b035_grass0" && key != "b035_grass1" {
		return false
	}
	b := indirectGrass053[key]
	if b == nil {
		b = &grassIndirect053{}
		indirectGrass053[key] = b
	}
	if b.VBO == 0 || b.Copies != grassCopies051() || b.Epoch != grassEpoch053 {
		b.Ranges = map[grassCell053]grassRange053{}
		var data []float32
		for _, g := range remaster.Instances {
			if g.Key != key {
				continue
			}
			start := len(data) / 8
			data = append(data, g.Packed052...)
			b.Ranges[grassCell053{g.X, g.Z}] = grassRange053{uint32(start), uint32(len(g.Packed052) / 8)}
		}
		if len(data) == 0 {
			return false
		}
		if b.VBO == 0 {
			glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&b.VBO)))
			glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&b.Commands)))
		}
		glExt("glBindBuffer", glArrayBuffer, uintptr(b.VBO))
		glExt("glBufferData", glArrayBuffer, uintptr(len(data)*4), uintptr(unsafe.Pointer(&data[0])), glStaticDraw)
		b.Copies = grassCopies051()
		b.Epoch = grassEpoch053
	}
	b.Draws = b.Draws[:0]
	for _, g := range groups {
		r, ok := b.Ranges[grassCell053{g.X, g.Z}]
		if !ok {
			return false
		}
		count := uint32(beta22InstanceCount(g))
		if count > r.Count {
			count = r.Count
		}
		if count > 0 {
			b.Draws = append(b.Draws, [5]uint32{uint32(p.Indices), count, 0, 0, r.Start})
		}
	}
	if len(b.Draws) == 0 {
		return true
	}
	setWorldFade(1)
	sceneInt("uLODSide052", 0)
	glExt("glBindBuffer", glArrayBuffer, uintptr(b.VBO))
	glExt("glVertexAttribPointer", loc, 4, glFloat, 0, 32, 0)
	glExt("glVertexAttribPointer", yaw, 1, glFloat, 0, 32, 16)
	glExt("glVertexAttribPointer", scaleLoc, 3, glFloat, 0, 32, 20)
	glExt("glBindBuffer", 0x8F3F, uintptr(b.Commands))
	glExt("glBufferData", 0x8F3F, uintptr(len(b.Draws)*20), uintptr(unsafe.Pointer(&b.Draws[0])), 0x88E0)
	glExt("glMultiDrawElementsIndirect", GL_TRIANGLES, 0x1405, 0, uintptr(len(b.Draws)), 20)
	glExt("glBindBuffer", 0x8F3F, 0)
	gpuWorld.draws++
	return true
}

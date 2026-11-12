package main

import "unsafe"

type grassBatch052 struct {
	VBO  uint32
	Data []float32
}

var grassBatches052 = map[string]*grassBatch052{}

// Culling retains the original spatial cells. Submit their already-validated
// instance records as one stream per grass type, instead of thousands of small
// driver calls. No resampling or new roots appear when the camera advances.
func batchGrass052(groups []remasterInstance) []remasterInstance {
	key := groups[0].Key
	if indirectGrassAvailable053() {
		return groups
	}
	if key != "b035_grass0" && key != "b035_grass1" && key != "b035_grass0_lod" && key != "b035_grass1_lod" {
		return groups
	}
	batch := grassBatches052[key]
	if batch == nil {
		batch = &grassBatch052{}
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&batch.VBO)))
		grassBatches052[key] = batch
	}
	batch.Data = batch.Data[:0]
	for _, g := range groups {
		count := beta22InstanceCount(g)
		if len(g.Packed052) < count*8 {
			return groups
		}
		batch.Data = append(batch.Data, g.Packed052[:count*8]...)
	}
	if len(batch.Data) == 0 {
		return nil
	}
	glExt("glBindBuffer", glArrayBuffer, uintptr(batch.VBO))
	glExt("glBufferData", glArrayBuffer, uintptr(len(batch.Data)*4), uintptr(unsafe.Pointer(&batch.Data[0])), 0x88E0)
	g := groups[0]
	g.VBO = batch.VBO
	g.Count = len(batch.Data) / 8
	g.Fade = 1
	g.Merged052 = true
	return []remasterInstance{g}
}

package main

import (
	"runtime"
	"unsafe"
)

type buildingKey026 struct {
	Asset      string
	PassRadius float64
}
type buildingCell026 struct {
	VBO   uint32
	Count int
}
type buildingBatch026 struct {
	VBO    uint32
	Cells  []buildingCell026
	Output [1]remasterInstance
}

var buildingCache026 = map[buildingKey026]*buildingBatch026{}
var buildingScratch026 [][8]float32
var buildingPass026 float64
var buildingStats026 struct{ Hits, Uploads, Bytes uint64 }

func sameBuildingCells026(cells []buildingCell026, groups []remasterInstance) bool {
	if len(cells) != len(groups) {
		return false
	}
	for i, g := range groups {
		if cells[i].VBO != g.VBO || cells[i].Count != g.Count {
			return false
		}
	}
	return true
}

func releaseBuildingCache026() {
	for _, entry := range buildingCache026 {
		if entry.VBO != 0 {
			glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(&entry.VBO)))
		}
	}
	buildingCache026 = map[buildingKey026]*buildingBatch026{}
	buildingScratch026 = nil
}

// City blocks retain 100 m CPU culling, but submit all visible instances of one
// building family together. Immutable cells are cached separately for the
// main view and each shadow radius; unchanged views perform no CPU copy/upload.
func beta012MergeBuildings(groups []remasterInstance) []remasterInstance {
	if len(groups) < 2 || groups[0].Kind != "building" {
		return groups
	}
	key := buildingKey026{groups[0].Key, buildingPass026}
	entry := buildingCache026[key]
	if entry == nil {
		entry = &buildingBatch026{}
		buildingCache026[key] = entry
	}
	if sameBuildingCells026(entry.Cells, groups) {
		buildingStats026.Hits++
		return entry.Output[:]
	}
	buildingScratch026 = buildingScratch026[:0]
	entry.Cells = entry.Cells[:0]
	for _, cell := range groups {
		buildingScratch026 = append(buildingScratch026, cell.Data...)
		entry.Cells = append(entry.Cells, buildingCell026{cell.VBO, cell.Count})
	}
	if len(buildingScratch026) == 0 {
		entry.Cells = nil
		return groups
	}
	if entry.VBO == 0 {
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&entry.VBO)))
	}
	glExt("glBindBuffer", glArrayBuffer, uintptr(entry.VBO))
	glExt("glBufferData", glArrayBuffer, uintptr(len(buildingScratch026)*32), uintptr(unsafe.Pointer(&buildingScratch026[0])), glStaticDraw)
	runtime.KeepAlive(buildingScratch026)
	buildingStats026.Uploads++
	buildingStats026.Bytes += uint64(len(buildingScratch026) * 32)
	g := groups[0]
	g.VBO = entry.VBO
	g.Count = len(buildingScratch026)
	g.Fade = 1
	entry.Output[0] = g
	return entry.Output[:]
}

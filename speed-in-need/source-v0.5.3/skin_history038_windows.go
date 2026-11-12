package main

import "unsafe"

type skinHistoryFrame038 struct {
	Current, Previous [][3]float32
	Frame             uint64
}

var skinHistory038 = map[string]*skinHistoryFrame038{}
var skinEntity038 string
var skinSamples038 uint64

func previousSkin038(key string, positions [][3]float32) [][3]float32 {
	h := skinHistory038[key]
	if h == nil {
		h = &skinHistoryFrame038{}
		skinHistory038[key] = h
	}
	if h.Frame == temporal038.Frame {
		return h.Previous
	}
	previousFrame := h.Frame
	h.Previous, h.Current = h.Current, h.Previous
	if cap(h.Current) < len(positions) {
		h.Current = make([][3]float32, len(positions))
	} else {
		h.Current = h.Current[:len(positions)]
	}
	copy(h.Current, positions)
	if previousFrame+1 != temporal038.Frame || len(h.Previous) != len(positions) {
		h.Previous = append(h.Previous[:0], positions...)
	}
	h.Frame = temporal038.Frame
	skinSamples038++
	return h.Previous
}
func bindPreviousSkin038(buffer *uint32, previous [][3]float32) {
	if hdrScene.motion == 0 || !hdrScene.active || sunShadow.drawing || len(previous) == 0 {
		return
	}
	if *buffer == 0 {
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(buffer)))
	}
	glExt("glBindBuffer", glArrayBuffer, uintptr(*buffer))
	glExt("glBufferData", glArrayBuffer, uintptr(len(previous))*12, uintptr(unsafe.Pointer(&previous[0])), 0x88E0)
	glExt("glEnableVertexAttribArray", 11)
	glExt("glVertexAttribPointer", 11, 3, glFloat, 0, 12, 0)
	sceneInt("uPreviousVertex038", 1)
}
func unbindPreviousSkin038() {
	sceneInt("uPreviousVertex038", 0)
	sceneInt("uObjectMotion038", 0)
	glExt("glDisableVertexAttribArray", 11)
}

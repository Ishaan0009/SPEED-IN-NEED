package main

var applicationFrame038 uint32

func latencyBegin038() {
	applicationFrame038++
	if outputFBO038 != 0 {
		nativeCall038("SIN038_FrameBegin", uintptr(applicationFrame038))
	}
}
func latencyMarker038(marker int) {
	if outputFBO038 != 0 {
		nativeCall038("SIN038_FrameMarker", uintptr(marker))
	}
}

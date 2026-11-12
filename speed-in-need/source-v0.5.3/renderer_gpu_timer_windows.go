//go:build windows

package main

import "unsafe"

// GL_TIME_ELAPSED measures actual GPU execution rather than CPU draw-submit
// time. Two query objects are polled without blocking the render thread.
const (
	glTimeElapsed          = 0x88BF
	glQueryResultAvailable = 0x8867
	glQueryResult          = 0x8866
)

type gpuTimerState struct {
	ids       [2]uint32
	submitted [2]bool
	next      int
	enabled   bool
	active    bool
}

var gpuTimer gpuTimerState

func initGPUTiming() {
	for _, name := range []string{"glGenQueries", "glBeginQuery", "glEndQuery", "glGetQueryObjectuiv", "glGetQueryObjectui64v", "glDeleteQueries"} {
		if !glExtAvailable(name) {
			return
		}
	}
	glExt("glGenQueries", 2, uintptr(unsafe.Pointer(&gpuTimer.ids[0])))
	gpuTimer.enabled = gpuTimer.ids[0] != 0 && gpuTimer.ids[1] != 0
}
func beginGPUTiming() {
	if !gpuTimer.enabled {
		return
	}
	// Only reuse the slot when the last submitted query has completed.
	idx := gpuTimer.next
	if gpuTimer.submitted[idx] {
		var ready uint32
		glExt("glGetQueryObjectuiv", uintptr(gpuTimer.ids[idx]), glQueryResultAvailable, uintptr(unsafe.Pointer(&ready)))
		if ready == 0 {
			return
		}
		var nanoseconds uint64
		glExt("glGetQueryObjectui64v", uintptr(gpuTimer.ids[idx]), glQueryResult, uintptr(unsafe.Pointer(&nanoseconds)))
		app.gpuMS = float64(nanoseconds) / 1_000_000
		app.gpuTiming = true
		gpuTimer.submitted[idx] = false
	}
	glExt("glBeginQuery", glTimeElapsed, uintptr(gpuTimer.ids[idx]))
	gpuTimer.active = true
}
func endGPUTiming() {
	if !gpuTimer.enabled || !gpuTimer.active {
		return
	}
	glExt("glEndQuery", glTimeElapsed)
	gpuTimer.submitted[gpuTimer.next] = true
	gpuTimer.next = (gpuTimer.next + 1) % len(gpuTimer.ids)
	gpuTimer.active = false
}
func shutdownGPUTiming() {
	if gpuTimer.enabled {
		glExt("glDeleteQueries", 2, uintptr(unsafe.Pointer(&gpuTimer.ids[0])))
	}
	gpuTimer = gpuTimerState{}
}

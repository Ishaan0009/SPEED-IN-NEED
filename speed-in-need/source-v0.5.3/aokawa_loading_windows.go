//go:build windows

package main

import "unsafe"

func drawAokawaLoading(progress float64, label string) {
	if qa.enabled || app.rc == 0 {
		return
	}
	useLighting(false)
	drawLoading036(progress, label)
	en(GL_DEPTH_TEST)
	presentFrame()
	// Keep window/cursor/close messages responsive; loading guards all menu mutations.
	var m msg
	for i := 0; i < 64 && win(user32, "PeekMessageW", uintptr(unsafe.Pointer(&m)), 0, 0, 0, PM_REMOVE) != 0; i++ {
		if m.message == 0x12 {
			app.run = false
			return
		}
		win(user32, "TranslateMessage", uintptr(unsafe.Pointer(&m)))
		win(user32, "DispatchMessageW", uintptr(unsafe.Pointer(&m)))
	}
}

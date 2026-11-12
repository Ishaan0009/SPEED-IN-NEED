package main

import (
	"math"
	"speedinneed/engine"
	"time"
	"unsafe"
)

var mini22Span = 430.0
var mini22Last time.Time

func beta013MiniSpan() float64 {
	now := time.Now()
	dt := math.Min(.05, now.Sub(mini22Last).Seconds())
	mini22Last = now
	braking := app.world != nil && (app.world.Car.Braking || app.world.Car.DriftActive)
	mini22Span += (engine.BrakeMapSpan(braking) - mini22Span) * (1 - math.Exp(-dt*5))
	return mini22Span
}

// Recenter while driving so free look continues past a window edge. Menus,
// focus loss and the map do not warp the pointer. No cursor motion steers.
func beta013CaptureLook() {
	if input020.Raw || qa.enabled || !beta020DrivingPointer() {
		return
	}
	type point struct{ X, Y int32 }
	p := point{int32(app.width / 2), int32(app.height / 2)}
	app.lastX, app.lastY = p.X, p.Y
	app.mouseInit = true
	win(user32, "ClientToScreen", app.hwnd, uintptr(unsafe.Pointer(&p)))
	win(user32, "SetCursorPos", uintptr(p.X), uintptr(p.Y))
}

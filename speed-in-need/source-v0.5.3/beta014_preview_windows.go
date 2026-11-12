package main

import (
	"math"
	"speedinneed/engine"
)

var preview014 struct {
	ready       bool
	screen, car int
	eye, target pos3
}

// The fitted camera follows a continuous automatic orbit. Mouse input belongs
// exclusively to the menu; camera fitting is damped to remove distance steps.
func beta014PreviewCamera(dt float64) (pos3, pos3) {
	angle := 5.6941366846 + math.Sin(beta.menuTime*.12)*.045
	pitch := 0.0
	if garageView037() && garage037.Touched {
		// Rotate the car for inspection; keep the camera facing the scan wall.
		pitch = engine.Clamp(garage037.Pitch, -.04, .10)
	}
	eye, target := beta013PreviewCamera(angle, pitch, float64(app.width)/float64(app.height))
	if !preview014.ready || preview014.screen != beta.screen || preview014.car != beta.carSelect {
		preview014.ready, preview014.screen, preview014.car = true, beta.screen, beta.carSelect
		preview014.eye, preview014.target = eye, target
	}
	k := 1 - math.Exp(-engine.Clamp(dt, 0, .1)*10)
	if garageView037() && garage037.Touched {
		k = 1
	}
	preview014.eye = add(preview014.eye, scale(sub(eye, preview014.eye), k))
	preview014.target = add(preview014.target, scale(sub(target, preview014.target), k))
	return garageCamera037(preview014.eye, preview014.target)
}

func beta014MenuPointer(hwnd uintptr, message uint32, x, y int) bool {
	if garagePointer037(message, x, y) {
		return true
	}
	switch message {
	case WM_LBUTTONDOWN:
		remaster.Drag, remaster.Dragged = true, false
		remaster.DownX, remaster.DownY = x, y
		return true
	case WM_MOUSEMOVE:
		remaster.MouseX, remaster.MouseY = x, y
		beta020Hover(x, y)
		if remaster.Drag && math.Abs(float64(x-remaster.DownX))+math.Abs(float64(y-remaster.DownY)) > 8 {
			remaster.Dragged = true
		}
		return true
	case 0x202:
		click := remaster.Drag && (!remaster.Dragged || buttonAt033(remaster.DownX, remaster.DownY) != "" && buttonAt033(remaster.DownX, remaster.DownY) == buttonAt033(x, y))
		remaster.Drag = false
		if click {
			remasterClick(x, y)
		}
		return true
	case WM_RBUTTONDOWN, WM_RBUTTONUP:
		return true
	}
	return false
}

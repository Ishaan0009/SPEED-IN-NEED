package main

import "testing"

func TestGarage037ViewportOrbitAndZoom(t *testing.T) {
	oldG, oldB, oldUI, oldW, oldH := garage037, beta, ui020, app.width, app.height
	defer func() { garage037 = oldG; beta = oldB; ui020 = oldUI; app.width = oldW; app.height = oldH }()
	app.width, app.height = 1920, 1080
	beta.screen = screenMaps
	ui020.Stage = 2
	ui020.Buttons = nil
	if !garagePointer037(WM_LBUTTONDOWN, 1200, 400) || !garagePointer037(WM_MOUSEMOVE, 1300, 450) || garage037.Yaw == 0 || garage037.Pitch == 0 {
		t.Fatal("garage viewport drag failed")
	}
	garagePointer037(0x202, 1300, 450)
	if garage037.Drag {
		t.Fatal("garage drag stuck")
	}
	garageZoom037(-100)
	if garage037.Zoom != .72 {
		t.Fatal("garage near limit")
	}
	garageZoom037(100)
	if garage037.Zoom != 1.42 {
		t.Fatal("garage far limit")
	}
	beta.screen = screenTitle
	if garagePointer037(WM_LBUTTONDOWN, 1200, 400) {
		t.Fatal("title menu captured garage input")
	}
}

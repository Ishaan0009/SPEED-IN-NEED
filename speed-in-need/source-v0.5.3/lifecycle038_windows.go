package main

import (
	"fmt"
	"os"
)

var lifecycle038 struct{ Frame, Width, Height int }

func qaLifecycle038() {
	if !qa.enabled || os.Getenv("SPEED_QA_LIFECYCLE") != "1" || beta.screen != screenDrive {
		return
	}
	q := &lifecycle038
	q.Frame++
	resize := func(w, h int) { win(user32, "SetWindowPos", app.hwnd, 0, 0, 0, uintptr(w), uintptr(h), 0x16) }
	switch q.Frame {
	case 1:
		q.Width, q.Height = app.width, app.height
	case 40:
		resize(1280, 720)
	case 60:
		qa.checks["038-live-resize-down"] = app.width < q.Width && outputFBO038 != 0
	case 90:
		resize(1600, 900)
	case 120:
		qa.checks["038-live-resize-up"] = app.width > 1200 && outputFBO038 != 0
	case 140:
		resize(q.Width, q.Height)
	case 190:
		toggleFullscreen()
	case 220:
		qa.checks["038-fullscreen-entry"] = outputFBO038 != 0
	case 240:
		toggleFullscreen()
	case 270:
		qa.checks["038-fullscreen-return"] = outputFBO038 != 0
	case 300, 340, 380, 420, 460:
		mode := map[int]string{300: "balanced", 340: "performance", 380: "dlaa", 420: "off", 460: "quality"}[q.Frame]
		os.Setenv("SPEED_QA_DLSS_MODE", mode)
		destroyHDR()
	case 330:
		qa.checks["038-balanced-transition"] = temporalMode038() == 2 && reconstructedReady038
	case 370:
		qa.checks["038-performance-transition"] = temporalMode038() == 1 && reconstructedReady038
	case 410:
		qa.checks["038-dlaa-transition"] = temporalMode038() == 6 && reconstructedReady038
	case 450:
		qa.checks["038-native-transition"] = temporalMode038() == 0 && !reconstructedReady038
	case 490:
		qa.checks["038-quality-return"] = temporalMode038() == 3 && reconstructedReady038
	case 500:
		os.Setenv("SPEED_QA_FSR_FG", "")
	case 525:
		f := fsrReport038()
		qa.checks["038-frame-generation-off"] = f["active"] == false
	case 550:
		os.Setenv("SPEED_QA_FSR_FG", "1")
	case 620:
		f := fsrReport038()
		qa.checks["038-frame-generation-resumed"] = f["error"] == "" && f["generated_frames"].(float64) > 0
		qa.checks["038-lifecycle-completed"] = temporalTransport038.Error == "" && rendererError038 == ""
	case 650:
		if os.Getenv("SPEED_QA_FORCE_FALLBACK") == "1" {
			fallback038(fmt.Errorf("QA-injected presentation failure"))
		}
	case 670:
		if os.Getenv("SPEED_QA_FORCE_FALLBACK") == "1" {
			qa.checks["038-fallback-draws-again"] = backend038.Name() == "OPENGL" && !hdrScene.failed
		}
	}
}

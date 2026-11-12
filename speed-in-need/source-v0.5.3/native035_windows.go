package main

import "os"

func applyHigh035() {
	applyRTX3060Profile034()
	app.quality = 2
	beta.cfg.Quality = 2
	beta.cfg.VSync = false
	beta.cfg.FrameCap = 0
	beta.cfg.LightingDetail = 1
	beta.cfg.MeshDetail = 2
	beta.cfg.ViewDistance = 1.15
	beta.cfg.TextureBudget = 2
	beta.cfg.AdaptiveQuality = false
}
func applyUltra035() {
	applyHigh035()
	app.quality = 3
	beta.cfg.Quality = 3
	beta.cfg.Shadows = 2
	beta.cfg.ShadowFilter = 2
	beta.cfg.LightingDetail = 2
	beta.cfg.MeshDetail = 3
	beta.cfg.TextureBudget = 3
	beta.cfg.ViewDistance = 1.35
	beta.cfg.Vegetation = 1.2
	beta.cfg.FoliageDensity = 1
}
func frameCapLabel035(fps int) string {
	if fps == 0 {
		return "UNCAPPED"
	}
	return fmtSpeed021(float64(fps)) + " FPS"
}
func applyIntegrated035() {
	applyAokawaPreset(0)
	beta.cfg.Lightweight = true
	beta.cfg.RenderScale = .6
	beta.cfg.Resolution = 0
	beta.cfg.TextureBudget = 0
	beta.cfg.TrafficDensity = 1
	beta.cfg.VSync = false
	beta.cfg.FrameCap = 0
	beta.cfg.AdaptiveQuality = true
}
func qaProfile035() {
	switch os.Getenv("SPEED_QA_035_PROFILE") {
	case "high":
		applyHigh035()
	case "ultra":
		applyUltra035()
	case "integrated":
		applyIntegrated035()
	}
}

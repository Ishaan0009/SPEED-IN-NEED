package main

import (
	"math"
	"speedinneed/engine"
)

func beta22InstanceCount(g remasterInstance) int {
	if g.Merged052 {
		return g.Count
	}
	if g.Count <= 0 {
		return 0
	}
	if g.Kind != "grass" && g.Kind != "bush" {
		return g.Count
	}
	density := beta.cfg.FoliageDensity
	if g.Kind == "grass" {
		copies := float64([]int{1, 2, 4, 8}[engine.ClampInt(app.quality, 0, 3)])

		uploaded := g.Copies051
		if uploaded <= 0 {
			uploaded = 8
		}
		density *= math.Min(copies, float64(uploaded)) / float64(uploaded)
	}
	if g.Kind == "bush" {
		density = math.Max(.5, density)
	}
	minimum := .025
	if beta.cfg.Lightweight {
		minimum = .0125
	}
	density = engine.Clamp(density*foliageFactor026(), minimum, 1)
	return engine.ClampInt(int(float64(g.Count)*density), 1, g.Count)
}

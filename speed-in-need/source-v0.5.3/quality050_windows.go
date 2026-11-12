package main

import "speedinneed/engine"

// One preset definition is shared by first-launch defaults and the settings UI.
// Individual saved settings are still loaded on top of these defaults.
func configureQuality050(c *nativeConfig, quality int) {
	quality = engine.ClampInt(quality, 0, 3)
	c.Quality = quality
	c.RenderScale = []float64{.67, 1, 1, 1}[quality]
	c.Shadows = quality
	c.ViewDistance = []float64{.60, .9, 1.2, 1.5}[quality]
	c.Vegetation = []float64{.5, .75, 1.1, 1.5}[quality]
	c.MeshDetail = quality
	c.GroundDetail = engine.ClampInt(quality, 0, 2)
	c.LightingDetail = quality
	c.ShadowFilter = engine.ClampInt(quality, 0, 2)
	c.FoliageDensity = []float64{.25, .50, .80, 1}[quality]
	c.Wind = true
	c.TextureBudget = quality
	c.TextureBias = []int{2, 1, 0, 0}[quality]
	c.Anisotropy = []int{4, 8, 16, 16}[quality]
	c.FXAA = true
	c.Bloom = quality >= 1
	c.MotionBlur = quality >= 2
	c.SSAO = quality >= 2
}

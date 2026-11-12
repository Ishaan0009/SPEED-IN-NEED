package main

import "testing"

func TestFirstLaunch050UsesTheSelectedMediumPreset(t *testing.T) {
	defaults := defaultCfg()
	explicit := nativeConfig{}
	configureQuality050(&explicit, 1)
	if defaults.Quality != 1 || defaults.RenderScale != explicit.RenderScale || defaults.ViewDistance != explicit.ViewDistance || defaults.Vegetation != explicit.Vegetation || defaults.FoliageDensity != explicit.FoliageDensity || defaults.TextureBias != explicit.TextureBias || defaults.TextureBudget != explicit.TextureBudget || defaults.Anisotropy != explicit.Anisotropy || defaults.Shadows != explicit.Shadows || defaults.ShadowFilter != explicit.ShadowFilter || defaults.MotionBlur != explicit.MotionBlur || defaults.SSAO != explicit.SSAO {
		t.Fatal("first launch no longer matches the selectable Medium graphics preset")
	}
	if defaults.FrameCap != 60 || !defaults.GamepadEnabled || defaults.MasterVolume <= 0 {
		t.Fatal("graphics defaults erased unrelated settings")
	}
}

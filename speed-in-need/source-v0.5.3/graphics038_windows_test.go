package main

import "testing"

func TestGPUBudget038UsesReportedMemoryNotModelNames(t *testing.T) {
	for _, tc := range []struct {
		dedicated, budget uint64
		cap               int
	}{{12 << 30, 11 << 30, 8192}, {6 << 30, 5 << 30, 2048}, {8 << 30, 7 << 30, 4096}, {2 << 30, 2 << 30, 1024}, {0, 0, 8192}} {
		if got := gpuTextureCapFor038(tc.dedicated, tc.budget); got != tc.cap {
			t.Fatal(tc, got)
		}
	}
}
func TestGraphics038UnavailableFeaturesCannotBeSelected(t *testing.T) {
	oldCfg, oldTab, oldSelect, oldGPU := beta.cfg, remaster.SettingsTab, beta.selection, gpu038
	defer func() { beta.cfg = oldCfg; remaster.SettingsTab = oldTab; beta.selection = oldSelect; gpu038 = oldGPU }()
	remaster.SettingsTab = 8
	gpu038 = capability038{DX12: true, DLSS: 0, SLInit: 0, RayTier: 12, NVIDIAFG: 6}
	for _, row := range []int{1, 2, 3, 4, 5, 6, 7, 9} {
		beta.selection = row
		before := beta.cfg
		if !adjustGraphics038(1) || beta.cfg != before || !graphicsRowDisabled038(row) {
			t.Fatal("capability row changed a feature", row)
		}
	}
	if len(graphicsValues038()) != len(graphicsLabels038) {
		t.Fatal("RTX setting layout")
	}
}

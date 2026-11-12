//go:build windows

package main

import "testing"

func TestBetaHeadlightToggleDoesNotToggleHelp(t *testing.T) {
	oldCfg, oldScreen, oldQA, oldHelp := beta.cfg, beta.screen, qa.enabled, app.showHelp
	defer func() { beta.cfg = oldCfg; beta.screen = oldScreen; qa.enabled = oldQA; app.showHelp = oldHelp }()
	qa.enabled = true
	beta.cfg = defaultCfg()
	beta.screen = screenDrive
	app.showHelp = false
	beta.bindWaiting = -1
	betaInput('H')
	if beta.cfg.Headlights || app.showHelp {
		t.Fatal("H must only disable headlights")
	}
	betaInput('H')
	if !beta.cfg.Headlights {
		t.Fatal("H did not restore headlights")
	}
	betaInput(0x70)
	if !app.showHelp || !beta.cfg.Headlights {
		t.Fatal("F1 help changed headlights")
	}
}
func TestBetaSettingsShareDrawAndHitRectangles(t *testing.T) {
	for i := 0; i < 15; i++ {
		r := settingRect(i)
		if r.W <= 0 || !r.Contains(r.X+r.W/2, r.Y+r.H/2) {
			t.Fatalf("setting %d unclickable", i)
		}
		for j := i + 1; j < 15; j++ {
			s := settingRect(j)
			if s.Contains(r.X+r.W/2, r.Y+r.H/2) {
				t.Fatalf("settings %d and %d overlap", i, j)
			}
		}
	}
}

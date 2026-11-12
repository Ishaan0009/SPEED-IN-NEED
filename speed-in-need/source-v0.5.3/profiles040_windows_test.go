package main

import (
	"encoding/json"
	"speedinneed/engine"
	"strings"
	"testing"
)

func profileUIFixture040(t *testing.T) {
	t.Helper()
	oldBeta, oldApp, oldRemaster, oldProfiles, oldPreview := beta, app, remaster, profiles040, preview014
	t.Cleanup(func() {
		beta, app, remaster, profiles040, preview014 = oldBeta, oldApp, oldRemaster, oldProfiles, oldPreview
	})
	beta.cfg = defaultCfg()
	app.tex = nil
	profiles040.Store = nil
	initializeProfiles040("")
}
func TestProfile040UISwitchingRestoresIndependentSettings(t *testing.T) {
	profileUIFixture040(t)
	if p := activeProfile040(); p == nil || !p.AllUnlocked040() {
		t.Fatal("default unlocked profile missing")
	}
	beta.cfg.MasterVolume = .23
	beta.cfg.CameraSensitivity = 1.7
	saveProfileSettings040()
	profileAction040("profiles040")
	profiles040.DraftName = "Fresh Driver"
	profileAction040("profile-create040")
	p := activeProfile040()
	if p.Name != "Fresh Driver" || beta.carSelect != 3 || beta.paint != 0 || p.CanRace040() {
		t.Fatal("new profile did not start with its own locked state")
	}
	if beta.cfg.MasterVolume == .23 || beta.cfg.CameraSensitivity == 1.7 {
		t.Fatal("new profile inherited previous driver settings")
	}
	beta.cfg.MasterVolume = .49
	saveProfileSettings040()
	id := p.ID
	profileAction040("profile-select040:ishaan")
	if beta.cfg.MasterVolume != .23 || beta.cfg.CameraSensitivity != 1.7 {
		t.Fatal("Ishaan settings were lost")
	}
	profileAction040("profile-select040:" + id)
	if beta.cfg.MasterVolume != .49 {
		t.Fatal("new profile settings were lost")
	}
	var cfg nativeConfig
	if e := json.Unmarshal(activeProfile040().Settings040(), &cfg); e != nil || cfg.SelectedCar != 3 {
		t.Fatal("profile did not serialize its own car selection")
	}
}
func TestProfile040UIUsesSuppliedGalleryOnly(t *testing.T) {
	profileUIFixture040(t)
	if len(profiles040.Avatars) == 0 {
		t.Fatal("supplied avatar manifest unavailable")
	}
	for _, avatar := range profiles040.Avatars {
		if !profiles040.Allowed[avatar.ID] || !strings.HasPrefix(avatar.Image, "assets/avatars040/") {
			t.Fatal("avatar is outside supplied gallery")
		}
	}
	profileAction040("profiles040")
	profileAction040("profile-avatar040:4")
	if activeProfile040().Avatar != profiles040.Avatars[4].ID {
		t.Fatal("avatar choice was not applied")
	}
	before := activeProfile040().Avatar
	profileAction040("profile-avatar040:9999")
	if activeProfile040().Avatar != before {
		t.Fatal("invalid avatar index changed portrait")
	}
}
func TestProfile040UINameEntryAndProtectedDeletion(t *testing.T) {
	profileUIFixture040(t)
	profileAction040("profiles040")
	profileAction040("profile-name040")
	for _, r := range strings.Repeat("A", 30) {
		profileChar040(r)
	}
	if len(profiles040.DraftName) != 24 {
		t.Fatal("name length limit failed")
	}
	profileKey040(0x08)
	if len(profiles040.DraftName) != 23 {
		t.Fatal("name backspace failed")
	}
	profileKey040(0x1b)
	if profiles040.Editing || profiles040.Page != 1 {
		t.Fatal("Escape should finish text entry first")
	}
	profileAction040("profile-delete040")
	if activeProfile040().Name != "Ishaan" {
		t.Fatal("protected profile removed")
	}
	profiles040.DraftName = "Temporary"
	profileAction040("profile-create040")
	id := activeProfile040().ID
	profileAction040("profile-delete040")
	if profiles040.Store.Find040(id) == nil {
		t.Fatal("profile deleted without confirmation")
	}
	profileAction040("profile-delete040")
	if profiles040.Store.Find040(id) != nil || activeProfile040().ID != "ishaan" {
		t.Fatal("confirmed profile deletion did not restore safe selection")
	}
}
func TestProfile040LockedSavedSelectionFallsBackSafely(t *testing.T) {
	profileUIFixture040(t)
	p, e := profiles040.Store.Create040("Saved Driver", "custom_001")
	if e != nil {
		t.Fatal(e)
	}
	cfg := defaultCfg()
	cfg.SelectedCar = 0
	cfg.Paint = 14
	cfg.GameMode = 1
	cfg.RaceDifficulty040 = 2
	raw, _ := json.Marshal(cfg)
	p.SetSettings040(raw)
	applyProfileSettings040()
	if beta.carSelect != 3 || beta.paint != 0 || remaster.Mode != 0 || beta.cfg.RaceDifficulty040 != 0 {
		t.Fatal("saved locked selections bypassed progression")
	}
	p.XP = engine.XPForLevel040(32)
	applyProfileSettings040()
	if beta.carSelect != 0 || beta.cfg.RaceDifficulty040 != 0 {
		t.Fatal("level32 car or difficulty gate wrong")
	}
}

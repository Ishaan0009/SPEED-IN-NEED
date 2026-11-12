package main

import "speedinneed/engine"

const qaProfilesFirst040 = 187

func registerQAProfiles040() {
	if len(remasterQANames) > qaProfilesFirst040 {
		return
	}
	remasterQANames = append(remasterQANames, "040-profile-manager", "040-profile-leaderboard", "040-new-profile-locked-garage", "040-profile-title")
}
func stageQAProfiles040() {
	if qa.stage < qaProfilesFirst040 || qa.stage > qaProfilesFirst040+3 {
		return
	}
	textures := profiles040.Textures
	initializeProfiles040("")
	if textures != nil {
		profiles040.Textures = textures
	}
	profiles040.MemoryOnly = true
	profiles040.Page = 0
	profiles040.Editing = false
	profiles040.DeletePending = ""
	profiles040.Cursor = 0
	s := profiles040.Store
	one, _ := s.Create040("Driver One", "custom_014")
	two, _ := s.Create040("Driver Two", "custom_029")
	one.XP = engine.XPForLevel040(7)
	one.NearPasses = 85
	one.RecordDrift040(4450, 17.8, false)
	two.XP = engine.XPForLevel040(10)
	two.NearPasses = 63
	two.RecordDrift040(12345, 15.6, false)
	two.RecordRaceResult040(1, 0)
	two.RecordRaceResult040(1, 1)
	two.RecordLap040(104.327, 2)
	ishaan := s.Find040("ishaan")
	ishaan.NearPasses = 144
	ishaan.RecordDrift040(20950, 21.4, false)
	ishaan.RecordLap040(98.035, 2)
	s.Select040("ishaan")
	applyProfileSettings040()
	selectScreen(screenTitle)
	app.started = false
	remaster.FullMap = false
	switch qa.stage {
	case qaProfilesFirst040:
		profiles040.Page = 1
		profiles040.DraftName = "NEW DRIVER"
		profiles040.Message = "NEW DRIVERS START WITH TRUENO / RACING UNLOCKS AT LEVEL 10"
	case qaProfilesFirst040 + 1:
		profiles040.Page = 2
		profiles040.Metric = "nearpass"
		profiles040.BoardFirst = 0
		profiles040.Message = "LOCAL DRIVER RECORDS / EACH PROFILE KEEPS ITS OWN PROGRESS"
	case qaProfilesFirst040 + 2:
		fresh, _ := s.Create040("New Driver", "custom_006")
		fresh.Settings = nil
		applyProfileSettings040()
		selectScreen(screenMaps)
		ui020.Stage = 2
		remaster.Row = 1
		remaster.Mode = 0
		beta.carSelect = 3
		preview014.ready = false
	case qaProfilesFirst040 + 3:
		profiles040.Page = 0
	}
}
func checksProfiles040() {
	if qa.stage < qaProfilesFirst040 || qa.stage > qaProfilesFirst040+3 {
		return
	}
	s := profiles040.Store
	if s == nil {
		qa.checks["040-profile-store-ready"] = false
		return
	}
	qa.checks["040-profile-store-ready"] = s.Active040() != nil && profiles040.MemoryOnly
	qa.checks["040-protected-driver-all-unlocked"] = s.Find040("ishaan") != nil && s.Find040("ishaan").CanRaceDifficulty040(2) && s.Find040("ishaan").CanUseCar040(0)
	qa.checks["040-supplied-avatar-gallery-ready"] = len(profiles040.Avatars) == 58 && profiles040.Allowed["custom_001"] && profiles040.Allowed["custom_058"]
	button := func(id string) bool {
		for _, b := range ui020.Buttons {
			if b.ID == id {
				return true
			}
		}
		return false
	}
	if qa.stage == qaProfilesFirst040 {
		qa.checks["040-profile-manager-controls-rendered"] = button("profile-create040") && button("profile-select040:ishaan") && button("profile-avatar040:0")
		qa.checks["040-avatar-atlas-uploaded"] = profiles040.Textures[profiles040.Atlas] != 0
	}
	if qa.stage == qaProfilesFirst040+1 {
		entries := s.Leaderboard040("nearpass")
		qa.checks["040-leaderboard-independent-records"] = len(entries) == 3 && entries[0].Name == "Ishaan" && entries[0].NearPasses == 144 && entries[1].BestDriftScore == 4450 && entries[2].BestLapSeconds == 104.327
		qa.checks["040-leaderboard-sort-controls-rendered"] = button("profile-sort040:nearpass") && button("profile-sort040:drift") && button("profile-sort040:lap")
	}
	if qa.stage == qaProfilesFirst040+2 {
		p := s.Active040()
		qa.checks["040-new-driver-only-trueno"] = p.Level040() == 1 && p.CanUseCar040(3) && !p.CanUseCar040(0) && !p.CanUseCar040(1) && !p.CanUseCar040(2) && !p.CanRace040() && beta.carSelect == 3 && p.PaintCount040(3, 18) == 1
	}
	if qa.stage == qaProfilesFirst040+3 {
		qa.checks["040-title-profile-and-leaderboard-actions"] = button("profiles040") && button("leaderboard040")
	}
}

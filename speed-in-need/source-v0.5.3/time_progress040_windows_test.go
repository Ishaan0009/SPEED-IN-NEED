package main

import (
	"encoding/json"
	"math"
	"speedinneed/engine"
	"testing"
)

func timeProgressFixture040(t *testing.T) {
	t.Helper()
	oldQA, oldCycle, oldRandom, oldProgress := qa, cycle034, randomHour040, progress040
	t.Cleanup(func() { qa, cycle034, randomHour040, progress040 = oldQA, oldCycle, oldRandom, oldProgress })
	profileUIFixture040(t)
	qa.enabled = true // Menu actions may call saveCfg; keep those writes disabled.
	cycle034.Elapsed, cycle034.Mode = 0, 0
	randomHour040 = math.NaN()
}

func TestTimeProgress040SelectedHourStaysFixedAndRandomAdvances(t *testing.T) {
	timeProgressFixture040(t)
	qa.enabled = false
	for mode, expected := range []float64{10, 17.8, 22} {
		remaster.Time = mode
		beta.cfg.RandomTime040 = false
		for _, elapsed := range []float64{0, 90, 9000, 1e9} {
			cycle034.Elapsed = elapsed
			if hour := cycleHour034(); hour != expected || cycle034.Elapsed != elapsed {
				t.Fatalf("selected time changed: mode %d elapsed %g hour %g", mode, elapsed, hour)
			}
		}
	}
	beta.cfg.RandomTime040 = true
	randomHour040 = 23.5
	cycle034.Elapsed = 90
	if hour := cycleHour034(); math.Abs(hour-.5) > 1e-8 {
		t.Fatal("random time did not advance and wrap midnight", hour)
	}
	cycle034.Elapsed = 180
	if hour := cycleHour034(); math.Abs(hour-1.5) > 1e-8 {
		t.Fatal("random time did not advance with elapsed simulation", hour)
	}
	startRandomTime040()
	if math.IsNaN(randomHour040) || randomHour040 < 0 || randomHour040 >= 24 || cycle034.Elapsed != 0 {
		t.Fatal("random mode did not initialize bounded starting time")
	}
}

func TestTimeProgress040ClockPausesOutsideDriving(t *testing.T) {
	timeProgressFixture040(t)
	qa.enabled = false
	beta.cfg.RandomTime040 = true
	remaster.Time, cycle034.Mode = 1, 1
	beta.screen = screenDrive
	app.paused, remaster.FullMap = false, false
	advanceCycle034(.025)
	if math.Abs(cycle034.Elapsed-.025) > 1e-8 {
		t.Fatal("random driving clock did not advance")
	}
	app.paused = true
	advanceCycle034(.05)
	app.paused, remaster.FullMap = false, true
	advanceCycle034(.05)
	remaster.FullMap, beta.screen = false, screenTitle
	advanceCycle034(.05)
	if math.Abs(cycle034.Elapsed-.025) > 1e-8 {
		t.Fatal("paused, map or title screen advanced world time")
	}
}

func TestTimeProgress040RandomSettingIsIndependentPerProfile(t *testing.T) {
	timeProgressFixture040(t)
	beta.cfg.RandomTime040 = true
	saveProfileSettings040()
	var cfg nativeConfig
	if err := json.Unmarshal(activeProfile040().Settings040(), &cfg); err != nil || !cfg.RandomTime040 {
		t.Fatal("random time setting was not serialized")
	}
	profileAction040("profiles040")
	profiles040.DraftName = "Fixed time driver"
	profileAction040("profile-create040")
	id := activeProfile040().ID
	beta.cfg.RandomTime040 = false
	saveProfileSettings040()
	profileAction040("profile-select040:ishaan")
	if !beta.cfg.RandomTime040 {
		t.Fatal("random mode lost after profile switch")
	}
	profileAction040("profile-select040:" + id)
	if beta.cfg.RandomTime040 {
		t.Fatal("random mode leaked into independent profile")
	}
}

func TestTimeProgress040LockedSelectionsDeniedAndCarCycling(t *testing.T) {
	timeProgressFixture040(t)
	p, err := profiles040.Store.Create040("Progress driver", "custom_001")
	if err != nil {
		t.Fatal(err)
	}
	applyProfileSettings040()
	if allowProgressAction040("mode1") || allowProgressAction040("car0") || allowProgressAction040("car1") || allowProgressAction040("car2") || allowProgressAction040("paint:1") {
		t.Fatal("fresh profile accepted a locked action")
	}
	if !allowProgressAction040("mode0") || !allowProgressAction040("car3") || !allowProgressAction040("paint:0") {
		t.Fatal("fresh profile denied its unlocked options")
	}
	for _, level := range []int{1, 4, 7, 10} {
		p.XP = engine.XPForLevel040(level)
		for class := 0; class < 4; class++ {
			for _, delta := range []int{-1, 1} {
				if next := nextProgressCar040(class, delta); !p.CanUseCar040(next) {
					t.Fatalf("level %d cycled from %d to locked car %d", level, class, next)
				}
			}
		}
	}
	p.XP = 0
	beta.carSelect, beta.paint = 0, 8
	if validateProgressSelection040() || beta.carSelect != 3 || beta.paint != 0 {
		t.Fatal("locked stored car did not fall back safely")
	}
	remaster.Mode = 1
	if validateProgressSelection040() || remaster.Mode != 0 {
		t.Fatal("locked race mode did not fall back")
	}
	p.XP = engine.XPForLevel040(10)
	remaster.Mode, beta.cfg.RaceDifficulty040 = 1, 2
	if validateProgressSelection040() || beta.cfg.RaceDifficulty040 != 0 {
		t.Fatal("locked race difficulty did not fall back to Easy")
	}
	if !validateProgressSelection040() {
		t.Fatal("safe unlocked selection was rejected")
	}
}

func TestTimeProgress040NewWorldDoesNotInheritAutosaveDelay(t *testing.T) {
	timeProgressFixture040(t)
	qa.enabled = false
	// The preceding drive saved at ten minutes. This fresh world must still
	// flush its first earned record after ten seconds, rather than ten minutes.
	progress040.LastSave, progress040.Dirty = 600, false
	app.world = &engine.World{Elapsed: 11, Car: engine.Car{Body: engine.ChassisState{Grounded: true}}, Remaster: &engine.RemasteredState{Mode: 1, BestLap: 100, Position: 1}}
	advanceProgress040(engine.PhysicsStep038)
	if activeProfile040().BestLapSeconds != 100 {
		t.Fatal("fixture did not earn lap record")
	}
	if progress040.Dirty || progress040.LastSave != app.world.Elapsed {
		t.Fatalf("new world inherited stale autosave clock: dirty %v lastsave %.1f elapsed %.1f", progress040.Dirty, progress040.LastSave, app.world.Elapsed)
	}
}

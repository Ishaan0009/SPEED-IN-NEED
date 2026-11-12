package engine

import (
	"encoding/json"
	"math"
	"strconv"
	"testing"
)

// A flat, in-memory world keeps progression tests independent of art assets,
// traffic randomness, rendering and filesystem state.
func driveEventFixture040(t *testing.T) (*World, *DriverProfile040) {
	t.Helper()
	d := &WorldDefinition{Grid: HeightGrid{W: 2, H: 2, MinX: -100, MinZ: -100, Step: 200, Values: make([]float32, 4)}, Roads: []RoadDefinition{{Points: []Vec{{0, -100}, {0, 100}}, Width: 12}}}
	w := &World{Remaster: &RemasteredState{Def: d}, Car: Car{Speed: 100 / 3.6, OnRoad: true, Body: ChassisState{Grounded: true}}, Traffic: []Traffic{{Pos: Vec{3, .2}, Speed: 8, Variant: 1}}}
	s := NewProfileStore040()
	p, err := s.Create040("Event driver", "custom_001")
	if err != nil {
		t.Fatal(err)
	}
	return w, p
}

func eventCount040(events []DriverEvent040, kind string) int {
	n := 0
	for _, event := range events {
		if event.Kind == kind {
			n++
		}
	}
	return n
}

func sweptPass040(s *DriveProgress040, w *World, p *DriverProfile040) []DriverEvent040 {
	w.Car.Pos = Vec{}
	s.Step040(w, p, PhysicsStep038)
	w.Car.Pos.Z = .4
	return s.Step040(w, p, PhysicsStep038)
}

func TestDriveEvents040SweptNearPassThresholdAndNoProximityFarm(t *testing.T) {
	for _, speed := range []float64{0, 99.999, 100, 180} {
		t.Run(strconv.FormatFloat(speed, 'f', -1, 64)+"kmh", func(t *testing.T) {
			w, p := driveEventFixture040(t)
			w.Car.Speed = speed / 3.6
			var observer DriveProgress040
			for i := 0; i < 600; i++ {
				if events := observer.Step040(w, p, PhysicsStep038); len(events) != 0 {
					t.Fatal("proximity awarded an event", events)
				}
			}
			if p.NearPasses != 0 || p.XP != 0 {
				t.Fatal("proximity farm changed profile")
			}
			events := sweptPass040(&observer, w, p)
			want := uint64(0)
			if speed >= 100 {
				want = 1
			}
			if p.NearPasses != want || p.XP != want*40 || eventCount040(events, "NEAR PASS") != int(want) {
				t.Fatalf("speed %.3f: events %+v, passes %d XP %d", speed, events, p.NearPasses, p.XP)
			}
		})
	}
}

func TestDriveEvents040NearPassCooldownAndUnsafeCrossings(t *testing.T) {
	w, p := driveEventFixture040(t)
	var observer DriveProgress040
	if eventCount040(sweptPass040(&observer, w, p), "NEAR PASS") != 1 {
		t.Fatal("initial swept pass missing")
	}
	// Returning ahead and passing again for almost eight seconds cannot farm.
	for i := 0; i < 470; i++ {
		sweptPass040(&observer, w, p)
	}
	if p.NearPasses != 1 {
		t.Fatalf("eight-second cooldown failed: %d passes", p.NearPasses)
	}
	for i := 0; i < 30; i++ {
		observer.Step040(w, p, PhysicsStep038)
	}
	if eventCount040(sweptPass040(&observer, w, p), "NEAR PASS") != 1 || p.NearPasses != 2 {
		t.Fatal("cooldown never reopened")
	}

	for _, scenario := range []struct {
		name                    string
		side, impulse, impact   float64
		changeVariant, teleport bool
	}{
		{"contact", 3, .2, 0, false, false}, {"traffic impact", 3, 0, 1, false, false},
		{"collision clearance", 1.8, 0, 0, false, false}, {"too far", 5, 0, 0, false, false},
		{"recycled vehicle", 3, 0, 0, true, false}, {"teleport", 3, 0, 0, false, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			w, p := driveEventFixture040(t)
			w.Traffic[0].Pos.X = scenario.side
			w.Car.ContactImpulse, w.Traffic[0].ImpactCooldown = scenario.impulse, scenario.impact
			var observer DriveProgress040
			observer.Step040(w, p, PhysicsStep038)
			if scenario.changeVariant {
				w.Traffic[0].Variant++
			}
			w.Car.Pos.Z = .4
			if scenario.teleport {
				w.Car.Pos.Z = 20
			}
			if events := observer.Step040(w, p, PhysicsStep038); len(events) != 0 || p.NearPasses != 0 || p.XP != 0 {
				t.Fatal("unsafe crossing counted", events)
			}
		})
	}
}

func TestDriveEvents040ProfileAndWorldIsolation(t *testing.T) {
	w, first := driveEventFixture040(t)
	_, second := driveEventFixture040(t)
	second.ID += "-second"
	var observer DriveProgress040
	observer.Step040(w, first, PhysicsStep038)
	w.Car.Pos.Z = .4
	if events := observer.Step040(w, second, PhysicsStep038); len(events) != 0 {
		t.Fatal("old profile's crossing leaked into new driver", events)
	}
	if eventCount040(sweptPass040(&observer, w, second), "NEAR PASS") != 1 || first.NearPasses != 0 || second.NearPasses != 1 {
		t.Fatal("profile pass totals not isolated")
	}
	clone := *w
	clone.Car.Pos.Z = .4
	if events := observer.Step040(&clone, first, PhysicsStep038); len(events) != 0 {
		t.Fatal("world transition inherited old crossing", events)
	}
	if eventCount040(sweptPass040(&observer, &clone, first), "NEAR PASS") != 1 || first.NearPasses != 1 {
		t.Fatal("new world retained another driver's cooldown")
	}
}

func TestDriveEvents040ContinuousDriftBankAndLandingOnce(t *testing.T) {
	w, p := driveEventFixture040(t)
	w.Traffic = nil
	w.Car.DriftScore = 900 // Historical lifetime score must not enter this run.
	var observer DriveProgress040
	observer.Step040(w, p, PhysicsStep038)
	w.Car.DriftActive, w.Car.DriftAngle = true, .2
	for tick := 0; tick < 120; tick++ {
		w.Car.DriftScore += 20
		if events := observer.Step040(w, p, PhysicsStep038); len(events) != 0 {
			t.Fatal("live drift banked early", events)
		}
	}
	if math.Abs(observer.Drift.Score-2400) > 1e-8 || math.Abs(observer.Drift.Seconds-1) > 1e-8 {
		t.Fatalf("live drift %+v", observer.Drift)
	}
	w.Car.DriftActive = false
	var bank []DriverEvent040
	for tick := 0; tick < 44; tick++ {
		bank = append(bank, observer.Step040(w, p, PhysicsStep038)...)
	}
	if eventCount040(bank, "DRIFT") != 1 || p.BestDriftScore != 2400 || math.Abs(p.BestDriftSeconds-1) > 1e-8 || p.XP != 120 {
		t.Fatalf("drift bank %+v, profile %+v", bank, p)
	}
	for tick := 0; tick < 120; tick++ {
		if events := observer.Step040(w, p, PhysicsStep038); len(events) != 0 {
			t.Fatal("bank repeated after stopping", events)
		}
	}
	w.Car.DriftActive = true
	for tick := 0; tick < 60; tick++ {
		w.Car.DriftScore += 10
		observer.Step040(w, p, PhysicsStep038)
	}
	w.Car.DriftActive = false
	for tick := 0; tick < 44; tick++ {
		observer.Step040(w, p, PhysicsStep038)
	}
	if p.BestDriftScore != 2400 || p.XP != 150 {
		t.Fatal("separate drift runs were combined or not awarded")
	}
	w.Car.Body.Grounded = false
	w.Car.Body.JumpDistance = 8
	for tick := 0; tick < 60; tick++ {
		observer.Step040(w, p, PhysicsStep038)
	}
	w.Car.Body.Grounded = true
	events := observer.Step040(w, p, PhysicsStep038)
	if eventCount040(events, "AIRTIME") != 1 || p.XP != 171 {
		t.Fatalf("stunt landing not banked once: events %+v XP %d", events, p.XP)
	}
	observer.Step040(w, p, PhysicsStep038)
	if p.XP != 171 {
		t.Fatal("grounded frame awarded repeated stunt XP")
	}
}

func TestDriveEvents040FixedSimulationIndependentOfRenderRate(t *testing.T) {
	for _, fps := range []int{30, 60, 144, 240} {
		w, p := driveEventFixture040(t)
		w.Traffic = nil
		var observer DriveProgress040
		observer.Step040(w, p, PhysicsStep038)
		var clock FixedClock038
		for frame := 0; frame < fps*2; frame++ {
			clock.Advance(1/float64(fps), func(dt float64) {
				w.Car.DriftActive = clock.Ticks < 120
				w.Car.DriftAngle = .2
				if w.Car.DriftActive {
					w.Car.DriftScore += 20
				}
				observer.Step040(w, p, dt)
			})
		}
		if clock.Ticks != 240 || p.BestDriftScore != 2400 || math.Abs(p.BestDriftSeconds-1) > 1e-8 || p.XP != 120 {
			t.Fatalf("%d FPS changed physics rewards: ticks %d profile %+v", fps, clock.Ticks, p)
		}
	}
}

func TestDriveEvents040RaceLapAndDifficultyUnlockOnce(t *testing.T) {
	w, p := driveEventFixture040(t)
	w.Traffic = nil
	w.Remaster.Mode, w.Remaster.BestLap = 1, 100
	w.Remaster.Finished, w.Remaster.Position = true, 1
	var observer DriveProgress040
	observer.Step040(w, p, PhysicsStep038)
	if p.BestLapSeconds != 0 || p.RaceWins != [3]uint64{} {
		t.Fatal("level-one driver bypassed locked racing")
	}
	p.XP = XPForLevel040(10)
	for difficulty := 0; difficulty < 3; difficulty++ {
		base, _ := driveEventFixture040(t)
		base.Traffic = nil
		base.Remaster.Mode, base.Remaster.RaceDifficulty040 = 1, difficulty
		base.Remaster.BestLap = 100 - float64(difficulty)*5
		base.Remaster.Finished, base.Remaster.Position = true, 1
		events := observer.Step040(base, p, PhysicsStep038)
		if eventCount040(events, "LAP") != 1 || eventCount040(events, "RACE FINISH") != 1 || p.RaceWins[difficulty] != 1 || p.BestLapSeconds != base.Remaster.BestLap {
			t.Fatalf("difficulty %d missing result %+v / %+v", difficulty, events, p)
		}
		observer.Step040(base, p, PhysicsStep038)
		if p.RaceWins[difficulty] != 1 {
			t.Fatal("finished race awarded more than once")
		}
		if difficulty < 2 && !p.CanRaceDifficulty040(difficulty+1) {
			t.Fatal("first place did not unlock next difficulty")
		}
	}
	if p.XP != XPForLevel040(10) {
		t.Fatal("race lap awarded free-roam XP")
	}
}

func TestDriveEvents040RejectedRaceDoesNotAnnounceSuccess(t *testing.T) {
	for _, scenario := range []struct {
		name                        string
		level, difficulty, position int
	}{
		{"racing locked", 1, 0, 1},
		{"difficulty locked", 10, 1, 1},
		{"invalid position", 10, 0, 0},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			w, p := driveEventFixture040(t)
			w.Traffic = nil
			p.XP = XPForLevel040(scenario.level)
			w.Remaster.Mode, w.Remaster.RaceDifficulty040 = 1, scenario.difficulty
			w.Remaster.Finished, w.Remaster.Position = true, scenario.position
			var observer DriveProgress040
			events := observer.Step040(w, p, PhysicsStep038)
			if eventCount040(events, "RACE FINISH") != 0 || observer.Finished || p.RaceWins != [3]uint64{} {
				t.Fatalf("rejected race reported successful: events %+v, finished %v, wins %v", events, observer.Finished, p.RaceWins)
			}
		})
	}
}

func TestDriveEvents040RacePassRecordsWithoutFreeRoamXP(t *testing.T) {
	w, p := driveEventFixture040(t)
	w.Remaster.Mode = 1
	p.XP = XPForLevel040(10)
	before := p.XP
	var observer DriveProgress040
	events := sweptPass040(&observer, w, p)
	if eventCount040(events, "NEAR PASS") != 1 || p.NearPasses != 1 || p.XP != before || events[0].XP != 0 {
		t.Fatalf("racing pass should count without XP: events %+v profile %+v", events, p)
	}
}

func TestProfile040IshaanDefaultAndExistingSaveRepair(t *testing.T) {
	check := func(t *testing.T, store *ProfileStore040) {
		t.Helper()
		p := store.Find040("ishaan")
		if p == nil || p.Name != "Ishaan" || p.Level040() != 50 || !p.AllUnlocked040() || !p.CanRaceDifficulty040(2) {
			t.Fatalf("protected level-50 driver not restored: %+v", p)
		}
		for class := 0; class < 4; class++ {
			if !p.CanUseCar040(class) {
				t.Fatal("protected driver's car locked", class)
			}
		}
		for _, row := range store.Leaderboard040("nearpass") {
			if row.ID == "ishaan" && row.Level != 50 {
				t.Fatal("leaderboard showed obsolete Ishaan level", row.Level)
			}
		}
	}
	fresh := NewProfileStore040()
	check(t, fresh)
	if fresh.Find040("ishaan").XP != XPForLevel040(50) {
		t.Fatal("default XP is not consistent with level 50")
	}
	for _, oldXP := range []uint64{0, XPForLevel040(50)} {
		existing := NewProfileStore040()
		old := existing.Find040("ishaan")
		old.Name, old.XP, old.NearPasses = "Legacy name", oldXP, 49
		old.Avatar = "../invalid"
		existing.ActiveID = "missing-active-profile"
		raw, err := json.Marshal(existing)
		if err != nil {
			t.Fatal(err)
		}
		loaded, err := decodeProfiles040(raw)
		if err != nil {
			t.Fatal(err)
		}
		check(t, loaded)
		p := loaded.Find040("ishaan")
		if loaded.ActiveID != "ishaan" || p.Avatar != "custom_001" || p.NearPasses != 49 || p.XP != oldXP {
			t.Fatalf("repair lost records or safe selection: %+v", loaded)
		}
	}
	missing, err := decodeProfiles040([]byte(`{"version":1,"activeID":"missing","profiles":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	check(t, missing)
	if missing.ActiveID != "ishaan" {
		t.Fatal("missing protected profile did not become safe active selection")
	}
}

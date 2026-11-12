package engine

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func newLockedProfile040(t *testing.T) (*ProfileStore040, *DriverProfile040) {
	t.Helper()
	s := NewProfileStore040()
	p, e := s.Create040("New driver", "avatar040_01")
	if e != nil {
		t.Fatal(e)
	}
	return s, p
}
func TestProfile040UnlocksAndRaceProgression(t *testing.T) {
	s, p := newLockedProfile040(t)
	ishaan := s.Find040("ishaan")
	if ishaan.Level040() != 50 {
		t.Fatal("Ishaan must display level50 while fully unlocked")
	}
	for class := 0; class < 4; class++ {
		if !ishaan.CanUseCar040(class) {
			t.Fatal("Ishaan car locked")
		}
		if p.CanUseCar040(class) != (class == 3) {
			t.Fatalf("new profile car %d availability", class)
		}
	}
	if p.PaintCount040(3, 18) != 1 || p.CanRace040() {
		t.Fatal("new profile must have one Trueno colour and locked racing")
	}
	for _, step := range []struct{ level, class int }{{12, 1}, {24, 2}, {32, 0}} {
		p.XP = XPForLevel040(step.level) - 1
		if p.CanUseCar040(step.class) {
			t.Fatal("premature car unlock")
		}
		p.XP++
		if !p.CanUseCar040(step.class) {
			t.Fatal("missing car unlock")
		}
	}
	if !p.CanRaceDifficulty040(0) || p.CanRaceDifficulty040(1) || p.CanRaceDifficulty040(2) {
		t.Fatal("level10 must start at Easy")
	}
	if p.RecordRaceResult040(1, 1) == nil {
		t.Fatal("locked Normal accepted")
	}
	if e := p.RecordRaceResult040(2, 0); e != nil {
		t.Fatal(e)
	}
	if p.CanRaceDifficulty040(1) {
		t.Fatal("second place unlocked Normal")
	}
	if e := p.RecordRaceResult040(1, 0); e != nil {
		t.Fatal(e)
	}
	if !p.CanRaceDifficulty040(1) || p.CanRaceDifficulty040(2) {
		t.Fatal("Easy first place progression wrong")
	}
	if e := p.RecordRaceResult040(1, 1); e != nil {
		t.Fatal(e)
	}
	if !p.CanRaceDifficulty040(2) {
		t.Fatal("Normal first did not unlock Hard")
	}
	for level := 0; level < 3; level++ {
		if !ishaan.CanRaceDifficulty040(level) {
			t.Fatal("Ishaan difficulty locked")
		}
	}
	if e := s.Delete040("ishaan"); e == nil {
		t.Fatal("protected Ishaan profile deleted")
	}
}
func TestProfile040EventsAndIndependentSettings(t *testing.T) {
	s, p := newLockedProfile040(t)
	first := []byte(`{"volume":0.3,"car":3}`)
	if e := p.SetSettings040(first); e != nil {
		t.Fatal(e)
	}
	first[2] = 'X'
	copy := p.Settings040()
	copy[2] = 'Y'
	if string(p.Settings040()) != `{"volume":0.3,"car":3}` {
		t.Fatal("settings alias caller storage")
	}
	other, e := s.Create040("Second", "avatar040_02")
	if e != nil {
		t.Fatal(e)
	}
	if len(other.Settings040()) != 0 || other.XP != 0 || other.NearPasses != 0 {
		t.Fatal("new profile inherited another profile data")
	}
	if e = other.SetSettings040([]byte(`{"volume":1}`)); e != nil {
		t.Fatal(e)
	}
	if ok, xp := p.RecordNearPass040(99, true); ok || xp != 0 {
		t.Fatal("slow pass counted")
	}
	if ok, xp := p.RecordNearPass040(150, false); !ok || xp != 0 || p.NearPasses != 1 {
		t.Fatal("race pass count/XP wrong")
	}
	if ok, xp := p.RecordNearPass040(150, true); !ok || xp != 40 || p.XP != 40 {
		t.Fatal("free pass XP wrong")
	}
	if xp := p.RecordDrift040(2000, 12, true); xp != 100 || p.BestDriftScore != 2000 || p.BestDriftSeconds != 12 {
		t.Fatal("drift record/XP wrong")
	}
	p.RecordDrift040(2000, 10, false)
	if p.BestDriftSeconds != 10 {
		t.Fatal("equal-score faster drift not recorded")
	}
	p.RecordDrift040(500, 3, true)
	if p.BestDriftScore != 2000 {
		t.Fatal("smaller drift replaced best")
	}
	before := p.XP
	for _, v := range []float64{math.NaN(), math.Inf(1), -1, 0} {
		p.RecordDrift040(v, 2, true)
		p.RecordStunt040(v, 1, true)
		p.RecordNearPass040(v, true)
	}
	if p.XP != before {
		t.Fatal("invalid events awarded XP")
	}
	if p.RecordStunt040(2, 1, true) != 0 || p.RecordStunt040(20, .1, true) != 0 || p.RecordStunt040(20, 1, false) != 0 {
		t.Fatal("invalid stunt awarded XP")
	}
	if p.RecordStunt040(20, 1, true) != 50 {
		t.Fatal("valid stunt not awarded")
	}
	if other.XP != 0 || other.NearPasses != 0 || other.BestDriftScore != 0 {
		t.Fatal("events leaked across profiles")
	}
	if e = p.SetAvatar040("outside", map[string]bool{"avatar040_01": true}); e == nil {
		t.Fatal("non-gallery avatar accepted")
	}
	if e = p.SetAvatar040("../outside", map[string]bool{"../outside": true}); e == nil {
		t.Fatal("avatar traversal accepted")
	}
}
func TestProfile040LeaderboardsUseCountsBestRunAndMinimumLap(t *testing.T) {
	s, a := newLockedProfile040(t)
	b, e := s.Create040("Other", "avatar040_02")
	if e != nil {
		t.Fatal(e)
	}
	a.NearPasses = 4
	b.NearPasses = 8
	a.BestDriftScore = 900
	b.BestDriftScore = 500
	a.XP, b.XP = XPForLevel040(10), XPForLevel040(10)
	if e = a.RecordLap040(112, 0); e != nil {
		t.Fatal(e)
	}
	a.RecordLap040(120, 0)
	b.RecordLap040(102, 0)
	if s.Leaderboard040("nearpass")[0].ID != b.ID || s.Leaderboard040("drift")[0].ID != a.ID || s.Leaderboard040("lap")[0].ID != b.ID {
		t.Fatal("leaderboard sort wrong")
	}
	if a.BestLapSeconds != 112 {
		t.Fatal("lap record did not retain minimum")
	}
	if e = a.RecordLap040(.1, 0); e == nil {
		t.Fatal("impossible lap time accepted")
	}
	if e = a.RecordLap040(90, 2); e == nil {
		t.Fatal("locked Hard lap accepted")
	}
	if s.Leaderboard040("lap")[2].BestLapSeconds != 0 {
		t.Fatal("no-lap profile sorted before timed laps")
	}
}
func TestProfile040ContinuousDriftStartsFromZeroAndCommitsOnce(t *testing.T) {
	var run DriftRun040
	for i := 0; i < 20; i++ {
		if _, done := run.Step040(true, 10, .1); done {
			t.Fatal("active drift committed")
		}
	}
	if run.Score != 200 || math.Abs(run.Seconds-2) > 1e-8 {
		t.Fatal("drift accumulation wrong")
	}
	run.Step040(false, 0, .1)
	run.Step040(true, 5, .1)
	if run.Score != 205 {
		t.Fatal("short grip gap reset drift")
	}
	var result DriftResult040
	done := false
	for i := 0; i < 4; i++ {
		result, done = run.Step040(false, 0, .1)
	}
	if !done || result.Score != 205 || run.Active {
		t.Fatal("finished drift was not committed once")
	}
	if _, done = run.Step040(false, 0, .1); done {
		t.Fatal("finished drift committed twice")
	}
	run.Step040(true, 3, .1)
	if run.Score != 3 {
		t.Fatal("new drift did not start from zero")
	}
}
func TestProfile040AtomicSaveBackupAndCorruptionRecovery(t *testing.T) {
	// A project-local test directory also exercises Windows atomic rename on
	// the actual game volume; no user settings or real profile save is touched.
	dir, e := os.MkdirTemp(".", ".profile040-test-")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	path := filepath.Join(dir, "profiles.json")
	s, p := newLockedProfile040(t)
	if e := SaveProfileStore040(path, s); e != nil {
		t.Fatal(e)
	}
	p.XP = 1234
	if e := SaveProfileStore040(path, s); e != nil {
		t.Fatal(e)
	}
	loaded, e := LoadProfileStore040(path)
	if e != nil || loaded.Active040().XP != 1234 {
		t.Fatalf("save reload failed: %v", e)
	}
	if e = os.WriteFile(path, []byte("interrupted JSON"), 0600); e != nil {
		t.Fatal(e)
	}
	recovered, e := LoadProfileStore040(path)
	if e != nil || !recovered.RecoveredFromBackup || recovered.Active040().XP != 0 {
		t.Fatalf("backup recovery failed: %v", e)
	}
	if e = SaveProfileStore040(path, recovered); e != nil {
		t.Fatal(e)
	}
	if _, e = decodeProfiles040([]byte(`{"version":99}`)); e == nil {
		t.Fatal("future format accepted")
	}
	if e = os.WriteFile(path, []byte("bad-primary"), 0600); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(path+".bak", []byte("bad-backup"), 0600); e != nil {
		t.Fatal(e)
	}
	if loaded, e = LoadProfileStore040(path); e == nil || loaded != nil {
		t.Fatal("corrupt saves silently replaced")
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "bad-primary" {
		t.Fatal("corrupt original was destroyed")
	}
	matches, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".*.tmp-*"))
	if len(matches) > 0 {
		t.Fatal("save left temporary files")
	}
	missing, e := LoadProfileStore040(filepath.Join(t.TempDir(), "missing.json"))
	if e != nil || missing.Active040().Name != "Ishaan" {
		t.Fatal("first-run default missing")
	}
}
func TestProfile040NameValidationAndRoundTrip(t *testing.T) {
	s := NewProfileStore040()
	for _, name := range []string{" ", "ishaan", "bad\nname", "../outside", "1234567890123456789012345"} {
		if _, e := s.Create040(name, "avatar040_00"); e == nil {
			t.Fatalf("invalid name accepted: %q", name)
		}
	}
	p, e := s.Create040("  Driver One  ", "avatar040_03")
	if e != nil || p.Name != "Driver One" {
		t.Fatal("valid profile creation failed")
	}
	if _, e = s.Create040("DRIVER ONE", "avatar040_03"); e == nil {
		t.Fatal("case-insensitive duplicate accepted")
	}
	raw, e := json.Marshal(s)
	if e != nil {
		t.Fatal(e)
	}
	round, e := decodeProfiles040(raw)
	if e != nil || round.Active040().ID != p.ID {
		t.Fatal("profile selection lost on reload")
	}
	if e = s.Delete040(p.ID); e != nil || s.Active040().Name != "Ishaan" {
		t.Fatal("active deletion did not safely select protected profile")
	}
}

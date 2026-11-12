package engine

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const ProfileVersion040 = 1
const MaximumProfiles040 = 64
const MaximumLevel040 = 50
const IshaanLevel040 = 50

var ErrProfileLocked040 = errors.New("this profile has not unlocked that activity")
var ErrProfileVersion040 = errors.New("unsupported profile save version")

// Profile settings are opaque game configuration JSON, copied on read and write.
// Events and saves are called on the game's main thread, independently of the
// displayed frame rate and any generated presentation frames.
type DriverProfile040 struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	Avatar            string          `json:"avatar"`
	XP                uint64          `json:"xp"`
	NearPasses        uint64          `json:"nearPasses"`
	BestDriftScore    float64         `json:"bestDriftScore"`
	BestDriftSeconds  float64         `json:"bestDriftSeconds"`
	BestLapSeconds    float64         `json:"bestLapSeconds"`
	BestLapDifficulty int             `json:"bestLapDifficulty"`
	RaceWins          [3]uint64       `json:"raceWins"`
	Settings          json.RawMessage `json:"settings,omitempty"`
	CreatedUTC        string          `json:"createdUTC"`
}

type ProfileStore040 struct {
	Version             int                 `json:"version"`
	ActiveID            string              `json:"activeID"`
	Profiles            []*DriverProfile040 `json:"profiles"`
	RecoveredFromBackup bool                `json:"-"`
}

func NewProfileStore040() *ProfileStore040 {
	return &ProfileStore040{Version: ProfileVersion040, ActiveID: "ishaan", Profiles: []*DriverProfile040{{ID: "ishaan", Name: "Ishaan", Avatar: "custom_001", XP: XPForLevel040(IshaanLevel040), CreatedUTC: time.Now().UTC().Format(time.RFC3339)}}}
}

func validProfileName040(name string) (string, error) {
	name = strings.TrimSpace(name)
	if !utf8.ValidString(name) || utf8.RuneCountInString(name) < 1 || utf8.RuneCountInString(name) > 24 {
		return "", errors.New("use a profile name of 1 to 24 characters")
	}
	for _, r := range name {
		if unicode.IsControl(r) || r == '/' || r == '\\' {
			return "", errors.New("profile names cannot contain control characters or slashes")
		}
	}
	return name, nil
}
func avatarKey040(key string) bool {
	if len(key) < 1 || len(key) > 80 {
		return false
	}
	for _, r := range key {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}
func (s *ProfileStore040) Find040(id string) *DriverProfile040 {
	for _, p := range s.Profiles {
		if p != nil && p.ID == id {
			return p
		}
	}
	return nil
}
func (s *ProfileStore040) Active040() *DriverProfile040 {
	if s == nil {
		return nil
	}
	return s.Find040(s.ActiveID)
}
func (s *ProfileStore040) Create040(name, avatar string) (*DriverProfile040, error) {
	name, err := validProfileName040(name)
	if err != nil {
		return nil, err
	}
	if len(s.Profiles) >= MaximumProfiles040 {
		return nil, errors.New("profile limit reached")
	}
	for _, p := range s.Profiles {
		if strings.EqualFold(p.Name, name) {
			return nil, errors.New("a profile with that name already exists")
		}
	}
	if !avatarKey040(avatar) {
		return nil, errors.New("choose an avatar from the gallery")
	}
	var random [12]byte
	if _, err = rand.Read(random[:]); err != nil {
		return nil, fmt.Errorf("create profile ID: %w", err)
	}
	p := &DriverProfile040{ID: hex.EncodeToString(random[:]), Name: name, Avatar: avatar, CreatedUTC: time.Now().UTC().Format(time.RFC3339)}
	s.Profiles = append(s.Profiles, p)
	s.ActiveID = p.ID
	return p, nil
}
func (s *ProfileStore040) Select040(id string) error {
	if s.Find040(id) == nil {
		return errors.New("profile not found")
	}
	s.ActiveID = id
	return nil
}
func (s *ProfileStore040) Delete040(id string) error {
	if id == "ishaan" {
		return errors.New("the Ishaan profile is protected")
	}
	for i, p := range s.Profiles {
		if p.ID == id {
			s.Profiles = append(s.Profiles[:i], s.Profiles[i+1:]...)
			if s.ActiveID == id {
				s.ActiveID = "ishaan"
			}
			return nil
		}
	}
	return errors.New("profile not found")
}
func (p *DriverProfile040) SetAvatar040(key string, allowed map[string]bool) error {
	if !avatarKey040(key) || !allowed[key] {
		return errors.New("choose an avatar from the supplied gallery")
	}
	p.Avatar = key
	return nil
}
func (p *DriverProfile040) SetSettings040(settings []byte) error {
	if len(settings) > 128*1024 || !json.Valid(settings) {
		return errors.New("profile settings must be valid JSON")
	}
	p.Settings = append(json.RawMessage(nil), settings...)
	return nil
}
func (p *DriverProfile040) Settings040() []byte { return append([]byte(nil), p.Settings...) }
func (p *DriverProfile040) AllUnlocked040() bool {
	return p != nil && p.ID == "ishaan" && p.Name == "Ishaan"
}
func XPForLevel040(level int) uint64 {
	level = ClampInt(level, 1, MaximumLevel040)
	n := uint64(level - 1)
	return 100 * n * n
}
func (p *DriverProfile040) Level040() int {
	if p == nil {
		return 1
	}
	if p.AllUnlocked040() {
		return IshaanLevel040
	}
	level := 1
	for level < MaximumLevel040 && p.XP >= XPForLevel040(level+1) {
		level++
	}
	return level
}
func CarUnlockLevel040(class int) int {
	if class < 0 || class > 3 {
		return MaximumLevel040 + 1
	}
	return [4]int{32, 12, 24, 1}[class]
}
func (p *DriverProfile040) CanUseCar040(class int) bool {
	return p != nil && class >= 0 && class < 4 && (p.AllUnlocked040() || p.Level040() >= CarUnlockLevel040(class))
}
func (p *DriverProfile040) PaintCount040(class, available int) int {
	if available <= 0 || !p.CanUseCar040(class) {
		return 0
	}

	if p.AllUnlocked040() {
		return available
	}
	// Every available finish unlocks by level 40, independently for each car.
	start := CarUnlockLevel040(class)
	return ClampInt(1+(p.Level040()-start)*(available-1)/(40-start), 1, available)
}
func (p *DriverProfile040) CanRace040() bool {
	return p != nil && (p.AllUnlocked040() || p.Level040() >= 10)
}
func (p *DriverProfile040) CanRaceDifficulty040(difficulty int) bool {
	if !p.CanRace040() || difficulty < 0 || difficulty > 2 {
		return false
	}
	return p.AllUnlocked040() || difficulty == 0 || p.RaceWins[difficulty-1] > 0
}
func (p *DriverProfile040) awardXP040(value uint64) uint64 {
	if p == nil || p.AllUnlocked040() {
		return 0
	}
	old := p.XP
	if ^uint64(0)-old < value {
		p.XP = ^uint64(0)
	} else {
		p.XP += value
	}
	return p.XP - old
}
func finitePositive040(v float64) bool { return v > 0 && !math.IsNaN(v) && !math.IsInf(v, 0) }

// The caller's geometry detector supplies one event per distinct close pass.
// Racing passes can enter the leaderboard; only free driving awards XP.
func (p *DriverProfile040) RecordNearPass040(speedKPH float64, freeRide bool) (bool, uint64) {
	if p == nil || !finitePositive040(speedKPH) || speedKPH < 100 {
		return false, 0
	}
	if p.NearPasses < ^uint64(0) {
		p.NearPasses++
	}
	if freeRide {
		return true, p.awardXP040(40)
	}
	return true, 0
}
func (p *DriverProfile040) RecordDrift040(score, seconds float64, freeRide bool) uint64 {
	if p == nil || !finitePositive040(score) || !finitePositive040(seconds) || seconds < .25 {
		return 0
	}
	if score > p.BestDriftScore || (score == p.BestDriftScore && (p.BestDriftSeconds == 0 || seconds < p.BestDriftSeconds)) {
		p.BestDriftScore, p.BestDriftSeconds = score, seconds
	}
	if !freeRide {
		return 0
	}
	return p.awardXP040(uint64(math.Min(2000, math.Max(1, math.Floor(score/20)))))
}
func (p *DriverProfile040) RecordStunt040(distance, airSeconds float64, freeRide bool) uint64 {
	if p == nil || !freeRide || !finitePositive040(distance) || !finitePositive040(airSeconds) || distance < 5 || airSeconds < .35 {
		return 0
	}
	return p.awardXP040(uint64(math.Min(400, math.Floor(distance*2+airSeconds*10))))
}
func (p *DriverProfile040) RecordLap040(seconds float64, difficulty int) error {
	if !p.CanRaceDifficulty040(difficulty) {
		return ErrProfileLocked040
	}
	if !finitePositive040(seconds) || seconds < 5 {
		return errors.New("invalid completed lap time")
	}
	if p.BestLapSeconds == 0 || seconds < p.BestLapSeconds {
		p.BestLapSeconds, p.BestLapDifficulty = seconds, difficulty
	}
	return nil
}
func (p *DriverProfile040) RecordRaceResult040(position, difficulty int) error {
	if !p.CanRaceDifficulty040(difficulty) {
		return ErrProfileLocked040
	}
	if position < 1 {
		return errors.New("invalid race position")
	}
	if position == 1 && p.RaceWins[difficulty] < ^uint64(0) {
		p.RaceWins[difficulty]++
	}
	return nil
}

type LeaderboardEntry040 struct {
	ID, Name, Avatar                                 string
	Level                                            int
	NearPasses                                       uint64
	BestDriftScore, BestDriftSeconds, BestLapSeconds float64
	BestLapDifficulty                                int
}

func (s *ProfileStore040) Leaderboard040(metric string) []LeaderboardEntry040 {
	var entries []LeaderboardEntry040
	for _, p := range s.Profiles {
		entries = append(entries, LeaderboardEntry040{p.ID, p.Name, p.Avatar, p.Level040(), p.NearPasses, p.BestDriftScore, p.BestDriftSeconds, p.BestLapSeconds, p.BestLapDifficulty})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		switch metric {
		case "drift":
			if a.BestDriftScore != b.BestDriftScore {
				return a.BestDriftScore > b.BestDriftScore
			}
			if a.BestDriftScore > 0 && a.BestDriftSeconds != b.BestDriftSeconds {
				return a.BestDriftSeconds < b.BestDriftSeconds
			}
		case "lap":
			if a.BestLapSeconds != b.BestLapSeconds {
				if a.BestLapSeconds == 0 {
					return false
				}
				if b.BestLapSeconds == 0 {
					return true
				}
				return a.BestLapSeconds < b.BestLapSeconds
			}
		default:
			if a.NearPasses != b.NearPasses {
				return a.NearPasses > b.NearPasses
			}
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})
	return entries
}

func normalizeProfiles040(s *ProfileStore040) error {
	if s.Version != ProfileVersion040 {
		return fmt.Errorf("%w: %d", ErrProfileVersion040, s.Version)
	}
	if len(s.Profiles) > MaximumProfiles040 {
		return errors.New("profile save exceeds profile limit")
	}
	ids, names := map[string]bool{}, map[string]bool{}
	hasIshaan := false
	for _, p := range s.Profiles {
		if p == nil || p.ID == "" || len(p.ID) > 80 || ids[p.ID] {
			return errors.New("invalid or duplicate profile ID")
		}
		ids[p.ID] = true
		name, err := validProfileName040(p.Name)
		if err != nil {
			return err
		}
		p.Name = name
		lower := strings.ToLower(name)
		if names[lower] {
			return errors.New("duplicate profile name")
		}
		names[lower] = true
		if p.ID == "ishaan" {
			p.Name = "Ishaan"
			hasIshaan = true
		} else if strings.EqualFold(name, "Ishaan") {
			return errors.New("Ishaan is a reserved profile name")
		}
		if !avatarKey040(p.Avatar) {
			p.Avatar = "custom_001"
		}
		if len(p.Settings) > 0 && (len(p.Settings) > 128*1024 || !json.Valid(p.Settings)) {
			return errors.New("invalid profile settings")
		}
		for _, v := range []float64{p.BestDriftScore, p.BestDriftSeconds, p.BestLapSeconds} {
			if v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
				return errors.New("invalid profile record")
			}
		}
		if p.BestLapDifficulty < 0 || p.BestLapDifficulty > 2 {
			p.BestLapDifficulty = 0
		}
	}
	if !hasIshaan {
		if len(s.Profiles) >= MaximumProfiles040 {
			return errors.New("missing protected profile in full save")
		}
		s.Profiles = append([]*DriverProfile040{NewProfileStore040().Profiles[0]}, s.Profiles...)
	}
	if s.Find040(s.ActiveID) == nil {
		s.ActiveID = "ishaan"
	}
	return nil
}
func decodeProfiles040(raw []byte) (*ProfileStore040, error) {
	if len(raw) > 2*1024*1024 {
		return nil, errors.New("profile save exceeds size limit")
	}
	var s ProfileStore040
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	if err := normalizeProfiles040(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Missing saves create a fresh store. Corrupt saves are never silently deleted.
// A valid backup is returned with RecoveredFromBackup=true; otherwise nil/error.
func LoadProfileStore040(path string) (*ProfileStore040, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return NewProfileStore040(), nil
	}
	if err == nil {
		if s, e := decodeProfiles040(raw); e == nil {
			return s, nil
		} else {
			err = e
		}
	}
	// An older game must not roll a newer-format save back to an old backup.
	if errors.Is(err, ErrProfileVersion040) {
		return nil, err
	}
	backup, backupErr := os.ReadFile(path + ".bak")
	if backupErr == nil {
		if s, e := decodeProfiles040(backup); e == nil {
			s.RecoveredFromBackup = true
			return s, nil
		}
	}
	return nil, fmt.Errorf("profile save could not be loaded; original and backup preserved: %w", err)
}
func atomicProfileFile040(path string, raw []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err = f.Write(raw); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
func SaveProfileStore040(path string, s *ProfileStore040) error {
	if path == "" || s == nil {
		return errors.New("profile save path/store is missing")
	}
	// Validate a separate snapshot so saving never mutates the selected profile.
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if _, err = decodeProfiles040(raw); err != nil {
		return err
	}
	if old, readErr := os.ReadFile(path); readErr == nil {
		if _, validErr := decodeProfiles040(old); validErr == nil {
			if err = atomicProfileFile040(path+".bak", old); err != nil {
				return fmt.Errorf("preserve profile backup: %w", err)
			}
		}
	} else if !os.IsNotExist(readErr) {
		return readErr
	}
	return atomicProfileFile040(path, append(raw, '\n'))
}

type DriftResult040 struct{ Score, Seconds float64 }
type DriftRun040 struct {
	Active                       bool
	Score, Seconds, QuietSeconds float64
}

// A brief loss of slip does not split a drift combo. Score starts at zero on
// each run and only positive measured score deltas are accumulated.
func (r *DriftRun040) Step040(drifting bool, scoreDelta, dt float64) (DriftResult040, bool) {
	if !finitePositive040(dt) {
		return DriftResult040{}, false
	}
	dt = math.Min(dt, .25)
	if drifting {
		if !r.Active {
			*r = DriftRun040{Active: true}
		}
		r.QuietSeconds = 0
		r.Seconds += dt
		if finitePositive040(scoreDelta) {
			r.Score += math.Min(scoreDelta, 1e6)
		}
		return DriftResult040{}, false
	}
	if !r.Active {
		return DriftResult040{}, false
	}
	r.QuietSeconds += dt
	if r.QuietSeconds < .35 {
		return DriftResult040{}, false
	}
	result := DriftResult040{r.Score, r.Seconds}
	*r = DriftRun040{}
	return result, true
}

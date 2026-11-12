package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
	"time"
)

var progress040 struct {
	Observer              engine.DriveProgress040
	Dirty                 bool
	LastSave, NoticeUntil float64
	Notice, Gate          string
	GateUntil             time.Time
	LastDrift             engine.DriftResult040
}

func advanceProgress040(dt float64) {
	if qa.enabled {
		return
	}
	p := activeProfile040()
	if p == nil {
		return
	}
	if progress040.Observer.World != app.world || progress040.Observer.ProfileID != p.ID {
		if progress040.Dirty {
			flushProfiles040()
		}
		progress040.Dirty = false
		progress040.LastSave = 0
		progress040.NoticeUntil = 0
		progress040.Notice = ""
		progress040.LastDrift = engine.DriftResult040{}
	}
	before := p.Level040()
	for _, event := range progress040.Observer.Step040(app.world, p, dt) {
		progress040.Dirty = true
		switch event.Kind {
		case "DRIFT":
			progress040.LastDrift = engine.DriftResult040{Score: event.Value, Seconds: event.Seconds}
			progress040.Notice = fmt.Sprintf("DRIFT BANKED / %.0f PTS / %.1f SEC", event.Value, event.Seconds)
		case "NEAR PASS":
			progress040.Notice = fmt.Sprintf("NEAR PASS / +%d XP", event.XP)
		case "AIRTIME":
			progress040.Notice = fmt.Sprintf("AIRTIME / %.1f M / +%d XP", event.Value, event.XP)
		case "LAP":
			progress040.Notice = fmt.Sprintf("LAP RECORDED / %.2f SEC", event.Value)
		case "RACE FINISH":
			progress040.Notice = fmt.Sprintf("FINISH / POSITION %.0f", event.Value)
		}
		progress040.NoticeUntil = app.world.Elapsed + 4
	}
	if p.Level040() > before {
		progress040.Notice = fmt.Sprintf("LEVEL %d / NEW REWARDS AVAILABLE", p.Level040())
		progress040.NoticeUntil = app.world.Elapsed + 5
	}
	if progress040.Dirty && app.world.Elapsed-progress040.LastSave > 10 {
		flushProfiles040()
		progress040.LastSave = app.world.Elapsed
		progress040.Dirty = false
	}
}
func progressGate040(message string) bool {
	progress040.Gate = message
	progress040.GateUntil = time.Now().Add(4 * time.Second)
	return false
}
func validateProgressSelection040() bool {
	p := activeProfile040()
	if p == nil {
		return true
	}
	if !p.CanUseCar040(beta.carSelect) {
		beta.carSelect = 3
		beta.paint = 0
		return progressGate040("THIS CAR IS LOCKED / EARN XP IN FREE ROAM")
	}
	if remaster.Mode == 1 && !p.CanRace040() {
		remaster.Mode = 0
		return progressGate040("CIRCUIT RACING UNLOCKS AT LEVEL 10")
	}
	if remaster.Mode == 1 && !p.CanRaceDifficulty040(beta.cfg.RaceDifficulty040) {
		beta.cfg.RaceDifficulty040 = 0
		return progressGate040("WIN THE PREVIOUS DIFFICULTY IN FIRST PLACE")
	}
	return true
}
func allowProgressAction040(id string) bool {
	p := activeProfile040()
	if p == nil {
		return true
	}
	if id == "mode1" && !p.CanRace040() {
		return progressGate040("CIRCUIT RACING / LOCKED UNTIL LEVEL 10")
	}
	if len(id) == 4 && id[:3] == "car" && id[3] >= '0' && id[3] <= '3' {
		class := int(id[3] - '0')
		if !p.CanUseCar040(class) {
			return progressGate040(fmt.Sprintf("CAR UNLOCKS AT LEVEL %d", engine.CarUnlockLevel040(class)))
		}
	}
	var paint int
	if _, e := fmt.Sscanf(id, "paint:%d", &paint); e == nil && setup050().Model == 0 && paint >= p.PaintCount040(beta.carSelect, len(paintPalette)) {
		return progressGate040("EARN MORE LEVELS TO UNLOCK THIS PAINT")
	}
	return true
}
func nextProgressCar040(current, delta int) int {
	p := activeProfile040()
	step := 1
	if delta < 0 {
		step = -1
	}
	result := (current + 4 + delta) % 4
	for i := 0; i < 4; i++ {
		if p == nil || p.CanUseCar040(result) {
			return result
		}
		result = (result + 4 + step) % 4
	}
	return current
}
func drawProgressNotice040(w, h float64) {
	if time.Now().Before(progress040.GateUntil) {
		headline035(progress040.Gate, w*.075, h*.805, h*.012, w*.8, cyan)
	}
}
func drawProgressHUD040(w, h float64) {
	p := activeProfile040()
	if p == nil || app.world == nil {
		return
	}
	drift := progress040.Observer.Drift
	if drift.Active {
		remasterText(fmt.Sprintf("DRIFT / %.0f", drift.Score), w*.40, h*.76, h*.021, cyan)
		remasterText(fmt.Sprintf("%.1f SEC", drift.Seconds), w*.46, h*.80, h*.010, white)
	}
	if app.world.Elapsed < progress040.NoticeUntil {
		headline035(progress040.Notice, w*.32, h*.10, h*.012, w*.5, white)
	}
	if !p.AllUnlocked040() {
		level := p.Level040()
		low, high := engine.XPForLevel040(level), engine.XPForLevel040(level+1)
		fraction := 0.
		if high > low {
			fraction = engine.Clamp(float64(p.XP-low)/float64(high-low), 0, 1)
		}
		remasterText(fmt.Sprintf("LEVEL %d", level), w*.45, h*.94, h*.010, mutedRefine)
		hudRect(w*.42, h*.97, w*.16, 2, .15, .18, .18, .7)
		hudRect(w*.42, h*.97, w*.16*fraction, 2, cyan[0], cyan[1], cyan[2], .9)
	}
}

var randomHour040 = math.NaN()

func startRandomTime040() {
	randomHour040 = math.Mod(float64(time.Now().UnixNano()%int64(24*time.Hour))/float64(time.Hour), 24)
	cycle034.Elapsed = 0
}
func visiblePaintCount040() int {
	available := 10
	if qa.enabled {
		available = len(paintPalette)
	}

	if p := activeProfile040(); p != nil {
		return p.PaintCount040(beta.carSelect, available)
	}
	return available
}
func skySettingName040() string {
	return []string{"AUTO / GPU BUDGET", "2K / FAST", "4K / HIGH", "8K / VERY HIGH", "16K / SOURCE"}[engine.ClampInt(beta.cfg.SkyDetail040, 0, 4)]
}

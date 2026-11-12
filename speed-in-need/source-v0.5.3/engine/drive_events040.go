package engine

import "math"

type DriverEvent040 struct {
	Kind           string
	Value, Seconds float64
	XP             uint64
}
type passState040 struct {
	Relative          Vec
	Heading, Cooldown float64
	Variant, Path     int
	Valid             bool
}
type DriveProgress040 struct {
	World                   *World
	ProfileID               string
	Drift                   DriftRun040
	LastDrift, LastLap      float64
	WasAirborne, Finished   bool
	AirSeconds, AirDistance float64
	passes                  map[int]passState040
}

// Observe simulation steps, never interpolated/generated graphics frames.
// A near pass requires crossing a nearby vehicle's longitudinal position;
// merely parking beside traffic or repeatedly touching it cannot earn points.
func (s *DriveProgress040) Step040(w *World, p *DriverProfile040, dt float64) []DriverEvent040 {
	if w == nil || w.Remaster == nil || p == nil || dt <= 0 {
		return nil
	}
	if s.World != w || s.ProfileID != p.ID {
		*s = DriveProgress040{World: w, ProfileID: p.ID, LastDrift: w.Car.DriftScore, passes: map[int]passState040{}}
	}
	var events []DriverEvent040
	c := &w.Car
	free := w.Remaster.Mode == 0
	driftDelta := math.Max(0, c.DriftScore-s.LastDrift)
	s.LastDrift = c.DriftScore
	if !c.OnRoad {
		driftDelta = 0
	}
	if result, complete := s.Drift.Step040(c.OnRoad && c.DriftActive && math.Abs(c.Speed) > 5 && math.Abs(c.DriftAngle) > .08, driftDelta, dt); complete {
		xp := p.RecordDrift040(result.Score, result.Seconds, free)
		events = append(events, DriverEvent040{"DRIFT", result.Score, result.Seconds, xp})
	}
	if !c.Body.Grounded {
		if !s.WasAirborne {
			s.AirSeconds = 0
			s.AirDistance = 0
		}
		s.WasAirborne = true
		s.AirSeconds += dt
		s.AirDistance = math.Max(s.AirDistance, c.Body.JumpDistance)
	} else if s.WasAirborne {
		xp := p.RecordStunt040(s.AirDistance, s.AirSeconds, free)
		if xp > 0 {
			events = append(events, DriverEvent040{"AIRTIME", s.AirDistance, s.AirSeconds, xp})
		}
		s.WasAirborne = false
	}
	forward := HeadingVec(c.Heading)
	for i, t := range w.Traffic {
		state := s.passes[i]
		state.Cooldown = math.Max(0, state.Cooldown-dt)
		relative := t.Pos.Sub(c.Pos)
		along := relative.X*forward.X + relative.Z*forward.Z
		if relative.Length() > 60 {
			state.Valid = false
			s.passes[i] = state
			continue
		}
		stable := state.Valid && state.Variant == t.Variant && state.Path == t.Path && state.Relative.Dist(relative) < 8 && math.Abs(math.Remainder(c.Heading-state.Heading, Tau)) < .25
		if stable && state.Cooldown == 0 && math.Abs(c.Speed)*3.6 >= 100 && t.Speed > 3 && c.OnRoad && c.ContactImpulse < .1 && t.ImpactCooldown <= 0 {
			before := state.Relative.X*forward.X + state.Relative.Z*forward.Z
			if before > 0 && along <= 0 {
				f := before / math.Max(.0001, before-along)
				cross := state.Relative.Add(relative.Sub(state.Relative).Mul(f))
				gap := cross.Length()
				if gap > 1.9 && gap < 3.25 && math.Abs(w.GroundHeight(t.Pos)-w.GroundHeight(c.Pos)) < 3 {
					if counted, xp := p.RecordNearPass040(math.Abs(c.Speed)*3.6, free); counted {
						events = append(events, DriverEvent040{"NEAR PASS", 1, 0, xp})
						state.Cooldown = 8
					}
				}
			}
		}
		state.Relative, state.Heading, state.Variant, state.Path, state.Valid = relative, c.Heading, t.Variant, t.Path, true
		s.passes[i] = state
	}
	if w.Remaster.Mode == 1 {
		if best := w.Remaster.BestLap; best > 0 && best != s.LastLap {
			if p.RecordLap040(best, w.Remaster.RaceDifficulty040) == nil {
				events = append(events, DriverEvent040{"LAP", best, 0, 0})
			}
			s.LastLap = best
		}
		if w.Remaster.Finished && !s.Finished {
			if p.RecordRaceResult040(w.Remaster.Position, w.Remaster.RaceDifficulty040) == nil {
				s.Finished = true
				events = append(events, DriverEvent040{"RACE FINISH", float64(w.Remaster.Position), 0, 0})
			}
		}
	}
	return events
}

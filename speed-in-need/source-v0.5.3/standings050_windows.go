package main

import (
	"fmt"
	"sort"
	"speedinneed/engine"
)

type standing050 struct {
	Name            string
	Local, Finished bool
	Time, Progress  float64
}

func standings050() []standing050 {
	q := &multiplayer050
	if q.Session == nil || app.world == nil || app.world.Remaster == nil {
		return nil
	}
	r := app.world.Remaster
	name := "YOU"
	if p := activeProfile040(); p != nil {
		name = p.Name
	}
	rows := []standing050{{Name: name, Local: true, Finished: r.Finished, Time: r.FinishTime, Progress: float64(r.Lap-1) + app.world.NetworkRouteProgress050(app.world.Car.Pos)}}
	for _, p := range q.View.Players {
		if p.ID == q.Session.ID() || !p.Ready {
			continue
		}
		c := p.State
		// The host validates remote checkpoint/lap fields before publishing them.
		if q.Session.IsHost() {
			if state := q.Race[p.ID]; state != nil {
				c.Lap = state.Lap
				c.Finished = state.Finished
				c.RaceTime = state.Time
			}
		}
		rows = append(rows, standing050{Name: p.Name, Finished: c.Finished, Time: c.RaceTime, Progress: float64(engine.ClampInt(c.Lap, 1, r.Laps+1)-1) + app.world.NetworkRouteProgress050(engine.Vec{X: c.X, Z: c.Z})})
	}
	for i, a := range r.Racers {
		rows = append(rows, standing050{Name: fmt.Sprintf("AI %d", i+1), Finished: a.Finished, Time: a.FinishTime, Progress: float64(a.Lap-1) + app.world.NetworkRouteProgress050(a.Car.Pos)})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.Finished != b.Finished {
			return a.Finished
		}
		if a.Finished {
			return a.Time < b.Time
		}
		return a.Progress > b.Progress
	})
	return rows
}
func racePosition050() (int, int) {
	if rows := standings050(); len(rows) > 0 {
		for i, r := range rows {
			if r.Local {
				return i + 1, len(rows)
			}
		}
	}
	r := app.world.Remaster
	return r.Position, len(r.Racers) + 1
}
func drawLANResults050(w, h float64) bool {
	if multiplayer050.Session == nil {
		return false
	}
	rows := standings050()
	hudRect(w*.20, h*.12, w*.60, h*.76, .008, .018, .028, .96)
	remasterCenter("LAN RACE / LIVE STANDINGS", h*.17, h*.026, white)
	for i, r := range rows {
		y := h * (.25 + float64(i)*.043)
		name := r.Name
		if r.Local {
			name += " / YOU"
		}
		state := "RACING"
		if r.Finished {
			state = fmt.Sprintf("%.2f SEC", r.Time)
		}
		remasterText(fmt.Sprintf("%02d / %s", i+1, name), w*.24, y, h*.013, white)
		remasterText(state, w*.64, y, h*.012, cyan)
	}
	remasterCenter("F8 / LOBBY AND NEXT RACE", h*.82, h*.014, cyan)
	return true
}

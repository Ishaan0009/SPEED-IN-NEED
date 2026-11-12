package main

import (
	"math"
	"sort"
	"speedinneed/engine"
	"speedinneed/lan"
)

var trafficSmoothing051 = map[int]lan.Car{}

// Give each driver a fair share of the bounded traffic budget, instead of
// filling it with the first twelve cars in the host's population array.
func selectTraffic051(w *engine.World, players []lan.Player) []int {
	observers := []engine.Vec{w.Car.Pos}
	for _, p := range players {
		if multiplayer050.Session != nil && p.ID != multiplayer050.Session.ID() && p.Ready {
			observers = append(observers, engine.Vec{X: p.State.X, Z: p.State.Z})
		}
	}
	selected := []int{}
	used := map[int]bool{}
	share := 12 / len(observers)
	for _, p := range observers {
		candidates := []int{}
		for i, c := range w.Traffic {
			if !used[i] && c.Pos.Dist(p) < 350 {
				candidates = append(candidates, i)
			}
		}
		sort.Slice(candidates, func(i, j int) bool {
			return w.Traffic[candidates[i]].Pos.Dist(p) < w.Traffic[candidates[j]].Pos.Dist(p)
		})
		for _, i := range candidates {
			if share <= 0 || len(selected) >= 12 {
				break
			}
			used[i] = true
			selected = append(selected, i)
			share--
		}
		share = 12 / len(observers)
	}
	return selected
}
func smoothTraffic051(c lan.Car, dt float64) lan.Car {
	old, ok := trafficSmoothing051[c.Entity051]
	if ok && math.Hypot(c.X-old.X, c.Z-old.Z) < 35 {
		k := 1 - math.Exp(-engine.Clamp(dt, 0, .1)*22)
		c.X = engine.Lerp(old.X, c.X, k)
		c.Z = engine.Lerp(old.Z, c.Z, k)
		c.H = old.H + math.Atan2(math.Sin(c.H-old.H), math.Cos(c.H-old.H))*k
	}
	if len(trafficSmoothing051) > 1024 {
		trafficSmoothing051 = map[int]lan.Car{}
	}
	trafficSmoothing051[c.Entity051] = c
	return c
}

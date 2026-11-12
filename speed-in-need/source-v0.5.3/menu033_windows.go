package main

import (
	"math"
	"speedinneed/engine"
	"strings"
)

var focus033 struct {
	ID            string
	Screen, Stage int
	Ready         bool
}

func buttonAt033(x, y int) string {
	for _, b := range ui020.Buttons {
		if float64(x) >= b.X && float64(x) <= b.X+b.W && float64(y) >= b.Y && float64(y) <= b.Y+b.H {
			return b.ID
		}
	}
	return ""
}
func syncMenuFocus033() {
	if focus033.Ready && focus033.Screen == beta.screen && focus033.Stage == ui020.Stage {
		return
	}
	focus033.Ready = true
	focus033.Screen, focus033.Stage = beta.screen, ui020.Stage
	focus033.ID = "next"
	if beta.screen == screenTitle {
		focus033.ID = "start"
	}
	if beta.screen == screenPause {
		focus033.ID = "resume"
	}
	if beta.screen == screenMaps {
		switch ui020.Stage {
		case 0:
			focus033.ID = []string{"mode0", "mode1"}[engine.ClampInt(remaster.Mode, 0, 1)]
		case 2:
			focus033.ID = "car+"
		case 3:
			focus033.ID = []string{"day", "afternoon", "night"}[engine.ClampInt(remaster.Time, 0, 2)]
		}
	}
}
func moveMenuFocus033(k uintptr) {
	if len(ui020.Buttons) == 0 {
		return
	}
	at := -1
	for i, b := range ui020.Buttons {
		if b.ID == focus033.ID {
			at = i
			break
		}
	}
	if at < 0 {
		focus033.ID = ui020.Buttons[0].ID
		return
	}
	if k == 9 {
		focus033.ID = ui020.Buttons[(at+1)%len(ui020.Buttons)].ID
		return
	}
	b := ui020.Buttons[at]
	bx, by := b.X+b.W*.5, b.Y+b.H*.5
	best, idx := math.Inf(1), -1
	for i, c := range ui020.Buttons {
		if i == at {
			continue
		}
		dx, dy := c.X+c.W*.5-bx, c.Y+c.H*.5-by
		primary, side := dy, math.Abs(dx)
		if k == 0x25 || k == 0x27 {
			primary, side = dx, math.Abs(dy)
		}
		if k == 0x25 || k == 0x26 {
			primary = -primary
		}
		if primary < 2 {
			continue
		}
		score := primary + side*2.6 + side*side/math.Max(40, primary)
		if score < best {
			best, idx = score, i
		}
	}
	if idx >= 0 {
		focus033.ID = ui020.Buttons[idx].ID
	}
}
func menuKey033(k uintptr) bool {
	if beta.screen != screenTitle && beta.screen != screenMaps && beta.screen != screenPause {
		return false
	}
	syncMenuFocus033()
	if k == VK_ESCAPE {
		if beta.screen == screenTitle {
			beta020Action("quit")
		} else if beta.screen == screenPause {
			beta020Action("resume")
		} else {
			beta020Action("back")
		}
		return true
	}
	if beta.screen == screenMaps {
		if k == 'Q' || k == 'E' {
			if ui020.Stage == 2 {
				delta := 1
				if k == 'Q' {
					delta = -1
				}
				remasterCycle(3, delta)
			}
			saveRemasterSettings()
			return true
		}
		if k == 0x25 || k == 0x27 {
			delta := 1
			if k == 0x25 {
				delta = -1
			}
			id := focus033.ID
			if ui020.Stage == 2 && (strings.HasPrefix(id, "car") || id == "") {
				cycleCar051(delta)
				preview014.ready = false
				saveRemasterSettings()
				return true
			}
			if strings.HasPrefix(id, "paint") {
				remasterCycle(3, delta)
				saveRemasterSettings()
				return true
			}
			if ui020.Stage == 0 && strings.HasPrefix(id, "mode") {
				remaster.Mode = (remaster.Mode + 1) % 2
				beta.mode, beta.mapSelect = remaster.Mode, remaster.Mode
				focus033.ID = []string{"mode0", "mode1"}[remaster.Mode]
				saveRemasterSettings()
				return true
			}
			if ui020.Stage == 3 && (id == "day" || id == "afternoon" || id == "night") {
				remaster.Time = (remaster.Time + 3 + delta) % 3
				focus033.ID = []string{"day", "afternoon", "night"}[remaster.Time]
				saveRemasterSettings()
				return true
			}
		}
	}
	if k == VK_RETURN || k == VK_SPACE {
		id := focus033.ID
		if id == "" {
			id = "next"
		}
		if beta.screen == screenMaps && (id == "next" || strings.HasPrefix(id, "car") || strings.HasPrefix(id, "mode") || id == "day" || id == "afternoon" || id == "night") {
			if strings.HasPrefix(id, "mode") || id == "day" || id == "afternoon" || id == "night" {
				beta020Action(id)
			}
			beta020Advance()
			return true
		}
		beta020Action(id)
		return true
	}
	if k == 9 || k >= 0x25 && k <= 0x28 {
		moveMenuFocus033(k)
		return true
	}
	return false
}
func menuHover033(x, y int) {
	if beta.screen == screenSettings {
		aokawaHover(x, y)
		return
	}
	syncMenuFocus033()
	if id := buttonAt033(x, y); id != "" {
		focus033.ID = id
	}
	if beta.screen == screenPause {
		for i, b := range ui020.Buttons {
			if b.ID == focus033.ID {
				beta.selection = i
				break
			}
		}
	}
}

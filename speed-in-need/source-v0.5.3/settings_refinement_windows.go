package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
	"strings"
)

var groupsRefine = []string{"DISPLAY", "GRAPHICS", "CONTROLS", "AUDIO", "MAP / GAMEPLAY"}

type settingRefine struct{ Tab, Row int }

var settingsRefine struct {
	Group, Cursor, First, LastTab, LastRow int
	Category                               bool
}

func init() {
	settingNames[4] = append(settingNames[4], "INVERT LOOK Y", "AUTO-CENTER CAMERA")
	settingNames[3] = append(settingNames[3], "SKY DETAIL / STREAMED HDR")
}
func groupRowsRefine(group int) []settingRefine {
	tabs := [][]int{{0}, {8, 1, 2, 3}, {4, 7}, {5}, {6}}
	var rows []settingRefine
	for _, tab := range tabs[engine.ClampInt(group, 0, 4)] {
		for row := range settingNames[tab] {
			if tab == 0 && row == 7 {
				continue
			}
			rows = append(rows, settingRefine{tab, row})
		}
	}
	if group == 1 {
		rows = append(rows, settingRefine{6, 1})
	}
	if group == 4 {
		rows = append(rows, settingRefine{0, 7})
	}
	filtered := rows[:0]
	for _, r := range rows {
		if group == 4 && r.Tab == 6 && r.Row == 1 {
			continue
		}
		filtered = append(filtered, r)
	}
	rows = filtered
	return rows
}
func syncSettingsRefine() {
	q := &settingsRefine
	if remaster.SettingsTab == q.LastTab && beta.selection == q.LastRow {
		return
	}
	for group := 0; group < 5; group++ {
		for i, row := range groupRowsRefine(group) {
			if row.Tab == remaster.SettingsTab && row.Row == beta.selection {
				q.Group, q.Cursor, q.First = group, i, 0
				q.LastTab, q.LastRow = remaster.SettingsTab, beta.selection
				return
			}
		}
	}
}
func selectSettingsRefine() {
	q := &settingsRefine
	rows := groupRowsRefine(q.Group)
	q.Cursor = engine.ClampInt(q.Cursor, 0, len(rows)-1)
	r := rows[q.Cursor]
	remaster.SettingsTab, beta.selection = r.Tab, r.Row
	q.LastTab, q.LastRow = r.Tab, r.Row
	if q.Cursor < q.First {
		q.First = q.Cursor
	}
	if q.Cursor >= q.First+9 {
		q.First = q.Cursor - 8
	}
}
func stepSettingsGroupRefine(delta int) {
	syncSettingsRefine()
	q := &settingsRefine
	q.Group = (q.Group + 5 + delta) % 5
	q.Cursor, q.First = 0, 0
	selectSettingsRefine()
}
func settingsKeyRefine(k uintptr) bool {
	syncSettingsRefine()
	q := &settingsRefine
	if q.Category {
		switch k {
		case 0x26, 0x25:
			stepSettingsGroupRefine(-1)
		case 0x28, 0x27:
			stepSettingsGroupRefine(1)
		case VK_RETURN:
			q.Category = false
			selectSettingsRefine()
		default:
			return false
		}
		return true
	}
	n := len(groupRowsRefine(q.Group))
	switch k {
	case 0x26:
		q.Cursor = (q.Cursor + n - 1) % n
		selectSettingsRefine()
	case 0x28:
		q.Cursor = (q.Cursor + 1) % n
		selectSettingsRefine()
	case 0x25:
		aokawaAdjustSetting(-1)
	case 0x27, VK_RETURN:
		aokawaAdjustSetting(1)
	default:
		return false
	}
	return true
}
func drawSettings033() {
	syncSettingsRefine()
	w, h := float64(app.width), float64(app.height)
	q := &settingsRefine
	// Two readable columns: compact category rail, then one ordered option list.
	roundRect032(w*.045, h*.065, w*.91, h*.87, 4, [4]float64{.025, .035, .042, .96})
	remasterText("SETTINGS", w*.073, h*.105, h*.027, white)
	if q.Category {
		remasterText("CHOOSE A CATEGORY", w*.073, h*.18, h*.011, mutedRefine)
		for i, name := range groupsRefine {
			y := h * (.29 + float64(i)*.105)
			color := mutedRefine
			if i == q.Group {
				color = white
				roundRect032(w*.073, y-h*.015, w*.80, h*.078, 4, [4]float64{.13, .18, .19, .85})
				hudRect(w*.073, y-h*.015, 3, h*.078, cyan[0], cyan[1], cyan[2], 1)
			}
			remasterText(name, w*.10, y, h*.018, color)
			remasterText(">", w*.84, y, h*.018, color)
		}
		headline035("SELECT: ARROWS / STICK     OPEN: ENTER / A     BACK: ESC / B", w*.073, h*.86, h*.010, w*.8, mutedRefine)
		return
	}
	for i, name := range groupsRefine {
		y := h * (.24 + float64(i)*.095)
		c := mutedRefine
		if i == q.Group {
			c = white
			hudRect(w*.068, y-7, 3, h*.05, cyan[0], cyan[1], cyan[2], 1)
		}
		headline035(name, w*.082, y, h*.014, w*.18, c)
	}
	rows := groupRowsRefine(q.Group)
	values := map[int][]string{}
	oldTab, oldRow := remaster.SettingsTab, beta.selection
	for row := 0; row < 9 && q.First+row < len(rows); row++ {
		i := q.First + row
		entry := rows[i]
		remaster.SettingsTab, beta.selection = entry.Tab, entry.Row
		if values[entry.Tab] == nil {
			values[entry.Tab] = aokawaSettingValues()
		}
		y := h * (.22 + float64(row)*.061)
		if i == q.Cursor {
			roundRect032(w*.285, y-8, w*.63, h*.05, 3, [4]float64{.12, .16, .17, .78})
			hudRect(w*.285, y-8, 2, h*.05, cyan[0], cyan[1], cyan[2], 1)
		}
		color := white
		if entry.Tab == 8 && graphicsRowDisabled038(entry.Row) {
			color = mutedRefine
		}
		label := settingNames[entry.Tab][entry.Row]
		if q.Group == 1 {
			label = tabs033[entry.Tab] + " / " + label
		}
		headline035(label, w*.305, y, h*.011, w*.31, color)
		value := values[entry.Tab][entry.Row]
		headline035(value, w*.66, y, h*.011, w*.235, cyan)
		if fraction, ok := settingLevel041(value); ok {
			hudRect(w*.66, y+h*.023, w*.17, 2, mutedRefine[0], mutedRefine[1], mutedRefine[2], .3)
			hudRect(w*.66, y+h*.023, w*.17*fraction, 2, cyan[0], cyan[1], cyan[2], .9)
			remasterText("-", w*.65, y+h*.017, h*.006, mutedRefine)
			remasterText("+", w*.84, y+h*.017, h*.006, mutedRefine)
		}
	}
	remaster.SettingsTab, beta.selection = oldTab, oldRow
	remasterText(fmt.Sprintf("%02d / %02d", q.Cursor+1, len(rows)), w*.85, h*.81, h*.010, mutedRefine)
	if beta.bindWaiting >= 0 || pad033.BindWaiting >= 0 {
		remasterText("PRESS A CONTROL / ESC CANCELS", w*.305, h*.86, h*.011, cyan)
	} else {
		headline035("ARROWS: SELECT / CHANGE    TAB OR LB/RB: CATEGORY    ESC: BACK", w*.305, h*.86, h*.010, w*.59, mutedRefine)
	}
}
func settingsClick033(x, y int) {
	w, h := float64(app.width), float64(app.height)
	fx, fy := float64(x)/w, float64(y)/h
	if settingsRefine.Category {
		if fx >= .073 && fx <= .873 && fy >= .275 && fy < .80 {
			settingsRefine.Group = engine.ClampInt(int((fy-.275)/.105), 0, 4)
			settingsRefine.Cursor, settingsRefine.First = 0, 0
			settingsRefine.Category = false
			selectSettingsRefine()
		}
		return
	}
	if fx >= .068 && fx < .27 && fy >= .22 && fy < .72 {
		settingsRefine.Group = engine.ClampInt(int((fy-.22)/.095), 0, 4)
		settingsRefine.Cursor, settingsRefine.First = 0, 0
		selectSettingsRefine()
		return
	}
	if fx >= .285 && fx < .915 && fy >= .212 && fy < .765 {
		syncSettingsRefine()
		row := int(math.Floor((fy-.212)/.061)) + settingsRefine.First
		if row < len(groupRowsRefine(settingsRefine.Group)) {
			settingsRefine.Cursor = row
			selectSettingsRefine()
			delta := 1
			if fx < .64 {
				delta = -1
			}
			aokawaAdjustSetting(delta)
		}
		return
	}
	if fy > .84 {
		aokawaSettingsKey(VK_ESCAPE)
	}
}

func settingLevel041(value string) (float64, bool) {
	v := strings.ToUpper(strings.TrimSpace(value))
	if strings.Contains(v, "UNAVAILABLE") || strings.Contains(v, "NOT ") || strings.Contains(v, "CONNECT") {
		return 0, false
	}
	if v == "OFF" || v == "ON" {
		return 0, false
	}
	if strings.HasSuffix(v, "%") {
		var n float64
		if _, err := fmt.Sscanf(v, "%f%%", &n); err == nil {
			return engine.Clamp(n/100, 0, 1), true
		}
	}
	return 0, false
}
func garagePadRefine(s padSample033, dt float64) {
	if !garageView037() {
		return
	}
	x, y := padDeadzone033(s.Axes[2], .18), padDeadzone033(s.Axes[3], .18)
	if math.Abs(x)+math.Abs(y) < .01 {
		return
	}
	garage037.Touched = true
	if s.Buttons&0x80 != 0 {
		garageZoom037(-y * engine.Clamp(dt, 0, .05) * .65)
		return
	}
	garage037.Yaw = math.Remainder(garage037.Yaw-x*engine.Clamp(dt, 0, .05)*1.6, engine.Tau)
	garage037.Pitch = engine.Clamp(garage037.Pitch-y*engine.Clamp(dt, 0, .05)*.35, -.24, .48)
}

package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
	"time"
)

var mutedRefine = [4]float64{.66, .71, .73, 1}

func titleEntryRefine(id, label string, x, y, w, h float64) {
	ui020.Buttons = append(ui020.Buttons, button020{id, x, y, w, h})
	active := focus033.ID == id || id == "start" && focus033.ID == ""
	hover := float64(remaster.MouseX) >= x && float64(remaster.MouseX) <= x+w && float64(remaster.MouseY) >= y && float64(remaster.MouseY) <= y+h
	target := 0.
	if active || hover {
		target = 1
	}
	ui021.Hover[id] += (target - ui021.Hover[id]) * (1 - math.Exp(-ui021.DT*16))
	e := ui021.Hover[id]
	hudRect(x, y+h-2, w*(.1+.9*e), 1, .79, .94, .57, .2+.7*e)
	color := mutedRefine
	if active || hover {
		color = white
	}
	remasterText(label, x+12*e, y+h*.22, h*.30, color)
	if active || hover {
		remasterText(">", x+w-h*.35, y+h*.22, h*.30, cyan)
	}
}
func drawTitle038(w, h float64) {
	defer titleProfile040(w, h)
	dis(GL_TEXTURE_2D)
	begin(GL_QUADS)
	col(.012, .018, .022, .98)
	vtx(0, 0, 0)
	vtx(0, h, 0)
	col(.012, .018, .022, 0)
	vtx(w*.68, h, 0)
	vtx(w*.68, 0, 0)
	end()
	remasterText("S / N", w*.065, h*.105, h*.021, cyan)
	remasterText("OPEN ROAD EXPERIENCE", w*.065, h*.19, h*.011, mutedRefine)
	headline035("SPEED", w*.06, h*.285, h*.10, w*.39, white)
	headline035("IN NEED", w*.065, h*.405, h*.055, w*.37, cyan)
	remasterText("YOUR CAR. YOUR ROAD.", w*.068, h*.515, h*.012, mutedRefine)
	for i, item := range [][2]string{{"start", "START DRIVING"}, {"settings", "SETTINGS"}, {"quit", "EXIT"}} {
		titleEntryRefine(item[0], item[1], w*.068, h*(.59+float64(i)*.073), w*.25, h*.054)
	}
	remasterText("BETA 0.5.3", w*.068, h*.93, h*.010, mutedRefine)
}
func drawModeChoice033(w, h float64) {
	for i, item := range [][3]string{{"OPEN WORLD", "Explore cities, coast and highlands", "FREE ROAM"}, {"CIRCUIT RACING", "Choose your rivals and chase the flag", "RACE"}} {
		x := w * (.065 + float64(i)*.32)
		y := h * .31
		roundRect032(x, y, w*.29, h*.30, 4, [4]float64{.04, .055, .06, .70})
		remasterText(fmt.Sprintf("0%d / %s", i+1, item[2]), x+w*.018, y+h*.025, h*.010, cyan)
		beta021Icon(i, x+w*.25, y+h*.075, h*.055)
		beta020Button(fmt.Sprintf("mode%d", i), item[0], x+w*.01, y+h*.11, w*.27, h*.065, remaster.Mode == i)
		if p := activeProfile040(); i == 1 && p != nil && !p.CanRace040() {
			lockRefine(x+w*.018, y+h*.255, h*.018)
			remasterText("LOCKED / LEVEL 10", x+w*.04, y+h*.26, h*.010, mutedRefine)
		}
		headline035(item[1], x+w*.018, y+h*.215, h*.011, w*.25, mutedRefine)
	}
	x, y := w*.715, h*.34
	roundRect032(x, y, w*.22, h*.24, 4, [4]float64{.035, .045, .05, .80})
	beta020Button("multiplayer050", "MULTIPLAYER / LAN", x+w*.01, y+h*.07, w*.20, h*.065, false)
	remasterText("4 PLAYERS / SAME WI-FI", x+w*.023, y+h*.175, h*.009, cyan)
	if remaster.Mode == 1 {
		level := engine.ClampInt(beta.cfg.RaceDifficulty040, 0, 2)
		beta020Button("difficulty040", "RIVALS / "+[]string{"EASY", "NORMAL", "HARD"}[level], w*.385, h*.65, w*.29, h*.045, true)
	}
}
func lockRefine(x, y, size float64) {
	dis(GL_TEXTURE_2D)
	col(.60, .66, .69, 1)
	gl("glLineWidth", b32(2))
	begin(0x0003)
	for i := 0; i <= 20; i++ {
		a := math.Pi + float64(i)*math.Pi/20
		vtx(x+size*.5+math.Cos(a)*size*.30, y+size*.40+math.Sin(a)*size*.34, 0)
	}
	end()
	roundRect032(x, y+size*.35, size, size*.7, 3, [4]float64{.36, .42, .44, 1})
	gl("glLineWidth", b32(1))
}

var paintNamesRefine = []string{"VERMILION", "COBALT", "PEARL", "ONYX", "SILVER", "GRAPHITE", "SUNBURST", "TANGERINE", "JADE", "VIOLET", "GLACIER", "GARNET", "COPPER", "FOREST", "IRIS", "ROSE", "MIDNIGHT", "CHAMPAGNE", "BURGUNDY", "ATLANTIC", "BRITISH GREEN", "IVORY", "BASALT", "TITANIUM", "DEEP BLUE", "PETROL", "BRONZE", "CITRUS", "SAKURA", "SLATE", "OXIDE", "OLIVE", "AUBERGINE", "FROST", "CARBON", "PLATINUM"}

func drawCarChoice036(w, h float64) { drawCarRemastered051(w, h) }
func drawLoadingOrbitLegacy050(progress float64, label string) {
	drawMenuBackdrop036()
	w, h := float64(app.width), float64(app.height)
	progress = engine.Clamp(progress, 0, 1)
	t := time.Since(loadingStart036).Seconds()
	cx, cy, r := w*.50, h*.47, h*.066
	dis(GL_TEXTURE_2D)
	gl("glLineWidth", b32(3))
	col(.18, .23, .25, 1)
	begin(0x0003)
	for i := 0; i <= 80; i++ {
		a := -math.Pi/2 + float64(i)*engine.Tau/80
		vtx(cx+math.Cos(a)*r, cy+math.Sin(a)*r, 0)
	}
	end()
	col(cyan[0], cyan[1], cyan[2], 1)
	begin(0x0003)
	for i := 0; i <= int(progress*80); i++ {
		a := -math.Pi/2 + float64(i)*engine.Tau/80
		vtx(cx+math.Cos(a)*r, cy+math.Sin(a)*r, 0)
	}
	end()
	gl("glLineWidth", b32(1))
	a := t*1.8 - math.Pi/2
	push()
	move(cx+math.Cos(a)*r, cy+math.Sin(a)*r, 0)
	rotate(a*180/math.Pi+90, 0, 0, 1)
	// Original small top-view car, constructed from simple filled shapes.
	roundRect032(-h*.009, -h*.018, h*.018, h*.036, 3, white)
	hudRect(-h*.006, -h*.009, h*.012, h*.009, .06, .12, .16, 1)
	hudRect(-h*.006, h*.005, h*.012, h*.006, .06, .12, .16, 1)
	pop()
	text := fmt.Sprintf("%02d", int(progress*100))
	remasterText(text, cx-remasterTextWidth(text, h*.022)/2, cy-h*.011, h*.022, white)
	remasterText("GETTING READY", cx-remasterTextWidth("GETTING READY", h*.014)/2, h*.59, h*.014, white)
	headline035(label, w*.32, h*.64, h*.010, w*.36, mutedRefine)
	remasterText("SPEED IN NEED", w*.065, h*.09, h*.018, white)
}

var hudRefine struct {
	Ready             bool
	Speed, RPM, Nitro float64
}

func drawSpeed033(c engine.Car) {
	if walker051.Active {
		w, h := float64(app.width), float64(app.height)
		headline035("ON FOOT / WASD: WALK / E: ENTER CAR", w*.64, h*.92, h*.011, w*.32, white)
		return
	}

	w, h := float64(app.width), float64(app.height)
	k := 1 - math.Exp(-engine.Clamp(navFrameDT033, .001, .1)*18)
	if !hudRefine.Ready {
		hudRefine.Speed = math.Abs(c.Speed) * 3.6
		hudRefine.Nitro = c.Nitro
		hudRefine.Ready = true
	}
	hudRefine.Speed += (math.Abs(c.Speed)*3.6 - hudRefine.Speed) * k
	hudRefine.RPM += (c.Gearbox.RPM/9000 - hudRefine.RPM) * k
	hudRefine.Nitro += (c.Nitro - hudRefine.Nitro) * k
	x, y, width := w*.835, h*.805, w*.12
	label := fmt.Sprintf("%03d", int(hudRefine.Speed+.5))
	remasterText(label, x, y, h*.065, white)
	remasterText("KM/H", x, y+h*.078, h*.010, mutedRefine)
	gear := "N"
	if c.Gearbox.Gear < 0 {
		gear = "R"
	} else if c.Gearbox.Gear > 0 {
		gear = fmt.Sprint(c.Gearbox.Gear)
	}
	remasterText(gear, x+width*.82, y+h*.048, h*.031, cyan)
	for i := 0; i < 24; i++ {
		fill := engine.Clamp(hudRefine.RPM*24-float64(i), 0, 1)
		color := mutedRefine
		if i >= 19 {
			color = [4]float64{1, .43, .37, 1}
		}
		hudRect(x+float64(i)*width/24, y-h*.035, width/24-3, h*.012, color[0], color[1], color[2], .18+.82*fill)
	}
	remasterText("NITRO", x, y+h*.115, h*.009, mutedRefine)
	for i := 0; i < 12; i++ {
		fill := engine.Clamp(hudRefine.Nitro*.12-float64(i), 0, 1)
		alpha := .18 + .72*fill
		if c.Boosting {
			alpha = math.Min(1, alpha+.10*(1+math.Sin(app.t*9-float64(i)*.5)))
		}
		hudRect(x+float64(i)*width/12, y+h*.140, width/12-3, h*.006, cyan[0], cyan[1], cyan[2], alpha)
	}
	if autopilot034 {
		remasterText("AUTOPILOT / P TO RELEASE", w*.40, h*.07, h*.011, cyan)
	}
}

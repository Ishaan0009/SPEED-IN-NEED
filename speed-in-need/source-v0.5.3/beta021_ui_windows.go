package main

import (
	"fmt"
	"math"
	"time"
)

var ui021 struct {
	Stage, Screen int
	Start, Last   time.Time
	DT            float64
	Hover         map[string]float64
}

func fmtSpeed021(value float64) string { return fmt.Sprintf("%.0f", value) }

func beta021UIFrame() {
	now := time.Now()
	ui021.DT = math.Min(.05, now.Sub(ui021.Last).Seconds())
	ui021.Last = now
	if ui021.Start.IsZero() || ui021.Stage != ui020.Stage || ui021.Screen != beta.screen {
		ui021.Start = now
		ui021.Stage = ui020.Stage
		ui021.Screen = beta.screen
	}
	if ui021.Hover == nil {
		ui021.Hover = map[string]float64{}
	}
}

func beta021Panel(x, y, w, h, alpha float64) {
	dis(GL_TEXTURE_2D)
	bevel := math.Min(4, math.Min(w, h)*.12)
	points := [][2]float64{{x + bevel, y}, {x + w - bevel, y}, {x + w, y + bevel}, {x + w, y + h - bevel}, {x + w - bevel, y + h}, {x + bevel, y + h}, {x, y + h - bevel}, {x, y + bevel}}
	begin(0x0006)
	col(.055, .065, .085, alpha)
	vtx(x+w*.5, y+h*.5, 0)
	for i := 0; i <= len(points); i++ {
		p := points[i%len(points)]
		t := (p[1] - y) / math.Max(1, h)
		col(.025-.012*t, .055-.025*t, .075-.03*t, alpha)
		vtx(p[0], p[1], 0)
	}
	end()
	gl("glLineWidth", b32(1))
	col(.24, .53, .56, .34*alpha)
	begin(0x0002)
	for _, p := range points {
		vtx(p[0], p[1], 0)
	}
	end()
}

func beta021Icon(kind int, x, y, size float64) {
	dis(GL_TEXTURE_2D)
	gl("glLineWidth", b32(2))
	col(1, .57, .29, .9)
	if kind == 0 {
		begin(0x0002)
		for _, p := range [][2]float64{{-.45, -.35}, {-.08, -.5}, {.4, -.29}, {.47, .17}, {.08, .48}, {-.4, .3}} {
			vtx(x+p[0]*size, y+p[1]*size, 0)
		}
		end()
		begin(0x0003)
		for i := 0; i < 20; i++ {
			t := float64(i) / 19
			vtx(x+math.Sin(t*7)*size*.22, y+(t-.5)*size*.6, 0)
		}
		end()
	}
	if kind == 1 {
		for j := 0; j < 3; j++ {
			for i := 0; i < 4; i++ {
				if (i+j)%2 == 0 {
					hudRect(x+float64(i-2)*size*.22, y+float64(j-1)*size*.22, size*.22, size*.22, 1, .57, .29, .9)
				}
			}
		}
		hudRect(x-size*.45, y-size*.27, 2, size*.95, 1, .57, .29, .9)
	}
	if kind == 2 {
		begin(0x0003)
		for i := 0; i <= 18; i++ {
			a := math.Pi + float64(i)/18*math.Pi
			vtx(x+math.Cos(a)*size*.24, y-size*.1+math.Sin(a)*size*.26, 0)
		}
		end()
		beta021Panel(x-size*.35, y-size*.1, size*.7, size*.5, .8)
	}
	if kind == 3 || kind == 4 {
		begin(0x0002)
		for i := 0; i < 28; i++ {
			a := float64(i) / 28 * math.Pi
			vtx(x+math.Cos(a)*size*.31, y-math.Sin(a)*size*.31, 0)
		}
		end()
		hudRect(x-size*.43, y, size*.86, 2, 1, .57, .29, .9)
		for i := 0; i < 5; i++ {
			a := math.Pi * (.1 + float64(i)*.2)
			begin(0x0001)
			vtx(x+math.Cos(a)*size*.4, y-math.Sin(a)*size*.4, 0)
			vtx(x+math.Cos(a)*size*.48, y-math.Sin(a)*size*.48, 0)
			end()
		}
	}
	if kind == 5 {
		begin(0x0003)
		for i := 0; i < 24; i++ {
			a := float64(i)/23*math.Pi*1.65 + math.Pi*.2
			vtx(x+math.Cos(a)*size*.35, y+math.Sin(a)*size*.35, 0)
		}
		end()
		hudRect(x+size*.2, y-size*.3, 3, 3, 1, .57, .29, 1)
	}
	gl("glLineWidth", b32(1))
}

func beta021MapChoice() {
	w, h := float64(app.width), float64(app.height)
	size := h * .56
	x := w * .15
	y := h * .23
	beta021Panel(x-14, y-14, size+28, size+28, .85)
	key, name := "beta020_atlas.png", "SPEED IN NEED"
	if remaster.Mode == 1 {
		key = "aokawa_circuit.png"
		name = "APEX CIRCUIT"
	}
	beta020TexturedPanel(key, x, y, size, size)
	remasterText(name, x+size+w*.05, h*.42, h*.032, white)
	hudRect(x+size+w*.05, h*.49, w*.17, 3, 1, .53, .23, 1)
}

func beta021LightingChoices() {
	w, h := float64(app.width), float64(app.height)
	names := []string{"GOLDEN HOUR", "SUNSET", "AFTER DARK"}
	hints := []string{"WARM LIGHT / CLEAR ROADS", "AMBER SKY / CITY LIGHTS", "MOONLIGHT / HEADLIGHTS"}
	for i, label := range names {
		x := w * (.07 + float64(i)*.295)
		beta021Panel(x, h*.27, w*.27, h*.43, .78)
		beta021Icon(i+3, x+w*.135, h*.405, h*.12)
		beta020Button([]string{"day", "afternoon", "night"}[i], label, x+w*.015, h*.55, w*.24, h*.075, remaster.Time == i && !beta.cfg.RandomTime040)
		remasterText(hints[i], x+w*.025, h*.655, h*.011, white)
	}
	rainLabel := "RAIN / OFF"
	if weather050.Rain {
		rainLabel = "RAIN / ON"
	}
	beta020Button("rain051", rainLabel, w*.34, h*.735, w*.32, h*.045, weather050.Rain)
	beta020Button("random-time040", "RANDOM / CHANGING SKY", w*.34, h*.79, w*.32, h*.04, beta.cfg.RandomTime040)
}

package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
)

func drawBeta22Tachometer(c engine.Car) {
	w, h := float64(app.width), float64(app.height)
	x, y, r := w*.911, h*.812, h*.108
	dis(GL_TEXTURE_2D)
	// Matte dial, fine perimeter and a calibrated RPM sweep.
	col(.012, .018, .026, .88)
	begin(0x0006)
	vtx(x, y, 0)
	for i := 0; i <= 80; i++ {
		a := float64(i) * math.Pi / 40
		vtx(x+math.Cos(a)*r, y+math.Sin(a)*r, 0)
	}
	end()
	gl("glLineWidth", b32(1.25))
	col(.80, .85, .87, .9)
	begin(0x0002)
	for i := 0; i < 80; i++ {
		a := float64(i) * math.Pi / 40
		vtx(x+math.Cos(a)*r, y+math.Sin(a)*r, 0)
	}
	end()
	angle := func(rpm float64) float64 { return (135 + rpm/9000*270) * math.Pi / 180 }
	for i := 0; i <= 45; i++ {
		rpm := float64(i) * 200
		a := angle(rpm)
		outer, inner := r*.94, r*.88
		if i%5 == 0 {
			inner = r * .79
		}
		if rpm >= 7000 {
			col(1, .23, .19, 1)
		} else {
			col(.88, .91, .91, .95)
		}
		gl("glLineWidth", b32(1.5))
		begin(0x0001)
		vtx(x+math.Cos(a)*inner, y+math.Sin(a)*inner, 0)
		vtx(x+math.Cos(a)*outer, y+math.Sin(a)*outer, 0)
		end()
		if i%5 == 0 {
			label := fmt.Sprint(i / 5)
			sz := h * .012
			tx := x + math.Cos(a)*r*.66
			ty := y + math.Sin(a)*r*.66
			remasterText(label, tx-remasterTextWidth(label, sz)*.5, ty-sz*.5, sz, white)
		}
	}
	a := angle(engine.Clamp(c.Gearbox.RPM, 0, 9000))
	col(1, .23, .16, 1)
	gl("glLineWidth", b32(2.6))
	begin(0x0001)
	vtx(x-math.Cos(a)*r*.12, y-math.Sin(a)*r*.12, 0)
	vtx(x+math.Cos(a)*r*.77, y+math.Sin(a)*r*.77, 0)
	end()
	remasterText("RPM X 1000", x-r*.38, y-r*.33, h*.008, white)
	speed := fmt.Sprintf("%03d", int(math.Abs(c.Speed)*3.6))
	sz := h * .032
	beta021Panel(x-r*.51, y+r*.16, r*1.03, r*.42, .96)
	remasterText(speed, x-remasterTextWidth(speed, sz)*.5, y+r*.20, sz, white)
	remasterText("KM/H", x-r*.20, y+r*.61, h*.009, white)
	gear := fmt.Sprint(c.Gearbox.Gear)
	if c.Gearbox.Gear < 0 {
		gear = "R"
	}
	remasterText(gear, x+r*.34, y-r*.08, h*.019, cyan)
	hudRect(x-r, y+r*1.16, r*2, 4, .12, .15, .19, .9)
	hudRect(x-r, y+r*1.16, r*2*engine.Clamp(c.Nitro/100, 0, 1), 4, .18, .85, .84, 1)
	remasterText("NITRO", x-r, y+r*1.25, h*.009, cyan)
	gl("glLineWidth", b32(1))
}

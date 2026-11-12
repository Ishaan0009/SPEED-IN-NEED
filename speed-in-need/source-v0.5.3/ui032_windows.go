package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
)

func roundRect032(x, y, w, h, r float64, c [4]float64) {
	dis(GL_TEXTURE_2D)
	col(c[0], c[1], c[2], c[3])
	begin(0x0006)
	vtx(x+w*.5, y+h*.5, 0)
	for corner := 0; corner <= 4; corner++ {
		i := corner % 4
		cx, cy := x+r, y+r
		if i == 1 || i == 2 {
			cx = x + w - r
		}
		if i == 2 || i == 3 {
			cy = y + h - r
		}
		start := math.Pi + float64(i)*math.Pi*.5
		for step := 0; step <= 8; step++ {
			a := start + float64(step)*math.Pi/16
			vtx(cx+math.Cos(a)*r, cy+math.Sin(a)*r, 0)
		}
	}
	end()
}
func drawTitle032(w, h float64) {
	dis(GL_TEXTURE_2D)
	begin(GL_QUADS)
	col(.012, .020, .030, .96)
	vtx(0, 0, 0)
	vtx(0, h, 0)
	col(.012, .020, .030, 0)
	vtx(w*.70, h, 0)
	vtx(w*.70, 0, 0)
	end()
	remasterText("SPEED", w*.07, h*.19, h*.095, white)
	remasterText("IN NEED", w*.07, h*.30, h*.080, white)
	remasterText("SPEED IN NEED", w*.075, h*.43, h*.021, cyan)
	remasterText("FIND YOUR NEXT ROAD", w*.075, h*.51, h*.015, [4]float64{.67, .73, .77, 1})
	beta020Button("start", "DRIVE", w*.07, h*.64, w*.25, h*.082, true)
	beta020Button("settings", "SETTINGS", w*.07, h*.744, w*.12, h*.055, false)
	beta020Button("quit", "EXIT", w*.20, h*.744, w*.12, h*.055, false)
	remasterText("BETA 0.5.3", w*.075, h*.93, h*.011, [4]float64{.48, .55, .60, 1})
}
func drawSpeed032(c engine.Car) {
	w, h := float64(app.width), float64(app.height)
	cx, cy, r := w*.90, h*.81, h*.105
	dis(GL_TEXTURE_2D)
	gl("glLineWidth", b32(2.2))
	begin(0x0003)
	for i := 0; i <= 60; i++ {
		a := (145 + float64(i)*250/60) * math.Pi / 180
		col(.72, .78, .82, .26)
		vtx(cx+math.Cos(a)*r, cy+math.Sin(a)*r, 0)
	}
	end()
	fill := engine.Clamp(c.Gearbox.RPM/8000, 0, 1)
	begin(0x0003)
	for i := 0; i <= int(fill*60); i++ {
		a := (145 + float64(i)*250/60) * math.Pi / 180
		col(.9, .96, .97, .92)
		if i > 51 {
			col(1, .32, .24, 1)
		}
		vtx(cx+math.Cos(a)*r, cy+math.Sin(a)*r, 0)
	}
	end()
	label := fmt.Sprintf("%d", int(math.Abs(c.Speed)*3.6))
	size := h * .061
	remasterText(label, cx-remasterTextWidth(label, size)*.5, cy-h*.014, size, white)
	gear := fmt.Sprint(c.Gearbox.Gear)
	if c.Gearbox.Gear < 0 {
		gear = "R"
	}
	remasterText(gear, cx-r*.65, cy-r*.45, h*.024, cyan)
	remasterText("KM/H", cx-r*.02, cy-r*.40, h*.012, [4]float64{.75, .82, .84, 1})
	hudRect(cx-r*.74, cy+r*.96, r*1.48, 3, .20, .25, .28, .6)
	hudRect(cx-r*.74, cy+r*.96, r*1.48*engine.Clamp(c.Nitro/100, 0, 1), 3, .30, .82, .83, .9)
	gl("glLineWidth", b32(1))
}
func drawRoadMinimap032(c engine.Car) {
	x, y, size := beta011MiniRect()
	cx, cy, r := x+size*.5, y+size*.5, size*.48
	dis(GL_TEXTURE_2D)
	col(.012, .020, .028, .42)
	begin(0x0006)
	vtx(cx, cy, 0)
	for i := 0; i <= 64; i++ {
		a := float64(i) * engine.Tau / 64
		vtx(cx+math.Cos(a)*r, cy+math.Sin(a)*r, 0)
	}
	end()
	span := 360 + math.Min(140, math.Abs(c.Speed)*2)
	scale := size / span
	f := engine.HeadingVec(c.Heading)
	right := engine.Vec{X: f.Z, Z: -f.X}
	gl("glLineWidth", b32(2.3))
	col(.88, .92, .92, .90)
	begin(0x0001)
	for _, road := range app.world.Remaster.Def.Roads {
		for i := 0; i+1 < len(road.Points); i++ {
			a := road.Points[i].Sub(c.Pos)
			b := road.Points[i+1].Sub(c.Pos)
			ax, ay := (a.X*right.X+a.Z*right.Z)*scale, -(a.X*f.X+a.Z*f.Z)*scale
			bx, by := (b.X*right.X+b.Z*right.Z)*scale, -(b.X*f.X+b.Z*f.Z)*scale
			dx, dy := bx-ax, by-ay
			den := dx*dx + dy*dy
			if den < 1e-8 {
				continue
			}
			bb := 2 * (ax*dx + ay*dy)
			cc := ax*ax + ay*ay - r*r
			disc := bb*bb - 4*den*cc
			if disc < 0 {
				continue
			}
			root := math.Sqrt(disc)
			lo := math.Max(0, (-bb-root)/(2*den))
			hi := math.Min(1, (-bb+root)/(2*den))
			if lo > hi {
				continue
			}
			vtx(cx+ax+dx*lo, cy+ay+dy*lo, 0)
			vtx(cx+ax+dx*hi, cy+ay+dy*hi, 0)
		}
	}
	end()
	col(.25, .88, .89, 1)
	begin(GL_TRIANGLES)
	vtx(cx, cy-9, 0)
	vtx(cx-6, cy+6, 0)
	vtx(cx+6, cy+6, 0)
	end()
	gl("glLineWidth", b32(1))
}

package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
	"unsafe"
)

var beta011InstanceScaleLoc int32

func drawBeta011NPC(t engine.Traffic) {
	old := beta.paint
	beta.paint = (t.Variant * 7) % len(paintPalette)
	drawPBRAsset(fmt.Sprintf("b011_npc%d", (t.Variant-4)%10), 0)
	beta.paint = old
}
func beta011CelestialDirection() pos3 {
	hour := sceneHour()
	angle := (hour - 6) / 24 * engine.Tau
	sun := norm(pos3{math.Cos(angle) * .83, math.Max(.035, math.Sin(angle)), .38})
	moon := norm(pos3{-math.Cos(angle) * .83, math.Max(.035, -math.Sin(angle)), .38})
	t := engine.Clamp((engine.NightAmount(hour)-.40)/.20, 0, 1)
	t = t * t * (3 - 2*t)
	return norm(add(scale(sun, 1-t), scale(moon, t)))
}
func drawBeta011Celestial() {
	dir := beta011CelestialDirection()
	key := "b011_sun"
	if engine.NightAmount(sceneHour()) > .5 {
		key = "b011_moon"
	}
	push()
	move(dir.x*900, dir.y*900, dir.z*900)
	resize(10, 10, 10)
	col(1, 1, 1, 1)
	drawPBRAsset(key, 0)
	pop()
}

type beta011Skid struct{ Vertices [12]float32 }

var beta011Skids struct {
	Marks     []beta011Skid
	Last      [2]engine.Vec
	LastWidth [2]engine.Vec
	Valid     bool
	VBO       uint32
	Frame     float64
}

func drawBeta011Skids(c engine.Car) {
	s := &beta011Skids
	active := c.OnRoad && c.Body.Grounded && (c.Burnout || c.DriftActive || c.Braking && math.Abs(c.Speed) > 4 || math.Abs(c.Speed) > 2 && (math.Abs(c.Rig.Wheels[2].SlipRatio) > .22 || math.Abs(c.Rig.Wheels[3].SlipRatio) > .22))
	if app.world.Elapsed != s.Frame {
		s.Frame = app.world.Elapsed
		if active {
			f := engine.HeadingVec(c.Heading)
			r := engine.Vec{X: f.Z, Z: -f.X}
			spec := engine.RemasteredVehicleSpec(app.world.CarClass).P
			for i, side := range []float64{-1, 1} {
				p := c.Pos.Sub(f.Mul(spec.Wheelbase * .46)).Add(r.Mul(side * spec.Track * .5))
				last := s.Last[i]
				if !s.Valid && c.Burnout {
					last = p.Sub(f.Mul(.08))
				}
				if (s.Valid || c.Burnout) && p.Dist(last) >= .06 && p.Dist(last) < 3 {
					width := r.Mul(.115)
					oldWidth := s.LastWidth[i]
					if oldWidth.Length() < .01 {
						oldWidth = width
					}
					a, b, cc, d := last.Sub(oldWidth), last.Add(oldWidth), p.Add(width), p.Sub(width)
					points := []engine.Vec{a, b, cc, d}
					mark := beta011Skid{}
					for j, q := range points {
						mark.Vertices[j*3] = float32(q.X)
						mark.Vertices[j*3+1] = float32(beta021SkidHeight(app.world, q))
						mark.Vertices[j*3+2] = float32(q.Z)
					}
					s.Marks = append(s.Marks, mark)
					s.Last[i] = p
					s.LastWidth[i] = width
				} else if !s.Valid || p.Dist(last) >= 3 {
					s.Last[i] = p
					s.LastWidth[i] = r.Mul(.115)
				}
			}
			s.Valid = true
			if len(s.Marks) > 4096 {
				copy(s.Marks, s.Marks[len(s.Marks)-4096:])
				s.Marks = s.Marks[:4096]
			}
		} else {
			s.Valid = false
		}
	}
	if len(s.Marks) == 0 {
		return
	}
	useLighting(false)
	dis(GL_TEXTURE_2D)
	en(GL_BLEND)
	gl("glDepthMask", 0)
	col(.045, .038, .029, .60)
	if s.VBO == 0 {
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&s.VBO)))
	}
	glExt("glBindBuffer", glArrayBuffer, uintptr(s.VBO))
	glExt("glBufferData", glArrayBuffer, uintptr(len(s.Marks)*48), uintptr(unsafe.Pointer(&s.Marks[0])), 0x88E0)
	gl("glEnableClientState", glVertexArray)
	gl("glVertexPointer", 3, glFloat, 12, 0)
	gl("glDrawArrays", GL_QUADS, 0, uintptr(len(s.Marks)*4))
	gl("glDisableClientState", glVertexArray)
	glExt("glBindBuffer", glArrayBuffer, 0)
	gl("glDepthMask", 1)
	useLighting(true)
}

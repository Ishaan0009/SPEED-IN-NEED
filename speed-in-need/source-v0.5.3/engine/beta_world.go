package engine

import "math"

type Hill struct {
	Pos            Vec
	Radius, Height float64
	Seed           int
}

func (h Hill) HeightAt(p Vec) float64 {
	d := p.Sub(h.Pos)
	a := math.Atan2(d.Z, d.X)
	noise := 1 + .09*math.Sin(a*3+float64(h.Seed)) + .045*math.Sin(a*7)
	r := d.Length() / (h.Radius * noise)
	if r >= 1 {
		return Terrain(p.X, p.Z) - .20
	}
	return Terrain(p.X, p.Z) - .20 + h.Height*math.Pow(1-r*r, 1.35)
}
func (w *World) buildHills() {
	w.Hills = nil
	if w.Map == 1 {
		return
	}
	for j := 0; j < 35; j++ {
		p := Vec{-445. + float64((j*139+29)%900), -415. + float64((j*197+31)%850)}
		if DistRoad(p, w.Routes) < 92 || w.IsWater(p, 30) || w.RampClearance(p, 40) {
			continue
		}
		w.Hills = append(w.Hills, Hill{p, 24. + float64((j*11)%29), 19. + float64((j*19)%42), j})
	}
}

// BuildingPlotClear checks the complete rendered 13 by 15 metre footprint,
// including the pavement, against every segment of every road.
func (w *World) BuildingPlotClear(p Vec) bool {
	if w.IsWater(p, 14) || w.RampClearance(p, 16) {
		return false
	}
	for _, r := range w.Routes {
		for i := 1; i < len(r.Points); i++ {
			a, b := r.Points[i-1], r.Points[i]
			// A conservative swept road radius around the building rectangle.
			steps := int(a.Dist(b)/2) + 1
			for j := 0; j <= steps; j++ {
				q := a.Add(b.Sub(a).Mul(float64(j) / float64(steps)))
				dx := math.Max(0, math.Abs(q.X-p.X)-6.5)
				dz := math.Max(0, math.Abs(q.Z-p.Z)-7.5)
				if math.Hypot(dx, dz) < 12 {
					return false
				}
			}
		}
	}
	for _, b := range w.Buildings {
		if math.Abs(b.X-p.X) < 17 && math.Abs(b.Z-p.Z) < 19 {
			return false
		}
	}
	return true
}

func (w *World) placeBuildings() {
	w.Buildings = nil
	target := 60
	if w.Map != 1 {
		target = 7
	}
	for i := 0; i < 900 && len(w.Buildings) < target; i++ {
		p := BuildingPosition(i, w.Routes)
		if i >= 180 {
			p = Vec{float64((i*101)%810 - 405), float64((i*137)%730 - 350)}
		}
		if w.BuildingPlotClear(p) {
			w.Buildings = append(w.Buildings, p)
		}
	}
	// Trees and boulders cannot occupy a building footprint either.
	trees := w.Trees[:0]
	for _, t := range w.Trees {
		clear := true
		for _, b := range w.Buildings {
			if math.Abs(t.Pos.X-b.X) < 10 && math.Abs(t.Pos.Z-b.Z) < 11 {
				clear = false
				break
			}
		}
		if clear {
			trees = append(trees, t)
		}
	}
	w.Trees = trees
	w.VisibleTrees = len(trees)
	rocks := w.Rocks[:0]
	for _, r := range w.Rocks {
		clear := true
		for _, b := range w.Buildings {
			if math.Abs(r.Pos.X-b.X) < 10+r.Radius && math.Abs(r.Pos.Z-b.Z) < 11+r.Radius {
				clear = false
				break
			}
		}
		if clear {
			rocks = append(rocks, r)
		}
	}
	w.Rocks = rocks
}

// LampHead uses the same base pivot, arm direction and fall rotation as the mesh.
func (p Pole) LampHead() (Vec, float64) {
	arm := Vec{math.Cos(p.Heading), -math.Sin(p.Heading)}.Mul(3.9)
	fall := p.FallDirection
	dot := arm.X*fall.X + arm.Z*fall.Z
	xy := arm.Add(fall.Mul(9.7*math.Sin(p.Angle) + dot*(math.Cos(p.Angle)-1)))
	return p.Pos.Add(xy), Terrain(p.Pos.X, p.Pos.Z) + 9.7*math.Cos(p.Angle) - dot*math.Sin(p.Angle)
}

func (p *Pole) Hit(velocity Vec) {
	if p.Falling {
		return
	}
	p.Falling = true
	p.FallDirection = velocity.Unit()
	p.AngularSpeed = Clamp(velocity.Length()*.055, .45, 2.8)
}

func (p *Pole) Step(dt float64) {
	if !p.Falling || p.Settled || dt <= 0 {
		return
	}
	// Uniform rod hinged at its breakaway base. Stop at terrain contact,
	// with a small dissipative rebound instead of sinking into the ground.
	p.AngularSpeed += (1.5*9.81/10*math.Sin(p.Angle) - .28*p.AngularSpeed) * dt
	p.Angle += p.AngularSpeed * dt
	limit := p.GroundLimit()
	if p.Angle >= limit {
		p.Angle = limit
		p.AngularSpeed *= -.16
		if math.Abs(p.AngularSpeed) < .16 {
			p.AngularSpeed = 0
			p.Settled = true
		}
	}
}

func (p Pole) GroundLimit() float64 {
	clear := func(angle float64) bool {
		for length := 2.; length <= 10; length += 2 {
			q := p.Pos.Add(p.FallDirection.Mul(length * math.Sin(angle)))
			y := Terrain(p.Pos.X, p.Pos.Z) + .10 + length*math.Cos(angle)
			if y < Terrain(q.X, q.Z)+.12 {
				return false
			}
		}
		return true
	}
	lo, hi := 0., math.Pi*.5+.15
	for i := 0; i < 16; i++ {
		mid := (lo + hi) * .5
		if clear(mid) {
			lo = mid
		} else {
			hi = mid
		}
	}
	return lo
}

// World light state is shared by visible lamps and shader light sources.
// The three authored environments retain their intended time of day.
func MapHour(mapID int) float64 { return [...]float64{18.5, 22, 16}[ClampInt(mapID, 0, 2)] }
func NightAmount(hour float64) float64 {
	if hour >= 19.5 || hour < 5.5 {
		return 1
	}
	if hour >= 17.5 {
		return Clamp((hour-17.5)/2, 0, 1)
	}
	if hour < 7.5 {
		return Clamp((7.5-hour)/2, 0, 1)
	}
	return 0
}
func AutomaticLights(hour float64) bool { return NightAmount(hour) > .2 }

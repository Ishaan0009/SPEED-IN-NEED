package engine

import "math"

type Walker051 struct {
	Active   bool
	Ped      Pedestrian
	Velocity Vec
	Vertical float64
}

func (w *World) ToggleWalker051(p *Walker051) bool {
	if w.Remaster == nil || w.Remaster.Mode != 0 || math.Abs(w.Car.Speed) > 2 {
		return false
	}
	if p.Active {
		if p.Ped.Pos.Dist(w.Car.Pos) > 3.6 {
			return false
		}
		p.Active = false
		return true
	}
	right := Vec{math.Cos(w.Car.Heading), -math.Sin(w.Car.Heading)}
	spec := VehicleVariantSpec050(w.CarClass, w.Car.Variant050)
	for _, sign := range []float64{-1, 1} {
		at := w.Car.Pos.Add(right.Mul(sign * (spec.HalfWidth + .65)))
		h := w.GroundHeight(at)
		if !w.walkerBlocked051(at, h, .29) {
			*p = Walker051{Active: true, Ped: Pedestrian{Pos: at, Height: h, Heading: w.Car.Heading}}
			p.Ped.Pose = p.Ped.walkingPose()
			return true
		}
	}
	return false
}
func (w *World) walkerBlocked051(pos Vec, y, radius float64) bool {
	if w.pedBlocked(pos, y, radius) {
		return true
	}
	spec := VehicleVariantSpec050(w.CarClass, w.Car.Variant050)
	if _, _, hit := pedBodyContact(pos, radius, w.Car.Pos, Vec{spec.HalfWidth, spec.HalfLength}, w.Car.Heading); hit {
		return true
	}
	cell := remasterMeshCell(pos)
	for x := cell.X - 1; x <= cell.X+1; x++ {
		for z := cell.Z - 1; z <= cell.Z+1; z++ {
			for _, id := range w.Remaster.triangleGrid[gridKey{x, z}] {
				t := w.Remaster.Def.CollisionTriangles[id]
				v := [3][3]float64{t.A, t.B, t.C}
				for _, height := range []float64{y + .4, y + 1.3} {
					var points [3]Vec
					n := 0
					for i := 0; i < 3; i++ {
						a, b := v[i], v[(i+1)%3]
						dy := b[1] - a[1]
						if math.Abs(dy) < 1e-8 {
							continue
						}
						u := (height - a[1]) / dy
						if u >= 0 && u <= 1 && n < 3 {
							points[n] = Vec{a[0] + (b[0]-a[0])*u, a[2] + (b[2]-a[2])*u}
							n++
						}
					}
					if n >= 2 {
						d := points[1].Sub(points[0])
						den := d.X*d.X + d.Z*d.Z
						u := 0.
						if den > 1e-9 {
							q := pos.Sub(points[0])
							u = Clamp((q.X*d.X+q.Z*d.Z)/den, 0, 1)
						}
						if pos.Dist(points[0].Add(d.Mul(u))) < radius {
							return true
						}
					}
				}
			}
		}
	}
	return false
}
func (w *World) StepWalker051(p *Walker051, direction Vec, sprint bool, dt float64) {
	if !p.Active || dt <= 0 {
		return
	}
	dt = math.Min(dt, 1./30)
	length := direction.Length()
	if length > 1 {
		direction = direction.Mul(1 / length)
	}
	speed := 2.0
	if sprint {
		speed = 4.4
	}
	target := direction.Mul(speed)
	p.Velocity = p.Velocity.Add(target.Sub(p.Velocity).Mul(1 - math.Exp(-dt*12)))
	old := p.Ped.Pos
	candidate := old.Add(p.Velocity.Mul(dt))
	can := func(at Vec) bool {
		return w.GroundHeight(at)-p.Ped.Height < .38 && !w.walkerBlocked051(at, p.Ped.Height, .29)
	}
	if can(candidate) {
		p.Ped.Pos = candidate
	} else {
		at := Vec{candidate.X, old.Z}
		if can(at) {
			p.Ped.Pos = at
		}
		at = Vec{p.Ped.Pos.X, candidate.Z}
		if can(at) {
			p.Ped.Pos = at
		}
	}
	ground := w.GroundHeight(p.Ped.Pos)
	p.Vertical -= 9.81 * dt
	p.Ped.Height += p.Vertical * dt
	if p.Ped.Height <= ground {
		p.Ped.Height = ground
		p.Vertical = 0
	}
	travel := old.Dist(p.Ped.Pos)
	p.Ped.Moving = travel > 1e-5
	p.Ped.Speed = travel / dt
	p.Ped.Phase += travel * 5.5
	if direction.Length() > .05 {
		heading := math.Atan2(direction.X, direction.Z)
		p.Ped.Heading += math.Atan2(math.Sin(heading-p.Ped.Heading), math.Cos(heading-p.Ped.Heading)) * (1 - math.Exp(-dt*14))
	}
	p.Ped.PreviousPose = p.Ped.Pose
	p.Ped.Pose = p.Ped.walkingPose()
}

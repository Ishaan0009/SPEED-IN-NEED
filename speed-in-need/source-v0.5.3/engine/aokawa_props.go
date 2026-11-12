package engine

import "math"

// Loose barriers exchange momentum with cars and settle on the same height field.
type DynamicProp struct {
	Pos, Velocity                                           Vec
	Height, VerticalSpeed, Heading, YawRate, Roll, RollRate float64
	Half                                                    Vec
	Mass                                                    float64
}

func (w *World) hitAokawaBarrier(c *Car, class int, index int, n Vec, penetration float64) {
	s := w.Remaster
	shape := s.Def.Colliders[index]
	p := s.DynamicProps[shape.Object]
	if p == nil {
		mass := 180.0
		if shape.Kind == "prop" {
			mass = 65
			if s.Def.Objects[shape.Object].Asset == "b021_cone" {
				mass = 4
			} else if s.Def.Objects[shape.Object].Asset == "b011_bin" {
				mass = 18
			}
		}
		p = &DynamicProp{Pos: shape.Center, Height: shape.MinY, Heading: shape.Heading, Half: shape.Half, Mass: mass}
		s.DynamicProps[shape.Object] = p
	}
	f := HeadingVec(c.Heading)
	r := Vec{f.Z, -f.X}
	v := f.Mul(c.Speed).Add(r.Mul(c.Lateral))
	closing := v.Sub(p.Velocity).X*n.X + v.Sub(p.Velocity).Z*n.Z
	carMass := VehicleVariantSpec050(class, c.Variant050).P.Mass
	if closing < 0 {
		impulse := -closing * 1.08 / (1/carMass + 1/p.Mass)
		v = v.Add(n.Mul(impulse / carMass))
		p.Velocity = p.Velocity.Sub(n.Mul(impulse / p.Mass))
		c.Speed = v.X*f.X + v.Z*f.Z
		c.Lateral = v.X*r.X + v.Z*r.Z
		p.YawRate = Clamp(p.YawRate+closing*.07, -3, 3)
		p.RollRate = Clamp(p.RollRate+math.Abs(closing)*.17, -3, 3)
		p.VerticalSpeed = math.Min(2, math.Abs(closing)*.08)
	}
	c.Pos = c.Pos.Add(n.Mul(penetration*p.Mass/(carMass+p.Mass) + .002))
	p.Pos = p.Pos.Sub(n.Mul(penetration * carMass / (carMass + p.Mass)))
}
func (w *World) stepAokawaProps(dt float64) {
	s := w.Remaster
	for _, p := range s.DynamicProps {
		if p.Velocity.Length() < .005 && math.Abs(p.VerticalSpeed) < .005 && math.Abs(p.RollRate) < .005 && math.Abs(p.YawRate) < .005 {
			continue
		}
		p.Pos = p.Pos.Add(p.Velocity.Mul(dt))
		speed := p.Velocity.Length()
		// Coulomb friction on the ground, independent of frame rate.
		ground := s.Def.Grid.Height(p.Pos) + p.Half.X*math.Abs(math.Sin(p.Roll))
		p.VerticalSpeed -= 9.81 * dt
		p.Height += p.VerticalSpeed * dt
		if p.Height < ground {
			p.Height = ground
			p.VerticalSpeed = math.Abs(p.VerticalSpeed) * .12
			if p.VerticalSpeed < .1 {
				p.VerticalSpeed = 0
			}
			if speed > 0 {
				p.Velocity = p.Velocity.Mul(math.Max(0, 1-4.5*dt/speed))
			}
		}
		p.Heading += p.YawRate * dt
		p.YawRate *= math.Exp(-dt * 2.0)
		p.Roll += p.RollRate * dt
		p.RollRate *= math.Exp(-dt * 3.2)
		p.Roll = Clamp(p.Roll, -math.Pi/2, math.Pi/2)
	}
}
func (w *World) aokawaDynamicContacts(c *Car, class int) {
	for object, p := range w.Remaster.DynamicProps {
		if c.Pos.Dist(p.Pos) > 7 || math.Abs(c.Body.Height-p.Height) > 2 {
			continue
		}
		spec := VehicleVariantSpec050(class, c.Variant050)
		n, depth, hit := contactOBB(c.Pos, Vec{spec.HalfWidth, spec.HalfLength}, c.Heading, p.Pos, p.Half, p.Heading)
		if hit {
			index := w.Remaster.Def.Objects[object].Collider
			if index >= 0 {
				w.hitAokawaBarrier(c, class, index, n, depth)
			}
		}
	}
}

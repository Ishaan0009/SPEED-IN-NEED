package engine

import "math"

func TrafficVelocity(t Traffic) Vec { return HeadingVec(t.Heading).Mul(t.Speed).Add(t.Impulse) }

// Collision velocity persists independently of route progress. The driver waits
// before resuming; recovery is a damped force rather than a positional snap.
func RecoverImpactedTraffic(t *Traffic, dt float64) {
	if t.ImpactCooldown > 0 {
		t.ImpactCooldown = math.Max(0, t.ImpactCooldown-dt)
		spring := .0
		if t.ImpactCooldown < .45 {
			spring = .65
		}
		t.Impulse = t.Impulse.Add(t.Offset.Mul(-spring * dt)).Mul(math.Exp(-dt * .9))
		t.Offset = t.Offset.Add(t.Impulse.Mul(dt))
		t.YawKick += t.ImpactYaw * dt
		t.ImpactYaw *= math.Exp(-dt * 2.8)
	} else {
		RecoverTraffic(t, dt)
		t.ImpactYaw *= math.Exp(-dt * 5)
	}
}

func (w *World) trafficStaticContacts(t *Traffic) {
	k := gridOf(t.Pos)
	y := w.GroundHeight(t.Pos)
	footprint := TrafficBodyFootprint(t.Variant)
	for x := k.X - 1; x <= k.X+1; x++ {
		for z := k.Z - 1; z <= k.Z+1; z++ {
			for _, i := range w.Remaster.staticGrid[gridKey{x, z}] {
				shape := w.Remaster.Def.Colliders[i]
				if shape.Object < 0 || shape.Object >= len(w.Remaster.Def.Objects) || !w.Remaster.Def.Objects[shape.Object].Solid {
					continue
				}
				if y+.7 < shape.MinY || y > shape.MaxY {
					continue
				}
				n, depth, hit := contactOBB(t.Pos, footprint, t.Heading, shape.Center, shape.Half, shape.Heading)
				if !hit {
					continue
				}
				push := n.Mul(depth + .002)
				t.Pos = t.Pos.Add(push)
				t.Offset = t.Offset.Add(push)
				y = w.GroundHeight(t.Pos)
				v := TrafficVelocity(*t)
				vn := v.X*n.X + v.Z*n.Z
				if vn < 0 {
					t.Impulse = t.Impulse.Sub(n.Mul(vn * 1.04))
				}
			}
		}
	}
}

// Predict the actual swept paths, including the player's crossing motion.
// A person on a parallel sidewalk is outside the driving corridor.
func (w *World) trafficHazardSpeed(t *Traffic, desired float64) float64 {
	f := HeadingVec(t.Heading)
	own := TrafficBodyFootprint(t.Variant)
	carF := HeadingVec(w.Car.Heading)
	carR := Vec{carF.Z, -carF.X}
	playerVelocity := carF.Mul(w.Car.Speed).Add(carR.Mul(w.Car.Lateral))
	check := func(pos, velocity Vec, radius, length float64) {
		delta := pos.Sub(t.Pos)
		along := delta.X*f.X + delta.Z*f.Z
		if along < -own.Z-1 || along > 70 {
			return
		}
		rel := velocity.Sub(f.Mul(t.Speed))
		v2 := rel.X*rel.X + rel.Z*rel.Z
		at := 0.
		if v2 > .01 {
			at = Clamp(-(delta.X*rel.X+delta.Z*rel.Z)/v2, 0, 2.5)
		}
		closest := delta.Add(rel.Mul(at))
		side := math.Abs(closest.X*f.Z - closest.Z*f.X)
		if side < own.X+radius+.3 && closest.Length() < own.Z+length+1.2 {
			gap := math.Max(0, along-own.Z-length-1)
			desired = math.Min(desired, math.Sqrt(2*6.5*gap))
			if at > 0 && at < 1.6 {
				desired = math.Min(desired, gap/(at+.5))
			}
			if at < .8 {
				desired = 0
			}
		}
	}
	check(w.Car.Pos, playerVelocity, VehicleVariantSpec050(w.CarClass, w.Car.Variant050).HalfWidth, VehicleVariantSpec050(w.CarClass, w.Car.Variant050).HalfLength)
	for _, i := range w.nearbyPedestrians(t.Pos, 60) {
		if i < 0 || i >= len(w.Pedestrians) {
			continue
		}
		p := &w.Pedestrians[i]
		if p.Pos.Dist(t.Pos) < 60 {
			check(p.Pos, p.Velocity(), .32, .32)
		}
	}
	return desired
}

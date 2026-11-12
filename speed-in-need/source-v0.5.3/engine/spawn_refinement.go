package engine

import "math"

// SetupSpawnPoint resolves a requested free-roam position without loading a
// second simulation in the menu. The final world validates tire support again.
func SetupSpawnPoint(d *WorldDefinition, request Vec, class int) (Vec, float64, bool) {
	if d == nil || !navFinite033(request) || len(d.Roads) == 0 {
		return Vec{}, 0, false
	}
	request.X = Clamp(request.X, d.Bounds[0].X+5, d.Bounds[1].X-5)
	request.Z = Clamp(request.Z, d.Bounds[0].Z+5, d.Bounds[1].Z-5)
	best, heading := math.Inf(1), 0.
	roadPoint := d.Spawn
	for _, road := range d.Roads {
		if road.Kind == "stunt" {
			continue
		}
		for i := 0; i+1 < len(road.Points); i++ {
			p, _ := navProject033(request, road.Points[i], road.Points[i+1])
			if distance := p.Dist(request); distance < best {
				best, roadPoint = distance, p
				f := road.Points[i+1].Sub(road.Points[i])
				heading = math.Atan2(f.X, f.Z)
			}
		}
	}
	clear := func(p Vec) bool {
		y := d.Grid.Height(p)
		if d.WaterGrid.W > 1 && d.WaterGrid.Height(p) > .5 && y < d.WaterHeight+.4 {
			return false
		}
		for _, water := range d.Water {
			if water.Contains(p) && best > 4 {
				return false
			}
		}
		for _, offset := range []Vec{{2, 0}, {-2, 0}, {0, 2}, {0, -2}} {
			if math.Abs(d.Grid.Height(p.Add(offset))-y) > 1.1 {
				return false
			}
		}
		spec := RemasteredVehicleSpec(class)
		for _, c := range d.Colliders {
			if c.MinY > y+2.5 || c.MaxY < y-.25 {
				continue
			}
			if c.Radius > 0 && p.Dist(c.Center) < c.Radius+math.Hypot(spec.HalfWidth, spec.HalfLength)+.3 {
				return false
			}
			if c.Half.X > 0 && c.Half.Z > 0 {
				if _, _, hit := contactOBB(p, Vec{spec.HalfWidth + .25, spec.HalfLength + .25}, heading, c.Center, c.Half, c.Heading); hit {
					return false
				}
			}
		}
		return true
	}
	if clear(request) {
		return request, heading, true
	}
	forward := HeadingVec(heading)
	for _, distance := range []float64{0, 6, -6, 12, -12, 24, -24, 40, -40} {
		p := roadPoint.Add(forward.Mul(distance))
		if clear(p) {
			return p, heading, true
		}
	}
	return d.Spawn, heading, false
}

func (w *World) ApplySetupSpawn(request Vec) bool {
	if w.Remaster == nil || w.Remaster.Mode != 0 {
		return false
	}
	p, h, ok := SetupSpawnPoint(w.Remaster.Def, request, w.CarClass)
	w.Car.Pos, w.Car.Heading = p, h
	w.Car.Speed, w.Car.Lateral, w.Car.YawRate = 0, 0, 0
	w.Car.Body, w.Car.Rig = ChassisState{}, SuspensionRig{}
	w.InitializeChassis()
	// Reject any remaining imported contact-mesh overlap before the first tick.
	before := w.Car.Pos
	w.remasteredMeshContacts(&w.Car, w.CarClass)
	if before.Dist(w.Car.Pos) > .25 {
		w.RecoverRemasteredCar()
		ok = false
	}
	w.InitializeChassis()
	return ok
}

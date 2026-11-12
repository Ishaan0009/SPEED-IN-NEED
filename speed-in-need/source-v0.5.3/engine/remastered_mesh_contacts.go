package engine

import "math"

func remasterMeshCell(p Vec) gridKey {
	return gridKey{int(math.Floor(p.X / 8)), int(math.Floor(p.Z / 8))}
}

func (s *RemasteredState) buildMeshContacts() {
	s.cameraGrid = map[gridKey][]int{}
	for i, t := range s.Def.CameraTriangles {
		lo := remasterMeshCell(Vec{math.Min(t.A[0], math.Min(t.B[0], t.C[0])) - .3, math.Min(t.A[2], math.Min(t.B[2], t.C[2])) - .3})
		hi := remasterMeshCell(Vec{math.Max(t.A[0], math.Max(t.B[0], t.C[0])) + .3, math.Max(t.A[2], math.Max(t.B[2], t.C[2])) + .3})
		if hi.X-lo.X > 90 || hi.Z-lo.Z > 90 {
			continue
		}
		for x := lo.X; x <= hi.X; x++ {
			for z := lo.Z; z <= hi.Z; z++ {
				k := gridKey{x, z}
				s.cameraGrid[k] = append(s.cameraGrid[k], i)
			}
		}
	}
	s.triangleGrid = map[gridKey][]int{}
	s.triangleBounds = make([][6]float64, len(s.Def.CollisionTriangles))
	for i, t := range s.Def.CollisionTriangles {
		s.triangleBounds[i] = [6]float64{
			math.Min(t.A[0], math.Min(t.B[0], t.C[0])), math.Min(t.A[1], math.Min(t.B[1], t.C[1])), math.Min(t.A[2], math.Min(t.B[2], t.C[2])),
			math.Max(t.A[0], math.Max(t.B[0], t.C[0])), math.Max(t.A[1], math.Max(t.B[1], t.C[1])), math.Max(t.A[2], math.Max(t.B[2], t.C[2])),
		}
		lo := remasterMeshCell(Vec{math.Min(t.A[0], math.Min(t.B[0], t.C[0])) - 3, math.Min(t.A[2], math.Min(t.B[2], t.C[2])) - 3})
		hi := remasterMeshCell(Vec{math.Max(t.A[0], math.Max(t.B[0], t.C[0])) + 3, math.Max(t.A[2], math.Max(t.B[2], t.C[2])) + 3})
		if hi.X-lo.X > 90 || hi.Z-lo.Z > 90 {
			continue
		}
		for x := lo.X; x <= hi.X; x++ {
			for z := lo.Z; z <= hi.Z; z++ {
				key := gridKey{x, z}
				s.triangleGrid[key] = append(s.triangleGrid[key], i)
			}
		}
	}
}
func horizontalTriangleSlice(t SurfaceTriangle, y float64) (Vec, Vec, bool) {
	points := [3][3]float64{t.A, t.B, t.C}
	var intersections [3]Vec
	count := 0
	for i := 0; i < 3; i++ {
		a, b := points[i], points[(i+1)%3]
		if (a[1] > y) == (b[1] > y) || math.Abs(a[1]-b[1]) < 1e-9 {
			continue
		}
		u := (y - a[1]) / (b[1] - a[1])
		p := Vec{Lerp(a[0], b[0], u), Lerp(a[2], b[2], u)}
		if count < 3 {
			intersections[count] = p
			count++
		}
	}
	if count < 2 || intersections[0].Dist(intersections[1]) < .001 {
		return Vec{}, Vec{}, false
	}
	return intersections[0], intersections[1], true
}
func (w *World) remasteredMeshContacts(c *Car, class int) {
	s := w.Remaster
	spec := VehicleVariantSpec050(class, c.Variant050)
	radius := math.Hypot(spec.HalfWidth, spec.HalfLength)
	for _, index := range s.triangleGrid[remasterMeshCell(c.Pos)] {
		bounds := s.triangleBounds[index]
		if c.Pos.X+radius < bounds[0] || c.Pos.X-radius > bounds[3] || c.Pos.Z+radius < bounds[2] || c.Pos.Z-radius > bounds[5] || c.Body.Height+.38 < bounds[1] || c.Body.Height-.35 > bounds[4] {
			continue
		}
		triangle := s.Def.CollisionTriangles[index]
		for _, y := range [3]float64{c.Body.Height - .35, c.Body.Height + .05, c.Body.Height + .38} {
			a, b, ok := horizontalTriangleSlice(triangle, y)
			if !ok {
				continue
			}
			delta := b.Sub(a)
			length := delta.Length()
			center := a.Add(b).Mul(.5)
			heading := math.Atan2(delta.X, delta.Z)
			n, overlap, hit := contactOBB(c.Pos, Vec{spec.HalfWidth, spec.HalfLength}, c.Heading, center, Vec{.025, length * .5}, heading)
			if !hit {
				continue
			}
			c.Pos = c.Pos.Add(n.Mul(overlap + .002))
			f := HeadingVec(c.Heading)
			r := Vec{f.Z, -f.X}
			velocity := f.Mul(c.Speed).Add(r.Mul(c.Lateral))
			approach := velocity.X*n.X + velocity.Z*n.Z
			if approach < 0 {
				velocity = velocity.Sub(n.Mul(approach * 1.02))
				c.Speed = velocity.X*f.X + velocity.Z*f.Z
				c.Lateral = velocity.X*r.X + velocity.Z*r.Z
				c.ContactImpulse = math.Max(c.ContactImpulse, -approach)
			}
			break
		}
	}
}
func (w *World) RemasteredWaterLevel(p Vec) (float64, bool) {
	if w.Remaster != nil && w.Remaster.Def.WaterGrid.W > 1 {
		return w.Remaster.Def.WaterHeight, w.IsWater(p, 0)
	}
	if w.Remaster == nil {
		return CoastWaterLevel, w.IsWater(p, 0)
	}
	for i, water := range w.Water {
		if water.Contains(p) {
			level := 4.
			if i < len(w.Remaster.Def.WaterLevels) {
				level = w.Remaster.Def.WaterLevels[i]
			}
			return level, true
		}
	}
	return 0, false
}

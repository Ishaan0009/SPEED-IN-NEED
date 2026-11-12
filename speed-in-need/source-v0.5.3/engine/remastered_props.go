package engine

import "math"

func (w *World) RemasteredCameraFraction(from, to [3]float64) float64 {
	if w.Remaster == nil {
		return 1
	}
	for i := 2; i <= 32; i++ {
		t := float64(i) / 32
		p := Vec{Lerp(from[0], to[0], t), Lerp(from[2], to[2], t)}
		y := Lerp(from[1], to[1], t)
		for _, index := range w.Remaster.cameraGrid[remasterMeshCell(p)] {
			if hit, ok := cameraTriangleFraction(from, to, w.Remaster.Def.CameraTriangles[index]); ok && hit <= t+.03125 {
				return math.Max(.12, hit-.06)
			}
		}
		// Imported cliff and cave meshes have triangle walls rather than box proxies.
		// Test the same spatially indexed wall slices used by vehicle contacts.
		for _, index := range w.Remaster.triangleGrid[remasterMeshCell(p)] {
			a, b, ok := horizontalTriangleSlice(w.Remaster.Def.CollisionTriangles[index], y)
			if !ok {
				continue
			}
			delta := b.Sub(a)
			u := Clamp((p.Sub(a).X*delta.X+p.Sub(a).Z*delta.Z)/math.Max(1e-9, delta.X*delta.X+delta.Z*delta.Z), 0, 1)
			if p.Dist(a.Add(delta.Mul(u))) < .32 {
				return math.Max(.12, t-.06)
			}
		}
		for _, index := range w.Remaster.surfaceGrid[remasterMeshCell(p)] {
			triangle := w.Remaster.Def.Surfaces[index]
			if !triangle.Support {
				continue
			}
			if hit, ok := cameraTriangleFraction(from, to, triangle); ok && hit <= t+.03125 {
				return math.Max(.12, hit-.06)
			}
		}
		for _, index := range w.Remaster.staticGrid[gridOf(p)] {
			c := w.Remaster.Def.Colliders[index]
			if w.Remaster.DynamicProps[c.Object] != nil {
				continue
			}
			if y < c.MinY-.3 || y > c.MaxY+.3 {
				continue
			}
			f := HeadingVec(c.Heading)
			r := Vec{f.Z, -f.X}
			d := p.Sub(c.Center)
			inside := false
			if c.Radius > 0 {
				inside = d.Length() < c.Radius+.3
			} else {
				inside = math.Abs(d.X*r.X+d.Z*r.Z) < c.Half.X+.3 && math.Abs(d.X*f.X+d.Z*f.Z) < c.Half.Z+.3
			}
			if inside {
				return math.Max(.12, t-.06)
			}
		}
	}
	return 1
}

// Segment/triangle intersection includes sloped opaque imported ground. A
// heightfield alone cannot describe a copied track with overhangs and layers.
func cameraTriangleFraction(from, to [3]float64, triangle SurfaceTriangle) (float64, bool) {
	sub := func(a, b [3]float64) [3]float64 { return [3]float64{a[0] - b[0], a[1] - b[1], a[2] - b[2]} }
	cross := func(a, b [3]float64) [3]float64 {
		return [3]float64{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
	}
	dot := func(a, b [3]float64) float64 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }
	d, e1, e2 := sub(to, from), sub(triangle.B, triangle.A), sub(triangle.C, triangle.A)
	h := cross(d, e2)
	det := dot(e1, h)
	if math.Abs(det) < 1e-9 {
		return 0, false
	}
	s := sub(from, triangle.A)
	u := dot(s, h) / det
	if u < 0 || u > 1 {
		return 0, false
	}
	q := cross(s, e1)
	v := dot(d, q) / det
	if v < 0 || u+v > 1 {
		return 0, false
	}
	t := dot(e2, q) / det
	return t, t > .08 && t < 1
}

func (w *World) stepRemasteredPoles(dt float64) {
	for _, p := range w.Remaster.FallenPoles {
		length := p.Length
		if length <= 0 {
			length = 9.7
		}
		if !p.Falling || p.Settled {
			continue
		}
		p.AngularSpeed += (1.5*9.81/length*math.Sin(p.Angle) - .30*p.AngularSpeed) * dt
		p.Angle += p.AngularSpeed * dt
		base := w.Remaster.Def.Grid.Height(p.Pos) + .12
		clear := func(a float64) bool {
			for l := 1.; l <= length; l += .8 {
				q := p.Pos.Add(p.FallDirection.Mul(l * math.Sin(a)))
				if base+l*math.Cos(a) < w.Remaster.Def.Grid.Height(q)+.10 {
					return false
				}
			}
			return true
		}
		lo, hi := 0., math.Pi*.5
		for i := 0; i < 12; i++ {
			mid := (lo + hi) * .5
			if clear(mid) {
				lo = mid
			} else {
				hi = mid
			}
		}
		if p.Angle > lo {
			p.Angle = lo
			p.AngularSpeed *= -.14
			if math.Abs(p.AngularSpeed) < .14 {
				p.AngularSpeed = 0
				p.Settled = true
			}
		}
	}
}

package engine

import "math"

// Apply once to the editable world. Roads and developed building footprints
// retain their authored elevation; the meadow blends into gentle uneven turf.
func (d *WorldDefinition) ApplyMeadowRelief051() {
	if d.ReliefReady051 || d.Grid.W < 3 {
		return
	}
	d.ReliefReady051 = true
	d.ReliefOriginal051 = d.Grid
	d.Grid.Values = append([]float32(nil), d.Grid.Values...)
	type segment struct {
		a, b Vec
		pad  float64
	}
	roads := map[gridKey][]segment{}
	buildings := map[gridKey][]PlacedObject{}
	cell := func(p Vec) gridKey { return gridKey{int(math.Floor(p.X / 40)), int(math.Floor(p.Z / 40))} }
	for _, road := range d.Roads {
		for i := 1; i < len(road.Points); i++ {
			a, b := road.Points[i-1], road.Points[i]
			pad := road.Width*.5 + 3 + d.Grid.Step*1.5
			if road.Kind == "town" {
				pad += 8
			}
			lo, hi := cell(Vec{math.Min(a.X, b.X) - pad - 22, math.Min(a.Z, b.Z) - pad - 22}), cell(Vec{math.Max(a.X, b.X) + pad + 22, math.Max(a.Z, b.Z) + pad + 22})
			for x := lo.X; x <= hi.X; x++ {
				for z := lo.Z; z <= hi.Z; z++ {
					k := gridKey{x, z}
					roads[k] = append(roads[k], segment{a, b, pad})
				}
			}
		}
	}
	for _, o := range d.Objects {
		if o.Kind == "building" {
			c := Vec{o.Pos[0], o.Pos[2]}
			pad := math.Max(18, o.Radius+8)
			lo, hi := cell(c.Sub(Vec{pad, pad})), cell(c.Add(Vec{pad, pad}))
			for x := lo.X; x <= hi.X; x++ {
				for z := lo.Z; z <= hi.Z; z++ {
					k := gridKey{x, z}
					buildings[k] = append(buildings[k], o)
				}
			}
		}
	}
	for z := 0; z < d.Grid.H; z++ {
		for x := 0; x < d.Grid.W; x++ {
			index := z*d.Grid.W + x
			base := float64(d.Grid.Values[index])
			if base < d.WaterHeight+1.5 {
				continue
			}
			p := Vec{d.Grid.MinX + float64(x)*d.Grid.Step, d.Grid.MinZ + float64(z)*d.Grid.Step}
			k := cell(p)
			blend := 1.
			for _, s := range roads[k] {
				v := s.b.Sub(s.a)
				den := v.X*v.X + v.Z*v.Z
				if den < 1e-8 {
					continue
				}
				q := p.Sub(s.a)
				u := Clamp((q.X*v.X+q.Z*v.Z)/den, 0, 1)
				distance := p.Dist(s.a.Add(v.Mul(u)))
				blend = math.Min(blend, Clamp((distance-s.pad)/20, 0, 1))
			}
			for _, o := range buildings[k] {
				distance := p.Dist(Vec{o.Pos[0], o.Pos[2]})
				blend = math.Min(blend, Clamp((distance-math.Max(12, o.Radius))/12, 0, 1))
			}
			for _, tunnel := range d.Tunnels {
				f := HeadingVec(tunnel.Heading)
				q := p.Sub(tunnel.Center)
				along := q.X*f.X + q.Z*f.Z
				side := q.X*f.Z - q.Z*f.X
				if math.Abs(along) < tunnel.Length*.5+40 && math.Abs(side) < 35 {
					blend = 0
				}
			}
			blend = blend * blend * (3 - 2*blend)
			wave := 1.9*math.Sin(p.X*.081)*math.Cos(p.Z*.067) + .48*math.Sin(p.Z*.19+p.X*.043) + 1.3*math.Cos((p.X-p.Z)*.037)
			d.Grid.Values[index] = float32(base + wave*blend)
		}
	}
	for i := range d.Objects {
		o := &d.Objects[i]
		switch o.Kind {
		case "tree", "bush", "ground-rock", "rock", "road-tree":
			o.Pos[1] += d.MeadowDelta051(Vec{o.Pos[0], o.Pos[2]})
		}
	}
	for i := range d.Colliders {
		c := &d.Colliders[i]
		if c.Object >= 0 && c.Object < len(d.Objects) {
			o := d.Objects[c.Object]
			switch o.Kind {
			case "tree", "bush", "ground-rock", "rock", "road-tree":
				delta := d.MeadowDelta051(c.Center)
				c.MinY += delta
				c.MaxY += delta
			}
		}
	}
	shift := func(t *SurfaceTriangle) {
		for _, p := range []*[3]float64{&t.A, &t.B, &t.C} {
			p[1] += d.MeadowDelta051(Vec{p[0], p[2]})
		}
	}
	for i := range d.Surfaces {
		if !d.Surfaces[i].Road {
			shift(&d.Surfaces[i])
		}
	}

}
func (d *WorldDefinition) MeadowDelta051(p Vec) float64 {
	if !d.ReliefReady051 {
		return 0
	}
	return d.Grid.Height(p) - d.ReliefOriginal051.Height(p)
}

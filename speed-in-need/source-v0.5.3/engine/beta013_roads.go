package engine

import "math"

// Do not paint a lane edge or centre line through the crossing carriageway.
// This uses the same spatial road segments as tire surface classification.
func (w *World) RoadMarkingAllowed(p, direction Vec, route int) bool {
	if w.Remaster == nil {
		return true
	}
	for _, seg := range w.Remaster.roadGrid[gridOf(p)] {
		if seg.Route == route || seg.Length < .01 {
			continue
		}
		f := seg.B.Sub(seg.A).Mul(1 / seg.Length)
		if math.Abs(f.X*direction.Z-f.Z*direction.X) < .35 {
			continue
		}
		delta := seg.B.Sub(seg.A)
		rel := p.Sub(seg.A)
		u := Clamp((rel.X*delta.X+rel.Z*delta.Z)/(seg.Length*seg.Length), 0, 1)
		if p.Dist(seg.A.Add(delta.Mul(u))) < seg.Width*.5+.6 {
			return false
		}
	}
	return true
}

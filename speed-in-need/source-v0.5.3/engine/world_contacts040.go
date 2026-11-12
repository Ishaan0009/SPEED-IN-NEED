package engine

import "math"

type CollisionFootprint040 struct {
	MinX, MinZ, Step float64
	W, H             int
	Cells            []uint8
}

// Only geometry occupied by actual low-level walls/columns blocks a car.
// Courtyards and basketball courts are not filled by an enclosing model AABB.
func (w *World) buildingContact040(c *Car, spec RemasteredVehicle, shape ColliderShape) (Vec, float64, bool, bool) {
	if shape.Kind != "building" || shape.Object < 0 || shape.Object >= len(w.Remaster.Def.Objects) {
		return Vec{}, 0, false, false
	}
	o := w.Remaster.Def.Objects[shape.Object]
	mask, ok := w.Remaster.Def.CollisionFootprints040[o.Asset]
	if !ok || mask.Step <= 0 || len(mask.Cells) != mask.W*mask.H {
		return Vec{}, 0, false, false
	}
	sx, sz := math.Abs(o.Scale[0]), math.Abs(o.Scale[2])
	if sx < .001 || sz < .001 {
		return Vec{}, 0, false, false
	}
	co, si := math.Cos(o.Heading), math.Sin(o.Heading)
	delta := c.Pos.Sub(Vec{o.Pos[0], o.Pos[2]})
	x, z := (delta.X*co-delta.Z*si)/sx, (delta.X*si+delta.Z*co)/sz
	radius := math.Hypot(spec.HalfWidth, spec.HalfLength) + mask.Step
	loX, hiX := ClampInt(int(math.Floor((x-radius/sx-mask.MinX)/mask.Step)), 0, mask.W-1), ClampInt(int(math.Floor((x+radius/sx-mask.MinX)/mask.Step)), 0, mask.W-1)
	loZ, hiZ := ClampInt(int(math.Floor((z-radius/sz-mask.MinZ)/mask.Step)), 0, mask.H-1), ClampInt(int(math.Floor((z+radius/sz-mask.MinZ)/mask.Step)), 0, mask.H-1)
	best := math.Inf(1)
	normal := Vec{}
	found := false
	for iz := loZ; iz <= hiZ; iz++ {
		for ix := loX; ix <= hiX; ix++ {
			if mask.Cells[iz*mask.W+ix] == 0 {
				continue
			}
			lx, lz := mask.MinX+(float64(ix)+.5)*mask.Step, mask.MinZ+(float64(iz)+.5)*mask.Step
			centre := Vec{o.Pos[0] + co*lx*sx + si*lz*sz, o.Pos[2] - si*lx*sx + co*lz*sz}
			if n, p, hit := contactOBB(c.Pos, Vec{spec.HalfWidth, spec.HalfLength}, c.Heading, centre, Vec{mask.Step * sx * .5, mask.Step * sz * .5}, o.Heading); hit && p < best {
				normal, best, found = n, p, true
			}
		}
	}
	return normal, best, found, true
}

func RacePace040(level int) (float64, float64) {
	switch ClampInt(level, 0, 2) {
	case 0:
		return .82, 100 / 3.6
	case 2:
		return 1.18, 300 / 3.6
	default:
		return 1.02, 200 / 3.6
	}
}

// Physical walls remain the primary boundary. This rule is a last-resort race
// containment guard for jumps/seams; free-roam driving is never constrained.
func (w *World) constrainRaceCorridor040(c *Car, class int) {
	if w.Remaster == nil || w.Remaster.Mode != 1 || len(w.Routes) == 0 {
		return
	}
	route := w.Routes[0]
	distance, hint := routeProjection(route, c.Pos, c.RaceBoundarySegment040)
	centre, f := route.AtDistance(distance)
	half := math.Max(3, w.Remaster.Def.Roads[0].Width*.5)
	if centre.Dist(c.Pos) > half*2 {
		distance, hint = routeProjection(route, c.Pos, -1)
		centre, f = route.AtDistance(distance)
	}
	c.RaceBoundarySegment040 = hint
	right := Vec{f.Z, -f.X}
	d := c.Pos.Sub(centre)
	side := d.X*right.X + d.Z*right.Z
	spec := VehicleVariantSpec050(class, c.Variant050)
	limit := half - spec.HalfWidth - .4
	if math.Abs(side) <= limit {
		return
	}
	c.Pos = c.Pos.Sub(right.Mul(side - math.Copysign(limit, side)))
	forward := HeadingVec(c.Heading)
	carRight := Vec{forward.Z, -forward.X}
	velocity := forward.Mul(c.Speed).Add(carRight.Mul(c.Lateral))
	out := velocity.X*right.X + velocity.Z*right.Z
	if out*side > 0 {
		velocity = velocity.Sub(right.Mul(out))
		c.Speed = velocity.X*forward.X + velocity.Z*forward.Z
		c.Lateral = velocity.X*carRight.X + velocity.Z*carRight.Z
	}
}

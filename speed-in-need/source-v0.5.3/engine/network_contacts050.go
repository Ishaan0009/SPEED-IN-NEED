package engine

// Each owner predicts its own contacts against the latest replicated poses.
// This is responsive LAN contact prediction, not internet anti-cheat authority.
func (w *World) ResolveNetworkContact050(other Car, class int) {
	c := &w.Car
	if c.Pos.Dist(other.Pos) > 6 || abs050(c.Body.Height-other.Body.Height) > 2 {
		return
	}
	a, b := VehicleVariantSpec050(w.CarClass, c.Variant050), VehicleVariantSpec050(class, other.Variant050)
	n, depth, hit := contactOBB(c.Pos, Vec{a.HalfWidth, a.HalfLength}, c.Heading, other.Pos, Vec{b.HalfWidth, b.HalfLength}, other.Heading)
	if !hit {
		return
	}
	c.Pos = c.Pos.Add(n.Mul(depth*.55 + .001))
	f := HeadingVec(c.Heading)
	right := Vec{f.Z, -f.X}
	v := f.Mul(c.Speed).Add(right.Mul(c.Lateral))
	ov := HeadingVec(other.Heading).Mul(other.Speed)
	closing := (v.X-ov.X)*n.X + (v.Z-ov.Z)*n.Z
	if closing < 0 {
		impulse := -closing / (1/a.P.Mass + 1/b.P.Mass) * 1.08
		v = v.Add(n.Mul(impulse / a.P.Mass))
		c.Speed = v.X*f.X + v.Z*f.Z
		c.Lateral = v.X*right.X + v.Z*right.Z
		c.ContactImpulse = Clamp(impulse/a.P.Mass, 0, 10)
	}
}
func abs050(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func (w *World) NetworkGatePassed050(pos Vec, next int) bool {
	return w.Remaster != nil && len(w.Remaster.Gates) > 0 && w.gatePassed(pos, next)
}

func (w *World) NetworkRouteProgress050(pos Vec) float64 {
	if len(w.Routes) == 0 || w.Routes[0].Len <= 0 {
		return 0
	}
	at, _ := routeProjection(w.Routes[0], pos, -1)
	return at / w.Routes[0].Len
}

// Human rows follow the complete AI grid rather than overlapping its last row.
func LANGrid050(ai, slot int) (back, lane float64) {
	back = 11 + float64((ClampInt(ai, 0, 7)+1)/2)*7 + float64(ClampInt(slot, 0, 3)/2)*7
	lane = -1.9
	if slot%2 == 1 {
		lane = 1.9
	}
	return
}

package engine

import "math"

// WheelHub026 follows the terrain plane at the rendered axle rather than the
// previous physics probe. The cylinder support includes slope, camber and steer.
// Travel limits keep imported hubs attached to their suspension on sharp crests.
func WheelHub026(c Car, anchor [3]float64, radius, halfWidth, steer float64, touching bool, surface func(Vec) float64) float64 {
	if !touching && !c.Body.Grounded {
		return anchor[1] - .10
	}
	pitch, roll := Clamp(c.Rig.Pitch, -1.05, 1.05), Clamp(c.Rig.Roll, -.75, .75)
	cp, sp, cr, sr := math.Cos(pitch), math.Sin(pitch), math.Cos(roll), math.Sin(roll)
	f := HeadingVec(c.Heading)
	right := Vec{f.Z, -f.X}
	cs, ss := math.Cos(steer), math.Sin(steer)
	// Tire rotation axis transformed through steer -> roll -> pitch -> yaw.
	ax, ay, az := cr*cs, cp*sr*cs+sp*ss, sp*sr*cs-cp*ss
	axis := right.Mul(ax).Add(f.Mul(az))
	hub := anchor[1]
	for iteration := 0; iteration < 3; iteration++ {
		x := cr*anchor[0] - sr*hub
		z := sp*(sr*anchor[0]+cr*hub) + cp*anchor[2]
		point := c.Pos.Add(right.Mul(x)).Add(f.Mul(z))
		gx := (surface(point.Add(Vec{.18, 0})) - surface(point.Add(Vec{-.18, 0}))) / .36
		gz := (surface(point.Add(Vec{0, .18})) - surface(point.Add(Vec{0, -.18}))) / .36
		length := math.Sqrt(1 + gx*gx + gz*gz)
		nx, ny, nz := -gx/length, 1/length, -gz/length
		dot := Clamp(nx*axis.X+ny*ay+nz*axis.Z, -1, 1)
		extent := (radius*math.Sqrt(math.Max(0, 1-dot*dot)) + halfWidth*math.Abs(dot)) / math.Max(.40, ny)
		wanted := surface(point) + extent + .048 - (c.Body.Height - .74)
		hub = (wanted + sp*anchor[2] - cp*sr*anchor[0]) / (cp * cr)
		hub = Clamp(hub, anchor[1]-.15, anchor[1]+.23)
	}
	return hub
}

func (w *World) VisualWheelHub026(c Car, anchor [3]float64, radius, halfWidth, steer float64, touching bool) float64 {
	return WheelHub026(c, anchor, radius, halfWidth, steer, touching, func(p Vec) float64 { return w.remasteredSupport(&c, p) })
}

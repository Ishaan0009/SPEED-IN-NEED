package engine

import "math"

// InterpolateCar draws between the last two physics snapshots. All expensive
// collision, traffic and scoring work remains strictly in the 120 Hz loop.
func InterpolateCar(previous, current Car, alpha float64) Car {
	alpha = Clamp(alpha, 0, 1)
	if previous.Pos.Dist(current.Pos) > 35 {
		return current
	} // teleport or map reset
	out := current
	out.Pos = previous.Pos.Add(current.Pos.Sub(previous.Pos).Mul(alpha))
	delta := math.Atan2(math.Sin(current.Heading-previous.Heading), math.Cos(current.Heading-previous.Heading))
	out.Heading = previous.Heading + delta*alpha
	out.Speed = Lerp(previous.Speed, current.Speed, alpha)
	out.Lateral = Lerp(previous.Lateral, current.Lateral, alpha)
	out.YawRate = Lerp(previous.YawRate, current.YawRate, alpha)
	out.Body.Height = Lerp(previous.Body.Height, current.Body.Height, alpha)
	out.Body.VerticalSpeed = Lerp(previous.Body.VerticalSpeed, current.Body.VerticalSpeed, alpha)
	// The legacy phase is wrapped to [-2pi,2pi]; interpolate across the
	// boundary using the shortest circular distance rather than rewinding.
	rollDelta := math.Remainder(current.WheelSpin-previous.WheelSpin, Tau)
	out.WheelSpin = previous.WheelSpin + rollDelta*alpha
	out.VisualSteer = Lerp(previous.VisualSteer, current.VisualSteer, alpha)
	out.Rig.Heave = Lerp(previous.Rig.Heave, current.Rig.Heave, alpha)
	out.Rig.Roll = Lerp(previous.Rig.Roll, current.Rig.Roll, alpha)
	out.Rig.Pitch = Lerp(previous.Rig.Pitch, current.Rig.Pitch, alpha)
	for i := range out.Rig.Wheels {
		out.Rig.Wheels[i].GroundHeight = Lerp(previous.Rig.Wheels[i].GroundHeight, current.Rig.Wheels[i].GroundHeight, alpha)
		out.Rig.Wheels[i].AngularVelocity = Lerp(previous.Rig.Wheels[i].AngularVelocity, current.Rig.Wheels[i].AngularVelocity, alpha)
		out.Rig.Wheels[i].Compression = Lerp(previous.Rig.Wheels[i].Compression, current.Rig.Wheels[i].Compression, alpha)
		out.Rig.Wheels[i].SteerAngle = Lerp(previous.Rig.Wheels[i].SteerAngle, current.Rig.Wheels[i].SteerAngle, alpha)
		out.Rig.Wheels[i].RollAngle = previous.Rig.Wheels[i].RollAngle + math.Remainder(current.Rig.Wheels[i].RollAngle-previous.Rig.Wheels[i].RollAngle, Tau)*alpha
	}
	return out
}

// CameraObstructionFraction tests tree-trunk and rock circles along the chase
// camera line, not tree foliage. Returns a [0,1] reduction along the ray.
func (w *World) CameraObstructionFraction(target, desired Vec) float64 {
	ray := desired.Sub(target)
	lenSquared := ray.X*ray.X + ray.Z*ray.Z
	if lenSquared < 1e-6 {
		return 1
	}
	length := math.Sqrt(lenSquared)
	limit := 1.0
	for _, i := range w.NearbyRockIndices(target, length+5) {
		r := w.Rocks[i]
		fraction := rayCircleEntry(target, ray, r.Pos, r.Radius+.6)
		if fraction < limit {
			limit = fraction
		}
	}
	for _, i := range w.NearbyTreeIndices(target, length+5) {
		t := w.Trees[i]
		fraction := rayCircleEntry(target, ray, t.Pos, t.Size*.42+.5)
		if fraction < limit {
			limit = fraction
		}
	}
	return Clamp(limit, 0, 1)
}
func rayCircleEntry(origin, ray, circle Vec, radius float64) float64 {
	a := ray.X*ray.X + ray.Z*ray.Z
	offset := origin.Sub(circle)
	b := 2 * (offset.X*ray.X + offset.Z*ray.Z)
	c := offset.X*offset.X + offset.Z*offset.Z - radius*radius
	disc := b*b - 4*a*c
	if disc < 0 || a < 1e-8 {
		return 1
	}
	root := math.Sqrt(disc)
	t := (-b - root) / (2 * a)
	if t >= 0 && t < 1 {
		return math.Max(0, t-.04)
	}
	return 1
}

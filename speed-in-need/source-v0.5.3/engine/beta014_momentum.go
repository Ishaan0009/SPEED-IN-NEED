package engine

// AdvancePlanarMomentum integrates forces in world space before expressing the
// result in the new chassis frame. Turning alone cannot create energy or turn
// the velocity vector: only tire forces change the direction of travel.
func AdvancePlanarMomentum(speed, lateral, heading, yawRate, longitudinalAcceleration, lateralAcceleration, dt float64) (float64, float64, float64) {
	f := HeadingVec(heading)
	r := Vec{f.Z, -f.X}
	velocity := f.Mul(speed + longitudinalAcceleration*dt).Add(r.Mul(lateral + lateralAcceleration*dt))
	heading += yawRate * dt
	f = HeadingVec(heading)
	r = Vec{f.Z, -f.X}
	return velocity.X*f.X + velocity.Z*f.Z, velocity.X*r.X + velocity.Z*r.Z, heading
}

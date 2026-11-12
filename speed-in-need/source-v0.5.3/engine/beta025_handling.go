package engine

import "math"

// CornerDriveReserve025 is calculated once per tire step. Driven patches keep
// room for lateral force without changing heading or the steering sign.
func CornerDriveReserve025(mu, speed, rack, wheelbase, recovery float64) float64 {
	if mu <= 0 || wheelbase <= 0 {
		return 1
	}
	lateral := math.Abs(speed * speed * math.Tan(rack) / wheelbase)
	use := Clamp(lateral/(mu*9.81), 0, .82) * GripRecovery024(recovery)
	return math.Sqrt(math.Max(0, 1-use*use))
}

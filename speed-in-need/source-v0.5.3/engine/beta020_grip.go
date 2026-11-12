package engine

// Normal driving uses grippier street/racing tires. The handbrake and its
// recovery retain the previous drift tune; upgrades cannot silently erase it.
func DrivingGripCoefficient(spec RemasteredVehicle, onRoad, handbrake bool, recovery float64) float64 {
	base := spec.Mu
	if !onRoad {
		if handbrake {
			return base * .69
		}
		return base * (.69 + .11*GripRecovery024(recovery))
	}
	if handbrake {
		return base
	}
	return base * (1 + .48*GripRecovery024(recovery))
}

// C1 continuity removes the change in recovery rate at either end of a slide.
func GripRecovery024(recovery float64) float64 {
	t := 1 - Clamp(recovery/1.05, 0, 1)
	return t * t * (3 - 2*t)
}

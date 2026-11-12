package engine

import "math"

// SpeedCameraShake is a continuous, bounded render offset in vehicle-local
// metres. It never feeds back into steering, physics or the camera spring.
func SpeedCameraShake(speed, seconds float64, cockpit bool) [3]float64 {
	if math.IsNaN(speed) || math.IsInf(speed, 0) || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return [3]float64{}
	}
	a := .048 * Clamp(math.Abs(speed)/75, 0, 1)
	if cockpit {
		a *= .38
	}
	return [3]float64{
		a * (.65*math.Sin(seconds*37) + .35*math.Sin(seconds*61+.7)),
		a * (.6*math.Sin(seconds*43+.4) + .4*math.Sin(seconds*71)),
		a * .25 * math.Sin(seconds*31+1.2),
	}
}

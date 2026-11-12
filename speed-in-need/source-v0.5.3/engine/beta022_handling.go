package engine

import "math"

// ProgressiveSteer advances a physical rack angle; it never changes heading.
// Small high-speed rack angles respond promptly; parking steering still builds
// progressively over a longer hold, with a faster return/counter.
func ProgressiveSteer(current, input, limit, speed, dt float64) float64 {
	input = Clamp(input, -1, 1)
	target := input * limit
	// Parking remains progressive. Once travelling, response stays prompt
	// through the 60–150km/h range instead of feeling like a heavy rack.
	timeToStop := Lerp(.30, .16, Clamp(math.Abs(speed)/(60./3.6), 0, 1))
	timeToStop = Lerp(timeToStop, .105, motorwaySteerBlend040(speed))
	rate := math.Max(.015, limit/timeToStop)
	rate *= steeringResponse041(speed)
	if math.Abs(input) < .02 {
		rate = math.Max(.04, limit/.105)
	} else if target*current < 0 {
		rate *= 1.8
	}
	if math.Abs(current) > limit {
		rate = math.Max(rate, 2.5)
	}
	return current + Clamp(target-current, -rate*dt, rate*dt)
}

// Faster steering rack movement, with a gradual return to motorway response.
// Tire forces and steering angle limits continue to bound lateral acceleration.
func steeringResponse041(speed float64) float64 {
	t := Clamp((math.Abs(speed)*3.6-120)/60, 0, 1)
	return Lerp(1.40, 1, t*t*(3-2*t))
}

func TrafficCornerPose(route Route, distance, direction, lane, speed float64) (Vec, float64) {
	return TrafficPose026(route, distance, direction, lane)
}

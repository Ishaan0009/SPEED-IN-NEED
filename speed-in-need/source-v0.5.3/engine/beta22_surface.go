package engine

import "math"

// OffroadResistance is the additional rolling/soil force, opposing velocity.
// At zero speed there is no artificial static force or reversal impulse.
func OffroadResistance(mass, normal, speed float64) float64 {
	if math.Abs(speed) < .01 {
		return 0
	}
	return math.Tanh(speed*2)*normal*.008 + mass*speed*.018
}

func BrakeMapSpan(braking bool) float64 {
	if braking {
		return 720
	}
	return 430
}

// WheelVerticalExtent projects a tire cylinder onto world up after steering,
// chassis roll and pitch. Tire spin does not change its supporting radius.
func WheelVerticalExtent(radius, halfWidth, pitch, roll, steer float64) float64 {
	axisY := math.Cos(pitch)*math.Sin(roll)*math.Cos(steer) + math.Sin(pitch)*math.Sin(steer)
	return radius*math.Sqrt(math.Max(0, 1-axisY*axisY)) + halfWidth*math.Abs(axisY)
}

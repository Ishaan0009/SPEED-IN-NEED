package engine

import (
	"math"
	"sort"
)

func routePoint026(r Route, s float64) Vec {
	if r.Closed {
		s = math.Mod(s+r.Len*2, r.Len)
		return r.sampleRaw(s)
	}
	if s < 0 {
		return r.Points[0].Add(r.Points[1].Sub(r.Points[0]).Unit().Mul(s))
	}
	if s > r.Len {
		n := len(r.Points)
		return r.Points[n-1].Add(r.Points[n-1].Sub(r.Points[n-2]).Unit().Mul(s - r.Len))
	}
	return r.sampleRaw(s)
}

// Integrating the polyline makes position and its tangent continuous at knots.
// The lane offset uses that same tangent, rather than rotating a sharp path.
func routeIntegral026(r Route, a, b float64) Vec {
	if r.Closed {
		for a < 0 {
			a += r.Len
			b += r.Len
		}
		for a >= r.Len {
			a -= r.Len
			b -= r.Len
		}
		if b > r.Len {
			return routeIntegral026(r, a, r.Len).Add(routeIntegral026(r, 0, b-r.Len))
		}
	}
	result := Vec{}
	index := sort.SearchFloat64s(r.Cumulative, a+1e-8)
	for a < b-1e-8 {
		next := b
		if index < len(r.Cumulative) && r.Cumulative[index] < b {
			next = r.Cumulative[index]
		}
		if next <= a+1e-8 {
			index++
			continue
		}
		p, q := routePoint026(r, a), routePoint026(r, next)
		result = result.Add(p.Add(q).Mul((next - a) * .5))
		a = next
		index++
	}
	return result
}

func trafficLane026(route Route, distance, direction, lane float64) Vec {
	span := Clamp(math.Abs(lane)*2.5+1.5, 6, 9)
	if route.Len < span*3 {
		span = route.Len * .2
	}
	centre := routeIntegral026(route, distance-span, distance+span).Mul(.5 / span)
	forward := routePoint026(route, distance+span).Sub(routePoint026(route, distance-span)).Unit().Mul(direction)
	return centre.Add(Vec{-forward.Z, forward.X}.Mul(lane))
}

func TrafficPose026(route Route, distance, direction, lane float64) (Vec, float64) {
	if len(route.Points) < 2 || route.Len < .01 {
		return Vec{}, 0
	}
	p := trafficLane026(route, distance, direction, lane)
	f := trafficLane026(route, distance+.08*direction, direction, lane).Sub(trafficLane026(route, distance-.08*direction, direction, lane))
	return p, math.Atan2(f.X, f.Z)
}

func InterpolateTraffic026(a, b Traffic, alpha float64) Traffic {
	if a.Pos.Dist(b.Pos) > 25 {
		return b
	}
	alpha = Clamp(alpha, 0, 1)
	b.Pos = a.Pos.Add(b.Pos.Sub(a.Pos).Mul(alpha))
	b.Heading = a.Heading + math.Remainder(b.Heading-a.Heading, Tau)*alpha
	b.WheelSpin = a.WheelSpin + math.Remainder(b.WheelSpin-a.WheelSpin, Tau)*alpha
	b.SteerAngle = Lerp(a.SteerAngle, b.SteerAngle, alpha)
	b.Pitch = Lerp(a.Pitch, b.Pitch, alpha)
	b.Roll = Lerp(a.Roll, b.Roll, alpha)
	return b
}

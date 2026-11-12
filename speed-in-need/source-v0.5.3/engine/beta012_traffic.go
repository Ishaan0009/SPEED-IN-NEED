package engine

import "math"

// Braking anticipation uses distance in metres and each body's actual length.
func (w *World) trafficRoadSpeed(t *Traffic) float64 {
	route := w.Routes[t.Path]
	limit := t.CruiseSpeed
	deceleration := 3.2
	if t.Variant == 0 {
		deceleration = 2.5
	}
	for _, ahead := range []float64{0, 8, 18, 32, 52, 75} {
		_, a := route.AtDistance(t.T + ahead*t.Direction)
		_, b := route.AtDistance(t.T + (ahead+10)*t.Direction)
		curve := math.Abs(math.Atan2(a.X*b.Z-a.Z*b.X, a.X*b.X+a.Z*b.Z)) / 10
		corner := math.Sqrt(9.81 * .33 / math.Max(curve, .0005))
		if curve > .06 {
			corner *= .8
		}
		if t.Variant == 0 {
			corner *= .85
		}
		limit = math.Min(limit, math.Sqrt(corner*corner+2*deceleration*math.Max(0, ahead-4)))
	}
	if !route.Closed {
		remaining := route.Len - t.T
		if t.Direction < 0 {
			remaining = t.T
		}
		if remaining < 28 {
			limit = math.Min(limit, math.Sqrt(5*math.Max(0, remaining)+16))
		}
	}
	return limit
}

// Snapshot all actors before integration. AI decisions do not depend on update order.
func (w *World) trafficSnapshot() []Traffic {
	s := w.Remaster
	if cap(s.TrafficSnapshot) < len(w.Traffic) {
		s.TrafficSnapshot = make([]Traffic, len(w.Traffic))
	}
	s.TrafficSnapshot = s.TrafficSnapshot[:len(w.Traffic)]
	copy(s.TrafficSnapshot, w.Traffic)
	return s.TrafficSnapshot
}

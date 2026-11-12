package engine

import "math"

// Continue across connected roads instead of making a U-turn at every route end.
func (w *World) connectAokawaTraffic(t *Traffic, overshoot float64) bool {
	current := w.Routes[t.Path]
	end := current.Points[len(current.Points)-1]
	if t.Direction < 0 {
		end = current.Points[0]
	}
	oldForward := HeadingVec(t.Heading)
	best := -10.
	path := -1
	distance, direction := 0., 1.
	for id, route := range w.Routes {
		if id == t.Path || w.Remaster.Def.Roads[id].Kind == "stunt" || !w.civilianRoute035(id) {
			continue
		}
		at, _ := routeProjection(route, end, -1)
		p, f := route.AtDistance(at)
		endAt := current.Len
		if t.Direction < 0 {
			endAt = 0
		}
		if p.Dist(end) > 7 || math.Abs(w.roadHeight035(t.Path, endAt)-w.roadHeight035(id, at)) > 2.5 {
			continue
		}
		for _, dir := range []float64{-1, 1} {
			if !route.Closed && ((dir > 0 && at > route.Len-5) || (dir < 0 && at < 5)) {
				continue
			}
			forward := f.Mul(dir)
			dot := oldForward.X*forward.X + oldForward.Z*forward.Z
			if dot < -.35 {
				continue
			}
			score := dot - float64(id)*.0001
			if score > best {
				best = score
				path = id
				distance = at
				direction = dir
			}
		}
	}
	if path < 0 {
		return false
	}
	oldPos, oldHeading := t.Pos, t.Heading
	t.Path = path
	t.Direction = direction
	t.T = Clamp(distance+overshoot*direction, 0, w.Routes[path].Len)
	wanted, heading := TrafficCornerPose(w.Routes[path], t.T, direction, t.Lane, t.Speed)
	t.Offset = oldPos.Sub(wanted)
	t.YawKick = math.Atan2(math.Sin(oldHeading-heading), math.Cos(oldHeading-heading))
	t.Speed = math.Min(t.Speed, 9)
	return true
}

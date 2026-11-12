package engine

import "math"

// ordinaryRoad035 separates seam suppression from genuine stunt-ramp flight.
func (w *World) ordinaryRoad035(c *Car) bool {
	if !c.OnRoad || w.Remaster == nil {
		return false
	}
	_, width, route := w.RemasteredRoadAt(c.Pos)
	if route < 0 || route >= len(w.Remaster.Def.Roads) || width < 3 {
		return false
	}
	return w.Remaster.Def.Roads[route].Kind != "stunt"
}

// BuildTrafficCore035 prunes dead-end road branches from the civilian network.
// It never edits road definitions, the waypoint navigator, or player access.
// Endpoint-to-polyline attachments handle authored T-junctions as well as joins.
func (w *World) BuildTrafficCore035() []bool {
	s := w.Remaster
	if len(s.TrafficCore035) == len(w.Routes) {
		return s.TrafficCore035
	}
	core := make([]bool, len(w.Routes))
	links := make([][2][]int, len(w.Routes))
	for i, r := range w.Routes {
		core[i] = len(r.Points) >= 2 && r.Len > 10 && (i >= len(s.Def.Roads) || s.Def.Roads[i].Kind != "stunt")
		if !core[i] || r.Closed {
			continue
		}
		for side, end := range []Vec{r.Points[0], r.Points[len(r.Points)-1]} {
			for j, other := range w.Routes {
				if i == j || len(other.Points) < 2 {
					continue
				}
				at, _ := routeProjection(other, end, -1)
				p, _ := other.AtDistance(at)
				endAt := 0.
				if side == 1 {
					endAt = r.Len
				}
				if p.Dist(end) <= 7 && math.Abs(w.roadHeight035(i, endAt)-w.roadHeight035(j, at)) < 2.5 {
					links[i][side] = append(links[i][side], j)
				}
			}
		}
	}
	for changed := true; changed; {
		changed = false
		for i, r := range w.Routes {
			if !core[i] || r.Closed {
				continue
			}
			for side := 0; side < 2; side++ {
				connected := false
				for _, j := range links[i][side] {
					connected = connected || core[j]
				}
				if !connected {
					core[i], changed = false, true
					break
				}
			}
		}
	}
	s.TrafficCore035 = core
	return core
}

func (w *World) roadHeight035(id int, at float64) float64 {
	r := w.Routes[id]
	if id < len(w.Remaster.Def.Roads) {
		e := w.Remaster.Def.Roads[id].Elevations
		if len(e) == len(r.Points) {
			for i := 1; i < len(r.Cumulative); i++ {
				if at <= r.Cumulative[i] {
					f := Clamp((at-r.Cumulative[i-1])/math.Max(.001, r.Cumulative[i]-r.Cumulative[i-1]), 0, 1)
					return Lerp(e[i-1], e[i], f)
				}
			}
			return e[len(e)-1]
		}
	}
	p, _ := r.AtDistance(at)
	return w.RemasteredGround(p)
}

func (w *World) civilianRoute035(id int) bool {
	if w.Remaster == nil || !w.Remaster.Def.Beta020 {
		return true
	}
	core := w.BuildTrafficCore035()
	return id >= 0 && id < len(core) && core[id]
}

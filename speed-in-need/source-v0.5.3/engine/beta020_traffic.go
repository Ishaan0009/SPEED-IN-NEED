package engine

import "math"

// The normal preset adds 1000 civilians to its previous 32-car population.
// Road-space allocation and right-hand lanes keep the initial population clear.
func (w *World) beta020TrafficPool() []Traffic {
	const target = 1056
	total := 0.0
	for i, r := range w.Routes {
		if r.Len >= 130 && w.Remaster.Def.Roads[i].Kind != "stunt" && w.civilianRoute035(i) {
			total += r.Len
		}
	}
	if total == 0 {
		return nil
	}
	pool := make([]Traffic, 0, target)
	distance := 0.0
	for ri, r := range w.Routes {
		if r.Len < 130 || w.Remaster.Def.Roads[ri].Kind == "stunt" || !w.civilianRoute035(ri) {
			continue
		}
		start := distance
		distance += r.Len
		begin, end := int(math.Round(start/total*target)), int(math.Round(distance/total*target))
		count := end - begin
		for i := 0; i < count; i++ {
			at := (float64(i) + .5) * r.Len / float64(count)
			direction := 1.0
			if i%2 == 1 {
				direction = -1
			}
			lane := math.Min(2.7, w.Remaster.Def.Roads[ri].Width*.25)
			cruise := 12 + float64((i+ri)%5)*1.8
			if w.Remaster.Def.Roads[ri].Kind == "town" {
				cruise = 7 + float64(i%4)
			}
			p, heading := TrafficCornerPose(r, at, direction, lane, cruise*.55)
			pool = append(pool, Traffic{Path: ri, T: at, Direction: direction, Lane: lane, Speed: cruise * .55, CruiseSpeed: cruise, Variant: (i*7 + ri) % 14, Pos: p, Heading: heading})
		}
	}
	return pool
}

func trafficCell020(p Vec) gridKey {
	return gridKey{int(math.Floor(p.X / 40)), int(math.Floor(p.Z / 40))}
}

func (s *RemasteredState) beta020TrafficGrid(snapshot []Traffic) {
	if s.TrafficGrid020 == nil {
		s.TrafficGrid020 = make(map[gridKey][]int)
	}
	// Clear only cells occupied last tick. Reusing bucket storage avoids both
	// per-tick allocations and scanning every historical cell after a long drive.
	for _, k := range s.TrafficCells038 {
		s.TrafficGrid020[k] = s.TrafficGrid020[k][:0]
	}
	s.TrafficCells038 = s.TrafficCells038[:0]
	for i, t := range snapshot {
		k := trafficCell020(t.Pos)
		if len(s.TrafficGrid020[k]) == 0 {
			s.TrafficCells038 = append(s.TrafficCells038, k)
		}
		s.TrafficGrid020[k] = append(s.TrafficGrid020[k], i)
	}
	if len(s.TrafficDesired020) != len(snapshot) {
		s.TrafficDesired020 = make([]float64, len(snapshot))
		for i, t := range snapshot {
			s.TrafficDesired020[i] = t.Speed
		}
	}
	s.TrafficTick020++
}

func (s *RemasteredState) beta020TrafficNearby(p Vec) []int {
	s.TrafficScratch020 = s.TrafficScratch020[:0]
	k := trafficCell020(p)
	for x := k.X - 2; x <= k.X+2; x++ {
		for z := k.Z - 2; z <= k.Z+2; z++ {
			s.TrafficScratch020 = append(s.TrafficScratch020, s.TrafficGrid020[gridKey{x, z}]...)
		}
	}
	return s.TrafficScratch020
}

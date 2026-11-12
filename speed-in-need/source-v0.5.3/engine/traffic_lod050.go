package engine

// Keep full-rate traffic around every participating driver. Far traffic still
// follows signals and routes, accumulating the actual elapsed time at 10 Hz.
func (w *World) observerDistanceSquared050(p Vec) float64 {
	delta := p.Sub(w.Car.Pos)
	best := delta.X*delta.X + delta.Z*delta.Z
	for _, observer := range w.NetworkObservers050 {
		d := p.Sub(observer)
		distance := d.X*d.X + d.Z*d.Z
		if distance < best {
			best = distance
		}
	}
	return best
}
func (w *World) trafficStep050(t *Traffic, index int, dt, distanceSquared float64) float64 {
	if !w.SimulationLOD050 {
		return dt
	}
	t.stepDebt050 += dt
	if distanceSquared > 400*400 && t.ImpactCooldown <= 0 && t.Offset.Length() < .05 && (w.Remaster.TrafficTick020+uint64(index))%12 != 0 {
		return 0
	}
	elapsed := t.stepDebt050
	t.stepDebt050 = 0
	return elapsed
}

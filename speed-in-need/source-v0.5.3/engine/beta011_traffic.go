package engine

// Density is bounded to keep short-block city traffic from multiplying per route.
func (w *World) SetTrafficDensity(level int) {
	if w.Remaster == nil || w.Remaster.Mode != 0 {
		return
	}
	s := w.Remaster
	if len(s.TrafficPool) == 0 {
		s.TrafficPool = append([]Traffic(nil), w.Traffic...)
	}
	count := []int{0, 16, 32, 56}[ClampInt(level, 0, 3)]
	if s.Def.Beta020 {
		count = []int{0, 256, 1032, 1056}[ClampInt(level, 0, 3)]
	}
	count = ClampInt(count, 0, len(s.TrafficPool))
	w.Traffic = make([]Traffic, count)
	for i := 0; i < count; i++ {
		index := i * len(s.TrafficPool) / count
		w.Traffic[i] = s.TrafficPool[index]
		w.Traffic[i].Variant = (i*7 + index) % 14
	}
}

func TrafficBodyFootprint(variant int) Vec {
	if variant >= 4 {
		return Vec{1.00, 2.16}
	}
	return [4]Vec{{1.23, 5.35}, {.98, 2.30}, {.99, 2.14}, {1.23, 3.52}}[variant%4]
}
func TrafficBodyMass(variant int) float64 {
	if variant >= 4 {
		return [10]float64{1280, 1390, 1430, 1620, 1650, 1400, 1370, 1210, 1580, 1420}[(variant-4)%10]
	}
	return [4]float64{11500, 1780, 1350, 6500}[variant%4]
}

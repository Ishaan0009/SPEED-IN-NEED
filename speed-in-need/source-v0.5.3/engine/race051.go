package engine

// Hard fields competitive performance cars, rather than asking a stock 132 kW
// drift coupe to keep up with the player's supercar. Physics remains shared.
func (w *World) SetRaceDifficulty051(level int) {
	if w.Remaster == nil {
		return
	}
	w.Remaster.RaceDifficulty040 = ClampInt(level, 0, 2)
	if level != 2 {
		return
	}
	for i := range w.Remaster.Racers {
		r := &w.Remaster.Racers[i]
		r.Class = []int{0, 2, 0, 2}[i%4]
		r.Car.Variant050 = []int{2, 3, 1, 0}[i%4]
		r.Car.Tyre050 = 1
		r.Car.Body.Initialized = false
		r.Car.TiresReady = false
	}
}

package engine

func (w *World) stuntRoad036(c *Car) bool {
	if w.Remaster == nil {
		return false
	}
	distance, width, id := w.RemasteredRoadAt(c.Pos)
	return id >= 0 && id < len(w.Remaster.Def.Roads) && distance < width*.5+1 && w.Remaster.Def.Roads[id].Kind == "stunt"
}

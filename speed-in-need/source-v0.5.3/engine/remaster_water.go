package engine

import "math"

// WaterArea represents the actual visible shoreline. The coastal renderer and
// the vehicle solver use the same bounds, instead of one global X-coordinate
// acting as an invisible wall outside the ocean.
type WaterArea struct{ Polygon []Vec }

const (
	CoastWaterLevel = -1.0
	CoastWaterMinX  = -680.0
	CoastWaterMaxX  = -205.0
	CoastWaterMinZ  = -540.0
	CoastWaterMaxZ  = 540.0
)

func WaterForMap(mapID int) []WaterArea {
	if mapID != 0 {
		return nil
	}
	return []WaterArea{{Polygon: []Vec{{CoastWaterMinX, CoastWaterMinZ}, {CoastWaterMaxX, CoastWaterMinZ}, {CoastWaterMaxX, CoastWaterMaxZ}, {CoastWaterMinX, CoastWaterMaxZ}}}}
}

func (w WaterArea) Contains(p Vec) bool {
	inside := false
	n := len(w.Polygon)
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		a, b := w.Polygon[i], w.Polygon[j]
		if (a.Z > p.Z) != (b.Z > p.Z) && p.X < (b.X-a.X)*(p.Z-a.Z)/(b.Z-a.Z)+a.X {
			inside = !inside
		}
	}
	return inside
}

func (w WaterArea) ClosestDry(p Vec) Vec {
	best := p
	min := math.Inf(1)
	outward := Vec{1, 0}
	// Positive signed area denotes CCW polygon and an outward right normal.
	signedArea := 0.0
	for i := 0; i < len(w.Polygon); i++ {
		a, b := w.Polygon[i], w.Polygon[(i+1)%len(w.Polygon)]
		signedArea += a.X*b.Z - b.X*a.Z
	}
	for i := 0; i < len(w.Polygon); i++ {
		a, b := w.Polygon[i], w.Polygon[(i+1)%len(w.Polygon)]
		line := b.Sub(a)
		t := Clamp((p.Sub(a).X*line.X+p.Sub(a).Z*line.Z)/math.Max(line.X*line.X+line.Z*line.Z, 1e-9), 0, 1)
		q := a.Add(line.Mul(t))
		if d := q.Dist(p); d < min {
			min, best = d, q
			outward = Vec{line.Z, -line.X}.Unit()
			if signedArea < 0 {
				outward = outward.Mul(-1)
			}
		}
	}
	return best.Add(outward.Mul(2.15))
}

func (w *World) RejectWater() {
	for _, water := range w.Water {
		if water.Contains(w.Car.Pos) {
			// An airborne car may fly over water; recover only at the surface.
			if w.Car.Body.Initialized && !w.Car.Body.Grounded && w.Car.Body.AirTime > .04 && w.Car.Body.Height > CoastWaterLevel+1.4 {
				continue
			}
			w.Car.Pos = water.ClosestDry(w.Car.Pos)
			w.Car.Speed *= .17
			w.Car.Lateral *= .23
			w.Car.Damage = Clamp(w.Car.Damage+0.5, 0, 100)
			w.InitializeChassis()
			return
		}
	}
}

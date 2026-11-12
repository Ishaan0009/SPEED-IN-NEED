package engine

import (
	"encoding/json"
	"errors"
	"math"
	"sort"
)

const RemasteredVersion = "0.3.0"

// WorldDefinition is built from the inspected GLB inputs. Render placements,
// drivable surfaces and collision geometry travel together in this document.
type HeightGrid struct {
	W, H             int
	MinX, MinZ, Step float64
	Values           []float32
}

func (g HeightGrid) Height(p Vec) float64 {
	if g.W < 2 || g.H < 2 || g.Step <= 0 || len(g.Values) != g.W*g.H {
		return Terrain(p.X, p.Z)
	}
	x := Clamp((p.X-g.MinX)/g.Step, 0, float64(g.W-1)-1e-7)
	z := Clamp((p.Z-g.MinZ)/g.Step, 0, float64(g.H-1)-1e-7)
	i, j := int(x), int(z)
	u, v := x-float64(i), z-float64(j)
	a, b, c, d := float64(g.Values[j*g.W+i]), float64(g.Values[j*g.W+i+1]), float64(g.Values[(j+1)*g.W+i]), float64(g.Values[(j+1)*g.W+i+1])
	if u+v <= 1 {
		return a + (b-a)*u + (c-a)*v
	}
	return d + (c-d)*(1-u) + (b-d)*(1-v)
}

type RoadDefinition struct {
	Elevations []float64
	Name       string
	Points     []Vec
	Width      float64
	Kind       string
}
type PlacedObject struct {
	Asset, Kind     string
	Pos             [3]float64
	Scale           [3]float64
	Heading, Radius float64
	Solid           bool
	Collider        int
}
type ColliderShape struct {
	Center                      Vec
	Half                        Vec
	Heading, MinY, MaxY, Radius float64
	Object                      int
	Kind                        string
}
type SurfaceTriangle struct {
	A, B, C [3]float64
	Road    bool
	Support bool
}
type TunnelDefinition struct {
	Center                               Vec
	Heading, Length, Width, StartY, EndY float64
	PortalApron, ApronStartY, ApronEndY  float64
	LiningInset                          float64
}

func (t TunnelDefinition) Coordinates(p Vec) (float64, float64) {
	f := HeadingVec(t.Heading)
	r := Vec{f.Z, -f.X}
	d := p.Sub(t.Center)
	return d.X*r.X + d.Z*r.Z, d.X*f.X + d.Z*f.Z
}
func (t TunnelDefinition) Floor(p Vec) (float64, bool) {
	side, along := t.Coordinates(p)
	half := t.Length * .5
	y := Lerp(t.StartY, t.EndY, Clamp(along/t.Length+.5, 0, 1))
	if t.PortalApron > 0 && math.Abs(along) > half {
		u := Clamp((math.Abs(along)-half)/t.PortalApron, 0, 1)
		if along < 0 {
			y = Lerp(t.StartY, t.ApronStartY, u)
		} else {
			y = Lerp(t.EndY, t.ApronEndY, u)
		}
	}
	return y, math.Abs(side) < t.Width*.5 && math.Abs(along) < half+t.PortalApron
}

type SignalDefinition struct {
	Direction                float64
	Pos                      Vec
	Route                    int
	Distance, Heading, Phase float64
}
type WorldDefinition struct {
	ReliefReady051         bool       `json:"-"`
	ReliefOriginal051      HeightGrid `json:"-"`
	CollisionFootprints040 map[string]CollisionFootprint040
	Beta026                bool
	Beta020                bool
	VegetationFile         string
	Atlas                  AtlasDefinition
	PassSpawn              Vec
	BeachSpawn             Vec
	PassHeading            float64
	BridgeSpans            []BridgeDefinition
	WaterGrid              HeightGrid
	WaterHeight            float64
	Name                   string
	AreaKM2                float64
	Grid                   HeightGrid
	Roads                  []RoadDefinition
	Objects                []PlacedObject
	Colliders              []ColliderShape
	Surfaces               []SurfaceTriangle
	Tunnels                []TunnelDefinition
	Signals                []SignalDefinition
	Water                  []WaterArea
	Spawn                  Vec
	SpawnHeading           float64
	Bounds                 [2]Vec
	TerrainTiles           []string
	CollisionTriangles     []SurfaceTriangle
	CameraTriangles        []SurfaceTriangle
	WaterLevels            []float64
}

type AtlasDefinition struct {
	Image          string
	OriginPixel    [2]float64
	Size           [2]int
	MetresPerPixel float64
}

type BridgeDefinition struct {
	Route      int
	Start, End Vec
	Y, Width   float64
}

func DecodeWorldDefinition(raw []byte) (*WorldDefinition, error) {
	d := new(WorldDefinition)
	if err := json.Unmarshal(raw, d); err != nil {
		return nil, err
	}
	if len(d.Roads) == 0 || d.Grid.W < 2 || len(d.Grid.Values) != d.Grid.W*d.Grid.H {
		return nil, errors.New("invalid remastered world data")
	}
	for _, r := range d.Roads {
		if len(r.Points) < 2 || r.Width < 3 {
			return nil, errors.New("invalid road definition")
		}
	}
	for i, c := range d.Colliders {
		if c.Object < 0 || c.Object >= len(d.Objects) || !d.Objects[c.Object].Solid {
			return nil, errors.New("collider without a visible solid object")
		}
		d.Colliders[i].Object = c.Object
	}
	return d, nil
}

type RemasteredState struct {
	RaceDifficulty040                                     int
	TrafficCore035                                        []bool
	SignalsByRoute021                                     [][]int
	SignalCount021                                        int
	TrafficPool                                           []Traffic
	TrafficGrid020                                        map[gridKey][]int
	TrafficScratch020                                     []int
	TrafficDesired020                                     []float64
	TrafficTick020                                        uint64
	TrafficCells038                                       []gridKey
	TrafficSnapshot                                       []Traffic
	DynamicProps                                          map[int]*DynamicProp
	Def                                                   *WorldDefinition
	Mode, Time, Laps, Opponents                           int
	Countdown, RaceTime, FinishTime, LastLapTime, BestLap float64
	Lap, NextGate, Position                               int
	PlayerProgress, PlayerLastRouteDistance               float64
	PlayerSegment                                         int
	Gates                                                 []Vec
	Racers                                                []Racer
	Started, Finished                                     bool
	CrossedStart                                          bool
	staticGrid                                            map[gridKey][]int
	surfaceGrid                                           map[gridKey][]int
	roadGrid                                              map[gridKey][]roadSegment
	contactMarks                                          []uint32
	contactSerial                                         uint32
	triangleGrid                                          map[gridKey][]int
	triangleBounds                                        [][6]float64
	cameraGrid                                            map[gridKey][]int
	FallenPoles                                           map[int]*Pole
	PickupCooldown                                        map[int]float64
	PickupIndices                                         []int
}
type Racer struct {
	DriftPulse040, DriftCooldown040, IntentionalDriftSeconds040 float64
	DriftPeakAngle040, BoostSeconds040                          float64
	IntentionalDrifts040                                        int
	LaneOffset                                                  float64
	NitroSpent                                                  float64
	StuckTime, RecoveryTime                                     float64
	Name                                                        string
	Class                                                       int
	Car                                                         Car
	Progress, RouteDistance, LastDistance, NitroTimer           float64
	Lap, NextGate, Segment                                      int
	Finished                                                    bool
	CrossedStart                                                bool
	FinishTime                                                  float64
}
type roadSegment struct {
	A, B           Vec
	Width          float64
	Route, Index   int
	Offset, Length float64
}

func NewRemasteredWorld(d *WorldDefinition, mode, car, timeOfDay, laps, opponents int) *World {
	w := &World{Map: ClampInt(mode, 0, 1), CarClass: ClampInt(car, 0, 3), DriveMode: 1}
	s := &RemasteredState{RaceDifficulty040: 1, Def: d, Mode: ClampInt(mode, 0, 1), Time: ClampInt(timeOfDay, 0, 2), Laps: ClampInt(laps, 1, 8), Opponents: ClampInt(opponents, 0, 7), Position: 1}
	w.Remaster = s
	for _, r := range d.Roads {
		w.Routes = append(w.Routes, newRoute(r.Name, append([]Vec(nil), r.Points...)))
	}
	w.Water = append([]WaterArea(nil), d.Water...)
	s.buildSpatial(w)
	s.buildMeshContacts()
	w.ResetRemastered()
	if d.Beta020 && mode == 0 {
		w.Traffic = w.beta020TrafficPool()
	}
	w.Remaster.TrafficPool = append([]Traffic(nil), w.Traffic...)
	w.SetTrafficDensity(2)
	return w
}
func (s *RemasteredState) buildSpatial(w *World) {
	s.staticGrid = map[gridKey][]int{}
	for i, c := range s.Def.Colliders {
		radius := math.Hypot(c.Half.X, c.Half.Z) + c.Radius
		lo, hi := gridOf(c.Center.Sub(Vec{radius, radius})), gridOf(c.Center.Add(Vec{radius, radius}))
		for x := lo.X; x <= hi.X; x++ {
			for z := lo.Z; z <= hi.Z; z++ {
				k := gridKey{x, z}
				s.staticGrid[k] = append(s.staticGrid[k], i)
			}
		}
	}
	s.surfaceGrid = map[gridKey][]int{}
	for i, t := range s.Def.Surfaces {
		lo := remasterMeshCell(Vec{math.Min(t.A[0], math.Min(t.B[0], t.C[0])) - .02, math.Min(t.A[2], math.Min(t.B[2], t.C[2])) - .02})
		hi := remasterMeshCell(Vec{math.Max(t.A[0], math.Max(t.B[0], t.C[0])) + .02, math.Max(t.A[2], math.Max(t.B[2], t.C[2])) + .02})
		if hi.X-lo.X > 100 || hi.Z-lo.Z > 100 {
			continue
		}
		for x := lo.X; x <= hi.X; x++ {
			for z := lo.Z; z <= hi.Z; z++ {
				k := gridKey{x, z}
				s.surfaceGrid[k] = append(s.surfaceGrid[k], i)
			}
		}
	}
	s.roadGrid = map[gridKey][]roadSegment{}
	for ri, r := range w.Routes {
		for i := 1; i < len(r.Points); i++ {
			a, b := r.Points[i-1], r.Points[i]
			width := s.Def.Roads[ri].Width
			seg := roadSegment{a, b, width, ri, i - 1, r.Cumulative[i-1], a.Dist(b)}
			pad := width*.5 + 10
			lo := gridOf(Vec{math.Min(a.X, b.X) - pad, math.Min(a.Z, b.Z) - pad})
			hi := gridOf(Vec{math.Max(a.X, b.X) + pad, math.Max(a.Z, b.Z) + pad})
			for x := lo.X; x <= hi.X; x++ {
				for z := lo.Z; z <= hi.Z; z++ {
					k := gridKey{x, z}
					s.roadGrid[k] = append(s.roadGrid[k], seg)
				}
			}
		}
	}
}
func (w *World) ResetRemastered() {
	s := w.Remaster
	if s == nil {
		return
	}
	w.Elapsed, w.Distance, w.Credits = 0, 0, 0
	w.Finished = false
	w.Challenge = 0
	w.Car = Car{Pos: s.Def.Spawn, Heading: s.Def.SpawnHeading, Nitro: 100, DriftCombo: 1, BrakeTemperature: 25, EngineTemperature: 80}
	w.Traffic = nil
	w.Trees = nil
	w.Rocks = nil
	w.Buildings = nil
	w.Poles = nil
	s.FallenPoles = map[int]*Pole{}
	s.DynamicProps = map[int]*DynamicProp{}
	s.PickupCooldown = map[int]float64{}
	s.PickupIndices = nil
	for i, o := range s.Def.Objects {
		if o.Kind == "pickup" {
			s.PickupIndices = append(s.PickupIndices, i)
		}
	}
	w.Ramps = nil
	w.Hills = nil
	s.Started, s.Finished, s.CrossedStart = false, false, false
	s.RaceTime, s.FinishTime, s.LastLapTime, s.BestLap = 0, 0, 0, 0
	s.Lap, s.NextGate, s.Position = 1, 1, 1
	s.PlayerProgress, s.PlayerLastRouteDistance = 0, 0
	s.Countdown = 0
	s.Racers = nil
	w.InitializeChassis()
	if s.Mode == 1 {
		s.Countdown = 3.5
		route := w.Routes[0]
		gateCount := int(math.Max(24, math.Ceil(route.Len/90)))
		s.Gates = nil
		for i := 0; i < gateCount; i++ {
			p, _ := route.AtDistance(float64(i) * route.Len / float64(gateCount))
			s.Gates = append(s.Gates, p)
		}
		// All starts share the circuit surface; the player starts behind the first row.
		playerStart := math.Max(0, route.Len-14)
		p, f := route.AtDistance(playerStart)
		r := Vec{f.Z, -f.X}
		w.Car.Pos = p.Add(r.Mul(-1.9))
		w.Car.Heading = math.Atan2(f.X, f.Z)
		s.PlayerLastRouteDistance = playerStart
		s.PlayerProgress = -14
		s.NextGate = 0
		w.InitializeChassis()
		for i := 0; i < s.Opponents; i++ {
			dist := math.Mod(route.Len-5-float64(i/2)*7+route.Len, route.Len)
			p, f := route.AtDistance(dist)
			lane := -1.9
			if i%2 == 1 {
				lane = 1.9
			}
			actor := Racer{Name: "RACER " + string(rune('A'+i)), Class: i % 4, RouteDistance: dist, LastDistance: dist, Progress: -5 - float64(i/2)*7, Lap: 1, NextGate: 0, LaneOffset: lane}
			actor.Car = Car{Pos: p.Add(Vec{f.Z, -f.X}.Mul(lane)), Heading: math.Atan2(f.X, f.Z), Nitro: 100, DriftCombo: 1, BrakeTemperature: 25, EngineTemperature: 80}
			old, oldClass := w.Car, w.CarClass
			w.Car, w.CarClass = actor.Car, actor.Class
			w.InitializeChassis()
			actor.Car = w.Car
			w.Car, w.CarClass = old, oldClass
			s.Racers = append(s.Racers, actor)
		}
	} else {
		// Population is tied to road length and districts, with deterministic spacing.
		for ri, r := range w.Routes {
			if s.Def.Roads[ri].Kind == "stunt" || r.Len < 130 {
				continue
			}
			count := ClampInt(int(r.Len/420), 2, 12)
			for i := 0; i < count; i++ {
				dist := (float64(i) + .37) * r.Len / float64(count)
				p, f := r.AtDistance(dist)
				direction := 1.
				if i%2 == 1 {
					direction = -1
					f = f.Mul(-1)
				}
				speed := 12. + float64((i+ri)%5)*1.8
				if s.Def.Roads[ri].Kind == "town" {
					speed = 8. + float64(i%3)
				}
				w.Traffic = append(w.Traffic, Traffic{Path: ri, T: dist, Speed: speed * .7, CruiseSpeed: speed, Direction: direction, Lane: 2.7, Variant: (i + ri) % 4, Pos: p.Add(Vec{-f.Z, f.X}.Mul(2.7)), Heading: math.Atan2(f.X, f.Z)})
			}
		}
	}
	w.SeedPedestrians()
}
func (w *World) RemasteredRoadAt(p Vec) (distance, width float64, route int) {
	if w.Remaster == nil {
		return DistRoad(p, w.Routes), 14, 0
	}
	best := math.Inf(1)
	width = 12
	route = -1
	for _, seg := range w.Remaster.roadGrid[gridOf(p)] {
		d := seg.B.Sub(seg.A)
		l2 := d.X*d.X + d.Z*d.Z
		if l2 < 1e-8 {
			continue
		}
		t := Clamp((p.Sub(seg.A).X*d.X+p.Sub(seg.A).Z*d.Z)/l2, 0, 1)
		q := seg.A.Add(d.Mul(t))
		dist := p.Dist(q)
		if dist < best {
			best, width, route = dist, seg.Width, seg.Route
		}
	}
	return best, width, route
}
func triangleHeight(t SurfaceTriangle, p Vec) (float64, bool) {
	ax, az := t.A[0], t.A[2]
	bx, bz := t.B[0], t.B[2]
	cx, cz := t.C[0], t.C[2]
	den := (bz-cz)*(ax-cx) + (cx-bx)*(az-cz)
	if math.Abs(den) < 1e-9 {
		return 0, false
	}
	a := ((bz-cz)*(p.X-cx) + (cx-bx)*(p.Z-cz)) / den
	b := ((cz-az)*(p.X-cx) + (ax-cx)*(p.Z-cz)) / den
	c := 1 - a - b
	if a < -.0001 || b < -.0001 || c < -.0001 {
		return 0, false
	}
	return a*t.A[1] + b*t.B[1] + c*t.C[1], true
}
func (w *World) RemasteredGround(p Vec) float64 {
	return w.remasteredGroundAt(p, math.Inf(-1))
}
func (w *World) remasteredSupport(c *Car, p Vec) float64 {
	reference := math.Inf(-1)
	if c.Body.Initialized {
		reference = c.Body.Height - .74
	}
	return w.remasteredGroundAt(p, reference)
}
func (w *World) remasteredGroundAt(p Vec, reference float64) float64 {
	s := w.Remaster
	h := s.Def.Grid.Height(p)
	// A bridge replaces the river bed across its whole deck. Sidewalk triangles
	// from an intersecting road must not pull a wheel down underneath the span.
	for _, b := range s.Def.BridgeSpans {
		delta := b.End.Sub(b.Start)
		length := delta.Length()
		if length < .01 {
			continue
		}
		f := delta.Mul(1 / length)
		q := p.Sub(b.Start)
		along := q.X*f.X + q.Z*f.Z
		side := math.Abs(q.X*f.Z - q.Z*f.X)
		if along >= -.001 && along <= length+.001 && side <= b.Width*.5+2.7 {
			if side > b.Width*.5+.02 {
				return b.Y + .18
			}
			return b.Y
		}
	}
	if s.Def.Name == "AOKAWA BAY" {
		distance := math.Max(math.Abs(p.X), math.Abs(p.Z)) - 1500
		if distance > 0 {
			// The distant mountain mesh uses this same continuous extension.
			return h + math.Sin(math.Min(1, distance/700)*math.Pi/2)*(160+100*math.Sin(p.X*.003+p.Z*.002)+80*math.Cos(p.Z*.004-p.X*.002))
		}
	}
	// A tunnel's road and the mountain above it are two separate support layers.
	// Cars landing on the restored mountain roof must not fall through to the bore.
	if !math.IsInf(reference, -1) && reference >= h-1 {
		for _, t := range s.Def.Tunnels {
			if floor, inside := t.Floor(p); inside && h > floor+4 {
				return h
			}
		}
	}
	if s.Mode == 1 && len(s.Def.Roads[0].Elevations) == len(s.Def.Roads[0].Points) {
		best := 30.
		road := s.Def.Roads[0]
		for _, seg := range s.roadGrid[gridOf(p)] {
			if seg.Route != 0 {
				continue
			}
			v := seg.B.Sub(seg.A)
			u := Clamp((p.Sub(seg.A).X*v.X+p.Sub(seg.A).Z*v.Z)/math.Max(1e-9, seg.Length*seg.Length), 0, 1)
			distance := p.Dist(seg.A.Add(v.Mul(u)))
			if distance < best {
				best = distance
				reference = Lerp(road.Elevations[seg.Index], road.Elevations[seg.Index+1], u)
			}
		}
	}
	best, bestScore := 0., math.Inf(1)
	for _, i := range s.surfaceGrid[remasterMeshCell(p)] {
		if y, ok := triangleHeight(s.Def.Surfaces[i], p); ok {
			if s.Def.Surfaces[i].Road || s.Def.Surfaces[i].Support {
				score := y
				if !math.IsInf(reference, -1) {
					score = math.Abs(y - reference)
					if y > reference+1.2 {
						score += 100
					}
				}
				if score < bestScore {
					bestScore = score
					best = y
				}
				continue
			}
			limit := h + .55
			if !math.IsInf(reference, -1) {
				limit = reference + 1.2
			}
			if y > h && y < limit {
				h = y
			}
		}
	}
	if !math.IsInf(bestScore, 1) {
		return best
	}
	for _, t := range s.Def.Tunnels {
		if y, ok := t.Floor(p); ok {
			_, along := t.Coordinates(p)
			if math.Abs(along) >= t.Length*.5 {
				return h
			}
			return y
		}
	}
	return h
}
func (w *World) InsideTunnel(p Vec) bool {
	if w.Remaster == nil {
		return false
	}
	for _, t := range w.Remaster.Def.Tunnels {
		if _, ok := t.Floor(p); ok {
			return true
		}
	}
	return false
}
func (w *World) TimeHour() float64 {
	if w.Remaster == nil {
		return MapHour(w.Map)
	}
	return [...]float64{16.2, 17.8, 22}[ClampInt(w.Remaster.Time, 0, 2)]
}
func (w *World) RemasteredSignalRed(signal SignalDefinition) bool {
	phase := math.Mod(w.Elapsed+signal.Phase, 36)
	return phase >= 14
}
func (r Route) ClosestDistance(p Vec, hint int) (float64, int) { return routeProjection(r, p, hint) }
func routeProjection(route Route, p Vec, hint int) (float64, int) {
	best := math.Inf(1)
	distance := 0.
	index := 0
	n := len(route.Points) - 1
	if n <= 0 {
		return 0, 0
	}
	// A local window handles every normal tick; global recovery is used after resets.
	start, end := 0, n
	if hint >= 0 && hint < n {
		start = hint - 14
		end = hint + 15
	}
	for k := start; k < end; k++ {
		i := k
		if route.Closed {
			i = (i%n + n) % n
		} else if i < 0 || i >= n {
			continue
		}
		a, b := route.Points[i], route.Points[i+1]
		v := b.Sub(a)
		l2 := v.X*v.X + v.Z*v.Z
		if l2 < 1e-9 {
			continue
		}
		u := Clamp((p.Sub(a).X*v.X+p.Sub(a).Z*v.Z)/l2, 0, 1)
		d := p.Dist(a.Add(v.Mul(u)))
		if d < best {
			best = d
			distance = route.Cumulative[i] + u*math.Sqrt(l2)
			index = i
		}
	}
	if best > 80 && hint >= 0 {
		return routeProjection(route, p, -1)
	}
	return distance, index
}
func (w *World) RacePositions() []int {
	if w.Remaster == nil {
		return nil
	}
	s := w.Remaster
	order := make([]int, len(s.Racers)+1)
	for i := range order {
		order[i] = i - 1
	}
	progress := func(i int) float64 {
		if i < 0 {
			return s.PlayerProgress
		}
		return s.Racers[i].Progress
	}
	finish := func(i int) (bool, float64) {
		if i < 0 {
			return s.Finished, s.FinishTime
		}
		return s.Racers[i].Finished, s.Racers[i].FinishTime
	}
	sort.SliceStable(order, func(i, j int) bool {
		a, at := finish(order[i])
		b, bt := finish(order[j])
		if a != b {
			return a
		}
		if a && b {
			return at < bt
		}
		return progress(order[i]) > progress(order[j])
	})
	return order
}

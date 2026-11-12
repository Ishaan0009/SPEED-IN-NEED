package engine

import (
	"math"
	"sort"
)

const Tau = math.Pi * 2

type Vec struct{ X, Z float64 }

func (a Vec) Add(b Vec) Vec     { return Vec{a.X + b.X, a.Z + b.Z} }
func (a Vec) Sub(b Vec) Vec     { return Vec{a.X - b.X, a.Z - b.Z} }
func (a Vec) Mul(k float64) Vec { return Vec{a.X * k, a.Z * k} }
func (a Vec) Length() float64   { return math.Hypot(a.X, a.Z) }
func (a Vec) Unit() Vec {
	if l := a.Length(); l > 0 {
		return a.Mul(1 / l)
	}
	return Vec{0, 1}
}
func (a Vec) Dist(b Vec) float64 { return a.Sub(b).Length() }
func Clamp(x, a, b float64) float64 {
	if x < a {
		return a
	}
	if x > b {
		return b
	}
	return x
}
func Lerp(a, b, t float64) float64   { return a + (b-a)*t }
func Smooth(a, b, t float64) float64 { return Lerp(a, b, Clamp(t, 0, 1)) }

type Route struct {
	Name       string
	Points     []Vec
	Len        float64
	Cumulative []float64
	Closed     bool
}

func newRoute(name string, p []Vec) Route {
	r := Route{Name: name, Points: p, Cumulative: make([]float64, len(p)), Closed: len(p) > 2 && p[0].Dist(p[len(p)-1]) < 0.001}
	for i := 1; i < len(p); i++ {
		r.Len += p[i].Dist(p[i-1])
		r.Cumulative[i] = r.Len
	}
	return r
}
func Catmull(p0, p1, p2, p3 Vec, t float64) Vec {
	t2 := t * t
	t3 := t * t * t
	return Vec{X: .5 * ((2 * p1.X) + (-p0.X+p2.X)*t + (2*p0.X-5*p1.X+4*p2.X-p3.X)*t2 + (-p0.X+3*p1.X-3*p2.X+p3.X)*t3), Z: .5 * ((2 * p1.Z) + (-p0.Z+p2.Z)*t + (2*p0.Z-5*p1.Z+4*p2.Z-p3.Z)*t2 + (-p0.Z+3*p1.Z-3*p2.Z+p3.Z)*t3)}
}
func SampleSpline(nodes []Vec, close bool, segments int) []Vec {
	out := make([]Vec, 0, len(nodes)*segments+1)
	count := len(nodes)
	if !close {
		count--
	}
	for i := 0; i < count; i++ {
		fetch := func(j int) Vec {
			if close {
				j = (j%len(nodes) + len(nodes)) % len(nodes)
			} else {
				if j < 0 {
					j = 0
				}
				if j >= len(nodes) {
					j = len(nodes) - 1
				}
			}
			return nodes[j]
		}
		for k := 0; k < segments; k++ {
			out = append(out, Catmull(fetch(i-1), fetch(i), fetch(i+1), fetch(i+2), float64(k)/float64(segments)))
		}
	}
	if close {
		out = append(out, out[0])
	} else {
		out = append(out, nodes[len(nodes)-1])
	}
	return out
}
func BuildRoutes() []Route {
	main := SampleSpline([]Vec{{0, -225}, {-22, -120}, {20, -35}, {90, 62}, {180, 75}, {242, 140}, {232, 230}, {132, 278}, {18, 246}, {-102, 185}, {-163, 75}, {-130, -65}, {-110, -180}}, true, 16)
	branch := SampleSpline([]Vec{{-22, -120}, {52, -153}, {166, -101}, {248, -30}, {180, 75}}, false, 17)
	cross := SampleSpline([]Vec{{-163, 75}, {-55, 28}, {20, -35}, {98, -22}, {242, 140}}, false, 17)
	return []Route{newRoute("Coastal Ring", main), newRoute("Valley Link", branch), newRoute("Canyon Pass", cross)}
}

// BuildRoutesForMap produces genuinely different road networks for each map.
// Every route has its own arc-length table, shared by rendering, NPCs and GPS.
func BuildRoutesForMap(mapID int) []Route {
	if mapID == 0 {
		return BuildRoutes()
	}
	if mapID == 1 { // Neon City: broad avenues and cross-town connectors.
		loop := SampleSpline([]Vec{{-220, -190}, {-220, -95}, {-218, 75}, {-200, 200}, {-110, 230}, {30, 235}, {170, 203}, {222, 85}, {222, -45}, {175, -187}, {30, -220}, {-110, -215}}, true, 14)
		avenue := SampleSpline([]Vec{{-218, 75}, {-120, 76}, {-5, 78}, {108, 84}, {222, 85}}, false, 18)
		cross := SampleSpline([]Vec{{-110, 230}, {-91, 145}, {-79, 62}, {-84, -80}, {-110, -215}}, false, 18)
		return []Route{newRoute("City Boulevard", loop), newRoute("Central Avenue", avenue), newRoute("Neon Crossroad", cross)}
	}
	// Canyon: distinct narrow serpentine switchback valley loop.
	loop := SampleSpline([]Vec{{-110, -238}, {-175, -175}, {-145, -92}, {-28, -40}, {95, 5}, {218, 84}, {172, 175}, {72, 213}, {-65, 141}, {-165, 160}, {-220, 57}, {-170, -65}}, true, 17)
	pass := SampleSpline([]Vec{{-145, -92}, {-40, -134}, {76, -105}, {170, -16}, {218, 84}}, false, 16)
	ridge := SampleSpline([]Vec{{-220, 57}, {-112, 13}, {-28, -40}, {50, 62}, {72, 213}}, false, 16)
	return []Route{newRoute("Canyon Loop", loop), newRoute("Ridge Pass", pass), newRoute("Rocky Switchbacks", ridge)}
}

func RawHeight(x, z float64) float64 {
	return 1.8*math.Sin(x*.009) + 1.4*math.Sin(z*.012) + .65*math.Sin(x*.026+z*.019)
}
func Terrain(x, z float64) float64 { return RawHeight(x, z) }
func (r Route) At(t float64) (Vec, Vec) {
	n := len(r.Points) - 1
	if n < 1 {
		return Vec{}, Vec{0, 1}
	}
	t = math.Mod(t, float64(n))
	if t < 0 {
		t += float64(n)
	}
	i := int(t)
	u := t - float64(i)
	a, b := r.Points[i], r.Points[i+1]
	dir := b.Sub(a).Unit()
	return Vec{Lerp(a.X, b.X, u), Lerp(a.Z, b.Z, u)}, dir
}

// AtDistance samples by metres along the road. Treating the raw spline index
// as metres used to make traffic speed and heading jump on uneven segments.
func (r Route) AtDistance(distance float64) (Vec, Vec) {
	if len(r.Points) < 2 || r.Len < 0.00001 {
		return Vec{}, Vec{0, 1}
	}
	if r.Closed {
		distance = math.Mod(distance, r.Len)
		if distance < 0 {
			distance += r.Len
		}
	} else {
		distance = Clamp(distance, 0, r.Len)
	}
	i := sort.Search(len(r.Cumulative), func(i int) bool { return r.Cumulative[i] > distance })
	if i < 1 {
		i = 1
	}
	if i >= len(r.Points) {
		i = len(r.Points) - 1
	}
	a, b := r.Points[i-1], r.Points[i]
	span := r.Cumulative[i] - r.Cumulative[i-1]
	t := 0.0
	if span > 0 {
		t = (distance - r.Cumulative[i-1]) / span
	}
	pos := a.Add(b.Sub(a).Mul(Clamp(t, 0, 1)))
	// Use a finite forward difference to smooth rotation across spline knots.
	delta := math.Max(.8, math.Min(2.0, r.Len/200))
	ahead, behind := distance+delta, distance-delta
	if r.Closed {
		ahead = math.Mod(ahead, r.Len)
		behind = math.Mod(behind+r.Len, r.Len)
	} else {
		ahead = Clamp(ahead, 0, r.Len)
		behind = Clamp(behind, 0, r.Len)
	}
	dir := r.sampleRaw(ahead).Sub(r.sampleRaw(behind)).Unit()
	return pos, dir
}
func (r Route) sampleRaw(distance float64) Vec {
	if len(r.Points) < 2 {
		return Vec{}
	}
	distance = Clamp(distance, 0, r.Len)
	i := sort.Search(len(r.Cumulative), func(i int) bool { return r.Cumulative[i] > distance })
	if i < 1 {
		i = 1
	}
	if i >= len(r.Points) {
		i = len(r.Points) - 1
	}
	span := r.Cumulative[i] - r.Cumulative[i-1]
	t := 0.0
	if span > 0 {
		t = (distance - r.Cumulative[i-1]) / span
	}
	return r.Points[i-1].Add(r.Points[i].Sub(r.Points[i-1]).Mul(Clamp(t, 0, 1)))
}

// RoadNormalAt produces one shared cross-road normal for a spline knot.
// Adjacent quad strips use identical edge vertices, preventing cracks and
// sharp texture changes when the road bends.
func RoadNormalAt(points []Vec, i int) Vec {
	if len(points) < 2 {
		return Vec{X: -1}
	}
	n := len(points)
	closed := n > 3 && points[0].Dist(points[n-1]) < .001
	var dir Vec
	if i <= 0 {
		if closed {
			dir = points[1].Sub(points[n-2]).Unit()
		} else {
			dir = points[1].Sub(points[0]).Unit()
		}
	} else if i >= n-1 {
		if closed {
			dir = points[1].Sub(points[n-2]).Unit()
		} else {
			dir = points[n-1].Sub(points[n-2]).Unit()
		}
	} else {
		before := points[i].Sub(points[i-1]).Unit()
		after := points[i+1].Sub(points[i]).Unit()
		dir = before.Add(after).Unit()
	}
	return Vec{X: -dir.Z, Z: dir.X}
}

func DistRoad(p Vec, routes []Route) float64 {
	best := 1e18
	for _, r := range routes {
		for i := 1; i < len(r.Points); i++ {
			a, b := r.Points[i-1], r.Points[i]
			v := b.Sub(a)
			den := v.X*v.X + v.Z*v.Z
			if den == 0 {
				continue
			}
			t := Clamp(((p.X-a.X)*v.X+(p.Z-a.Z)*v.Z)/den, 0, 1)
			d := p.Sub(a.Add(v.Mul(t)))
			squared := d.X*d.X + d.Z*d.Z
			if squared < best {
				best = squared
			}
		}
	}
	return math.Sqrt(best)
}

type Car struct {
	Variant050, Tyre050, Rim050, Livery050 int
	Wetness050                             float64
	RaceBoundarySegment040                 int
	EngineRunning                          bool
	BrakeInput, RearLockTimer              float64
	RearLocked                             bool
	TiresReady                             bool
	OffroadBlend                           float64
	TireForces                             [4]TireForces
	CoastingLights                         bool

	Braking                                                            bool
	DriftActive                                                        bool
	DriftRecovery, BrakeTemperature, EngineTemperature, ContactImpulse float64
	Pos                                                                Vec
	Heading, Speed                                                     float64
	YawRate, DriftAngle                                                float64
	WheelSpin, VisualSteer                                             float64
	Lateral                                                            float64
	Nitro                                                              float64
	ReverseHold                                                        float64
	Damage                                                             float64
	DriftScore                                                         float64
	DriftCombo                                                         float64
	OnRoad                                                             bool
	Boosting                                                           bool
	Burnout                                                            bool
	Gearbox                                                            Transmission
	Rig                                                                SuspensionRig
	Body                                                               ChassisState
}
type Pickup struct {
	Pos      Vec
	Cooldown float64
}
type Traffic struct {
	stepDebt050                        float64
	WheelSpin, SteerAngle, Pitch, Roll float64
	StalledSeconds                     float64
	Braking                            bool
	Path                               int
	T, Speed                           float64
	CruiseSpeed                        float64
	Direction                          float64
	Lane                               float64
	Variant                            int
	Pos                                Vec
	Heading                            float64
	NearCooldown                       float64
	// Off-route displacement and velocity after physical contact with the player.
	// Smoothly recover to the planned road centreline instead of teleporting.
	Offset         Vec
	Impulse        Vec
	YawKick        float64
	ImpactYaw      float64
	ImpactCooldown float64
	AvoidOffset    float64
}
type Pole struct {
	Pos, FallDirection           Vec
	Heading, Angle, AngularSpeed float64
	Length                       float64
	Falling, Settled             bool
}
type Tree struct {
	Pos  Vec
	Size float64
	Type int
}
type Rock struct {
	Pos            Vec
	Radius, Height float64
}

type Input struct {
	Throttle, Brake, Steer float64
	Handbrake, Boost       bool
	// EngineGated distinguishes controller ignition from keyboard/AI driving.
	EngineGated, EngineEnabled bool
}
type World struct {
	NetworkClient050    bool // The host alone advances AI/traffic; clients predict their own car.
	SimulationLOD050    bool
	NetworkObservers050 []Vec
	Pedestrians         []Pedestrian
	SidewalkPaths       []SidewalkPath
	pedestrianTick      uint64
	pedestrianGrid      map[[2]int][]int
	pedestrianGridCount int
	pedestrianScratch   []int
	Remaster            *RemasteredState
	Hills               []Hill
	// CarClass: 0=296 track car, 1=classic 911; DriveMode: 0=Tour, 1=Sport, 2=Drift.
	CarClass, DriveMode                             int
	Car                                             Car
	Routes                                          []Route
	Pickups                                         []Pickup
	Traffic                                         []Traffic
	Trees                                           []Tree
	VisibleTrees                                    int // only drawn trees may cause vehicle collisions
	Buildings                                       []Vec
	Poles                                           []Pole
	Rocks                                           []Rock
	Water                                           []WaterArea
	Ramps                                           []Ramp
	Elapsed                                         float64
	Map                                             int
	Distance                                        float64
	NearMisses                                      int
	Challenge                                       int
	ChallengeTime                                   float64
	Checkpoint                                      int
	Checkpoints                                     []Vec
	Finished                                        bool
	Credits                                         int
	CollisionCells                                  map[gridKey]*obstacleCell
	gridTreeCount, gridRockCount, gridBuildingCount int
}

func rng(seed *uint32) float64 {
	*seed = *seed*1664525 + 1013904223
	return float64((*seed>>8)&0xffffff) / float64(0xffffff)
}
func NewWorld(mapId int) *World {
	mapId = ClampInt(mapId, 0, 2)
	w := &World{Routes: BuildRoutesForMap(mapId), Map: mapId, DriveMode: 1}
	w.Reset()
	return w
}
func (w *World) Reset() {
	if w.Remaster != nil {
		w.ResetRemastered()
		return
	}
	w.Elapsed = 0
	w.Distance = 0
	w.NearMisses = 0
	w.ChallengeTime = 0
	w.Finished = false
	w.Checkpoint = 0
	w.Credits = 0
	w.Car = Car{Pos: Vec{0, -218}, Heading: -0.25, Nitro: 65, DriftCombo: 1}
	if w.Map != 0 {
		start, tangent := w.Routes[0].AtDistance(15)
		w.Car.Pos = start
		w.Car.Heading = math.Atan2(tangent.X, tangent.Z)
	}
	w.Pickups = nil
	w.Traffic = nil
	w.Trees = nil
	w.Buildings = nil
	w.Poles = nil
	w.Rocks = nil
	w.Water = WaterForMap(w.Map)
	w.Ramps = BuildRamps(w.Map)
	w.buildHills()
	for k, r := range w.Routes {
		for i := 10; i < len(r.Points)-8; i += 18 {
			p, dir := r.At(float64(i))
			n := Vec{-dir.Z, dir.X}
			w.Pickups = append(w.Pickups, Pickup{Pos: p.Add(n.Mul(float64((i%3)-1) * 2.5))})
		}
		// Stunt connectors are reserved for the player; civilians stay on the ring.
		if k != 0 {
			continue
		}
		for j := 0; j < 21; j++ {
			variant := (j + k*3) % 7
			maxCruise := 7.5 + float64((j*3+k)%5)*.7 // 27-37 km/h urban civilian traffic
			if variant%4 == 3 {
				maxCruise = 5.5 + float64(j%3)*.55
			} // slow truck
			distance := (float64(j) + .35*float64(k)) / 21 * r.Len
			initialPos, initialDir := r.AtDistance(distance)
			lane := 3.1 // right-hand driving lane relative to heading
			direction := 1.0
			if j%2 == 0 {
				direction = -1
			}
			if direction < 0 {
				initialDir = initialDir.Mul(-1)
			}
			w.Traffic = append(w.Traffic, Traffic{Path: k, T: distance, Speed: maxCruise * .78, CruiseSpeed: maxCruise, Lane: lane, Direction: direction, Variant: variant, Pos: initialPos.Add(Vec{-initialDir.Z, initialDir.X}.Mul(lane)), Heading: math.Atan2(initialDir.X, initialDir.Z)})
		}
	}
	seed := uint32(799 + w.Map*133)
	for i := 0; i < 1200 && len(w.Trees) < 360; i++ {
		x := rng(&seed)*760 - 370
		z := rng(&seed)*720 - 330
		p := Vec{x, z}
		if DistRoad(p, w.Routes) > 13 && !w.IsWater(p, 5) && !w.RampClearance(p, 10) {
			w.Trees = append(w.Trees, Tree{Pos: p, Size: 0.75 + rng(&seed)*1.8, Type: int(rng(&seed) * 3)})
		}
	}
	// Roadside rocks are visually rendered and collidable using the same placements.
	for j := 0; j < 42; j++ {
		x := -435. + float64((j*151+43)%870)
		z := -418. + float64((j*173+55)%850)
		pt := Vec{X: x, Z: z}
		if DistRoad(pt, w.Routes) < 22 || pt.Dist(w.Car.Pos) < 20 || w.IsWater(pt, 5) || w.RampClearance(pt, 10) {
			continue
		}
		w.Rocks = append(w.Rocks, Rock{Pos: pt, Radius: 2.5 + float64(j%5)*.7, Height: 2.2 + float64(j%4)*.7})
	}
	w.VisibleTrees = len(w.Trees)
	w.placeBuildings()
	for k, r := range w.Routes {
		for i := 4; i < len(r.Points)-4; i += 12 {
			p, dir := r.At(float64(i))
			side := Vec{-dir.Z, dir.X}
			base := p.Add(side.Mul(11.2 + float64(k%2)))
			if DistRoad(base, w.Routes) > 10 && !w.IsWater(base, 2) && !w.RampClearance(base, 3) {
				w.Poles = append(w.Poles, Pole{Pos: base, Heading: math.Atan2(side.Z, -side.X)})
			}
		}
	}
	w.RebuildCollisionGrid()
	w.InitializeChassis()
}
func (w *World) StartChallenge(mode int) {
	w.Challenge = mode
	w.ChallengeTime = 0
	w.Finished = false
	w.Checkpoint = 0
	w.Checkpoints = nil
	if mode == 1 {
		r := w.Routes[0]
		for i := 12; i < len(r.Points)-2; i += 19 {
			w.Checkpoints = append(w.Checkpoints, r.Points[i])
		}
	}
}

type HandlingTuning struct {
	SteerLock, YawResponse, LateralGrip, EngineAcceleration, SpeedCap float64
}

// HandlingProfile adapts the approachable street/grip balance from the two
// uploaded open-source vehicle studies to the existing native simcade model.
// This is a simplified tuning system, not Godot VehicleBody or tire Pacejka physics.
// MapRelative returns right/forward distances in metres for a heading-up GPS.
// Positive right goes screen-right; positive forward goes screen-up.
func MapRelative(from, to Vec, heading float64) (right, forward float64) {
	d := to.Sub(from)
	f := HeadingVec(heading)
	r := Vec{f.Z, -f.X}
	return d.X*r.X + d.Z*r.Z, d.X*f.X + d.Z*f.Z
}

// CarElevation aligns each model's tire contact point to asphalt and grass.
// Terrain is drawn at -0.065m, asphalt at +0.235m; Porsche wheels are
// centred 0.09m above the 296 wheels, so its root requires less lift.
func CarElevation(distance float64, modelID int) float64 {
	nearRoad, offRoad := .25, -.04
	if modelID == 1 {
		nearRoad, offRoad = .15, -.14
	}
	transition := Clamp((distance-7.2)/3.5, 0, 1)
	transition = transition * transition * (3 - 2*transition)
	return Lerp(nearRoad, offRoad, transition)
}

func HandlingProfile(carClass, mode int) HandlingTuning {
	t := HandlingTuning{SteerLock: .295, YawResponse: 1.0, LateralGrip: 1.0, EngineAcceleration: 10.1, SpeedCap: 45.5}
	if carClass == 1 {
		t.SteerLock = .265
		t.YawResponse = .87
		t.LateralGrip = .92
		t.EngineAcceleration = 8.7
		t.SpeedCap = 42
	}
	switch mode {
	case 0: // TOUR: stable keyboard/free-drive handling
		t.SteerLock *= .78
		t.YawResponse *= .85
		t.LateralGrip *= 1.27
		t.EngineAcceleration *= .85
	case 2: // DRIFT: less rear grip, quicker yaw, still recovers with neutral input
		t.SteerLock *= 1.16
		t.YawResponse *= 1.13
		t.LateralGrip *= .82
		t.EngineAcceleration *= 1.04
	}
	return t
}

// Tire force approximation built independently from first principles:
// lateral cornering stiffness with a saturating tire-grip limit. Positive
// slip angle produces an opposing tire force. No VDrift code is copied.
func TireLateralForce(slipAngle, stiffness, maxForce float64) float64 {
	return Clamp(-stiffness*math.Atan(slipAngle), -maxForce, maxForce)
}

// Traffic recovery behaves like a damped lateral spring after a collision.
func RecoverTraffic(t *Traffic, dt float64) {
	if t.ImpactCooldown > 0 {
		t.ImpactCooldown = math.Max(0, t.ImpactCooldown-dt)
	}
	// Relax towards the driving lane while retaining inertia for a brief push.
	accel := t.Offset.Mul(-1.9).Sub(t.Impulse.Mul(3.1))
	t.Impulse = t.Impulse.Add(accel.Mul(dt))
	t.Offset = t.Offset.Add(t.Impulse.Mul(dt))
	if t.Offset.Length() > 7.0 {
		t.Offset = t.Offset.Unit().Mul(7.0)
	}
	if t.Offset.Length() < .015 && t.Impulse.Length() < .04 {
		t.Offset, t.Impulse = Vec{}, Vec{}
	}
	t.YawKick *= math.Exp(-dt * 2.6)
}

func (w *World) Step(in Input, dt float64) {
	if w.Remaster != nil {
		w.StepRemastered(in, dt)
		return
	}
	dt = Clamp(dt, 0, .025)
	if w.Finished || dt <= 0 {
		return
	}
	c := &w.Car
	w.Elapsed += dt
	w.ChallengeTime += dt
	steer := Clamp(in.Steer, -1, 1)
	tuning := HandlingProfile(w.CarClass, w.DriveMode)
	// A deliberate "grip first" simcade model. The old integrator preserved
	// large lateral velocity indefinitely, making the car drift without input.
	// At neutral input the tire forces now settle lateral speed and yaw quickly.
	c.OnRoad = DistRoad(c.Pos, w.Routes) < 8.6
	grip := 1.0
	if !c.OnRoad {
		grip = .62
	}
	speed := c.Speed
	absSpeed := math.Abs(speed)
	wasAirborne := c.Body.Initialized && !c.Body.Grounded
	if !c.Body.Initialized {
		w.InitializeChassis()
	}
	// Physically bounded road-wheel angle, scaled down at high speeds.
	steerAngle := steer * (tuning.SteerLock / (1 + absSpeed*.06))
	c.VisualSteer = Lerp(c.VisualSteer, steerAngle, 1-math.Exp(-dt*12))
	// Bicycle-model target angular velocity with chassis yaw inertia.
	wheelbase := VehicleSpec(w.CarClass).Wheelbase
	targetYaw := speed * math.Tan(steerAngle) / wheelbase
	// Drifting is user initiated: handbrake, throttle-induced power over,
	// or sharp steering with sufficient speed. Driving straight never drifts.
	// Drift only below 100 km/h, where keyboard countersteering is manageable.
	// At high speed the handbrake slows the car first instead of snapping sideways.
	powerOver := in.Throttle > .92 && absSpeed > 12 && absSpeed < 25 && math.Abs(steer) > .82
	drifting := (in.Handbrake || powerOver) && absSpeed > 7 && absSpeed < 27.5 && math.Abs(steer) > .12
	response := 3.4 * grip * tuning.YawResponse
	if drifting {
		response = 5.5 * grip * tuning.YawResponse
	}
	c.YawRate += (targetYaw - c.YawRate) * (1 - math.Exp(-dt*response))
	if math.Abs(steer) < .025 {
		c.YawRate *= math.Exp(-dt * 2.8)
	}
	c.YawRate = Clamp(c.YawRate, -2.1, 2.1)
	if wasAirborne {
		// Tires cannot create steering forces without ground contact.
		c.YawRate = c.Body.AirYawRate * math.Exp(-dt*.12)
		c.Body.AirYawRate = c.YawRate
	}
	// Two-axle tire response based on front/rear slip angles, with a much
	// smaller rear grip budget while the handbrake is engaged. This improves
	// countersteer and prevents permanent sideways momentum on corner exit.
	lf, lr := 1.18, 1.54
	v := math.Max(4.0, absSpeed)
	frontSlip := math.Atan2(c.Lateral+c.YawRate*lf, v) - steerAngle
	rearSlip := math.Atan2(c.Lateral-c.YawRate*lr, v)
	frontGrip := 6.1 * grip * tuning.LateralGrip
	rearGrip := 6.4 * grip * tuning.LateralGrip
	if drifting {
		rearGrip *= .36
	} else if in.Handbrake {
		rearGrip *= .74
	}
	frontForce := TireLateralForce(frontSlip, 28, frontGrip)
	rearForce := TireLateralForce(rearSlip, 28, rearGrip)
	// Relaxation is deliberately slower than the full rigid-body tire model
	// to remain controllable on digital keyboard steering.
	physicalLat := frontForce + rearForce - speed*c.YawRate
	if wasAirborne {
		c.Lateral *= math.Exp(-dt * .035)
	} else if drifting {
		c.Lateral += (physicalLat*.18 + speed*c.YawRate*.88 - c.Lateral*1.45) * dt
	} else {
		c.Lateral += (physicalLat*.24 - c.Lateral*(4.8*grip*tuning.LateralGrip)) * dt
	}
	if absSpeed < 3 {
		c.YawRate *= math.Exp(-dt * 4)
		c.Lateral *= math.Exp(-dt * 9)
	}
	c.Lateral = Clamp(c.Lateral, -24, 24)
	c.Heading += c.YawRate * dt
	// Stationary burnout: rear brakes hold the axle while throttle stays live.
	// It generates rear-tire smoke without killing engine power or moving the car.
	c.Burnout = !wasAirborne && in.Handbrake && in.Throttle > .35 && math.Abs(c.Speed) < 2.5 && in.Brake == 0
	acceleration := DriveAcceleration(c, in, dt, w.CarClass)
	if wasAirborne {
		acceleration = 0
	}
	preBrakeSpeed := c.Speed
	c.Boosting = in.Boost && in.Brake == 0 && c.Nitro > 0 && c.Speed > 2
	if c.Boosting {
		if wasAirborne {
			// Thrust follows the car's nose in flight, including vertical lift.
			acceleration += 22 * math.Cos(c.Rig.Pitch)
			c.Body.VerticalSpeed += -22 * math.Sin(c.Rig.Pitch) * dt
		} else {
			acceleration += 38
		}
		c.Nitro = math.Max(0, c.Nitro-29*dt)
	}
	acceleration -= c.Speed*math.Abs(c.Speed)*.0024 + c.Speed*.047
	if !c.OnRoad && !wasAirborne {
		acceleration -= c.Speed * .19
	}
	// Keep speed continuous when nitro ends. Old code instantly clamped the
	// vehicle from boost speed to the normal cap within a single frame.
	if !c.Boosting && c.Speed > tuning.SpeedCap {
		acceleration -= 8 + (c.Speed-tuning.SpeedCap)*.30
	}
	if c.Burnout {
		acceleration = 0
		c.Speed = 0
		c.Lateral *= math.Exp(-dt * 11)
		c.YawRate *= math.Exp(-dt * 8)
	}
	c.Speed = Clamp(c.Speed+acceleration*dt, -13.0, 66.)
	if c.Boosting && c.Speed > 66 {
		c.Speed = 66
	}
	if !c.Boosting && c.Speed <= tuning.SpeedCap {
		c.Speed = math.Min(c.Speed, tuning.SpeedCap)
	}
	if !wasAirborne && in.Brake > 0 && c.ReverseHold <= .25 && preBrakeSpeed >= 0 && c.Speed < 0 {
		c.Speed = 0 // do not instantly reverse while braking from forward motion
	} else if !wasAirborne && in.Brake > 0 && preBrakeSpeed < 0 && c.Speed > 0 {
		c.Speed = 0
	}
	if !wasAirborne && in.Handbrake && preBrakeSpeed > 0 && c.Speed < 0 {
		c.Speed = 0
	}
	if wasAirborne {
		// Preserve world-space momentum while the chassis rotates in flight.
		f := HeadingVec(c.Heading)
		r := Vec{f.Z, -f.X}
		velocity := c.Body.AirVelocity
		if velocity.Length() < .001 {
			velocity = f.Mul(preBrakeSpeed).Add(r.Mul(c.Lateral))
		}
		velocity = velocity.Mul(math.Exp(-dt * .028))
		if c.Boosting {
			velocity = velocity.Add(f.Mul(22 * math.Cos(c.Rig.Pitch) * dt))
		}
		if velocity.Length() > 90 {
			velocity = velocity.Unit().Mul(90)
		}
		c.Body.AirVelocity = velocity
		c.Speed = velocity.X*f.X + velocity.Z*f.Z
		c.Lateral = velocity.X*r.X + velocity.Z*r.Z
	}
	// Wheel roll is derived from travel distance, independent of frame rate.
	c.WheelSpin += c.Speed * dt / .34
	if c.WheelSpin > Tau || c.WheelSpin < -Tau {
		c.WheelSpin = math.Mod(c.WheelSpin, Tau)
	}
	c.DriftAngle = math.Atan2(c.Lateral, math.Max(4, math.Abs(c.Speed))) * 180 / math.Pi
	if math.Abs(c.DriftAngle) > 10 && math.Abs(c.Speed) > 10 && math.Abs(c.YawRate) > .08 {
		c.DriftCombo = Clamp(c.DriftCombo+dt*.34, 1, 5)
		c.DriftScore += math.Abs(c.DriftAngle) * math.Abs(c.Speed) * dt * .7 * c.DriftCombo
	} else {
		c.DriftCombo = Lerp(c.DriftCombo, 1, Clamp(dt*2, 0, 1))
	}
	dir := Vec{math.Sin(c.Heading), math.Cos(c.Heading)}
	right := Vec{dir.Z, -dir.X}
	before := c.Pos
	c.Pos = c.Pos.Add(dir.Mul(c.Speed * dt)).Add(right.Mul(c.Lateral * dt))
	w.StepChassis(in, acceleration, dt)
	// No hard world-bound clamps: these previously felt like invisible walls.
	// Terrain outside the detailed centre remains drivable; road cues encourage returning.
	w.Distance += c.Pos.Dist(before)
	for i := range w.Pickups {
		p := &w.Pickups[i]
		if p.Cooldown > 0 {
			p.Cooldown -= dt
		}
		if p.Cooldown <= 0 && c.Pos.Dist(p.Pos) < 4 && c.Body.Height-w.GroundHeight(p.Pos) < 4 {
			c.Nitro = Clamp(c.Nitro+35, 0, 100)
			p.Cooldown = 30
			w.Credits += 20
		}
	}
	// Traffic follows measured arc length, approaches its cruise speed gently, and
	// slows behind vehicles on the same route/lane instead of racing the player.
	for i := range w.Traffic {
		t := &w.Traffic[i]
		if t.NearCooldown > 0 {
			t.NearCooldown -= dt
		}
		route := w.Routes[t.Path]
		targetSpeed := t.CruiseSpeed
		// Give way when the player blocks the lane ahead rather than driving
		// blindly until the collision handler fires.
		forwardNPC := HeadingVec(t.Heading)
		toPlayer := c.Pos.Sub(t.Pos)
		ahead := toPlayer.X*forwardNPC.X + toPlayer.Z*forwardNPC.Z
		side := math.Abs(toPlayer.X*forwardNPC.Z - toPlayer.Z*forwardNPC.X)
		if ahead > 0 && ahead < 23 && side < 3.1 && c.Body.Height-w.GroundHeight(c.Pos) < 3 {
			playerVelocity := HeadingVec(c.Heading).Mul(c.Speed)
			leaderSpeed := math.Max(0, playerVelocity.X*forwardNPC.X+playerVelocity.Z*forwardNPC.Z)
			targetSpeed = math.Min(targetSpeed, PredictiveFollowingSpeed(t.Speed, leaderSpeed, ahead-4.5))
		}
		for j := range w.Traffic {
			if i == j {
				continue
			}
			lead := w.Traffic[j]
			if lead.Path != t.Path || math.Abs(lead.Lane-t.Lane) > 1.5 || lead.Direction != t.Direction {
				continue
			}
			gap := (lead.T - t.T) * t.Direction
			if route.Closed {
				gap = math.Mod(gap+route.Len, route.Len)
			}
			if gap >= 0 && gap < 24 {
				targetSpeed = math.Min(targetSpeed, PredictiveFollowingSpeed(t.Speed, lead.Speed, gap-5))
			}
		}
		// Give NPCs a narrow lateral escape path when the player blocks the lane.
		avoid := 0.0
		if ahead > 7 && ahead < 19 && side < 2.9 {
			avoid = -.65
		}
		t.AvoidOffset = Lerp(t.AvoidOffset, avoid, Clamp(dt*1.4, 0, 1))
		t.Speed += (targetSpeed - t.Speed) * Clamp(dt*1.45, 0, 1)
		t.Speed = Clamp(t.Speed, 0, t.CruiseSpeed)
		t.T += t.Speed * dt * t.Direction
		if route.Closed {
			t.T = math.Mod(t.T+route.Len, route.Len)
		} else if t.T > route.Len || t.T < 0 {
			// Open connectors have endpoints; turn around instead of teleporting.
			if t.T > route.Len {
				t.T = 2*route.Len - t.T
			} else {
				t.T = -t.T
			}
			t.T = Clamp(t.T, 0, route.Len)
			t.Direction = -t.Direction
		}
		point, dir := route.AtDistance(t.T)
		if t.Direction < 0 {
			dir = dir.Mul(-1)
		}
		RecoverTraffic(t, dt)
		base := point.Add(Vec{-dir.Z, dir.X}.Mul(t.Lane + t.AvoidOffset))
		t.Pos = base.Add(t.Offset)
		t.Heading = math.Atan2(dir.X, dir.Z) + t.YawKick
		deltaTraffic := c.Pos.Sub(t.Pos)
		dist := deltaTraffic.Length()
		// Traffic contact uses an oriented footprint instead of an oversized
		// circular blocker, so close overtakes no longer feel like invisible walls.
		forward := HeadingVec(t.Heading)
		rightTraffic := Vec{forward.Z, -forward.X}
		long := deltaTraffic.X*forward.X + deltaTraffic.Z*forward.Z
		lat := deltaTraffic.X*rightTraffic.X + deltaTraffic.Z*rightTraffic.Z
		halfLen, halfWid := 2.35, 1.08
		if t.Variant%4 == 3 {
			halfLen, halfWid = 3.05, 1.22
		}
		bufferLong, bufferWid := 1.05, .88
		trafficHeight := 1.65
		if t.Variant%4 == 3 {
			trafficHeight = 3.2
		}
		if math.Abs(long) < halfLen+bufferLong && math.Abs(lat) < halfWid+bufferWid && c.Body.Height-.5 < Terrain(t.Pos.X, t.Pos.Z)+trafficHeight {
			// Separate via least penetration, but only impulse once per contact.
			overLong := halfLen + bufferLong - math.Abs(long)
			overWid := halfWid + bufferWid - math.Abs(lat)
			n := rightTraffic
			if overLong < overWid {
				n = forward
				if long < 0 {
					n = n.Mul(-1)
				}
				c.Pos = c.Pos.Add(n.Mul(overLong * .88))
			} else {
				if lat < 0 {
					n = n.Mul(-1)
				}
				c.Pos = c.Pos.Add(n.Mul(overWid * .88))
			}
			if t.ImpactCooldown <= 0 {
				carForward := HeadingVec(c.Heading)
				carRight := Vec{carForward.Z, -carForward.X}
				playerVelocity := carForward.Mul(c.Speed).Add(carRight.Mul(c.Lateral))
				trafficVelocity := forward.Mul(t.Speed).Add(t.Impulse)
				relative := playerVelocity.Sub(trafficVelocity)
				closing := -relative.X*n.X - relative.Z*n.Z
				if closing > .3 {
					playerMass, otherMass := 1550.0, 1480.0
					if w.CarClass == 1 {
						playerMass = 1370
					}
					if t.Variant%4 == 3 {
						otherMass = 7400
					} // truck reacts much less
					impulse := (1.08 * closing) / (1/playerMass + 1/otherMass)
					playerVelocity = playerVelocity.Add(n.Mul(impulse / playerMass))
					t.Impulse = t.Impulse.Sub(n.Mul(impulse / otherMass))
					c.Speed = Clamp(playerVelocity.X*carForward.X+playerVelocity.Z*carForward.Z, -13, 66)
					c.Lateral = Clamp(playerVelocity.X*carRight.X+playerVelocity.Z*carRight.Z, -15, 15)
					t.YawKick = Clamp(t.YawKick+(lat/math.Max(halfWid, .5))*closing*.005*(1480/otherMass), -.25, .25)
					c.Damage = Clamp(c.Damage+math.Min(22, closing*.4), 0, 100)
					t.ImpactCooldown = .20
				}
			}
		} else if dist < 5.2 && math.Abs(c.Speed) > 13 && t.NearCooldown <= 0 {
			w.NearMisses++
			w.Credits += 15
			t.NearCooldown = 8
		}
	}

	// The shoreline is a polygon owned by the same map data as its water mesh.
	w.RejectWater()
	// Solid boulders and ridges match the exact visual placement/size.
	for _, rockIndex := range w.NearbyRockIndices(c.Pos, 9) {
		rock := w.Rocks[rockIndex]
		minDist := rock.Radius + 1.05
		if d := c.Pos.Dist(rock.Pos); d < minDist && c.Body.Height-.5 < Terrain(rock.Pos.X, rock.Pos.Z)+rock.Height {
			n := c.Pos.Sub(rock.Pos).Unit()
			c.Pos = rock.Pos.Add(n.Mul(minDist))
			c.Speed *= .21
			c.Lateral *= .35
			c.Damage = Clamp(c.Damage+math.Min(12, math.Abs(c.Speed)*.3+3), 0, 100)
		}
	}
	// The large rendered mountain belts are impassable obstacles; avoid
	// a mismatch where the model looks solid but the car can pass through.
	for _, hill := range w.Hills {
		if c.Body.Height-.5 > hill.HeightAt(c.Pos)+1 {
			continue
		}
		if d := c.Pos.Dist(hill.Pos); d < hill.Radius*.68+1.1 {
			normal := c.Pos.Sub(hill.Pos).Unit()
			c.Pos = hill.Pos.Add(normal.Mul(hill.Radius*.68 + 1.1))
			c.Speed *= .18
			c.Lateral *= .4
			c.Damage = Clamp(c.Damage+7, 0, 100)
		}
	}
	visible := ClampInt(w.VisibleTrees, 0, len(w.Trees))
	for _, treeIndex := range w.NearbyTreeIndices(c.Pos, 6) {
		if treeIndex >= visible {
			continue
		}
		tree := w.Trees[treeIndex]
		if c.Pos.Dist(tree.Pos) < tree.Size*.42+1.18 && c.Body.Height-.5 < Terrain(tree.Pos.X, tree.Pos.Z)+tree.Size*4.5 {
			n := c.Pos.Sub(tree.Pos).Unit()
			c.Pos = tree.Pos.Add(n.Mul(tree.Size*.42 + 1.18))
			c.Speed *= .16
			c.Lateral *= .3
			c.Damage = Clamp(c.Damage+7, 0, 100)
		}
	}
	{
		for _, index := range w.NearbyBuildingIndices(c.Pos, 14) {
			b := w.Buildings[index]
			// Building visuals are axis-aligned 13m x 15m blocks. A former 7.5m
			// circular blocker extended well beyond some visible building faces.
			dx, dz := c.Pos.X-b.X, c.Pos.Z-b.Z
			halfX, halfZ := 6.5+1.0, 7.5+1.0
			if math.Abs(dx) < halfX && math.Abs(dz) < halfZ {
				overlapX, overlapZ := halfX-math.Abs(dx), halfZ-math.Abs(dz)
				if overlapX < overlapZ {
					sign := 1.0
					if dx < 0 {
						sign = -1
					}
					c.Pos.X = b.X + sign*halfX
				} else {
					sign := 1.0
					if dz < 0 {
						sign = -1
					}
					c.Pos.Z = b.Z + sign*halfZ
				}
				c.Speed *= .42
				c.Lateral *= .4
			}
		}
	}

	for i := range w.Poles {
		p := &w.Poles[i]
		if !p.Falling && c.Pos.Dist(p.Pos) < 1.75 && math.Abs(c.Speed) > 3 && c.Body.Height-.5 < Terrain(p.Pos.X, p.Pos.Z)+10 {
			p.Hit(HeadingVec(c.Heading).Mul(c.Speed).Add(Vec{math.Cos(c.Heading), -math.Sin(c.Heading)}.Mul(c.Lateral)))
			// Lightweight breakaway lamp post: smaller impulse than a tree.
			c.Speed *= .89
			c.Lateral *= .91
		}
		p.Step(dt)
	}
	if w.Challenge == 1 && w.Checkpoint < len(w.Checkpoints) && c.Pos.Dist(w.Checkpoints[w.Checkpoint]) < 13 {
		w.Checkpoint++
		w.Credits += 45
		if w.Checkpoint >= len(w.Checkpoints) {
			w.Finished = true
			w.Credits += int(math.Max(100, 850-w.ChallengeTime*3))
		}
	}
	if w.Challenge == 2 && w.ChallengeTime >= 75 {
		w.Finished = true
		w.Credits += int(c.DriftScore * .07)
	}
	if w.Challenge == 3 && w.ChallengeTime >= 75 {
		w.Finished = true
		w.Credits += w.NearMisses * 80
	}
}
func HeadingVec(h float64) Vec { return Vec{math.Sin(h), math.Cos(h)} }

func ClampInt(x, a, b int) int {
	if x < a {
		return a
	}
	if x > b {
		return b
	}
	return x
}

func BuildingPosition(i int, routes []Route) Vec {
	r := routes[i%len(routes)]
	t := float64((i*17 + 11) % (len(r.Points) - 1))
	p, d := r.At(t)
	side := Vec{-d.Z, d.X}
	sign := 1.
	if i%2 == 0 {
		sign = -1
	}
	return p.Add(side.Mul(sign * (24 + float64(i%3)*8)))
}

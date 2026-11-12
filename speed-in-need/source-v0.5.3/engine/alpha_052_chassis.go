package engine

import (
	"math"
	"sync"
)

// ChassisState owns vertical momentum and angular inertia. Horizontal tire
// handling remains the existing tuned simcade model; this is not a full
// six-degree-of-freedom tire/rigid-body replacement.
type ChassisState struct {
	Height, VerticalSpeed, PitchRate, RollRate       float64
	AirYawRate, AirTime, JumpDistance, LandingImpact float64
	Grounded, Initialized                            bool
	LaunchPosition                                   Vec
	AirVelocity                                      Vec
}

type Ramp struct {
	Min, Max                       Vec
	Name                           string
	Start                          Vec
	Heading, Length, Width, Height float64
	RouteID                        int
	Path                           Route
}

var rampCache [3]struct {
	once  sync.Once
	ramps []Ramp
}

func BuildRamps(mapID int) []Ramp {
	id := ClampInt(mapID, 0, 2)
	entry := &rampCache[id]
	entry.once.Do(func() { entry.ramps = buildRoadRamps(id) })
	return append([]Ramp(nil), entry.ramps...)
}
func buildRoadRamps(mapID int) []Ramp {
	routes := BuildRoutesForMap(mapID)
	out := make([]Ramp, 0, 2)
	for i := 1; i <= 2; i++ {
		route := routes[i]
		length := 60.
		height := 15.
		if i == 2 {
			length = 52
			height = 12
		}
		start, bestScore := 35., math.Inf(1)
		// Pick an approach clear of crossings and with gentle curvature. A
		// ramp must not rise across another avenue at an intersection.
		for candidate := 35.; candidate < route.Len-length-30; candidate += 8 {
			score := math.Abs(candidate-route.Len*.3) * .002
			_, first := route.AtDistance(candidate)
			for d := 0.; d <= length; d += 5 {
				p, tangent := route.AtDistance(candidate + d)
				score += (1 - first.X*tangent.X - first.Z*tangent.Z) * 8
				for k, other := range routes {
					if k == i {
						continue
					}
					gap := DistRoad(p, []Route{other})
					score += math.Max(0, 32-gap) * 10
				}
			}
			if score < bestScore {
				start = candidate
				bestScore = score
			}
		}
		points := make([]Vec, 61)
		for j := range points {
			points[j], _ = route.AtDistance(start + length*float64(j)/60)
		}
		path := newRoute("stunt deck", points)
		low, high := points[0], points[0]
		for _, p := range points {
			low.X = math.Min(low.X, p.X)
			low.Z = math.Min(low.Z, p.Z)
			high.X = math.Max(high.X, p.X)
			high.Z = math.Max(high.Z, p.Z)
		}
		p, d := path.AtDistance(0)
		out = append(out, Ramp{Min: low, Max: high, Name: []string{"SKYLINE LAUNCH", "RETURN FLIGHT"}[i-1], Start: p, Heading: math.Atan2(d.X, d.Z), Length: path.Len, Width: 12, Height: height, RouteID: i, Path: path})
	}
	return out
}
func (r Ramp) Sample(along float64) (Vec, Vec) {
	if len(r.Path.Points) > 1 {
		p, d := r.Path.AtDistance(Clamp(along, 0, r.Length))
		if along < 0 {
			p = p.Add(d.Mul(along))
		}
		if along > r.Length {
			p = p.Add(d.Mul(along - r.Length))
		}
		return p, d
	}
	f := HeadingVec(r.Heading)
	return r.Start.Add(f.Mul(along)), f
}
func (r Ramp) Coordinates(p Vec) (side, along float64) {
	best := math.Inf(1)
	for i := 1; i < len(r.Path.Points); i++ {
		a, b := r.Path.Points[i-1], r.Path.Points[i]
		v := b.Sub(a)
		l := r.Path.Cumulative[i] - r.Path.Cumulative[i-1]
		if l < 1e-6 {
			continue
		}
		t := Clamp((p.Sub(a).X*v.X+p.Sub(a).Z*v.Z)/(l*l), 0, 1)
		q := a.Add(v.Mul(t))
		delta := q.Sub(p)
		dist := delta.X*delta.X + delta.Z*delta.Z
		if dist < best {
			best = dist
			f := v.Mul(1 / l)
			d := p.Sub(q)
			side = d.X*f.Z - d.Z*f.X
			along = r.Path.Cumulative[i-1] + t*l
		}
	}
	for _, end := range []float64{0, r.Length} {
		q, f := r.Sample(end)
		d := p.Sub(q)
		t := d.X*f.X + d.Z*f.Z
		if (end == 0 && t < 0) || (end == r.Length && t > 0) {
			s := d.X*f.Z - d.Z*f.X
			if s*s < best {
				side = s
				along = end + t
				best = s * s
			}
		}
	}
	return
}
func (r Ramp) DeckHeight(p Vec, along float64) float64 {
	t := Clamp(along/r.Length, 0, 1)
	return Terrain(p.X, p.Z) + .24 + r.Height*t*t
}
func (r Ramp) Surface(p Vec) (float64, bool) {
	if p.X < r.Min.X-r.Width*.5-.01 || p.X > r.Max.X+r.Width*.5+.01 || p.Z < r.Min.Z-r.Width*.5-.01 || p.Z > r.Max.Z+r.Width*.5+.01 {
		return 0, false
	}
	side, along := r.Coordinates(p)
	if math.Abs(side) > r.Width*.5+.001 || along < -.001 || along > r.Length+.001 {
		return 0, false
	}
	return r.DeckHeight(p, along), true
}
func (w *World) RampClearance(p Vec, margin float64) bool {
	for _, r := range w.Ramps {
		s, a := r.Coordinates(p)
		if math.Abs(s) < r.Width*.5+margin && a > -margin && a < r.Length+145+margin {
			return true
		}
	}
	return false
}

func (w *World) IsWater(p Vec, margin float64) bool {
	if w.Remaster != nil && w.Remaster.Def.WaterGrid.W > 1 {
		g := w.Remaster.Def.WaterGrid
		if g.Height(p) > .45 {
			return true
		}
		for _, v := range []Vec{{margin, 0}, {-margin, 0}, {0, margin}, {0, -margin}} {
			if margin > 0 && g.Height(p.Add(v)) > .45 {
				return true
			}
		}
		return false
	}
	for _, water := range w.Water {
		if water.Contains(p) {
			return true
		}
		if margin > 0 {
			for _, offset := range []Vec{{margin, 0}, {-margin, 0}, {0, margin}, {0, -margin}} {
				if water.Contains(p.Add(offset)) {
					return true
				}
			}
		}
	}
	return false
}

func (w *World) GroundHeight(p Vec) float64 {
	if w.Remaster != nil {
		return w.RemasteredGround(p)
	}
	if w.IsWater(p, 0) {
		return CoastWaterLevel
	}
	// Smooth road shoulders match the rendered road/grass transition.
	h := Terrain(p.X, p.Z) + CarElevation(DistRoad(p, w.Routes), 0) - .015
	for _, hill := range w.Hills {
		if p.Dist(hill.Pos) < hill.Radius*1.15 {
			h = math.Max(h, hill.HeightAt(p))
		}
	}
	for _, r := range w.Ramps {
		if rh, ok := r.Surface(p); ok && rh > h {
			h = rh
		}
	}
	return h
}

func (w *World) TerrainHeight(p Vec) float64 {
	if w.Remaster != nil {
		return w.Remaster.Def.Grid.Height(p)
	}
	if w.IsWater(p, 0) {
		return math.Min(Terrain(p.X, p.Z), CoastWaterLevel-3)
	}
	return Terrain(p.X, p.Z)
}

func (w *World) InitializeChassis() {
	c := &w.Car
	c.Body = ChassisState{Height: w.GroundHeight(c.Pos) + .74, Grounded: true, Initialized: true}
	params := VehicleSpec(w.CarClass)
	if w.Remaster != nil {
		params = VehicleVariantSpec050(w.CarClass, c.Variant050).P
	}
	staticCompression := params.Mass * 9.81 / (4 * params.SpringRate)
	f := HeadingVec(c.Heading)
	r := Vec{f.Z, -f.X}
	for i := range c.Rig.Wheels {
		x, z := -params.Track*.5, -params.Wheelbase*.5
		if i%2 == 1 {
			x = -x
		}
		if i < 2 {
			z = -z
		}
		compression := staticCompression
		load := params.Mass * 9.81 / 4
		if w.Remaster != nil {
			spec := VehicleVariantSpec050(w.CarClass, c.Variant050)
			fraction := 1 - spec.FrontWeight
			z = -params.Wheelbase * spec.FrontWeight
			if i < 2 {
				z = params.Wheelbase * (1 - spec.FrontWeight)
				fraction = spec.FrontWeight
			}
			load = params.Mass * 9.81 * fraction / 2
			compression = load / params.SpringRate
		}
		c.Rig.Wheels[i].Compression = compression
		c.Rig.Wheels[i].GroundHeight = w.GroundHeight(c.Pos.Add(r.Mul(x)).Add(f.Mul(z)))
		c.Rig.Wheels[i].NormalLoad = load
		c.Rig.Wheels[i].Touching = true
	}
}

func (w *World) StepChassis(in Input, longitudinalAcceleration, dt float64) {
	if dt <= 0 {
		return
	}
	c := &w.Car
	if !c.Body.Initialized {
		w.InitializeChassis()
	}
	b := &c.Body
	p := VehicleSpec(w.CarClass)
	forward := HeadingVec(c.Heading)
	right := Vec{forward.Z, -forward.X}
	staticCompression := p.Mass * 9.81 / (4 * p.SpringRate)
	restLength := .74 - p.WheelRadius + staticCompression
	force, rollMoment, pitchMoment := 0., 0., 0.
	contacts := 0
	for i := 0; i < 4; i++ {
		x, z := -p.Track*.5, -p.Wheelbase*.5
		if i%2 == 1 {
			x = -x
		}
		if i < 2 {
			z = -z
		}
		position := c.Pos.Add(right.Mul(x)).Add(forward.Mul(z))
		ground := w.GroundHeight(position)
		wheel := &c.Rig.Wheels[i]
		mountY := b.Height + x*math.Sin(c.Rig.Roll) - z*math.Sin(c.Rig.Pitch)
		compression := Clamp(ground+p.WheelRadius+restLength-mountY, 0, .34)
		// Ground-relative damping accounts for the velocity up a ramp.
		groundSpeed := Clamp((ground-wheel.GroundHeight)/dt, -55, 55)
		mountSpeed := b.VerticalSpeed + x*b.RollRate - z*b.PitchRate
		load := 0.
		if compression > 0 {
			load = Clamp(p.SpringRate*compression-p.Damping*(mountSpeed-groundSpeed), 0, p.Mass*9.81*1.5)
			contacts++
		}
		wheel.Velocity = (compression - wheel.Compression) / dt
		wheel.Compression, wheel.NormalLoad, wheel.GroundHeight = compression, load, ground
		wheel.Touching = compression > 0
		wheel.SteerAngle = 0
		if i < 2 {
			wheel.SteerAngle = AckermannSteer(c.VisualSteer, p.Wheelbase, p.Track, math.Copysign(1, x))
		}
		force += load
		rollMoment += x * load
		pitchMoment -= z * load
		contactSpeed := c.Speed - c.YawRate*x
		angular := contactSpeed / p.WheelRadius
		if c.Burnout && i >= 2 {
			angular = 16 + 34*Clamp(c.Gearbox.RPM/p.RedlineRPM, 0, 1)
			wheel.SlipRatio = 2
		} else if !wheel.Touching {
			angular = wheel.AngularVelocity * math.Exp(-dt*.15)
			if i >= 2 && in.Throttle > 0 {
				angular = Lerp(angular, 90*in.Throttle, 1-math.Exp(-dt*2))
			}
			wheel.SlipRatio *= math.Exp(-dt * 3)
		} else if in.Handbrake && i >= 2 {
			angular *= math.Exp(-dt * 65)
			wheel.SlipRatio = -1
		} else {
			if i >= 2 && in.Throttle > .8 && math.Abs(c.Speed) < 12 {
				angular *= 1.04
			}
			wheel.SlipRatio = (angular*p.WheelRadius - contactSpeed) / math.Max(2, math.Abs(contactSpeed))
		}
		wheel.AngularVelocity = angular
		wheel.RollAngle += angular * dt
	}
	wasGrounded := b.Grounded
	b.Grounded = contacts > 0
	if b.Grounded {
		// Forces at four mounts transfer load. Anti-roll bars and angular
		// damping remove the former exaggerated, speed-driven seesaw pose.
		rollMoment += p.Mass*c.Speed*c.YawRate*.30 - c.Rig.Roll*15000 - b.RollRate*5200
		pitchMoment += -p.Mass*Clamp(longitudinalAcceleration, -20, 40)*.26 - b.PitchRate*6500
	} else {
		b.AirTime += dt
		b.JumpDistance = c.Pos.Dist(b.LaunchPosition)
		rollMoment = -b.RollRate * 70
		pitchMoment = -b.PitchRate * 110
	}
	if wasGrounded && !b.Grounded {
		b.AirYawRate = c.YawRate
		b.AirVelocity = forward.Mul(c.Speed).Add(right.Mul(c.Lateral))
		b.AirTime = 0
		b.LaunchPosition = c.Pos
	}
	if !wasGrounded && b.Grounded {
		b.LandingImpact = math.Max(0, -b.VerticalSpeed)
		c.Damage = Clamp(c.Damage+math.Max(0, b.LandingImpact-12)*.3, 0, 100)
	}
	rollInertia := p.Mass * (p.Track*p.Track + 1.1) / 12
	pitchInertia := p.Mass * (p.Wheelbase*p.Wheelbase + 1.1) / 12
	b.RollRate = Clamp(b.RollRate+rollMoment/rollInertia*dt, -1.8, 1.8)
	b.PitchRate = Clamp(b.PitchRate+pitchMoment/pitchInertia*dt, -2.3, 2.3)
	c.Rig.Roll = Clamp(c.Rig.Roll+b.RollRate*dt, -.42, .42)
	c.Rig.Pitch = Clamp(c.Rig.Pitch+b.PitchRate*dt, -1.25, 1.25)
	if math.Abs(c.Rig.Roll) >= .42 {
		b.RollRate *= .5
	}
	if math.Abs(c.Rig.Pitch) >= 1.25 {
		b.PitchRate *= .5
	}
	b.VerticalSpeed += Clamp(force/p.Mass-9.81, -9.81, 60) * dt
	b.Height += b.VerticalSpeed * dt
	// A bump stop bounds penetration after unusually hard landings.
	minimum := w.GroundHeight(c.Pos) + .46
	if b.Height < minimum {
		b.Height = minimum
		b.VerticalSpeed = math.Max(0, -b.VerticalSpeed*.08)
	}
	c.Rig.Heave = b.Height - w.GroundHeight(c.Pos) - .74
	c.Rig.HeaveSpeed = b.VerticalSpeed
}

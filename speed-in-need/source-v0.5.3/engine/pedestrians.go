package engine

import "math"

type PedVec struct{ X, Y, Z float64 }

func (a PedVec) Add(b PedVec) PedVec  { return PedVec{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
func (a PedVec) Sub(b PedVec) PedVec  { return PedVec{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func (a PedVec) Mul(s float64) PedVec { return PedVec{a.X * s, a.Y * s, a.Z * s} }
func (a PedVec) Length() float64      { return math.Sqrt(a.X*a.X + a.Y*a.Y + a.Z*a.Z) }
func (a PedVec) Dot(b PedVec) float64 { return a.X*b.X + a.Y*b.Y + a.Z*b.Z }
func (a PedVec) XZ() Vec              { return Vec{a.X, a.Z} }
func (a PedVec) Unit() PedVec {
	if l := a.Length(); l > 1e-8 {
		return a.Mul(1 / l)
	}
	return PedVec{0, 1, 0}
}

const PedJointCount = 15

type PedBone struct {
	A, B     int
	Radius   float64
	Material int
}

// Pelvis, chest, head; shoulders/elbows/hands; hips/knees/feet.
var PedBones = [...]PedBone{
	{0, 1, .17, 1}, {1, 2, .105, 0}, {1, 3, .105, 1}, {3, 4, .075, 1}, {4, 5, .058, 0},
	{1, 6, .105, 1}, {6, 7, .075, 1}, {7, 8, .058, 0}, {0, 9, .105, 2}, {9, 10, .10, 2},
	{10, 11, .075, 2}, {0, 12, .105, 2}, {12, 13, .10, 2}, {13, 14, .075, 2},
}
var pedRest = [PedJointCount]PedVec{
	{0, .91, 0}, {0, 1.43, 0}, {0, 1.66, 0}, {-.23, 1.48, 0}, {-.29, 1.18, .06}, {-.28, .90, .10},
	{.23, 1.48, 0}, {.29, 1.18, .06}, {.28, .90, .10}, {-.12, .87, 0}, {-.13, .49, .045}, {-.13, .09, .055},
	{.12, .87, 0}, {.13, .49, .045}, {.13, .09, .055},
}

type pedConstraint struct {
	A, B     int
	Min, Max float64
}

var pedConstraints = func() []pedConstraint {
	var out []pedConstraint
	for _, b := range PedBones {
		l := pedRest[b.A].Sub(pedRest[b.B]).Length()
		out = append(out, pedConstraint{b.A, b.B, l, l})
	}
	// Cross braces preserve torso/hip volume; elbow/knee spans limit extension.
	for _, p := range [][2]int{{3, 6}, {9, 12}, {3, 12}, {6, 9}, {3, 9}, {6, 12}} {
		l := pedRest[p[0]].Sub(pedRest[p[1]]).Length()
		out = append(out, pedConstraint{p[0], p[1], l, l})
	}
	for _, p := range [][3]int{{3, 4, 5}, {6, 7, 8}, {9, 10, 11}, {12, 13, 14}} {
		l := pedRest[p[0]].Sub(pedRest[p[1]]).Length() + pedRest[p[1]].Sub(pedRest[p[2]]).Length()
		out = append(out, pedConstraint{p[0], p[2], l * .22, l * .985})
	}
	return out
}()

type SidewalkPath struct {
	A, B   Vec
	Length float64
}
type PedParticle struct {
	Pos, Previous       PedVec
	InverseMass, Radius float64
}
type Pedestrian struct {
	ID, Path                                           int
	Pos                                                Vec
	Height, Heading, Distance, Direction, Speed, Phase float64
	Pose, PreviousPose                                 [PedJointCount]PedVec
	Particles                                          [PedJointCount]PedParticle
	Ragdoll, Sleeping, Moving                          bool
	RagdollAge, SleepTime, Recover, PreviousDT         float64
	groundAt, avoidUntil                               float64
}

func (p *Pedestrian) Velocity() Vec {
	if p.Ragdoll {
		if p.Sleeping {
			return Vec{}
		}
		return p.Particles[0].Pos.Sub(p.Particles[0].Previous).Mul(1 / math.Max(1./120, p.PreviousDT)).XZ()
	}
	if !p.Moving {
		return Vec{}
	}
	return HeadingVec(p.Heading).Mul(p.Speed)
}
func (p *Pedestrian) walkingPose() [PedJointCount]PedVec {
	pose := pedRest
	// Distance-driven stance moves backwards at walking speed, keeping the
	// planted shoe still in world space. Only the recovery foot lifts.
	if p.Moving {
		stride := math.Pi / 5.5
		bob := .008*math.Cos(p.Phase*2) - .055
		for i := range pose {
			if i != 11 && i != 14 {
				pose[i].Y += bob
			}
		}
		for side, joints := range [][3]int{{9, 10, 11}, {12, 13, 14}} {
			phase := math.Mod(p.Phase+float64(side)*math.Pi, 2*math.Pi)
			if phase < 0 {
				phase += 2 * math.Pi
			}
			u := phase / math.Pi
			travel, lift := stride*(.5-u), 0.0
			if u > 1 {
				u -= 1
				travel = stride * (-.5 + u*u*(3-2*u))
				lift = math.Sin(u*math.Pi) * .11
			}
			hip, knee, foot := joints[0], joints[1], joints[2]
			pose[foot].Z += travel
			pose[foot].Y += lift
			// Bend knees toward the travel direction; two-bone IK retains
			// anatomical leg lengths instead of stretching the imported skin.
			a, b := pedRest[hip].Sub(pedRest[knee]).Length(), pedRest[knee].Sub(pedRest[foot]).Length()
			delta := pose[foot].Sub(pose[hip])
			length := math.Min(delta.Length(), a+b-.001)
			axis := delta.Unit()
			along := (a*a - b*b + length*length) / (2 * length)
			bend := math.Sqrt(math.Max(0, a*a-along*along))
			forward := PedVec{Y: axis.Z, Z: -axis.Y}
			pose[knee] = pose[hip].Add(axis.Mul(along)).Add(forward.Mul(bend))
		}
		swing := math.Sin(p.Phase) * .15
		pose[4].Z -= swing * .55
		pose[5].Z -= swing
		pose[7].Z += swing * .55
		pose[8].Z += swing
	}
	f := HeadingVec(p.Heading)
	r := Vec{f.Z, -f.X}
	for i, q := range pose {
		pose[i] = PedVec{p.Pos.X + r.X*q.X + f.X*q.Z, p.Height + q.Y, p.Pos.Z + r.Z*q.X + f.Z*q.Z}
	}
	return pose
}

// Circle against an oriented horizontal body, including the inside case.
func pedBodyContact(pos Vec, radius float64, center, half Vec, heading float64) (Vec, float64, bool) {
	f := HeadingVec(heading)
	r := Vec{f.Z, -f.X}
	d := pos.Sub(center)
	x, z := d.X*r.X+d.Z*r.Z, d.X*f.X+d.Z*f.Z
	qx, qz := Clamp(x, -half.X, half.X), Clamp(z, -half.Z, half.Z)
	dx, dz := x-qx, z-qz
	distance := math.Hypot(dx, dz)
	if distance >= radius {
		return Vec{}, 0, false
	}
	if distance > 1e-7 {
		return r.Mul(dx / distance).Add(f.Mul(dz / distance)), radius - distance, true
	}
	if half.X-math.Abs(x) < half.Z-math.Abs(z) {
		return r.Mul(math.Copysign(1, x)), radius + half.X - math.Abs(x), true
	}
	return f.Mul(math.Copysign(1, z)), radius + half.Z - math.Abs(z), true
}
func (w *World) pedBlocked(pos Vec, y, radius float64) bool {
	k := gridOf(pos)
	for x := k.X - 1; x <= k.X+1; x++ {
		for z := k.Z - 1; z <= k.Z+1; z++ {
			for _, i := range w.Remaster.staticGrid[gridKey{x, z}] {
				c := w.Remaster.Def.Colliders[i]
				if c.Object < 0 || c.Object >= len(w.Remaster.Def.Objects) || !w.Remaster.Def.Objects[c.Object].Solid || c.MaxY < y+.1 || c.MinY > y+1.7 {
					continue
				}
				if _, _, hit := pedBodyContact(pos, radius, c.Center, c.Half, c.Heading); hit {
					return true
				}
			}
		}
	}
	return false
}

// Walk only on authored sidewalk triangles, with feet on their actual height.
func (w *World) pedestrianPavement(pos Vec) (float64, bool) {
	for _, b := range w.Remaster.Def.BridgeSpans {
		v := b.End.Sub(b.Start)
		length := v.Length()
		if length < .01 {
			continue
		}
		f := v.Mul(1 / length)
		q := pos.Sub(b.Start)
		along := q.X*f.X + q.Z*f.Z
		side := math.Abs(q.X*f.Z - q.Z*f.X)
		if along >= 0 && along <= length && side <= b.Width*.5+2.7 {
			if side > b.Width*.5+.25 && side < b.Width*.5+2.45 {
				return b.Y + .18, true
			}
			return 0, false
		}
	}
	ground := w.Remaster.Def.Grid.Height(pos)
	height := math.Inf(-1)
	for _, i := range w.Remaster.surfaceGrid[remasterMeshCell(pos)] {
		t := w.Remaster.Def.Surfaces[i]
		if !t.Support || t.Road {
			continue
		}
		if y, ok := triangleHeight(t, pos); ok && y >= ground-.15 && y < ground+1.5 {
			height = math.Max(height, y)
		}
	}
	return height, !math.IsInf(height, -1) && math.Abs(height-w.GroundHeight(pos)) < .03
}

func (w *World) SeedPedestrians() {
	defer w.rebuildPedestrianGrid()
	w.Pedestrians = nil
	w.SidewalkPaths = nil
	w.pedestrianTick = 0
	if w.Remaster == nil || w.Remaster.Mode != 0 {
		return
	}
	for ri, road := range w.Remaster.Def.Roads {
		if road.Kind != "town" || ri >= len(w.Routes) {
			continue
		}
		route := w.Routes[ri]
		for at := 8.; at+12 < route.Len-8; at += 28 {
			a, af := route.AtDistance(at)
			b, bf := route.AtDistance(math.Min(at+24, route.Len-8))
			if af.X*bf.X+af.Z*bf.Z < .94 {
				continue
			}
			for _, sign := range []float64{-1, 1} {
				offset := sign * (road.Width*.5 + 2.0)
				aa := a.Add(Vec{af.Z, -af.X}.Mul(offset))
				bb := b.Add(Vec{bf.Z, -bf.X}.Mul(offset))
				mid := aa.Add(bb).Mul(.5)
				distance, width, _ := w.RemasteredRoadAt(mid)
				if distance < width*.5+.65 || distance > width*.5+3.6 {
					continue
				}
				h := w.GroundHeight(mid)
				if _, wet := w.RemasteredWaterLevel(mid); wet && h < w.Remaster.Def.WaterHeight+.4 {
					continue
				}
				clear := true
				for u := 0.; u <= 1.001; u += .125 {
					q := aa.Add(bb.Sub(aa).Mul(u))
					y, paved := w.pedestrianPavement(q)
					if !paved || w.pedBlocked(q, y, .36) {
						clear = false
						break
					}
				}
				if !clear {
					continue
				}
				w.SidewalkPaths = append(w.SidewalkPaths, SidewalkPath{aa, bb, aa.Dist(bb)})
			}
		}
	}
	count := min(640, len(w.SidewalkPaths))
	for i := 0; i < count; i++ {
		path := i * len(w.SidewalkPaths) / count
		lane := w.SidewalkPaths[path]
		direction := 1.
		if i%2 == 1 {
			direction = -1
		}
		distance := lane.Length * (.18 + .64*float64((i*17)%31)/31)
		pos := lane.A.Add(lane.B.Sub(lane.A).Mul(distance / lane.Length))
		f := lane.B.Sub(lane.A).Unit().Mul(direction)
		y, _ := w.pedestrianPavement(pos)
		p := Pedestrian{ID: i, Path: path, Pos: pos, Height: y, Heading: math.Atan2(f.X, f.Z), Distance: distance, Direction: direction, Speed: 1.05 + .1*float64(i%5), Phase: float64(i) * 1.7, Moving: true}
		p.Pose = p.walkingPose()
		p.PreviousPose = p.Pose
		w.Pedestrians = append(w.Pedestrians, p)
	}
}

func (p *Pedestrian) BeginRagdoll(velocity PedVec) {
	p.Ragdoll, p.Sleeping = true, false
	p.RagdollAge, p.SleepTime, p.Recover = 0, 0, 0
	p.PreviousDT = 1. / 120
	masses := [PedJointCount]float64{14, 18, 5, 2, 2, 1, 2, 2, 1, 4, 7, 3, 4, 7, 3}
	for i, pos := range p.Pose {
		radius := .085
		if i == 0 {
			radius = .13
		}
		if i == 1 {
			radius = .18
		}
		if i == 2 {
			radius = .13
		}
		spin := PedVec{(pos.Y - p.Height - .9) * .5, 0, 0}
		p.Particles[i] = PedParticle{pos, pos.Sub(velocity.Add(spin).Mul(p.PreviousDT)), 1 / masses[i], radius}
	}
}
func (w *World) pedVehicleHit(p *Pedestrian, dt float64) {
	if p.Ragdoll {
		return
	}
	hit := func(pos Vec, heading float64, half Vec, height float64, velocity Vec, mass float64, apply func(Vec)) {
		if p.Ragdoll || p.Pos.Dist(pos) > half.Length()+1 || math.Abs(p.Height-height) > 1.6 {
			return
		}
		n, _, contact := pedBodyContact(p.Pos, .30, pos, half, heading)
		if !contact {
			return
		}
		v := p.Velocity()
		relative := velocity.Sub(v)
		closing := relative.X*n.X + relative.Z*n.Z
		if closing < .7 {
			return
		}
		impulse := closing * 1.08 / (1/75. + 1/mass)
		apply(n.Mul(-impulse / mass))
		v = v.Add(n.Mul(impulse / 75))
		p.BeginRagdoll(PedVec{v.X, math.Min(3, closing*.12), v.Z})
	}
	c := &w.Car
	spec := VehicleVariantSpec050(w.CarClass, w.Car.Variant050)
	f := HeadingVec(c.Heading)
	r := Vec{f.Z, -f.X}
	hit(c.Pos, c.Heading, Vec{spec.HalfWidth, spec.HalfLength}, c.Body.Height-.74, f.Mul(c.Speed).Add(r.Mul(c.Lateral)), spec.P.Mass, func(d Vec) {
		c.Speed += d.X*f.X + d.Z*f.Z
		c.Lateral += d.X*r.X + d.Z*r.Z
		c.ContactImpulse = math.Max(c.ContactImpulse, d.Length())
	})
	for i := range w.Traffic {
		t := &w.Traffic[i]
		if t.Pos.Dist(p.Pos) > 9 {
			continue
		}
		hit(t.Pos, t.Heading, TrafficBodyFootprint(t.Variant), w.GroundHeight(t.Pos), TrafficVelocity(*t), TrafficBodyMass(t.Variant), func(d Vec) { t.Impulse = t.Impulse.Add(d); t.ImpactCooldown = math.Max(t.ImpactCooldown, .6) })
	}
	_ = dt
}

func constrainPed(p *Pedestrian, c pedConstraint) {
	a, b := &p.Particles[c.A], &p.Particles[c.B]
	delta := b.Pos.Sub(a.Pos)
	l := delta.Length()
	if l < 1e-8 {
		return
	}
	target := Clamp(l, c.Min, c.Max)
	correction := delta.Mul((l - target) / (l * (a.InverseMass + b.InverseMass)))
	a.Pos = a.Pos.Add(correction.Mul(a.InverseMass))
	b.Pos = b.Pos.Sub(correction.Mul(b.InverseMass))
}

// A sphere against the full vehicle box, including the roof and underneath.
// Each limb transfers its own impulse back to the vehicle instead of allowing
// a fallen body to pass through a moving car after the first impact.
func pedParticleVehicle(a *PedParticle, dt float64, pos Vec, heading float64, half Vec, bottom, top float64, velocity Vec, mass float64, apply func(Vec)) {
	f := HeadingVec(heading)
	r := Vec{f.Z, -f.X}
	d := a.Pos.XZ().Sub(pos)
	x, z := d.X*r.X+d.Z*r.Z, d.X*f.X+d.Z*f.Z
	y := a.Pos.Y
	delta := PedVec{x - Clamp(x, -half.X, half.X), y - Clamp(y, bottom, top), z - Clamp(z, -half.Z, half.Z)}
	l := delta.Length()
	if l >= a.Radius {
		return
	}
	depth := a.Radius - l
	if l > 1e-7 {
		delta = delta.Mul(1 / l)
	} else {
		xd, zd := half.X-math.Abs(x), half.Z-math.Abs(z)
		yd := math.Min(y-bottom, top-y)
		if xd < zd && xd < yd {
			delta = PedVec{math.Copysign(1, x), 0, 0}
			depth = a.Radius + xd
		} else if zd < yd {
			delta = PedVec{0, 0, math.Copysign(1, z)}
			depth = a.Radius + zd
		} else {
			delta = PedVec{0, math.Copysign(1, y-(bottom+top)*.5), 0}
			depth = a.Radius + yd
		}
	}
	n := PedVec{r.X*delta.X + f.X*delta.Z, delta.Y, r.Z*delta.X + f.Z*delta.Z}
	v := a.Pos.Sub(a.Previous).Mul(1 / dt)
	a.Pos = a.Pos.Add(n.Mul(depth + .001))
	relative := v.Sub(PedVec{velocity.X, 0, velocity.Z})
	vn := relative.Dot(n)
	if vn < 0 {
		j := -vn * 1.02 / (a.InverseMass + 1/mass)
		v = v.Add(n.Mul(j * a.InverseMass))
		apply(Vec{-n.X * j / mass, -n.Z * j / mass})
	}
	a.Previous = a.Pos.Sub(v.Mul(dt))
}

func (w *World) ragdollVehicleContacts(p *Pedestrian, dt float64) {
	c := &w.Car
	spec := VehicleVariantSpec050(w.CarClass, w.Car.Variant050)
	if c.Pos.Dist(p.Pos) < spec.HalfLength+3 {
		for i := range p.Particles {
			f := HeadingVec(c.Heading)
			r := Vec{f.Z, -f.X}
			pedParticleVehicle(&p.Particles[i], dt, c.Pos, c.Heading, Vec{spec.HalfWidth, spec.HalfLength}, c.Body.Height-.62, c.Body.Height+.55, f.Mul(c.Speed).Add(r.Mul(c.Lateral)), spec.P.Mass, func(d Vec) { c.Speed += d.X*f.X + d.Z*f.Z; c.Lateral += d.X*r.X + d.Z*r.Z })
		}
	}
	for j := range w.Traffic {
		t := &w.Traffic[j]
		half := TrafficBodyFootprint(t.Variant)
		if t.Pos.Dist(p.Pos) > half.Z+3 {
			continue
		}
		height := 1.5
		if t.Variant == 0 {
			height = 3.0
		}
		if t.Variant == 3 {
			height = 2.6
		}
		ground := w.GroundHeight(t.Pos)
		for i := range p.Particles {
			pedParticleVehicle(&p.Particles[i], dt, t.Pos, t.Heading, half, ground+.12, ground+height, TrafficVelocity(*t), TrafficBodyMass(t.Variant), func(d Vec) { t.Impulse = t.Impulse.Add(d); t.ImpactCooldown = math.Max(t.ImpactCooldown, .6) })
		}
	}
}
func (w *World) stepPedRagdoll(p *Pedestrian, dt float64) {
	p.RagdollAge += dt
	if p.Sleeping {
		wake := w.Car.Pos.Dist(p.Pos) < 7 && math.Abs(w.Car.Speed) > 1
		if !wake {
			for _, t := range w.Traffic {
				if t.Pos.Dist(p.Pos) < 9 && TrafficVelocity(t).Length() > 1 {
					wake = true
					break
				}
			}
		}
		if wake {
			p.Sleeping = false
			p.SleepTime = 0
		} else {
			return
		}
	}
	var ground [PedJointCount]float64
	for i := range p.Particles {
		a := &p.Particles[i]
		v := a.Pos.Sub(a.Previous).Mul(1 / math.Max(1e-5, p.PreviousDT))
		v = v.Mul(math.Exp(-dt * .9))
		if v.Length() > 35 {
			v = v.Unit().Mul(35)
		}
		a.Previous = a.Pos
		a.Pos = a.Pos.Add(v.Mul(dt)).Add(PedVec{0, -9.81 * dt * dt, 0})
		ground[i] = w.GroundHeight(a.Pos.XZ())
	}
	p.PreviousDT = dt
	for iteration := 0; iteration < 8; iteration++ {
		for _, c := range pedConstraints {
			constrainPed(p, c)
		}
		for i := range p.Particles {
			a := &p.Particles[i]
			floor := ground[i] + a.Radius
			if a.Pos.Y < floor {
				a.Pos.Y = floor
			}
		}
		// The whole limb capsule is supported, not only its joint centres.
		for _, b := range PedBones {
			a, c := &p.Particles[b.A], &p.Particles[b.B]
			for _, t := range []float64{.25, .5, .75} {
				y := Lerp(a.Pos.Y, c.Pos.Y, t)
				floor := Lerp(ground[b.A], ground[b.B], t) + b.Radius
				if y < floor {
					d := floor - y
					a.Pos.Y += d * (1 - t)
					c.Pos.Y += d * t
				}
			}
		}
	}
	w.ragdollVehicleContacts(p, dt)
	// Restore joints after vehicle separation, then enforce terrain contact.
	for iteration := 0; iteration < 4; iteration++ {
		for _, c := range pedConstraints {
			constrainPed(p, c)
		}
		for i := range p.Particles {
			a := &p.Particles[i]
			a.Pos.Y = math.Max(a.Pos.Y, ground[i]+a.Radius)
		}
	}
	maxMotion := 0.
	for i := range p.Particles {
		a := &p.Particles[i]
		k := gridOf(a.Pos.XZ())
		for _, index := range w.Remaster.staticGrid[k] {
			c := w.Remaster.Def.Colliders[index]
			if c.Object < 0 || c.Object >= len(w.Remaster.Def.Objects) || !w.Remaster.Def.Objects[c.Object].Solid || a.Pos.Y+a.Radius < c.MinY || a.Pos.Y-a.Radius > c.MaxY {
				continue
			}
			if n, depth, hit := pedBodyContact(a.Pos.XZ(), a.Radius, c.Center, c.Half, c.Heading); hit {
				a.Pos.X += n.X * depth
				a.Pos.Z += n.Z * depth
			}
		}
		if a.Pos.Y < ground[i]+a.Radius+.025 {
			v := a.Pos.Sub(a.Previous)
			a.Previous = PedVec{a.Pos.X - v.X*.78, a.Pos.Y, a.Pos.Z - v.Z*.78}
		}
		maxMotion = math.Max(maxMotion, a.Pos.Sub(a.Previous).Length()/dt)
		p.Pose[i] = a.Pos
	}
	p.Pos = p.Particles[0].Pos.XZ()
	p.Height = w.GroundHeight(p.Pos)
	if maxMotion < .12 {
		p.SleepTime += dt
	} else {
		p.SleepTime = 0
	}
	if p.SleepTime > 1.5 {
		p.Sleeping = true
	}
}

func (w *World) StepPedestrians(dt float64) {
	if w.Remaster == nil || w.Remaster.Mode != 0 || dt <= 0 {
		return
	}
	w.pedestrianTick++
	if w.pedestrianTick%12 == 1 {
		w.rebuildPedestrianGrid()
	}
	for i := range w.Pedestrians {
		p := &w.Pedestrians[i]
		distance := p.Pos.Dist(w.Car.Pos)
		step := dt
		if distance > 190 {
			if w.pedestrianTick%8 != uint64(i)%8 {
				continue
			}
			step *= 8
		}
		p.PreviousPose = p.Pose
		if p.Ragdoll {
			w.stepPedRagdoll(p, step)
			continue
		}
		if p.Path < 0 || p.Path >= len(w.SidewalkPaths) {
			continue
		}
		lane := w.SidewalkPaths[p.Path]
		if lane.Length < .01 {
			continue
		}
		nextDistance := p.Distance + p.Direction*p.Speed*step
		if nextDistance > lane.Length {
			nextDistance = lane.Length
			p.Direction = -1
		}
		if nextDistance < 0 {
			nextDistance = 0
			p.Direction = 1
		}
		target := lane.A.Add(lane.B.Sub(lane.A).Mul(nextDistance / lane.Length))
		delta := target.Sub(p.Pos)
		if delta.Length() > p.Speed*step {
			delta = delta.Unit().Mul(p.Speed * step)
		}
		candidate := p.Pos.Add(delta)
		if distance < 170 && w.Elapsed >= p.avoidUntil && w.pedestrianTick%12 == uint64(i)%12 && w.pedBlocked(candidate, p.Height, .30) {
			p.Direction = -p.Direction
			p.avoidUntil = w.Elapsed + .7
		}
		p.Moving = w.Elapsed >= p.avoidUntil
		if !p.Moving {
			candidate = p.Pos
			delta = Vec{}
		} else {
			p.Distance = nextDistance
		}
		p.Pos = candidate
		f := lane.B.Sub(lane.A).Unit().Mul(p.Direction)
		p.Heading = math.Atan2(f.X, f.Z)
		if w.Elapsed >= p.groundAt {
			if y, ok := w.pedestrianPavement(p.Pos); ok {
				p.Height = y
			} else {
				p.Direction = -p.Direction
			}
			p.groundAt = w.Elapsed + .10
		}
		p.Phase += delta.Length() * 5.5
		p.Pose = p.walkingPose()
		if distance < 100 {
			w.pedVehicleHit(p, step)
			if p.Ragdoll {
				w.stepPedRagdoll(p, step)
			}
		}
	}
}

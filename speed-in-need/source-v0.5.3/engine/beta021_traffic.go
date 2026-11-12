package engine

import "math"

// Signals stop an approach, never a car that has already passed the stop line.
func (w *World) trafficRedGap021(t *Traffic) float64 {
	best := math.Inf(1)
	route := w.Routes[t.Path]
	s := w.Remaster
	if s.SignalsByRoute021 == nil || s.SignalCount021 != len(s.Def.Signals) {
		s.SignalsByRoute021 = make([][]int, len(w.Routes))
		s.SignalCount021 = len(s.Def.Signals)
		for i, signal := range s.Def.Signals {
			if signal.Route >= 0 && signal.Route < len(w.Routes) {
				s.SignalsByRoute021[signal.Route] = append(s.SignalsByRoute021[signal.Route], i)
			}
		}
	}
	for _, index := range s.SignalsByRoute021[t.Path] {
		signal := s.Def.Signals[index]
		if signal.Route != t.Path || signal.Direction != 0 && signal.Direction != t.Direction || !w.RemasteredSignalRed(signal) {
			continue
		}
		gap := (signal.Distance-t.T)*t.Direction - TrafficBodyFootprint(t.Variant).Z - 1
		if route.Closed {
			gap = math.Mod(gap+route.Len, route.Len)
		}
		if gap >= -.5 && gap < best {
			best = gap
		}
	}
	return best
}

// Resolve an actual future crossing, rather than stopping for a car on a
// neighbouring road or on the opposite side of the same hairpin.
func trafficCrossingYield021(t, other Traffic) bool {
	if t.Path == other.Path || other.Speed < .6 {
		return false
	}
	f, g := HeadingVec(t.Heading), HeadingVec(other.Heading)
	cross := f.X*g.Z - f.Z*g.X
	if math.Abs(cross) < .45 {
		return false
	}
	delta := other.Pos.Sub(t.Pos)
	a := (delta.X*g.Z - delta.Z*g.X) / cross
	b := (delta.X*f.Z - delta.Z*f.X) / cross
	if a < 0 || a > 26 || b < 0 || b > 26 {
		return false
	}
	ta, tb := a/math.Max(2, t.Speed), b/math.Max(2, other.Speed)
	return math.Abs(ta-tb) < 1.25 && (b < 3 || other.Path < t.Path)
}

func (w *World) turnAtDeadEnd021(t *Traffic) {
	oldPos, oldHeading := t.Pos, t.Heading
	t.T = Clamp(t.T, .05, w.Routes[t.Path].Len-.05)
	t.Direction = -t.Direction
	p, f := w.Routes[t.Path].AtDistance(t.T)
	f = f.Mul(t.Direction)
	t.Offset = oldPos.Sub(p.Add(Vec{-f.Z, f.X}.Mul(t.Lane)))
	t.YawKick = math.Atan2(math.Sin(oldHeading-math.Atan2(f.X, f.Z)), math.Cos(oldHeading-math.Atan2(f.X, f.Z)))
	t.Speed = math.Min(3, t.Speed)
}

func (w *World) recycleBlockedTraffic021(index int) bool {
	t := &w.Traffic[index]
	if t.Pos.Dist(w.Car.Pos) < 220 {
		return false
	}
	for trial := 0; trial < 12; trial++ {
		ri := (index*37 + trial*17 + int(w.Remaster.TrafficTick020/120)) % len(w.Routes)
		route := w.Routes[ri]
		if route.Len < 130 || w.Remaster.Def.Roads[ri].Kind == "stunt" || !w.civilianRoute035(ri) {
			continue
		}
		at := math.Mod(float64(index*73+trial*113), route.Len-30) + 15
		direction := 1.0
		if index%2 != 0 {
			direction = -1
		}
		p, f := route.AtDistance(at)
		f = f.Mul(direction)
		p = p.Add(Vec{-f.Z, f.X}.Mul(t.Lane))
		if p.Dist(w.Car.Pos) < 240 {
			continue
		}
		clear := true
		for _, j := range w.Remaster.beta020TrafficNearby(p) {
			if j != index && w.Traffic[j].Pos.Dist(p) < TrafficBodyFootprint(t.Variant).Z+TrafficBodyFootprint(w.Traffic[j].Variant).Z+8 {
				clear = false
				break
			}
		}
		if !clear {
			continue
		}
		t.Path = ri
		t.T = at
		t.Direction = direction
		t.Pos = p
		t.Heading = math.Atan2(f.X, f.Z)
		t.Offset = Vec{}
		t.Impulse = Vec{}
		t.YawKick = 0
		t.Speed = 5
		t.StalledSeconds = 0
		return true
	}
	return false
}

func (w *World) stepRemasteredTraffic(dt float64) {
	s := w.Remaster
	if !s.Def.Beta020 {
		w.stepRemasteredTraffic020(dt)
		return
	}
	tickDT050 := dt
	snapshot := w.trafficSnapshot()
	s.beta020TrafficGrid(snapshot)
	for i := range w.Traffic {
		t := &w.Traffic[i]
		playerDistanceSquared038 := w.observerDistanceSquared050(t.Pos)
		dt := w.trafficStep050(t, i, tickDT050, playerDistanceSquared038)
		if dt <= 0 {
			continue
		}
		oldPos, oldHeading := t.Pos, t.Heading
		route := w.Routes[t.Path]
		// Older dead-end logic permanently set CruiseSpeed to zero.
		if t.CruiseSpeed < 1 && !route.Closed && (t.T < .1 || t.T > route.Len-.1) {
			t.CruiseSpeed = 10
		}
		desired := s.TrafficDesired020[i]
		redGap := w.trafficRedGap021(t)
		if len(w.Traffic) < 128 || playerDistanceSquared038 < 22500 || dt > tickDT050*1.5 || s.TrafficTick020%6 == uint64(i)%6 {
			desired = w.trafficRoadSpeed(t)
			if playerDistanceSquared038 < 90000 {
				_, h0 := TrafficPose026(route, t.T, t.Direction, t.Lane)
				look := Clamp(t.Speed*1.5, 9, 24)
				_, h1 := TrafficPose026(route, t.T+look*t.Direction, t.Direction, t.Lane)
				curvature := math.Abs(math.Remainder(h1-h0, Tau)) / look
				if curvature > .002 {
					desired = math.Min(desired, math.Sqrt(3.0/curvature))
				}
			}
			nearby := s.beta020TrafficNearby(t.Pos)
			f := HeadingVec(t.Heading)
			ownLength := TrafficBodyFootprint(t.Variant).Z
			check := func(pos Vec, otherSpeed, otherLength, halfWidth float64) {
				delta := pos.Sub(t.Pos)
				ahead := delta.X*f.X + delta.Z*f.Z
				side := math.Abs(delta.X*f.Z - delta.Z*f.X)
				if ahead > 0 && ahead < 65 && side < TrafficBodyFootprint(t.Variant).X+halfWidth-.15 {
					desired = math.Min(desired, PredictiveFollowingSpeed(t.Speed, math.Max(0, otherSpeed), ahead-ownLength-otherLength))
				}
			}
			pf := HeadingVec(w.Car.Heading)
			check(w.Car.Pos, w.Car.Speed*(pf.X*f.X+pf.Z*f.Z), VehicleVariantSpec050(w.CarClass, w.Car.Variant050).HalfLength, VehicleVariantSpec050(w.CarClass, w.Car.Variant050).HalfWidth)
			for _, j := range nearby {
				if i == j {
					continue
				}
				other := snapshot[j]
				if other.Path == t.Path && other.Direction == t.Direction {
					gap := (other.T - t.T) * t.Direction
					if route.Closed {
						gap = math.Mod(gap+route.Len, route.Len)
					}
					if gap > 0 && gap < 70 {
						desired = math.Min(desired, PredictiveFollowingSpeed(t.Speed, other.Speed, gap-ownLength-TrafficBodyFootprint(other.Variant).Z))
					}
				} else {
					of := HeadingVec(other.Heading)
					if f.X*of.X+f.Z*of.Z > .82 {
						check(other.Pos, other.Speed, TrafficBodyFootprint(other.Variant).Z, TrafficBodyFootprint(other.Variant).X)
					}
					if trafficCrossingYield021(*t, other) {
						delta := other.Pos.Sub(t.Pos)
						desired = math.Min(desired, math.Sqrt(5*math.Max(0, delta.Length()-8)))
					}
				}
			}
			if redGap < 55 {
				desired = math.Min(desired, math.Sqrt(6.4*math.Max(0, redGap-.45)))
			}
			desired = w.trafficHazardSpeed(t, desired)
			s.TrafficDesired020[i] = desired
		}
		if t.ImpactCooldown > .45 {
			desired = 0
		}
		change := Clamp(desired-t.Speed, -6.5*dt, 2.0*dt)
		t.Braking = change < -.05*dt
		t.Speed = math.Max(0, t.Speed+change)
		t.T += t.Speed * t.Direction * dt
		if route.Closed {
			t.T = math.Mod(t.T+route.Len, route.Len)
		} else if t.T > route.Len || t.T < 0 {
			overshoot := math.Max(t.T-route.Len, -t.T)
			if !w.connectAokawaTraffic(t, overshoot) {
				w.turnAtDeadEnd021(t)
			}
			route = w.Routes[t.Path]
		}
		// Distant traffic keeps the route's existing smooth tangent. The extra
		// axle-span samples are reserved for cars whose corner pose is visible.
		wanted, forward := route.AtDistance(t.T)
		forward = forward.Mul(t.Direction)
		roadHeading := math.Atan2(forward.X, forward.Z)
		wanted = wanted.Add(Vec{-forward.Z, forward.X}.Mul(t.Lane))
		if playerDistanceSquared038 < 40000 {
			wanted, roadHeading = TrafficCornerPose(route, t.T, t.Direction, t.Lane, t.Speed)
		}
		RecoverImpactedTraffic(t, dt)
		t.Pos = wanted.Add(t.Offset)
		if t.ImpactCooldown > 0 || t.Offset.Length() > .2 {
			w.trafficStaticContacts(t)
		}
		angle := math.Atan2(math.Sin(roadHeading+t.YawKick-t.Heading), math.Cos(roadHeading+t.YawKick-t.Heading))
		if playerDistanceSquared038 < 40000 && math.Abs(t.YawKick) < .02 && t.Offset.Length() < .05 {
			t.Heading += Clamp(angle, -math.Max(.8, t.Speed/6)*dt, math.Max(.8, t.Speed/6)*dt)
		} else {
			t.Heading += Clamp(angle, -1.5*dt, 1.5*dt)
		}
		motion := t.Pos.Sub(oldPos).Length()
		if motion < 2 {
			displacement := t.Pos.Sub(oldPos)
			f := HeadingVec(t.Heading)
			t.WheelSpin = math.Mod(t.WheelSpin+(displacement.X*f.X+displacement.Z*f.Z)/.34, Tau)
		}
		if playerDistanceSquared038 < 40000 {
			yawRate := math.Remainder(t.Heading-oldHeading, Tau) / dt
			wantedSteer := Clamp(math.Atan(2.65*yawRate/math.Max(2, t.Speed)), -.60, .60)
			t.SteerAngle += Clamp(wantedSteer-t.SteerAngle, -3.2*dt, 3.2*dt)
			forward := HeadingVec(t.Heading)
			right := Vec{forward.Z, -forward.X}
			half := math.Min(3.4, TrafficBodyFootprint(t.Variant).Z*.65)
			pitch := -math.Atan2(w.GroundHeight(t.Pos.Add(forward.Mul(half)))-w.GroundHeight(t.Pos.Sub(forward.Mul(half))), 2*half)
			roll := math.Atan2(w.GroundHeight(t.Pos.Add(right.Mul(.8)))-w.GroundHeight(t.Pos.Sub(right.Mul(.8))), 1.6)
			t.Pitch = Lerp(t.Pitch, Clamp(pitch, -.45, .45), 1-math.Exp(-dt*9))
			t.Roll = Lerp(t.Roll, Clamp(roll, -.42, .42), 1-math.Exp(-dt*9))
		}
		if t.Speed < .25 && redGap > 60 {
			t.StalledSeconds += dt
		} else {
			t.StalledSeconds = 0
		}
		// An offscreen jam may recycle a route record; visible cars never teleport.
		if t.StalledSeconds > 24 && (uint64(i)+s.TrafficTick020)%120 == 0 {
			w.recycleBlockedTraffic021(i)
		}
	}
}

func predictiveFollowingSpeed020(followerSpeed, leaderSpeed, bumperGap float64) float64 {
	if bumperGap <= 0 {
		return 0
	}
	safeGap := 2.3 + followerSpeed*1.25
	if bumperGap >= safeGap+18 {
		return followerSpeed + 30
	}
	errorDist := bumperGap - safeGap
	closing := math.Max(0, followerSpeed-leaderSpeed)
	timeToCollision := 999.0
	if closing > 0.01 {
		timeToCollision = bumperGap / closing
	}
	target := math.Max(0, leaderSpeed+errorDist*.9)
	if timeToCollision < 2.2 {
		target = math.Min(target, math.Max(0, leaderSpeed-2.5))
	}
	return target
}

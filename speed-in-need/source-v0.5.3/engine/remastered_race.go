package engine

import "math"

func (w *World) stepRemasteredRace(dt float64) {
	s := w.Remaster
	route := w.Routes[0]
	for i := range s.Racers {
		actor := &s.Racers[i]
		if s.Countdown > 0 {
			w.stepRemasteredVehicle(&actor.Car, actor.Class, Input{Brake: 1, Handbrake: true}, dt, false)
			continue
		}
		distance, segment := routeProjection(route, actor.Car.Pos, actor.Segment)
		actor.Segment = segment
		actor.RouteDistance = distance
		look := Clamp(8+math.Abs(actor.Car.Speed)*.40, 10, 38)
		target, tangent := route.AtDistance(distance + look)
		_, far := route.AtDistance(distance + look + 35)
		curve := math.Abs(math.Atan2(tangent.X*far.Z-tangent.Z*far.X, tangent.X*far.X+tangent.Z*far.Z)) / 35
		spec := VehicleVariantSpec050(actor.Class, actor.Car.Variant050)
		spec.Mu *= TyreGrip050(actor.Car.Tyre050, actor.Car.Wetness050, 0)
		targetSpeed := math.Min(spec.TopSpeed*.98, math.Sqrt(DrivingGripCoefficient(spec, true, false, 0)*9.81*.88/math.Max(.00035, curve)))
		pace, topSpeed := RacePace040(s.RaceDifficulty040)
		topSpeed = math.Min(spec.TopSpeed*.98, topSpeed)
		targetSpeed = Clamp(targetSpeed*pace, 8, topSpeed)
		// Plan braking before a corner rather than discovering it at the front bumper.
		for _, ahead := range []float64{0, 12, 25, 45, 70, 105, 150} {
			_, a := route.AtDistance(distance + ahead)
			_, b := route.AtDistance(distance + ahead + 12)
			curvature := math.Abs(math.Atan2(a.X*b.Z-a.Z*b.X, a.X*b.X+a.Z*b.Z)) / 12
			cornerGrip := DrivingGripCoefficient(spec, true, false, 0) * []float64{.48, .72, .92}[ClampInt(s.RaceDifficulty040, 0, 2)]
			cornerSpeed := math.Sqrt(cornerGrip * 9.81 / math.Max(.0003, curvature))
			targetSpeed = math.Min(targetSpeed, math.Sqrt(cornerSpeed*cornerSpeed+2*9.2*math.Max(0, ahead-5)))
		}
		lane := -1.9
		if actor.Finished {
			targetSpeed = 12
			lane = 4.8
		}
		if i%2 == 1 && !actor.Finished {
			lane = 1.9
		}
		if curve > .015 {
			lane *= .35
		}
		centre, currentForward := route.AtDistance(distance)
		currentRight := Vec{currentForward.Z, -currentForward.X}
		nearestBlock := math.Inf(1)
		pass := func(pos Vec, speed float64) {
			delta := pos.Sub(centre)
			ahead := delta.X*currentForward.X + delta.Z*currentForward.Z
			side := delta.X*currentRight.X + delta.Z*currentRight.Z
			if ahead > 0 && ahead < 40 && ahead < nearestBlock && math.Abs(side-lane) < 2.2 && speed < math.Max(7, actor.Car.Speed-2) {
				nearestBlock = ahead
				if side >= 0 {
					lane = -4.8
				} else {
					lane = 4.8
				}
			}
		}
		pass(w.Car.Pos, w.Car.Speed)
		for j := range s.Racers {
			if j != i {
				pass(s.Racers[j].Car.Pos, s.Racers[j].Car.Speed)
			}
		}
		// Passing must stay inside the corner's usable road. The old pass
		// override requested the full outside lane even after corner planning
		// had reduced it, sending a recovering car straight back into a wall.
		laneLimit := 3.0
		if curve > .006 {
			laneLimit = 1.6
		}
		lane = Clamp(lane, -laneLimit, laneLimit)
		if actor.Car.ContactImpulse > .12 || actor.RecoveryTime > 0 {
			lane = 0
		}
		actor.LaneOffset = Lerp(actor.LaneOffset, lane, 1-math.Exp(-dt*7))
		target = target.Add(Vec{tangent.Z, -tangent.X}.Mul(actor.LaneOffset))
		toTarget := target.Sub(actor.Car.Pos)
		headingError := math.Atan2(math.Sin(math.Atan2(toTarget.X, toTarget.Z)-actor.Car.Heading), math.Cos(math.Atan2(toTarget.X, toTarget.Z)-actor.Car.Heading))
		crossTrack := math.Abs(actor.Car.Pos.Sub(centre).X*currentRight.X + actor.Car.Pos.Sub(centre).Z*currentRight.Z)
		// Reduce entry speed when the actual car has fallen behind its path,
		// including rear-heavy cars whose yaw response differs from the 296.
		targetSpeed *= Clamp(1-math.Abs(headingError)*.27, .60, 1)
		if crossTrack > 5 {
			targetSpeed = math.Min(targetSpeed, math.Max(7, 20-(crossTrack-5)*1.4))
		}
		steeringAngle := math.Atan2(2*spec.P.Wheelbase*math.Sin(headingError), math.Max(6, toTarget.Length()))
		steeringSpec := spec
		steeringSpec.Mu = Lerp(DrivingGripCoefficient(spec, true, false, actor.Car.DriftRecovery), DrivingGripCoefficient(spec, false, false, actor.Car.DriftRecovery), actor.Car.OffroadBlend)
		maxSteer := RacingSteerLimit(steeringSpec, actor.Car.Speed, false)
		input := Input{Throttle: Clamp((targetSpeed-actor.Car.Speed)*.34+.62, 0, 1), Steer: Clamp(steeringAngle/math.Max(.008, maxSteer), -1, 1)}
		if actor.Car.Speed > targetSpeed+1.5 {
			input.Brake = Clamp((actor.Car.Speed-targetSpeed)*.16, 0, 1)
			input.Throttle = 0
		}
		// The same braking model responds to nearby competitors; AI has no speed teleport.
		forward := HeadingVec(actor.Car.Heading)
		consider := func(pos Vec, speed float64) {
			d := pos.Sub(actor.Car.Pos)
			ahead := d.X*forward.X + d.Z*forward.Z
			side := math.Abs(d.X*forward.Z - d.Z*forward.X)
			if ahead > 0 && ahead < 24 && side < 1.65 {
				safe := math.Max(0, speed+(ahead-7-actor.Car.Speed*.45)*.8)
				if actor.Car.Speed > safe {
					input.Brake = math.Max(input.Brake, Clamp((actor.Car.Speed-safe)*.18, 0, 1))
					input.Throttle = 0
				}
			}
		}
		consider(w.Car.Pos, w.Car.Speed)
		for j := range s.Racers {
			if i != j {
				consider(s.Racers[j].Car.Pos, s.Racers[j].Car.Speed)
			}
		}
		if math.Abs(actor.Car.Speed) < 1.8 {
			actor.StuckTime += dt
		} else {
			actor.StuckTime = math.Max(0, actor.StuckTime-dt*2)
		}
		if actor.StuckTime > 1.6 && actor.RecoveryTime <= 0 {
			actor.RecoveryTime = 3.6
			actor.StuckTime = 0
		}
		if actor.RecoveryTime > 0 {
			actor.RecoveryTime -= dt
			if actor.RecoveryTime > 2.0 {
				// Back away at a controlled speed, with steering proportional to
				// the path error. Full reverse/full lock used to swing a nearly
				// aligned car into the other wall and spend its forward phase
				// merely braking out of reverse.
				input = Input{Brake: Clamp((3+actor.Car.Speed)*.45, 0, .65), Steer: -Clamp(headingError*1.6, -.6, .6)}
				if actor.Car.Speed < -3.3 {
					input.Throttle = .25
					input.Brake = 0
				}
			} else {
				input = Input{Throttle: 1, Steer: Clamp(headingError*2.5, -1, 1)}
			}
		}
		if actor.RecoveryTime <= 0 && input.Throttle == 0 && input.Brake > 0 && math.Abs(actor.Car.Speed) < .6 {
			input.Handbrake = true
		}
		actor.NitroTimer = math.Max(0, actor.NitroTimer-dt)
		if curve < .005 && targetSpeed > actor.Car.Speed+1.5 && actor.Car.Speed > 7 && math.Abs(headingError) < .25 && crossTrack < 2.8 && nearestBlock > 22 && actor.Car.Nitro > 12 && actor.NitroTimer <= 0 && input.Brake == 0 {
			input.Boost = true
			duty := 1.8
			if s.RaceDifficulty040 == 0 {
				duty = 1.2
			} else if s.RaceDifficulty040 == 2 {
				duty = 4.5
			}
			if math.Mod(s.RaceTime+float64(i)*.6, 6) > duty {
				input.Boost = false
			}
		}
		w.controlledAIDrift040(actor, &input, curve, headingError, crossTrack, nearestBlock, dt)
		previousNitro := actor.Car.Nitro
		w.stepRemasteredVehicle(&actor.Car, actor.Class, input, dt, false)
		actor.NitroSpent += math.Max(0, previousNitro-actor.Car.Nitro)
		if actor.Car.Boosting {
			actor.BoostSeconds040 += dt
		}
		if actor.DriftPulse040 > 0 || actor.Car.DriftRecovery > 0 {
			actor.DriftPeakAngle040 = math.Max(actor.DriftPeakAngle040, math.Abs(actor.Car.DriftAngle))
		}
		if actor.Finished {
			continue
		}
		current, _ := routeProjection(route, actor.Car.Pos, actor.Segment)
		diff := current - actor.LastDistance
		if diff < -route.Len*.5 {
			diff += route.Len
		}
		if diff > route.Len*.5 {
			diff -= route.Len
		}
		if math.Abs(diff) < math.Max(2, math.Abs(actor.Car.Speed)*dt*3) {
			actor.Progress += diff
		}
		actor.LastDistance = current
		w.advanceRacerGate(actor)
	}
	// Resolve competitor-to-competitor overlap with momentum and distinct masses.
	for i := 0; i < len(s.Racers); i++ {
		for j := i + 1; j < len(s.Racers); j++ {
			a, b := &s.Racers[i], &s.Racers[j]
			if a.Car.Pos.Dist(b.Car.Pos) > 6 {
				continue
			}
			sa, sb := VehicleVariantSpec050(a.Class, a.Car.Variant050), VehicleVariantSpec050(b.Class, b.Car.Variant050)
			n, p, hit := contactOBB(a.Car.Pos, Vec{sa.HalfWidth, sa.HalfLength}, a.Car.Heading, b.Car.Pos, Vec{sb.HalfWidth, sb.HalfLength}, b.Car.Heading)
			if hit {
				a.Car.Pos = a.Car.Pos.Add(n.Mul(p*.5 + .001))
				b.Car.Pos = b.Car.Pos.Sub(n.Mul(p*.5 + .001))
				av, bv := HeadingVec(a.Car.Heading).Mul(a.Car.Speed), HeadingVec(b.Car.Heading).Mul(b.Car.Speed)
				relative := av.Sub(bv)
				closing := relative.X*n.X + relative.Z*n.Z
				if closing < 0 {
					impulse := -closing / (1/sa.P.Mass + 1/sb.P.Mass)
					av = av.Add(n.Mul(impulse / sa.P.Mass))
					bv = bv.Sub(n.Mul(impulse / sb.P.Mass))
					af, bf := HeadingVec(a.Car.Heading), HeadingVec(b.Car.Heading)
					a.Car.Speed = av.X*af.X + av.Z*af.Z
					b.Car.Speed = bv.X*bf.X + bv.Z*bf.Z
				}
			}
		}
	}
}

// Opponents use the same handbrake, tires and recovery as the player. A short
// intentional pulse is only allowed on a clear, flat, wide medium-speed bend;
// the path follower then supplies its ordinary countersteering correction.
func (w *World) controlledAIDrift040(actor *Racer, in *Input, curve, error, cross, block, dt float64) {
	actor.DriftCooldown040 = math.Max(0, actor.DriftCooldown040-dt)
	s := w.Remaster
	unsafe := actor.Finished || actor.RecoveryTime > 0 || !actor.Car.Body.Grounded || actor.Car.ContactImpulse > .10 || cross > 2.5 || math.Abs(error) > .33 || block < 24 || in.Brake > .35
	if unsafe {
		actor.DriftPulse040 = 0
		return
	}
	if actor.DriftPulse040 <= 0 && actor.DriftCooldown040 <= 0 && s.RaceDifficulty040 > 0 && s.Def.Roads[0].Width >= 16 && actor.Car.Speed > 14 && actor.Car.Speed < 27 && curve > .0045 && curve < .018 && cross < 1.4 && math.Abs(error) > .025 && math.Abs(error) < .20 && in.Brake < .15 && math.Abs(in.Steer) > .08 {
		clear := actor.Car.Pos.Dist(w.Car.Pos) > 13
		for i := range s.Racers {
			if &s.Racers[i] != actor && actor.Car.Pos.Dist(s.Racers[i].Car.Pos) < 13 {
				clear = false
				break
			}
		}
		f := HeadingVec(actor.Car.Heading)
		grade := math.Abs(w.GroundHeight(actor.Car.Pos.Add(f.Mul(2)))-w.GroundHeight(actor.Car.Pos.Sub(f.Mul(2)))) / 4
		if clear && grade < .16 {
			actor.DriftPulse040 = .16
			actor.DriftCooldown040 = 18
			if s.RaceDifficulty040 == 2 {
				actor.DriftPulse040 = .22
				actor.DriftCooldown040 = 12
			}
			actor.IntentionalDrifts040++
		}
	}
	if actor.DriftPulse040 > 0 {
		actor.DriftPulse040 = math.Max(0, actor.DriftPulse040-dt)
		actor.IntentionalDriftSeconds040 += dt
		in.Handbrake = true
		in.Boost = false
		in.Brake = 0
		in.Throttle = Clamp(in.Throttle, .35, .65)
	}
}
func (w *World) gatePassed(pos Vec, next int) bool {
	s := w.Remaster
	if len(s.Gates) == 0 {
		return false
	}
	gate := s.Gates[next%len(s.Gates)]
	_, tangent := w.Routes[0].AtDistance(float64(next) * w.Routes[0].Len / float64(len(s.Gates)))
	delta := pos.Sub(gate)
	return delta.Length() < 32 && delta.X*tangent.X+delta.Z*tangent.Z >= 0
}
func (w *World) advanceRacerGate(r *Racer) {
	s := w.Remaster
	if !s.Started || r.Finished || !w.gatePassed(r.Car.Pos, r.NextGate) {
		return
	}
	if r.NextGate == 0 {
		if r.CrossedStart {
			r.Lap++
			if r.Lap > s.Laps {
				r.Finished = true
				r.FinishTime = s.RaceTime
			}
		}
		r.CrossedStart = true
	}
	r.NextGate = (r.NextGate + 1) % len(s.Gates)
}
func (w *World) updatePlayerRace(dt float64) {
	s := w.Remaster
	if !s.Started || s.Finished {
		return
	}
	route := w.Routes[0]
	current, segment := routeProjection(route, w.Car.Pos, s.PlayerSegment)
	s.PlayerSegment = segment
	diff := current - s.PlayerLastRouteDistance
	if diff < -route.Len*.5 {
		diff += route.Len
	}
	if diff > route.Len*.5 {
		diff -= route.Len
	}
	if math.Abs(diff) < math.Max(2, math.Abs(w.Car.Speed)*dt*3) {
		s.PlayerProgress += diff
	}
	s.PlayerLastRouteDistance = current
	if w.gatePassed(w.Car.Pos, s.NextGate) {
		if s.NextGate == 0 {
			if s.CrossedStart {
				lapTime := s.RaceTime - s.LastLapTime
				if s.BestLap == 0 || lapTime < s.BestLap {
					s.BestLap = lapTime
				}
				s.LastLapTime = s.RaceTime
				s.Lap++
				if s.Lap > s.Laps {
					s.Finished = true
					s.FinishTime = s.RaceTime
					w.Finished = true
				}
			}
			s.CrossedStart = true
		}
		s.NextGate = (s.NextGate + 1) % len(s.Gates)
	}
}
func (w *World) stepRemasteredTraffic020(dt float64) {
	s := w.Remaster
	snapshot := w.trafficSnapshot()
	s.beta020TrafficGrid(snapshot)
	for i := range w.Traffic {
		t := &w.Traffic[i]
		route := w.Routes[t.Path]
		desired := s.TrafficDesired020[i]
		if len(w.Traffic) < 128 || t.Pos.Dist(w.Car.Pos) < 180 || s.TrafficTick020%6 == uint64(i)%6 {
			desired = w.trafficRoadSpeed(t)
			nearby := s.beta020TrafficNearby(t.Pos)
			if t.ImpactCooldown > 0 {
				t.ImpactCooldown = math.Max(0, t.ImpactCooldown-dt)
			}
			f := HeadingVec(t.Heading)
			check := func(pos Vec, speed, length float64) {
				delta := pos.Sub(t.Pos)
				ahead := delta.X*f.X + delta.Z*f.Z
				side := math.Abs(delta.X*f.Z - delta.Z*f.X)
				if ahead > 0 && ahead < 65 && side < TrafficBodyFootprint(t.Variant).X+1.0 {
					bodyLength := TrafficBodyFootprint(t.Variant).Z
					desired = math.Min(desired, predictiveFollowingSpeed020(t.Speed, speed, ahead-bodyLength-length-1.6))
				}
			}
			playerForward := HeadingVec(w.Car.Heading)
			check(w.Car.Pos, w.Car.Speed*(playerForward.X*f.X+playerForward.Z*f.Z), VehicleVariantSpec050(w.CarClass, w.Car.Variant050).HalfLength)
			for _, j := range nearby {
				if i != j {
					other := snapshot[j]
					of := HeadingVec(other.Heading)
					check(other.Pos, other.Speed*(of.X*f.X+of.Z*f.Z), TrafficBodyFootprint(other.Variant).Z)
				}
			}
			for _, signal := range s.Def.Signals {
				if signal.Route != t.Path || (signal.Direction != 0 && signal.Direction != t.Direction) || !w.RemasteredSignalRed(signal) {
					continue
				}
				gap := (signal.Distance-t.T)*t.Direction - TrafficBodyFootprint(t.Variant).Z - 1.0
				if route.Closed {
					gap = math.Mod(gap+route.Len, route.Len)
				}
				if gap >= -.5 && gap < 55 {
					desired = math.Min(desired, math.Sqrt(2*3.2*math.Max(0, gap-.8)))
				}
			}
			// Give way to already crossing traffic; deterministic road priority avoids deadlock.
			for _, j := range nearby {
				if i == j {
					continue
				}
				other := snapshot[j]
				if other.Path == t.Path || other.Speed < .4 {
					continue
				}
				of := HeadingVec(other.Heading)
				dot := f.X*of.X + f.Z*of.Z
				if math.Abs(dot) > .65 {
					continue
				}
				delta := other.Pos.Sub(t.Pos)
				ahead := delta.X*f.X + delta.Z*f.Z
				if ahead < 2 || ahead > 22 || delta.Length() > 25 {
					continue
				}
				pred := other.Pos.Add(of.Mul(other.Speed * .7))
				rel := pred.Sub(t.Pos.Add(f.Mul(t.Speed * .7)))
				if rel.Length() < 6 && (other.Path < t.Path || ahead < 7) {
					desired = math.Min(desired, math.Sqrt(5*math.Max(0, ahead-7)))
				}
			}
			s.TrafficDesired020[i] = desired
		}
		change := Clamp(desired-t.Speed, -4.2*dt, 1.9*dt)
		t.Braking = change < -.05*dt
		t.Speed = math.Max(0, t.Speed+change)
		t.T += t.Speed * t.Direction * dt
		if route.Closed {
			t.T = math.Mod(t.T+route.Len, route.Len)
		} else {
			if t.T > route.Len {
				if w.connectAokawaTraffic(t, t.T-route.Len) {
					route = w.Routes[t.Path]
				} else {
					t.T = route.Len
					t.Speed = 0
					t.CruiseSpeed = 0
				}
			}
			if t.T < 0 {
				if w.connectAokawaTraffic(t, -t.T) {
					route = w.Routes[t.Path]
				} else {
					t.T = 0
					t.Speed = 0
					t.CruiseSpeed = 0
				}
			}
		}
		p, tangent := route.AtDistance(t.T)
		if t.Direction < 0 {
			tangent = tangent.Mul(-1)
		}
		RecoverTraffic(t, dt)
		t.Offset = t.Offset.Mul(math.Exp(-dt * .8))
		t.Pos = p.Add(Vec{-tangent.Z, tangent.X}.Mul(t.Lane)).Add(t.Offset)
		t.Heading = math.Atan2(tangent.X, tangent.Z) + t.YawKick
		// During a lane/route transition the offset is moving too. Orient the
		// visible body and lamps along that actual motion, not the new centreline.
		motion := t.Pos.Sub(snapshot[i].Pos)
		if t.Speed > .4 && t.Impulse.Length() < .5 && motion.Length() > 1e-5 && math.Abs(t.YawKick) < .15 {
			t.Heading = math.Atan2(motion.X, motion.Z)
		}
	}
}
func (w *World) RecoverRemasteredCar() {
	s := w.Remaster
	if s == nil {
		return
	}
	route := w.Routes[0]
	distance, _ := routeProjection(route, w.Car.Pos, -1)
	if s.Mode == 0 {
		best := math.Inf(1)
		for _, candidate := range w.Routes {
			at, _ := routeProjection(candidate, w.Car.Pos, -1)
			p, _ := candidate.AtDistance(at)
			if d := p.Dist(w.Car.Pos); d < best {
				best = d
				route = candidate
				distance = at
			}
		}
	}
	if s.Mode == 1 && len(s.Gates) > 0 {
		gate := (s.NextGate + len(s.Gates) - 1) % len(s.Gates)
		distance = float64(gate) * route.Len / float64(len(s.Gates))
	}
	p, f := route.AtDistance(distance)
	w.Car.Pos = p
	w.Car.Heading = math.Atan2(f.X, f.Z)
	w.Car.Speed, w.Car.Lateral, w.Car.YawRate = 0, 0, 0
	w.Car.Body = ChassisState{}
	w.Car.Rig = SuspensionRig{}
	w.InitializeChassis()
}

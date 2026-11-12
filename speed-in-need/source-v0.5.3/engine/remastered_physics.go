package engine

import "math"

type RemasteredVehicle struct {
	Name                                              string
	P                                                 VehicleParameters
	HalfLength, HalfWidth, CGHeight, YawInertia       float64
	CdA, DownforceArea, PowerKW, TopSpeed, BoostForce float64
	Mu, FrontStiffness, RearStiffness, FrontWeight    float64
	AWD, Paintable                                    bool
}

func RemasteredVehicleSpec(id int) RemasteredVehicle {
	switch ClampInt(id, 0, 3) {
	case 1:
		return RemasteredVehicle{"911 TURBO CLASSIC", VehicleSpec(1), 2.16, .86, .43, 2250, .76, .18, 235, 70, 1850, 1.13, 85000, 103000, .43, false, true}
	case 2:
		p := VehicleParameters{Mass: 1490, Wheelbase: 2.93, Track: 1.76, WheelRadius: .32, IdleRPM: 950, RedlineRPM: 8200, PeakTorque: 610, FinalDrive: 3.55, Ratios: [6]float64{3.12, 2.12, 1.55, 1.20, .96, .78}, ReverseRatio: 3.1, SpringRate: 48000, Damping: 4400}
		return RemasteredVehicle{"CUSTOM GT", p, 2.32, .93, .39, 2750, .68, .88, 370, 87, 2100, 1.18, 98000, 114000, .46, true, true}
	case 3:
		p := VehicleParameters{Mass: 1080, Wheelbase: 2.40, Track: 1.43, WheelRadius: .31, IdleRPM: 950, RedlineRPM: 7800, PeakTorque: 185, FinalDrive: 4.30, Ratios: [6]float64{3.58, 2.02, 1.38, 1, .86, .72}, ReverseRatio: 3.48, SpringRate: 34500, Damping: 3150}
		return RemasteredVehicle{"TRUENO DRIFT", p, 2.08, .85, .43, 1550, .72, .06, 132, 54, 1350, 1.08, 69000, 76000, .51, false, false}
	default:
		return RemasteredVehicle{"296 CHALLENGE", VehicleSpec(0), 2.30, .94, .36, 2800, .69, 1.12, 390, 85, 2200, 1.22, 105000, 123000, .45, false, true}
	}
}
func (w *World) StepRemastered(in Input, dt float64) {
	if dt <= 0 || math.IsNaN(dt) || math.IsInf(dt, 0) {
		return
	}
	// Bounded substeps preserve contact stability if an external caller supplies a large frame.
	if dt > 1./120+1e-8 {
		n := int(math.Ceil(dt / (1. / 120)))
		if n > 120 {
			n = 120
		}
		for i := 0; i < n; i++ {
			w.StepRemastered(in, dt/float64(n))
		}
		return
	}
	s := w.Remaster
	w.Elapsed += dt
	w.stepRemasteredPoles(dt)
	w.stepAokawaProps(dt)
	for _, id := range s.PickupIndices {
		if s.PickupCooldown[id] > 0 {
			s.PickupCooldown[id] = math.Max(0, s.PickupCooldown[id]-dt)
			continue
		}
		o := s.Def.Objects[id]
		if w.Car.Pos.Dist(Vec{o.Pos[0], o.Pos[2]}) < 2.2 && math.Abs(w.Car.Body.Height-o.Pos[1]) < 3 {
			w.Car.Nitro = math.Min(100, w.Car.Nitro+40)
			s.PickupCooldown[id] = 25
		}
	}
	if s.Mode == 1 && s.Countdown > 0 {
		s.Countdown = math.Max(0, s.Countdown-dt)
		in = Input{Brake: 1, Handbrake: true}
		if s.Countdown == 0 {
			s.Started = true
		}
	}
	if s.Finished {
		in.Throttle = 0
		in.Boost = false
		in.Brake = 1
	}
	if s.Started && !s.Finished {
		s.RaceTime += dt
	}
	before := w.Car.Pos
	w.stepRemasteredVehicle(&w.Car, w.CarClass, in, dt, true)
	w.Distance += before.Dist(w.Car.Pos)
	if !w.NetworkClient050 {
		if s.Mode == 1 {
			w.stepRemasteredRace(dt)
		} else {
			w.stepRemasteredTraffic(dt)
		}
	}
	w.resolveActorContacts(dt)
	w.StepPedestrians(dt)
	if s.Mode == 1 {
		w.updatePlayerRace(dt)
		order := w.RacePositions()
		for i, id := range order {
			if id < 0 {
				s.Position = i + 1
			}
		}
	}
}
func (w *World) stepRemasteredVehicle(c *Car, class int, in Input, dt float64, player bool) {
	spec := VehicleVariantSpec050(class, c.Variant050)
	spec.Mu *= TyreGrip050(c.Tyre050, c.Wetness050, c.OffroadBlend)
	p := spec.P
	if !c.Body.Initialized {
		old, oldClass := w.Car, w.CarClass
		w.Car, w.CarClass = *c, class
		w.InitializeChassis()
		*c = w.Car
		w.Car, w.CarClass = old, oldClass
	}
	in.Throttle = Clamp(in.Throttle, 0, 1)
	in.Brake = Clamp(in.Brake, 0, 1)
	in.Steer = Clamp(in.Steer, -1, 1)
	c.EngineRunning = !in.EngineGated || in.EngineEnabled
	if !c.EngineRunning {
		in.Throttle, in.Boost = 0, false
	}
	c.ContactImpulse = math.Max(0, c.ContactImpulse-dt*18)
	// The service pedal uses four-wheel braking with tire slip control. Rear
	// axle locking is an intentional handbrake action, including on a fresh
	// full-pressure pedal press while steering or using the accelerator.
	c.RearLockTimer = 0
	c.BrakeInput = in.Brake
	c.RearLocked = in.Handbrake
	c.DriftActive = c.RearLocked && math.Abs(c.Speed) > 5 && c.Body.Grounded
	if c.DriftActive {
		c.DriftRecovery = .82
	} else {
		c.DriftRecovery = math.Max(0, c.DriftRecovery-dt*1.25)
	}
	c.Burnout = c.Body.Grounded && ((in.Throttle > .35 && in.Handbrake && math.Abs(c.Speed) < 3.5) || (in.Throttle > .7 && in.Brake > .7 && math.Abs(c.Speed) < .8))
	distance, width, _ := w.RemasteredRoadAt(c.Pos)
	c.OnRoad = distance < width*.5+.25
	// The short graded connectors join each tunnel mouth to the existing road.
	for _, t := range w.Remaster.Def.Tunnels {
		side, along := t.Coordinates(c.Pos)
		if math.Abs(side) < t.Width*.5 && math.Abs(along) < t.Length*.5+35 {
			c.OnRoad = true
		}
	}
	if w.Remaster.Mode == 1 {
		for _, i := range w.Remaster.surfaceGrid[remasterMeshCell(c.Pos)] {
			t := w.Remaster.Def.Surfaces[i]
			if !t.Road {
				continue
			}
			if _, ok := triangleHeight(t, c.Pos); ok {
				c.OnRoad = true
				break
			}
		}
	}
	speed := math.Abs(c.Speed)
	// The centre's distance blends contact quality at a verge; a car no longer
	// loses almost half its grip in the single tick it crosses the road edge.
	offroadTarget := Clamp((distance-width*.5+.55)/2.4, 0, 1)
	if c.OnRoad && distance > width*.5+.25 {
		offroadTarget = 0
	} // tunnel / race deck
	if !c.TiresReady {
		c.OffroadBlend = offroadTarget
	} else {
		c.OffroadBlend += (offroadTarget - c.OffroadBlend) * (1 - math.Exp(-dt/0.16))
	}
	grip := Lerp(DrivingGripCoefficient(spec, true, c.RearLocked, c.DriftRecovery), DrivingGripCoefficient(spec, false, c.RearLocked, c.DriftRecovery), c.OffroadBlend)
	steeringSpec := spec
	steeringSpec.Mu = grip
	maxSteer := RacingSteerLimit(steeringSpec, speed, c.DriftActive || c.DriftRecovery > 0)
	desiredSteer := in.Steer * maxSteer
	rate := Lerp(2.0, .65, Clamp(speed/55, 0, 1))
	if desiredSteer*c.VisualSteer < 0 {
		rate *= 1.65
	}
	if math.Abs(in.Steer) < .02 {
		rate = 2.7
	}
	// Relax the large drift steering angle immediately on Space release. Retaining
	// it for the recovery timer previously caused a second snap on regripping.
	if !c.DriftActive {
		desiredSteer = Clamp(desiredSteer, -RacingSteerLimit(steeringSpec, speed, false), RacingSteerLimit(steeringSpec, speed, false))
	}
	if player {
		steerDT := dt
		if !c.DriftActive && c.DriftRecovery == 0 {
			steerDT *= 1.25
		}
		c.VisualSteer = ProgressiveSteer(c.VisualSteer, desiredSteer/math.Max(.001, maxSteer), maxSteer, speed, steerDT)
	} else {
		c.VisualSteer += Clamp(desiredSteer-c.VisualSteer, -rate*dt, rate*dt)
	}
	f := HeadingVec(c.Heading)
	right := Vec{f.Z, -f.X}
	g := &c.Gearbox
	if g.Gear == 0 {
		g.Gear = 1
		g.RPM = p.IdleRPM
	}
	g.ShiftTimer = math.Max(0, g.ShiftTimer-dt)
	if in.Throttle > .05 && c.Speed > -.6 {
		g.Gear = ClampInt(g.Gear, 1, 6)
		c.ReverseHold = 0
	}
	if in.Brake > .05 && in.Throttle == 0 && math.Abs(c.Speed) < .65 && !in.Handbrake {
		c.ReverseHold += dt
		if c.ReverseHold >= .10 {
			g.Gear = -1
			g.ShiftTimer = 0
		}
	} else if in.Brake == 0 {
		c.ReverseHold = 0
	}
	ratio := p.Ratios[ClampInt(g.Gear-1, 0, 5)]
	// Select launch gear from road speed before clutch slip raises engine RPM.
	// Otherwise a stalled uphill car can remain in a high gear indefinitely.
	if g.Gear > 1 && speed < 5 {
		g.Gear = 1
		g.ShiftTimer = 0
		ratio = p.Ratios[0]
	}
	if g.Gear < 0 {
		ratio = p.ReverseRatio
	}
	g.RPM = math.Max(p.IdleRPM, speed/p.WheelRadius*ratio*p.FinalDrive*60/Tau)
	throttle := in.Throttle
	if g.Gear < 0 {
		throttle = in.Brake
	}
	if speed < 2 && throttle > .1 && g.Gear <= 1 {
		g.RPM = math.Max(g.RPM, p.IdleRPM+throttle*1800)
	}
	if g.Gear > 0 && g.ShiftTimer == 0 && in.Brake == 0 {
		if g.RPM > p.RedlineRPM*.91 && g.Gear < 6 {
			g.Gear++
			g.ShiftTimer = .19
		} else if g.RPM < p.RedlineRPM*.40 && g.Gear > 1 {
			g.Gear--
			g.ShiftTimer = .15
		}
	}
	// Re-evaluate the engaged ratio after a shift, before applying wheel torque.
	if g.Gear > 0 {
		ratio = p.Ratios[ClampInt(g.Gear-1, 0, 5)]
	} else {
		ratio = p.ReverseRatio
	}
	g.RPM = math.Max(p.IdleRPM, speed/p.WheelRadius*ratio*p.FinalDrive*60/Tau)
	if speed < 2 && throttle > .1 && g.Gear <= 1 {
		g.RPM = math.Max(g.RPM, p.IdleRPM+throttle*1800)
	}
	engineForce := engineTorque(g.RPM, p.PeakTorque, p.RedlineRPM) * ratio * p.FinalDrive * .88 / p.WheelRadius * throttle
	if !c.EngineRunning {
		engineForce = 0
		g.RPM = 0
	}
	engineForce = math.Min(engineForce, spec.PowerKW*1000*.88/math.Max(7, speed))
	if g.ShiftTimer > 0 {
		engineForce *= .22
	}
	if g.RPM > p.RedlineRPM*1.05 {
		engineForce *= .15
	}
	if c.EngineTemperature > 125 {
		engineForce *= Clamp(1-(c.EngineTemperature-125)*.012, .65, 1)
	}
	direction := 1.
	if g.Gear < 0 {
		direction = -1
	}
	engineForce *= direction
	braking := 0.
	c.CoastingLights = in.Throttle < .05 && speed > 2 && g.Gear > 0
	c.Braking = (g.Gear >= 0 && in.Brake > .05) || (g.Gear < 0 && in.Throttle > .05)
	if (g.Gear >= 0 && in.Brake > 0 && speed > 0) || (g.Gear < 0 && in.Throttle > 0 && speed > 0) {
		pedal := in.Brake
		if g.Gear < 0 {
			pedal = in.Throttle
		}
		// Deliberate simcade brake assist requested by the player. This raises
		// longitudinal service-brake authority, not normal cornering/drift grip.
		braking = math.Min(speed*p.Mass/dt, spec.Mu*p.Mass*9.81*1.12*1.55*1.30) * pedal
		if c.BrakeTemperature > 550 {
			braking *= Clamp(1-(c.BrakeTemperature-550)*.0007, .65, 1)
		}
	}
	c.Boosting = in.Boost && in.Brake == 0 && c.Nitro > .05 && c.Speed > 2 && !c.Burnout
	boostForce := 0.
	if c.Boosting {
		boostForce = NitroDriveForce038(spec, g.RPM, ratio, speed, g.ShiftTimer)
		if !c.Body.Grounded {
			// Preserve the established, bounded stunt air assist. It has no tire
			// contact or clutch and is halved in the airborne branch below.
			boostForce = NitroDriveForce038(spec, g.RPM, ratio, speed, 0)
		}
		c.Nitro = math.Max(0, c.Nitro-dt*18)
	} else if !in.Handbrake && !in.Boost {
		c.Nitro = math.Min(100, c.Nitro+dt*2.0)
	}
	slopeX := w.remasteredSupport(c, c.Pos.Add(Vec{.5, 0})) - w.remasteredSupport(c, c.Pos.Add(Vec{-.5, 0}))
	slopeZ := w.remasteredSupport(c, c.Pos.Add(Vec{0, .5})) - w.remasteredSupport(c, c.Pos.Add(Vec{0, -.5}))
	slopeNormal := math.Sqrt(1 + slopeX*slopeX + slopeZ*slopeZ)
	normal := p.Mass*9.81/slopeNormal + .5*1.225*spec.DownforceArea*speed*speed
	mu := grip
	if braking > 0 {
		// Rear locking must not reduce the service pedal's four-wheel stopping
		// budget. Handbrake-only drift keeps the original reduced rear grip.
		mu = math.Max(mu, Lerp(DrivingGripCoefficient(spec, true, false, 0), DrivingGripCoefficient(spec, false, false, 0), c.OffroadBlend))
	}
	tractionFraction := 1 - spec.FrontWeight
	if spec.AWD {
		tractionFraction = .88
	}
	engineForce = Clamp(engineForce, -normal*mu*tractionFraction*1.15, normal*mu*tractionFraction*1.15)
	resistance := .5 * 1.225 * spec.CdA * c.Speed * speed
	if speed > .08 {
		resistance += math.Copysign(normal*.014, c.Speed)
	}
	if c.OffroadBlend > 0 {
		// Deformable turf dissipates energy through the tires and soil. Keep
		// lateral grip separate: more rolling resistance does not mean asphalt grip.
		soil := 1.0
		if c.Body.Grounded && w.Remaster.Def.Grid.Height(c.Pos) > 55 {
			soil = .84
		}
		resistance += OffroadResistance(p.Mass, normal, c.Speed) * c.OffroadBlend * soil
	}
	if throttle < .03 && speed > .5 && (!in.Handbrake || in.Brake > .05) {
		resistance += math.Copysign(p.Mass*(.12+.16*g.RPM/p.RedlineRPM), c.Speed)
	}
	gradeRaw := Clamp(slopeX*f.X+slopeZ*f.Z, -4, 4)
	grade := gradeRaw / slopeNormal
	longForce := engineForce + boostForce - resistance - math.Copysign(braking, c.Speed) - p.Mass*9.81*grade
	if in.Handbrake && speed > .1 {
		longForce -= math.Copysign(math.Min(speed*p.Mass/dt, normal*(1-spec.FrontWeight)*mu*.38), c.Speed)
	}
	if speed < .25 {
		hold := 0.
		if in.Handbrake {
			hold += normal * (1 - spec.FrontWeight) * mu * .65
		}
		if g.Gear >= 0 {
			hold += normal * mu * .96 * in.Brake
		} else {
			hold += normal * mu * .96 * in.Throttle
		}
		if hold > 0 && math.Abs(longForce) < hold {
			longForce = -c.Speed * p.Mass / dt
		}
	}
	if g.Gear < 0 && c.Speed < -11 {
		longForce += p.Mass * math.Min(6, (-11-c.Speed)*3)
	}
	if c.Speed > spec.TopSpeed {
		longForce -= p.Mass * (c.Speed - spec.TopSpeed) * 2
	}
	acceleration := longForce / p.Mass
	airborne := !c.Body.Grounded
	if airborne {
		velocity := c.Body.AirVelocity
		if velocity.Length() < .01 {
			velocity = f.Mul(c.Speed).Add(right.Mul(c.Lateral))
		}
		boostForce *= .5
		airDrag := .5 * 1.225 * spec.CdA * velocity.Length() / p.Mass
		velocity = velocity.Mul(math.Max(0, 1-airDrag*dt))
		if c.Boosting {
			velocity = velocity.Add(f.Mul(boostForce / p.Mass * math.Cos(c.Rig.Pitch) * dt))
			c.Body.VerticalSpeed -= boostForce / p.Mass * math.Sin(c.Rig.Pitch) * dt
		}
		c.YawRate = Lerp(c.YawRate, c.Body.AirYawRate+in.Steer*.12, 1-math.Exp(-dt*1.8))
		c.Heading += c.YawRate * dt
		f = HeadingVec(c.Heading)
		right = Vec{f.Z, -f.X}
		c.Body.AirVelocity = velocity
		c.Speed = velocity.X*f.X + velocity.Z*f.Z
		c.Lateral = velocity.X*right.X + velocity.Z*right.Z
	} else {
		fx, fy, moment := racingTireStep(c, spec, engineForce+boostForce, braking, normal, mu, dt, in)
		longForce = fx - resistance - p.Mass*9.81*grade
		// Static contact holding must be applied after the tire solve, otherwise
		// the earlier hill-hold force is overwritten by this grounded branch.
		servicePedal := in.Brake
		if g.Gear < 0 {
			servicePedal = in.Throttle
		}
		idleHold := in.Throttle < .01 && in.Brake < .01 && math.Abs(c.Speed) < .18 && math.Abs(c.Lateral) < .12
		parkHold := in.Handbrake && in.Throttle < .01 && math.Abs(c.Speed) < .25
		if math.Abs(grade) < mu*.80 && (idleHold || parkHold || servicePedal > .05 && math.Abs(c.Speed) < .12) {
			longForce = -c.Speed * p.Mass / dt
			fy = -c.Lateral * p.Mass / dt
			moment = -c.YawRate * spec.YawInertia / dt
		}
		if g.Gear < 0 && c.Speed < -11 {
			longForce += p.Mass * math.Min(6, (-11-c.Speed)*3)
		}
		if c.Speed > spec.TopSpeed {
			longForce -= p.Mass * (c.Speed - spec.TopSpeed) * 2
		}
		acceleration = longForce / p.Mass
		u := c.Speed
		c.YawRate = Clamp(c.YawRate+Clamp(moment/spec.YawInertia, -12, 12)*dt, -2.4, 2.4)
		c.Speed, c.Lateral, c.Heading = AdvancePlanarMomentum(c.Speed, c.Lateral, c.Heading, c.YawRate, acceleration, fy/p.Mass, dt)
		if braking > 0 && c.Speed*u < 0 {
			c.Speed = 0
		}
		if math.Abs(u) < .5 && in.Handbrake && in.Throttle == 0 {
			c.Speed *= math.Exp(-dt * 25)
		}
		// Contact-patch forces converge smoothly to rest without low-speed chatter.
		if math.Abs(u) < 1 {
			c.Lateral *= math.Exp(-dt * 12)
			if !c.Burnout || math.Abs(in.Steer) < .15 {
				c.YawRate *= math.Exp(-dt * 10)
			}
		}
		c.YawRate = Clamp(c.YawRate, -2.4, 2.4)
		f = HeadingVec(c.Heading)
		right = Vec{f.Z, -f.X}
		if c.Burnout && (!in.Handbrake || math.Abs(in.Steer) < .15) {
			c.Speed = 0
			c.Lateral = 0
			c.YawRate = 0
		}
		if in.Throttle == 0 && in.Brake == 0 && speed < .10 && math.Abs(grade) < .012 {
			c.Speed = 0
		}
	}
	c.Speed = Clamp(c.Speed, -13, 100)
	c.Lateral = Clamp(c.Lateral, -35, 35)
	c.Pos = c.Pos.Add(f.Mul(c.Speed * dt)).Add(right.Mul(c.Lateral * dt))
	w.stepRemasteredChassis(c, class, in, acceleration, normal-p.Mass*9.81, dt)
	w.remasteredStaticContacts(c, class)
	w.aokawaDynamicContacts(c, class)
	w.remasteredMeshContacts(c, class)
	w.constrainRaceCorridor040(c, class)
	if waterLevel, inWater := w.RemasteredWaterLevel(c.Pos); inWater && !w.InsideTunnel(c.Pos) {
		// Deep water is explicitly represented and rendered. Recover at the last road,
		// never by colliding with an invisible rectangular boundary.
		if c.Body.Height < waterLevel+.15 {
			best := math.Inf(1)
			p, f := w.Routes[0].AtDistance(0)
			for _, route := range w.Routes {
				rd, _ := routeProjection(route, c.Pos, -1)
				point, direction := route.AtDistance(rd)
				distance := point.Dist(c.Pos)
				if distance < best && w.GroundHeight(point) > waterLevel+1 {
					best = distance
					p = point
					f = direction
				}
			}

			c.Pos = p
			c.Heading = math.Atan2(f.X, f.Z)
			c.Speed, c.Lateral = 0, 0
			c.Body.Initialized = false
		}
	}
	c.DriftAngle = math.Atan2(c.Lateral, math.Max(3, math.Abs(c.Speed))) * 180 / math.Pi
	if player && (c.DriftActive || c.DriftRecovery > 0) && math.Abs(c.DriftAngle) > 7 && speed > 7 {
		c.DriftCombo = Clamp(c.DriftCombo+dt*.25, 1, 5)
		c.DriftScore += math.Abs(c.DriftAngle) * speed * dt * .25 * c.DriftCombo
	} else {
		c.DriftCombo = Lerp(c.DriftCombo, 1, 1-math.Exp(-dt*2))
	}
	c.WheelSpin = math.Mod(c.WheelSpin+c.Speed/p.WheelRadius*dt, Tau)
	brakePower := braking * speed
	c.BrakeTemperature = math.Max(25, c.BrakeTemperature+brakePower*dt/22000-(c.BrakeTemperature-25)*(.012+speed*.0014)*dt)
	enginePower := math.Abs(engineForce * c.Speed)
	c.EngineTemperature = Clamp(c.EngineTemperature+enginePower*.66*dt/240000-(c.EngineTemperature-78)*(.03+speed*.0015)*dt, 25, 155)
}
func (w *World) stepRemasteredChassis(c *Car, class int, in Input, longAcceleration, downforce, dt float64) {
	spec := VehicleVariantSpec050(class, c.Variant050)
	p := spec.P
	b := &c.Body
	f := HeadingVec(c.Heading)
	r := Vec{f.Z, -f.X}

	// Suspension roll is relative to the ground bank, rather than world-up.
	// A global 11-degree limit left the uphill tires hanging over side slopes.
	leftSupport := w.remasteredSupport(c, c.Pos.Sub(r.Mul(p.Track*.5)))
	rightSupport := w.remasteredSupport(c, c.Pos.Add(r.Mul(p.Track*.5)))
	terrainRoll := Clamp(math.Atan2((rightSupport-leftSupport)*math.Cos(c.Rig.Pitch), p.Track), -.55, .55)
	ordinaryRoad := w.ordinaryRoad035(c)
	guardSupport := ordinaryRoad || ((in.Handbrake || c.DriftActive || math.Abs(c.Lateral) > 4) && !w.stuntRoad036(c))
	roadGradeSpeed := 0.
	if guardSupport {
		// The lateral velocity matters during a drift. Treating only forward
		// motion as ground velocity made shoulders inject damper energy.
		roadGradeSpeed = (w.remasteredSupport(c, c.Pos.Add(f.Mul(1.5))) - w.remasteredSupport(c, c.Pos.Sub(f.Mul(1.5)))) / 3 * c.Speed
		roadGradeSpeed += (w.remasteredSupport(c, c.Pos.Add(r.Mul(1.5))) - w.remasteredSupport(c, c.Pos.Sub(r.Mul(1.5)))) / 3 * c.Lateral
	}
	force, rollMoment, pitchMoment := 0., 0., 0.
	contacts := 0
	for i := 0; i < 4; i++ {
		x, z := -p.Track*.5, -p.Wheelbase*.5
		if i%2 == 1 {
			x = -x
		}
		fraction := 1 - spec.FrontWeight
		z = -p.Wheelbase * spec.FrontWeight
		if i < 2 {
			z = p.Wheelbase * (1 - spec.FrontWeight)
			fraction = spec.FrontWeight
		}
		static := p.Mass * 9.81 * fraction / (2 * p.SpringRate)
		rest := .74 - p.WheelRadius + static
		cr, sr, cp, sp := math.Cos(c.Rig.Roll), math.Sin(c.Rig.Roll), math.Cos(c.Rig.Pitch), math.Sin(c.Rig.Pitch)
		localX, localY := cr*x-sr*.74, sr*x+cr*.74
		pos := c.Pos.Add(r.Mul(localX)).Add(f.Mul(sp*localY + cp*z))
		ground := w.remasteredSupport(c, pos)
		wheel := &c.Rig.Wheels[i]
		mount := b.Height - .74 + cp*localY - sp*z
		compression := Clamp(ground+p.WheelRadius+rest-mount, 0, .32)
		groundSpeed := Clamp((ground-wheel.GroundHeight)/dt, -18, 18)
		// Tiny deck seams must not be interpreted as an 18 m/s moving floor.
		// Real road grades retain their tangent velocity; stunt ramps are exempt.
		if guardSupport {
			groundSpeed = math.Min(groundSpeed, math.Max(1.2, roadGradeSpeed+1.2))
		}
		mountSpeed := b.VerticalSpeed + cp*(cr*x-sr*.74)*b.RollRate + (-sp*localY-cp*z)*b.PitchRate
		load := 0.
		if compression > 0 {
			load = Clamp(p.SpringRate*compression-p.Damping*(mountSpeed-groundSpeed), 0, p.Mass*9.81*.8)
			contacts++
		}
		wheel.Velocity = (compression - wheel.Compression) / dt
		wheel.Compression, wheel.GroundHeight, wheel.NormalLoad, wheel.Touching = compression, ground, load, compression > 0
		wheel.SteerAngle = 0
		if i < 2 {
			wheel.SteerAngle = AckermannSteer(c.VisualSteer, p.Wheelbase, p.Track, math.Copysign(1, x))
		}
		if !b.Grounded {
			wheel.AngularVelocity *= math.Exp(-dt * .12)
			wheel.RollAngle = math.Mod(wheel.RollAngle+wheel.AngularVelocity*dt, Tau)
		}
		force += load
		rollMoment += x * load
		pitchMoment -= z * load
	}
	wasGrounded := b.Grounded
	b.Grounded = contacts > 0
	if b.Grounded {
		rollMoment += p.Mass*c.Speed*c.YawRate*spec.CGHeight*.24 - (c.Rig.Roll-terrainRoll)*35000 - b.RollRate*11000
		pitchMoment += -p.Mass*Clamp(longAcceleration, -12, 12)*spec.CGHeight*.45 - b.PitchRate*11800
	} else {
		b.AirTime += dt
		b.JumpDistance = c.Pos.Dist(b.LaunchPosition)
		rollMoment = -b.RollRate * 120
		pitchMoment = -b.PitchRate * 160
	}
	if wasGrounded && !b.Grounded {
		b.AirVelocity = f.Mul(c.Speed).Add(r.Mul(c.Lateral))
		b.AirYawRate = c.YawRate
		b.AirTime = 0
		b.LaunchPosition = c.Pos
	}
	if !wasGrounded && b.Grounded {
		b.LandingImpact = math.Max(0, -b.VerticalSpeed)
		c.Damage = Clamp(c.Damage+math.Max(0, b.LandingImpact-12)*.25, 0, 100)
	}
	rollInertia := p.Mass * (p.Track*p.Track + 1.1) / 12
	pitchInertia := p.Mass * (p.Wheelbase*p.Wheelbase + 1.1) / 12
	b.RollRate = Clamp(b.RollRate+rollMoment/rollInertia*dt, -1.3, 1.3)
	b.PitchRate = Clamp(b.PitchRate+pitchMoment/pitchInertia*dt, -1.7, 1.7)
	c.Rig.Roll = Clamp(c.Rig.Roll+b.RollRate*dt, math.Max(-.75, terrainRoll-.19), math.Min(.75, terrainRoll+.19))
	c.Rig.Pitch = Clamp(c.Rig.Pitch+b.PitchRate*dt, -1.05, 1.05)
	if math.Abs(c.Rig.Roll-terrainRoll) >= .19 {
		b.RollRate *= .3
	}
	if !b.Grounded {
		downforce = 0
	}
	b.VerticalSpeed += Clamp((force-downforce)/p.Mass-9.81, -12, 35) * dt
	b.Height += b.VerticalSpeed * dt
	if wasGrounded && b.Grounded && guardSupport {
		b.VerticalSpeed = math.Min(b.VerticalSpeed, math.Max(.75, roadGradeSpeed+.75))
	}
	minimum := w.remasteredSupport(c, c.Pos) + .70
	if b.Height < minimum {
		b.Height = minimum
		// Non-penetration is a velocity constraint as well as a position constraint.
		// Carry the deck's upward tangent velocity into the jump at a ramp lip.
		// The rotated axle spacing is shorter on a slope. Dividing its height
		// difference by the unrotated wheelbase understated downhill velocity,
		// repeatedly turning ground contact into a tiny artificial jump.
		gx := w.remasteredSupport(c, c.Pos.Add(Vec{.5, 0})) - w.remasteredSupport(c, c.Pos.Add(Vec{-.5, 0}))
		gz := w.remasteredSupport(c, c.Pos.Add(Vec{0, .5})) - w.remasteredSupport(c, c.Pos.Add(Vec{0, -.5}))
		velocity := f.Mul(c.Speed).Add(r.Mul(c.Lateral))
		tangentVelocity := gx*velocity.X + gz*velocity.Z
		b.VerticalSpeed = Clamp(tangentVelocity, -25, 25)
		if guardSupport {
			b.VerticalSpeed = math.Min(b.VerticalSpeed, math.Max(.75, roadGradeSpeed+.75))
		}
		b.Grounded = true
	}
	c.Rig.Heave = b.Height - w.remasteredSupport(c, c.Pos) - .74
	c.Rig.HeaveSpeed = b.VerticalSpeed
}
func contactOBB(a Vec, ah Vec, angle float64, b Vec, bh Vec, bangle float64) (Vec, float64, bool) {
	af := HeadingVec(angle)
	ar := Vec{af.Z, -af.X}
	bf := HeadingVec(bangle)
	br := Vec{bf.Z, -bf.X}
	delta := a.Sub(b)
	best := math.Inf(1)
	normal := Vec{}
	dot := func(a, b Vec) float64 { return a.X*b.X + a.Z*b.Z }
	for _, axis := range []Vec{ar, af, br, bf} {
		ra := math.Abs(dot(ar, axis))*ah.X + math.Abs(dot(af, axis))*ah.Z
		rb := math.Abs(dot(br, axis))*bh.X + math.Abs(dot(bf, axis))*bh.Z
		separation := dot(delta, axis)
		overlap := ra + rb - math.Abs(separation)
		if overlap <= 0 {
			return Vec{}, 0, false
		}
		if overlap < best {
			best = overlap
			normal = axis
			if separation < 0 {
				normal = normal.Mul(-1)
			}
		}
	}
	return normal, best, true
}
func contactCircle(c Vec, half Vec, heading float64, p Vec, radius float64) (Vec, float64, bool) {
	f := HeadingVec(heading)
	right := Vec{f.Z, -f.X}
	d := p.Sub(c)
	x, z := d.X*right.X+d.Z*right.Z, d.X*f.X+d.Z*f.Z
	qx, qz := Clamp(x, -half.X, half.X), Clamp(z, -half.Z, half.Z)
	dx, dz := qx-x, qz-z
	dist := math.Hypot(dx, dz)
	if dist >= radius {
		return Vec{}, 0, false
	}
	if dist > 1e-8 {
		return right.Mul(dx / dist).Add(f.Mul(dz / dist)), radius - dist, true
	}
	px, pz := half.X-math.Abs(x), half.Z-math.Abs(z)
	if px < pz {
		return right.Mul(-math.Copysign(1, x)), radius + px, true
	}
	return f.Mul(-math.Copysign(1, z)), radius + pz, true
}
func (w *World) remasteredStaticContacts(c *Car, class int) {
	s := w.Remaster
	spec := VehicleVariantSpec050(class, c.Variant050)
	if len(s.contactMarks) != len(s.Def.Colliders) {
		s.contactMarks = make([]uint32, len(s.Def.Colliders))
	}
	s.contactSerial++
	if s.contactSerial == 0 {
		for i := range s.contactMarks {
			s.contactMarks[i] = 0
		}
		s.contactSerial = 1
	}
	cell := gridOf(c.Pos)
	for x := cell.X - 1; x <= cell.X+1; x++ {
		for z := cell.Z - 1; z <= cell.Z+1; z++ {
			for _, index := range s.staticGrid[gridKey{x, z}] {
				if s.contactMarks[index] == s.contactSerial {
					continue
				}
				s.contactMarks[index] = s.contactSerial
				shape := s.Def.Colliders[index]
				if s.DynamicProps[shape.Object] != nil {
					continue
				}
				if pole := s.FallenPoles[shape.Object]; pole != nil && pole.Falling {
					// A released breakaway pole cannot become a rigid wall on the
					// following physics tick while its fall animation is still upright.
					continue
				}
				if c.Body.Height-.55 > shape.MaxY || c.Body.Height+.7 < shape.MinY {
					continue
				}
				var n Vec
				var penetration float64
				var hit bool
				if shape.Radius > 0 {
					n, penetration, hit = contactCircle(c.Pos, Vec{spec.HalfWidth, spec.HalfLength}, c.Heading, shape.Center, shape.Radius)
				} else {
					var detailed bool
					n, penetration, hit, detailed = w.buildingContact040(c, spec, shape)
					if !detailed {
						n, penetration, hit = contactOBB(c.Pos, Vec{spec.HalfWidth, spec.HalfLength}, c.Heading, shape.Center, shape.Half, shape.Heading)
					}
				}
				if !hit {
					continue
				}
				if shape.Kind == "rock" && shape.MaxY-shape.MinY < .40 && shape.MaxY-w.GroundHeight(shape.Center) < .42 {
					c.Body.VerticalSpeed = math.Max(c.Body.VerticalSpeed, math.Min(.75, .18+math.Abs(c.Speed)*.025))
					c.Speed *= math.Exp(-.65 / 120)
					c.Lateral *= math.Exp(-.65 / 120)
					continue
				}
				if shape.Kind == "barrier" || shape.Kind == "prop" {
					w.hitAokawaBarrier(c, class, index, n, penetration)
					continue
				}
				f := HeadingVec(c.Heading)
				r := Vec{f.Z, -f.X}
				v := f.Mul(c.Speed).Add(r.Mul(c.Lateral))
				closing := v.X*n.X + v.Z*n.Z
				breakaway := shape.Kind == "lamp" || shape.Kind == "signal" || shape.Kind == "sign" || shape.Kind == "road-tree"
				if breakaway && -closing > .025 && s.FallenPoles[shape.Object] == nil {
					pole := &Pole{Pos: shape.Center, Heading: shape.Heading, Length: math.Max(.5, shape.MaxY-shape.MinY)}
					pole.Hit(v)
					s.FallenPoles[shape.Object] = pole
					v = v.Add(n.Mul(-closing * .16))
					c.Speed = v.X*f.X + v.Z*f.Z
					c.Lateral = v.X*r.X + v.Z*r.Z
					continue
				}
				c.Pos = c.Pos.Add(n.Mul(penetration + .002))
				if closing < 0 {
					impulse := -(1.04) * closing
					v = v.Add(n.Mul(impulse))
					// Tangential velocity is preserved, allowing an immediate reverse/turn away.
					c.Speed = v.X*f.X + v.Z*f.Z
					c.Lateral = v.X*r.X + v.Z*r.Z
					c.ContactImpulse = math.Max(c.ContactImpulse, -closing)
					c.Damage = Clamp(c.Damage+math.Max(0, -closing-3)*.25, 0, 100)
				}
			}
		}
	}
}
func (w *World) resolveActorContacts(dt float64) {
	s := w.Remaster
	playerSpec := VehicleVariantSpec050(w.CarClass, w.Car.Variant050)
	contact := func(other *Car, class int, mass float64) {
		spec := RemasteredVehicleSpec(class)
		if s.Mode == 0 {
			// Main visible body footprints, excluding mirrors and ornamental protrusions.
			footprint := TrafficBodyFootprint(class)
			spec.HalfWidth = footprint.X
			spec.HalfLength = footprint.Z
		}
		if w.Car.Body.Height-other.Body.Height > 2 || w.Car.Body.Height-other.Body.Height < -2 {
			return
		}
		n, penetration, hit := contactOBB(w.Car.Pos, Vec{playerSpec.HalfWidth, playerSpec.HalfLength}, w.Car.Heading, other.Pos, Vec{spec.HalfWidth, spec.HalfLength}, other.Heading)
		if !hit {
			return
		}
		m1, m2 := playerSpec.P.Mass, mass
		w.Car.Pos = w.Car.Pos.Add(n.Mul(penetration*m2/(m1+m2) + .002))
		other.Pos = other.Pos.Sub(n.Mul(penetration * m1 / (m1 + m2)))
		f, r := HeadingVec(w.Car.Heading), Vec{math.Cos(w.Car.Heading), -math.Sin(w.Car.Heading)}
		of, orr := HeadingVec(other.Heading), Vec{math.Cos(other.Heading), -math.Sin(other.Heading)}
		v := f.Mul(w.Car.Speed).Add(r.Mul(w.Car.Lateral))
		ov := of.Mul(other.Speed).Add(orr.Mul(other.Lateral))
		closing := v.Sub(ov).X*n.X + v.Sub(ov).Z*n.Z
		if closing >= 0 {
			return
		}
		impulse := -closing * 1.05 / (1/m1 + 1/m2)
		v = v.Add(n.Mul(impulse / m1))
		ov = ov.Sub(n.Mul(impulse / m2))
		w.Car.Speed = v.X*f.X + v.Z*f.Z
		w.Car.Lateral = v.X*r.X + v.Z*r.Z
		other.Speed = ov.X*of.X + ov.Z*of.Z
		other.Lateral = ov.X*orr.X + ov.Z*orr.Z
		other.ContactImpulse = -closing
		arm := other.Pos.Sub(w.Car.Pos).Mul(.5)
		torque := arm.X*n.Z - arm.Z*n.X
		w.Car.YawRate += Clamp(torque*impulse/playerSpec.YawInertia, -.65, .65)
		other.YawRate -= Clamp(torque*impulse/math.Max(500, m2*(spec.HalfLength*spec.HalfLength+spec.HalfWidth*spec.HalfWidth)/3), -.8, .8)
		w.Car.ContactImpulse = math.Max(w.Car.ContactImpulse, -closing)
		_ = dt
	}
	if s.Mode == 1 {
		for i := range s.Racers {
			r := &s.Racers[i]
			contact(&r.Car, r.Class, RemasteredVehicleSpec(r.Class).P.Mass)
		}
	} else {
		for i := range w.Traffic {
			t := &w.Traffic[i]
			if t.Pos.Dist(w.Car.Pos) > TrafficBodyFootprint(t.Variant).Length()+playerSpec.HalfLength+1 {
				continue
			}
			mass := TrafficBodyMass(t.Variant)
			f, r := HeadingVec(t.Heading), Vec{math.Cos(t.Heading), -math.Sin(t.Heading)}
			velocity := f.Mul(t.Speed).Add(t.Impulse)
			other := Car{Pos: t.Pos, Heading: t.Heading, Speed: velocity.X*f.X + velocity.Z*f.Z, Lateral: velocity.X*r.X + velocity.Z*r.Z, Body: ChassisState{Height: w.GroundHeight(t.Pos) + .74}}
			old := other.Pos
			contact(&other, t.Variant, mass)
			t.Offset = t.Offset.Add(other.Pos.Sub(old))
			t.Pos = other.Pos
			if other.ContactImpulse > .01 {
				after := f.Mul(other.Speed).Add(r.Mul(other.Lateral))
				t.Speed = math.Max(0, math.Min(t.Speed, other.Speed))
				t.Impulse = after.Sub(f.Mul(t.Speed))
				t.ImpactYaw += other.YawRate
				t.ImpactCooldown = 1.1
			}
		}
	}
}

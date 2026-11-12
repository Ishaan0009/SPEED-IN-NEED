package engine

import "math"

// TireForces is telemetry in the tire frame, in N, radians and m/s.
type TireForces struct{ Longitudinal, Lateral, SlipAngle, ContactSpeed, NormalLoad, MuX, MuY float64 }

// RacingSteerLimit is shared by the player and AI. Keyboard steering follows
// the available grip envelope instead of demanding a spin at motorway speeds.
func RacingSteerLimit(spec RemasteredVehicle, speed float64, drifting bool) float64 {
	q := Clamp(math.Abs(speed)/65, 0, 1)
	angle := (.070 + (.60-.070)*(1-q)*(1-q))
	aero := 1 + .5*1.225*spec.DownforceArea*speed*speed/(spec.P.Mass*9.81)
	gripAngle := math.Atan(spec.P.Wheelbase * spec.Mu * 9.81 * aero * 1.10 / math.Max(8, speed*speed))
	if drifting {
		angle = math.Min(angle, math.Max(.10, gripAngle*2.35+.028))
	} else {
		// Keep useful rack authority through ordinary road speeds. The old
		// electronic envelope was already closing at 60km/h even though tire
		// forces and the yaw envelope separately enforce available grip.
		// Additional motorway attenuation now fades in after 150km/h.
		high := motorwaySteerBlend040(speed)
		mechanical := Lerp(.60, .070, high)
		angle = math.Min(mechanical, gripAngle*Lerp(1.25, 1, high)+.006*(1-high))
	}
	return angle
}

func motorwaySteerBlend040(speed float64) float64 {
	t := Clamp((math.Abs(speed)-150./3.6)/(120./3.6), 0, 1)
	return t * t * (3 - 2*t)
}

// Four contact patches generate chassis forces and moments. An implicit slip
// compliance term stabilizes wheel inertia near rest; four substeps resolve
// drive/brake torque without snapping wheel rotation to chassis speed.
func racingTireStep(c *Car, spec RemasteredVehicle, driveForce, brakeForce, normal, mu, dt float64, in Input) (float64, float64, float64) {
	p := spec.P
	frontCG, rearCG := p.Wheelbase*(1-spec.FrontWeight), p.Wheelbase*spec.FrontWeight
	u, v, r := c.Speed, c.Lateral, c.YawRate
	loadTotal := 0.0
	for _, wheel := range c.Rig.Wheels {
		if wheel.Touching {
			loadTotal += wheel.NormalLoad
		}
	}
	var loads [4]float64
	for i := 0; i < 4; i++ {
		base := normal * spec.FrontWeight * .5
		if i >= 2 {
			base = normal * (1 - spec.FrontWeight) * .5
		}
		if loadTotal > 1 {
			base = normal * c.Rig.Wheels[i].NormalLoad / loadTotal
		}
		if !c.Rig.Wheels[i].Touching {
			base = 0
		}
		loads[i] = base
	}
	if !c.TiresReady {
		for i := range c.Rig.Wheels {
			c.Rig.Wheels[i].AngularVelocity = u / p.WheelRadius
		}
		c.TiresReady = true
	}
	fxTotal, fyTotal, moment := 0.0, 0.0, 0.0
	cornerReserve := CornerDriveReserve025(mu, u, c.VisualSteer, p.Wheelbase, c.DriftRecovery)
	for sub := 0; sub < 4; sub++ {
		fxSum, fySum, mz := 0.0, 0.0, 0.0
		h := dt / 4
		for i := 0; i < 4; i++ {
			wheel := &c.Rig.Wheels[i]
			side := -p.Track * .5
			if i%2 == 1 {
				side = -side
			}
			along := frontCG
			angle := AckermannSteer(c.VisualSteer, p.Wheelbase, p.Track, math.Copysign(1, side))
			stiffness := spec.FrontStiffness * .5
			if i >= 2 {
				along = -rearCG
				angle = 0
				stiffness = spec.RearStiffness * .5
			}
			if !c.RearLocked && !in.Handbrake {
				gripBlend := GripRecovery024(c.DriftRecovery)
				if i < 2 {
					stiffness *= 1 + .70*gripBlend
				} else {
					stiffness *= 1 + .88*gripBlend
				}
			}
			cs, sn := math.Cos(angle), math.Sin(angle)
			vx, vy := u-r*side, v+r*along
			xw, yw := vx*cs+vy*sn, -vx*sn+vy*cs
			alpha := math.Atan2(yw, math.Max(1, math.Abs(xw)))
			muY := mu
			if (c.RearLocked || in.Handbrake) && i >= 2 {
				muY *= .68
			}
			drive := 0.0
			if spec.AWD {
				drive = driveForce * p.WheelRadius * .25
			} else if i >= 2 {
				drive = driveForce * p.WheelRadius * .5
			}
			if !in.Handbrake && !c.Burnout {
				limit := loads[i] * mu * p.WheelRadius * cornerReserve
				drive = Clamp(drive, -limit, limit)
			}
			brake := brakeForce * p.WheelRadius * .32
			if i >= 2 {
				brake = brakeForce * p.WheelRadius * .18
			}
			inertia := 1.8
			omega := wheel.AngularVelocity + drive/inertia*h
			lock := (in.Handbrake || c.RearLocked) && i >= 2
			if lock {
				omega = 0
			} else {
				omega = math.Copysign(math.Max(0, math.Abs(omega)-brake/inertia*h), omega)
			}
			fx, fy, slip := 0.0, 0.0, 0.0
			if loads[i] > .1 {
				capX, capY := mu*loads[i], muY*loads[i]
				if brakeForce > 0 {
					capX *= 1.55 * 1.30
				}
				if lock {
					capX *= .52
					if brakeForce > 0 {
						capX /= .52
					}
					if !in.Handbrake {
						capX = mu * loads[i] * .90
					}
				}
				den := math.Max(.75, math.Max(math.Abs(xw), math.Abs(p.WheelRadius*omega)))
				slip = (p.WheelRadius*omega - xw) / den
				compliance := 35000 * p.WheelRadius * p.WheelRadius * h / inertia
				if lock {
					compliance = 0
				}
				fx = capX * math.Tanh(35000*(p.WheelRadius*omega-xw)/(den+compliance)/capX)
				// Service brakes act immediately at the contact patch. Wheel-inertia
				// compliance must not delay pedal response or lock the steering axle.
				serviceBrake := brakeForce > 0 && math.Abs(xw) > .03
				if serviceBrake {
					// ABS allocates service braking to the measured contact loads.
					// The old fixed bias wasted rear grip and lengthened stops.
					bias := loads[i] / math.Max(1, normal)
					fx = -math.Copysign(math.Min(brakeForce*bias, capX), xw)
					omega = xw / p.WheelRadius * (1 - .10*math.Abs(fx)/math.Max(1, capX))
					slip = (p.WheelRadius*omega - xw) / math.Max(.75, math.Abs(xw))
				}
				// Implicit patch compliance prevents stiff low-speed tires from
				// alternating between opposite saturated forces each fixed step.
				effective := stiffness / (1 + stiffness*dt*(1/p.Mass+along*along/spec.YawInertia)/math.Max(2, math.Abs(xw)))
				fy = -capY * math.Tanh(effective*alpha/capY)
				// A locked rear patch spends most of its friction budget braking.
				q := math.Hypot(fx/capX, fy/capY)
				if q > 1 {
					fx /= q
					fy /= q
				}
				if !lock && !serviceBrake {
					omega -= p.WheelRadius * fx / inertia * h
				}
			}
			if lock {
				omega = 0
				slip = -math.Copysign(1, xw)
			}
			wheel.AngularVelocity = Clamp(omega, -500, 500)
			wheel.SlipRatio = slip
			wheel.SteerAngle = angle
			wheel.RollAngle = math.Mod(wheel.RollAngle+omega*h, Tau)
			c.TireForces[i] = TireForces{fx, fy, alpha, xw, loads[i], mu, muY}
			bx, by := fx*cs-fy*sn, fx*sn+fy*cs
			fxSum += bx
			fySum += by
			mz += along*by - side*bx
		}
		fxTotal += fxSum * .25
		fyTotal += fySum * .25
		moment += mz * .25
	}
	// Explicit game handling assist: a bounded yaw moment, never a heading edit.
	target := u * math.Tan(c.VisualSteer) / p.Wheelbase
	gripRate := mu * normal / p.Mass / math.Max(4, math.Abs(u))
	target = Clamp(target, -gripRate, gripRate)
	beta := math.Atan2(v, math.Max(3, math.Abs(u)))
	strength := 1.0
	if c.DriftActive {
		strength = .20
	} else if c.DriftRecovery > 0 {
		strength = .25 + .75*GripRecovery024(c.DriftRecovery)
	}
	// A bounded yaw moment stabilizes grip without changing the heading directly.
	gain := Lerp(10.5, 9.3, Clamp(math.Abs(u)/60, 0, 1))
	if c.DriftRecovery > 0 && math.Abs(c.VisualSteer) < .01 {
		gain = 9.5
	}
	assistLimit := p.Mass * 5
	if c.Burnout && in.Handbrake && math.Abs(in.Steer) >= .15 {
		target = in.Steer * in.Throttle * 1.05
		gain, strength, assistLimit = 12, 1, p.Mass*9
	}
	assist := spec.YawInertia * (gain*(target-r) + 3.2*beta) * strength
	moment += Clamp(assist, -assistLimit, assistLimit)
	if !c.RearLocked && !in.Handbrake {
		blend := GripRecovery024(c.DriftRecovery)
		reserve := math.Max(0, mu*normal-math.Abs(fyTotal)) * .60
		fyTotal += Clamp(-v*p.Mass*3.7*blend, -reserve, reserve)
	}
	// Arcade assistance shares the same total contact budget as the tires.
	if c.Burnout && in.Handbrake && math.Abs(in.Steer) >= .15 {
		// Controlled low-speed donut assist uses bounded patch force and yaw
		// moment. The rear axle remains locked; there are no heading edits.
		targetSpeed := 1.7 * in.Throttle
		fxTotal = Clamp((targetSpeed-u)*p.Mass*3, -normal*mu*.28, normal*mu*.28)
	}
	budget := mu * normal
	if brakeForce > 0 {
		budget *= 1.55 * 1.30
	}
	if budget > 1 {
		q := math.Hypot(fxTotal, fyTotal) / budget
		if q > 1 {
			fxTotal /= q
			fyTotal /= q
			moment /= q
		}
	}
	return fxTotal, fyTotal, moment
}

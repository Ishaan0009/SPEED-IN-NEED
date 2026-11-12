package engine

import "math"

// WheelContact is calculated separately for each wheel. It is deliberately
// independent of the graphics representation and can be logged in debug mode.
type WheelContact struct {
	Compression, Velocity, NormalLoad        float64
	GroundHeight, AngularVelocity, SlipRatio float64
	RollAngle, SteerAngle                    float64
	Touching                                 bool
}

// SuspensionRig uses four ground probes and sprung-mass dynamics. Heading/yaw
// still use the existing simcade chassis; this is not yet a six-DoF solver.
type SuspensionRig struct {
	Heave, HeaveSpeed, Roll, Pitch float64
	Wheels                         [4]WheelContact // front-left, front-right, rear-left, rear-right
}

type Transmission struct {
	Gear       int // -1=reverse, 1..6=forward
	RPM        float64
	ShiftTimer float64
}

type VehicleParameters struct {
	Mass, Wheelbase, Track, WheelRadius         float64
	IdleRPM, RedlineRPM, PeakTorque, FinalDrive float64
	Ratios                                      [6]float64
	ReverseRatio                                float64
	SpringRate, Damping                         float64
}

func VehicleSpec(carClass int) VehicleParameters {
	if carClass >= 2 {
		return RemasteredVehicleSpec(carClass).P
	}
	if carClass == 1 { // air-cooled rear-engine 911: lighter, more elastic response
		return VehicleParameters{Mass: 1370, Wheelbase: 2.45, Track: 1.57, WheelRadius: .34, IdleRPM: 900,
			RedlineRPM: 7000, PeakTorque: 495, FinalDrive: 3.52,
			Ratios: [6]float64{3.38, 2.18, 1.53, 1.14, .89, .71}, ReverseRatio: 3.20,
			SpringRate: 40500, Damping: 3500}
	}
	return VehicleParameters{Mass: 1550, Wheelbase: 2.72, Track: 1.67, WheelRadius: .34, IdleRPM: 900,
		RedlineRPM: 7800, PeakTorque: 660, FinalDrive: 3.46,
		Ratios: [6]float64{3.05, 2.10, 1.49, 1.15, .94, .77}, ReverseRatio: 3.0,
		SpringRate: 47000, Damping: 4100}
}

func engineTorque(rpm, peak, redline float64) float64 {
	// Smooth cam/induction torque curve, still useful when low-rpm shifting.
	ratio := Clamp(rpm/redline, 0, 1.3)
	return peak * Clamp(.74+.36*math.Sin(math.Min(ratio, 1)*math.Pi*.82), .45, 1.08)
}

// DriveAcceleration integrates a genuine negative-ratio reverse gear and
// automatic gear selection, without making S reverse a vehicle moving forward.
// Service brakes act on all four tires. Handbrake drag remains independent.
func DriveAcceleration(c *Car, input Input, dt float64, carClass int) float64 {
	params := VehicleSpec(carClass)
	gearbox := &c.Gearbox
	if gearbox.Gear == 0 {
		gearbox.Gear = 1
		gearbox.RPM = params.IdleRPM
	}
	gearbox.ShiftTimer = math.Max(0, gearbox.ShiftTimer-dt)
	if input.Throttle > .05 {
		c.ReverseHold = 0
		if gearbox.Gear < 0 && math.Abs(c.Speed) < .7 {
			gearbox.Gear = 1
			gearbox.ShiftTimer = .20
		}
	}
	if input.Brake > .05 && c.Speed <= .25 && input.Throttle == 0 {
		c.ReverseHold += dt
		if c.ReverseHold > .24 && gearbox.Gear > 0 && gearbox.ShiftTimer == 0 {
			gearbox.Gear = -1
			gearbox.ShiftTimer = .12
		}
	}
	if input.Brake == 0 {
		c.ReverseHold = 0
	}
	ratio := params.Ratios[ClampInt(gearbox.Gear-1, 0, len(params.Ratios)-1)]
	if gearbox.Gear < 0 {
		ratio = params.ReverseRatio
	}
	driveRPM := math.Abs(c.Speed) / params.WheelRadius * ratio * params.FinalDrive * 60 / (2 * math.Pi)
	gearbox.RPM = math.Max(params.IdleRPM, driveRPM)
	if math.Abs(c.Speed) < 2 && input.Throttle > .1 {
		gearbox.RPM = math.Max(gearbox.RPM, params.IdleRPM+input.Throttle*2200)
	}
	if gearbox.Gear > 0 && gearbox.ShiftTimer <= 0 && input.Brake == 0 {
		if gearbox.RPM > params.RedlineRPM*.88 && gearbox.Gear < 6 {
			gearbox.Gear++
			gearbox.ShiftTimer = .22
		}
		if gearbox.RPM < params.RedlineRPM*.27 && gearbox.Gear > 1 {
			gearbox.Gear--
			gearbox.ShiftTimer = .19
		}
	}
	accelerator := Clamp(input.Throttle, 0, 1)
	if gearbox.Gear < 0 {
		accelerator = Clamp(input.Brake, 0, 1)
	}
	force := engineTorque(gearbox.RPM, params.PeakTorque, params.RedlineRPM) * ratio * params.FinalDrive * .88 / params.WheelRadius
	if gearbox.ShiftTimer > .05 {
		force *= .14
	}
	acceleration := force / params.Mass * accelerator
	if gearbox.Gear < 0 {
		acceleration = -acceleration
	}
	// Foot brake while driving forward; a reversed car can brake with W.
	if gearbox.Gear >= 0 && input.Brake > .05 {
		acceleration -= 19.5 * input.Brake
	}
	if gearbox.Gear < 0 && input.Throttle > .05 {
		acceleration += 18 * input.Throttle
	}
	if input.Handbrake {
		if c.Speed > .1 {
			acceleration -= 5.8
		} else if c.Speed < -.1 {
			acceleration += 5.8
		}
	}
	return acceleration
}

// UpdateSuspension performs four independent height samples every 120 Hz tick.
// Heave is spring/damper-integrated and pitch/roll are driven by terrain and
// lateral/longitudinal weight transfer; the rigid planar solver remains active.
func UpdateSuspension(c *Car, routes []Route, carClass int, dt float64) {
	params := VehicleSpec(carClass)
	forward := HeadingVec(c.Heading)
	right := Vec{forward.Z, -forward.X}
	centreHeight := Terrain(c.Pos.X, c.Pos.Z)
	baseY := centreHeight + .75 + c.Rig.Heave
	roadHere := DistRoad(c.Pos, routes) < 7.3
	totalForce := 0.0
	leftHeight, rightHeight, frontHeight, rearHeight := 0., 0., 0., 0.
	for i := 0; i < 4; i++ {
		front := i < 2
		lateral := -1.0
		if i%2 == 1 {
			lateral = 1
		}
		longitudinal := -1.0
		if front {
			longitudinal = 1
		}
		p := c.Pos.Add(forward.Mul(params.Wheelbase * .5 * longitudinal)).Add(right.Mul(params.Track * .5 * lateral))
		ground := Terrain(p.X, p.Z)
		if roadHere {
			ground += .18
		}
		mount := baseY - .25
		springGap := mount - ground - params.WheelRadius
		compression := Clamp(.24-springGap, 0, .32)
		wheel := &c.Rig.Wheels[i]
		vel := (compression - wheel.Compression) / math.Max(dt, 1e-5)
		force := Clamp(params.SpringRate*compression+params.Damping*vel, 0, params.Mass*9.81)
		wheel.Compression, wheel.Velocity, wheel.NormalLoad = compression, vel, force
		wheel.Touching = compression > 0
		wheel.GroundHeight = ground
		// The front wheels follow Ackermann geometry: the inside wheel
		// steers more sharply and travels a shorter arc than the outside.
		wheel.SteerAngle = 0
		if front {
			wheel.SteerAngle = AckermannSteer(c.VisualSteer, params.Wheelbase, params.Track, lateral)
		}
		linearContactSpeed := c.Speed - c.YawRate*lateral*params.Track*.5
		angularSpeed := linearContactSpeed / params.WheelRadius
		if c.Burnout && !front {
			// Rear-wheel smoke needs a visibly spinning tire. Holding the
			// chassis fixed is distinct from stopping the powertrain.
			angularSpeed = 16 + 34*Clamp(c.Gearbox.RPM/params.RedlineRPM, 0, 1)
			wheel.SlipRatio = 2.0
		} else if !wheel.Touching {
			// A wheel in the air retains some rotation instead of snapping
			// to zero, which also prevents visible one-frame judder.
			angularSpeed = wheel.AngularVelocity * math.Exp(-dt*1.1)
			wheel.SlipRatio *= math.Exp(-dt * 4)
		} else {
			wheel.SlipRatio *= math.Exp(-dt * 10)
		}
		wheel.AngularVelocity = angularSpeed
		wheel.RollAngle += angularSpeed * dt
		totalForce += force
		if i%2 == 0 {
			leftHeight += ground
		} else {
			rightHeight += ground
		}
		if front {
			frontHeight += ground
		} else {
			rearHeight += ground
		}
	}
	// Allow gentle body bounce but prevent numerical explosions at coarse terrain.
	verticalAccel := Clamp(totalForce/params.Mass-9.81, -16, 16)
	c.Rig.HeaveSpeed = Clamp(c.Rig.HeaveSpeed+verticalAccel*dt, -2, 2)
	c.Rig.Heave = Clamp(c.Rig.Heave+c.Rig.HeaveSpeed*dt, -.18, .26)
	if c.Rig.Heave <= -.18 && c.Rig.HeaveSpeed < 0 {
		c.Rig.HeaveSpeed = 0
	}
	if c.Rig.Heave >= .26 && c.Rig.HeaveSpeed > 0 {
		c.Rig.HeaveSpeed = 0
	}
	targetRoll := Clamp((rightHeight-leftHeight)/math.Max(params.Track, 1)-c.Speed*c.YawRate*.006, -.16, .16)
	targetPitch := Clamp((frontHeight-rearHeight)/math.Max(params.Wheelbase, 1)-c.Speed*.0018, -.12, .12)
	blend := 1 - math.Exp(-dt*8)
	c.Rig.Roll = Lerp(c.Rig.Roll, targetRoll, blend)
	c.Rig.Pitch = Lerp(c.Rig.Pitch, targetPitch, blend)
}

// AckermannSteer is a finite inner/outer-angle approximation for a steered
// front axle. Side=-1 left, +1 right, in the vehicle's LOCAL coordinates.
func AckermannSteer(centreAngle, wheelbase, track, side float64) float64 {
	if math.Abs(centreAngle) < 1e-7 || wheelbase <= 0 || track <= 0 {
		return centreAngle
	}
	t := math.Tan(centreAngle)
	return math.Atan(t / (1 - side*track*.5/wheelbase*t))
}

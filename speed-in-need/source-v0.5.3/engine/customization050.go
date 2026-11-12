package engine

// VehicleVariant050 keeps model-specific physics independent from UI and files.
func VehicleVariantSpec050(base, variant int) RemasteredVehicle {
	if variant >= 15 && variant <= 18 {
		s := RemasteredVehicleSpec(variant - 15)
		s.Paintable = true
		if variant == 17 {
			s.Name = "MCLAREN P1 GTR"
		}
		return s
	}
	s := RemasteredVehicleSpec(base)
	if variant == 1 || variant == 13 {
		s = RemasteredVehicleSpec(2)
	}
	if variant == 2 || variant == 3 {
		s = RemasteredVehicleSpec(0)
	}
	switch variant {
	case 1:
		s.Name = "M4 WIDEBODY"
		s.P.Mass = 1710
		s.P.Wheelbase = 2.85
		s.P.PeakTorque = 650
		s.PowerKW = 390
		s.TopSpeed = 82
		s.HalfLength = 2.425
		s.HalfWidth = 1.0
		s.AWD = true
		s.CdA = .76
	case 2:
		s.Name = "EB110 SUPER SPORT"
		s.P.Mass = 1560
		s.P.Wheelbase = 2.55
		s.P.PeakTorque = 650
		s.PowerKW = 450
		s.TopSpeed = 97
		s.HalfLength = 2.2
		s.HalfWidth = .97
		s.AWD = true
		s.CdA = .66
	case 3:
		s.Name = "LP670"
		s.P.Mass = 1580
		s.P.Wheelbase = 2.67
		s.P.PeakTorque = 660
		s.PowerKW = 493
		s.TopSpeed = 95
		s.HalfLength = 2.35
		s.HalfWidth = 1.02
		s.AWD = true
		s.CdA = .75
	default:
		if variant >= 4 {
			s = garageVariant051(s, variant)
		}
	}

	if variant == 1 || variant == 2 || variant == 3 || variant == 13 || variant == 14 {
		s.YawInertia = s.P.Mass * (s.HalfLength*s.HalfLength + s.HalfWidth*s.HalfWidth) / 3
		s.FrontStiffness = s.P.Mass * 56
		s.RearStiffness = s.P.Mass * 63
		s.P.SpringRate = s.P.Mass * 28
		s.P.Damping = s.P.Mass * 2.6
		if variant == 1 {
			s.P.Track = 1.62
			s.P.WheelRadius = .335
			s.CGHeight = .45
			s.FrontWeight = .52
			s.Mu = 1.14
			s.DownforceArea = .22
		}
		if variant == 2 {
			s.P.Track = 1.60
			s.P.WheelRadius = .325
			s.CGHeight = .35
			s.FrontWeight = .43
			s.Mu = 1.17
			s.DownforceArea = .40
		}
		if variant == 3 {
			s.P.Track = 1.69
			s.P.WheelRadius = .35
			s.CGHeight = .37
			s.FrontWeight = .43
			s.Mu = 1.18
			s.DownforceArea = .62
		}
		if variant == 13 {
			s.CGHeight = .46
			s.FrontWeight = .55
			s.P.FinalDrive = 3.55
			s.DownforceArea = .20
		}
		if variant == 14 {
			s.CGHeight = .62
			s.FrontWeight = .53
			s.Mu = 1.03
			s.DownforceArea = .02
			s.P.SpringRate = s.P.Mass * 22
			s.P.Damping = s.P.Mass * 2.5
		}
	}
	return s
}
func TyreGrip050(tyre int, wet, offroad float64) float64 {
	dry := []float64{1, 1.10, .96, .88}[ClampInt(tyre, 0, 3)]
	rain := []float64{.78, .62, .96, .84}[ClampInt(tyre, 0, 3)]
	ground := []float64{1, .82, 1.02, 1.30}[ClampInt(tyre, 0, 3)]
	return Lerp(Lerp(dry, rain, Clamp(wet, 0, 1)), ground, Clamp(offroad, 0, 1))
}

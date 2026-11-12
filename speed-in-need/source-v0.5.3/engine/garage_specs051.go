package engine

type garageSpec051 struct {
	Name                                                        string
	Mass, Power, Speed, Wheelbase, Track, Radius, Width, Length float64
	AWD                                                         bool
}

var garageSpecs051 = []garageSpec051{
	{"COMPACT", 1200, 145, 58, 2.9419, 1.6183, 0.3430, 1.1352, 2.3000, false},
	{"COUPE", 1450, 210, 66, 2.6181, 1.4385, 0.3259, 1.0223, 2.3000, false},
	{"HATCHBACK", 1300, 175, 63, 2.7949, 1.6180, 0.3488, 1.0124, 2.3000, false},
	{"OFFROAD", 2050, 225, 59, 2.7099, 1.7165, 0.4176, 1.1443, 2.3000, true},
	{"PICKUP", 2050, 225, 59, 2.8365, 1.2954, 0.3338, 0.9180, 2.3000, true},
	{"SEDAN", 1450, 210, 66, 2.7596, 1.4726, 0.3216, 1.0244, 2.3000, false},
	{"SUV", 2050, 225, 59, 2.7239, 1.4552, 0.3750, 1.0313, 2.3000, true},
	{"WAGON", 1450, 210, 66, 2.7641, 1.4704, 0.3195, 1.0233, 2.3000, false},
	{"RX7 STYLISED", 1260, 206, 70, 2.6500, 1.5500, 0.3300, 1.0511, 2.1400, false},
	{"SKYLINE R34", 1560, 300, 82, 2.68, 1.52, .35, 1.02, 2.325, true},
	{"4X4 JEEP", 1900, 150, 55, 2.6571, 1.79, .4086, 1.19, 2.25, true},
}

func garageVariant051(s RemasteredVehicle, variant int) RemasteredVehicle {
	i := variant - 4
	if i < 0 || i >= len(garageSpecs051) {
		return s
	}
	q := garageSpecs051[i]
	s.Name = q.Name
	s.P.Mass = q.Mass
	s.P.Wheelbase = q.Wheelbase
	s.P.Track = q.Track
	s.P.WheelRadius = q.Radius
	s.PowerKW = q.Power
	s.P.PeakTorque = q.Power * 1.7
	s.TopSpeed = q.Speed
	s.AWD = q.AWD
	s.HalfWidth = q.Width
	s.HalfLength = q.Length
	s.CdA = .76
	s.Mu = 1.12
	s.YawInertia = q.Mass * (q.Length*q.Length + q.Width*q.Width) / 3
	s.P.SpringRate = q.Mass * 30
	s.P.Damping = q.Mass * 2.8
	return s
}

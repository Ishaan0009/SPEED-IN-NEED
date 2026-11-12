package main

import (
	"math"
	"speedinneed/engine"
)

type cabin053 struct{ X, Z, Head, Wheel, Roof float64 }

func cabinProfile053(id, variant int) cabin053 {
	if variant >= 15 && variant <= 18 {
		id = variant - 15
	}
	p := []cabin053{{.305, -.14, .81, .678, 1.10}, {.34, -.10, 1.04, .81, 1.30}, {.34, -.28, .84, .76, 1.13}, {-.36, -.14, 1.01, .764, 1.26}}[engine.ClampInt(id, 0, 3)]
	switch variant {
	case 1:
		p = cabin053{.32, -.12, 1.06, .87, 1.37}
	case 2:
		p = cabin053{.30, .02, .76, .58, 1.07}
	case 3:
		p = cabin053{.32, -.08, .86, .66, 1.167}
	case 13:
		p = cabin053{-.34, -.20, 1.04, .80, 1.355}
	case 14:
		p = cabin053{.40, -.12, 1.49, 1.15, 1.813}
	}
	return p
}
func seatedDriverPose053(id, variant int, steer float64) [engine.PedJointCount]engine.PedVec {
	if variant == 0 {
		return seatedDriverPose032(id, steer)
	}
	base := seatedDriverPose032(0, 0)
	p := cabinProfile053(id, variant)
	dx, dz, dy := p.X-base[2].X, p.Z-base[0].Z, p.Head-base[2].Y
	for i := range base {
		base[i].X += dx
		base[i].Z += dz
		base[i].Y += dy
	}
	for _, j := range []int{5, 8} {
		base[j].Y = p.Wheel
	}
	for _, j := range []int{11, 14} {
		base[j].Y = math.Max(.12, base[j].Y)
	}
	return fitDriverArms035(base, steer)
}
func vehicleLamp053(c engine.Car, id int, front bool, side float64) pos3 {
	key := modelKey050(c, id, false)
	if a := pbrAssets[key]; a != nil {
		tag := "lamp_tail"
		if front {
			tag = "lamp_head"
		}
		if side < 0 {
			tag += "_left"
		} else {
			tag += "_right"
		}
		if p, ok := a.Anchors[tag]; ok {
			return pos3{p[0], p[1], p[2]}
		}
	}
	spec := engine.VehicleVariantSpec050(id, c.Variant050)
	z := spec.HalfLength * .92
	if !front {
		z = -z
	}
	return pos3{side * spec.HalfWidth * .72, .62, z}
}

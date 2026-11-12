//go:build windows

package main

import (
	"math"
	"speedinneed/engine"
	"unsafe"
)

func updateRemasterLights(c engine.Car, view [16]float32) {
	transform := func(p pos3, w float64) pos3 {
		return pos3{float64(view[0])*p.x + float64(view[4])*p.y + float64(view[8])*p.z + float64(view[12])*w, float64(view[1])*p.x + float64(view[5])*p.y + float64(view[9])*p.z + float64(view[13])*w, float64(view[2])*p.x + float64(view[6])*p.y + float64(view[10])*p.z + float64(view[14])*w}
	}
	var positions, directions, colours, parameters [36]float32
	count := 0
	addLight := func(p, d, color pos3, distance, outer, inner float64) {
		if count >= []int{4, 8, 10, 12}[engine.ClampInt(lightingDetail026(), 0, 3)] {
			return
		}
		p = transform(p, 1)
		d = norm(transform(d, 0))
		for j, v := range []pos3{p, d, color, {distance, outer, inner}} {
			arr := []*[36]float32{&positions, &directions, &colours, &parameters}[j]
			arr[count*3] = float32(v.x)
			arr[count*3+1] = float32(v.y)
			arr[count*3+2] = float32(v.z)
		}
		count++
	}
	sceneFloat("uTunnelFloor", -10000)
	if app.world.Remaster != nil && (beta.screen == screenDrive || remasterSceneOverlay()) && !sunShadow.drawing {
		for _, t := range app.world.Remaster.Def.Tunnels {
			sceneVector("uTunnelCenter", pos3{t.Center.X, (t.StartY + t.EndY) * .5, t.Center.Z})
			sceneVector("uTunnelForward", pos3{math.Sin(t.Heading), 0, math.Cos(t.Heading)})
			sceneVector("uTunnelExtent", pos3{t.Width * .5, 7.6, t.Length*.5 - t.LiningInset - 1})
			sceneFloat("uTunnelApron", t.PortalApron)
			sceneFloat("uTunnelClipEnd", t.Length*.5-.02)
			sceneFloat("uTunnelRoadEnd", t.Length*.5+t.PortalApron-.1)
			sceneFloat("uTunnelSlope", (t.EndY-t.StartY)/t.Length)
			sceneFloat("uTunnelFloor", (t.StartY+t.EndY)*.5)
		}
		spec := engine.VehicleVariantSpec050(app.world.CarClass, c.Variant050)
		f := engine.HeadingVec(c.Heading)
		lampEnergy := .18 + .82*engine.NightAmount(sceneHour())
		if app.world.InsideTunnel(c.Pos) {
			lampEnergy = 1
		}
		if beta.cfg.Headlights {
			for _, side := range []float64{-1, 1} {
				p3 := aokawaToWorld(c, vehicleLamp053(c, app.world.CarClass, true, side))
				p := engine.Vec{X: p3.x, Z: p3.z}
				addLight(pos3{p.X, p3.y, p.Z}, pos3{f.X * math.Cos(c.Rig.Pitch), -math.Sin(c.Rig.Pitch) - .065, f.Z * math.Cos(c.Rig.Pitch)}, scale(pos3{.92, .97, 1}, lampEnergy*1.6), 145, .86, .986)
			}
		}
		if c.Speed < -.1 || c.Gearbox.Gear < 0 {
			p := c.Pos.Sub(f.Mul(spec.HalfLength))
			addLight(pos3{p.X, c.Body.Height - .15, p.Z}, pos3{-f.X, -.32, -f.Z}, scale(pos3{1, .98, .91}, .018+.065*engine.NightAmount(sceneHour())), 8, .66, .94)
		}
		if c.Braking || c.CoastingLights {
			p := c.Pos.Sub(f.Mul(spec.HalfLength))
			addLight(pos3{p.X, c.Body.Height - .18, p.Z}, pos3{-f.X, -.2, -f.Z}, scale(pos3{.32, .004, .001}, lampEnergy), 10, .4, .9)
		}
		if engine.AutomaticLights(sceneHour()) || app.world.InsideTunnel(c.Pos) {
			type candidate struct {
				p                          engine.Vec
				angle, y, length, distance float64
			}
			near := [4]candidate{}
			for i := range near {
				near[i].distance = 1e9
			}
			insert := func(v candidate) {
				for i := range near {
					if v.distance < near[i].distance {
						for j := len(near) - 1; j > i; j-- {
							near[j] = near[j-1]
						}
						near[i] = v
						break
					}
				}
			}
			for _, t := range app.world.Traffic {
				d := t.Pos.Dist(c.Pos)
				if d > 105 {
					continue
				}
				length, y := 2.1, .66
				if t.Variant == 0 {
					length = 5.25
					y = 1.3
				}
				if t.Variant == 3 {
					length = 3.3
					y = 1.0
				}
				insert(candidate{t.Pos, t.Heading, app.world.GroundHeight(t.Pos) + y, length, d})
			}
			for _, t := range app.world.Remaster.Racers {
				d := t.Car.Pos.Dist(c.Pos)
				if d < 105 {
					insert(candidate{t.Car.Pos, t.Car.Heading, t.Car.Body.Height - .13, engine.RemasteredVehicleSpec(t.Class).HalfLength * .94, d})
				}
			}
			for _, v := range near {
				if v.distance > 105 {
					continue
				}
				f := engine.HeadingVec(v.angle)
				r := engine.Vec{X: f.Z, Z: -f.X}
				for _, side := range []float64{-1, 1} {
					p := v.p.Add(f.Mul(v.length)).Add(r.Mul(side * .72))
					addLight(pos3{p.X, v.y, p.Z}, pos3{f.X, -.065, f.Z}, pos3{1.5, 1.48, 1.38}, 105, .85, .983)
				}
			}
		}
	}
	for i, name := range []string{"uLightPos[0]", "uLightDir[0]", "uLightColor[0]", "uLightParams[0]"} {
		a := []*[36]float32{&positions, &directions, &colours, &parameters}[i]
		if loc := sceneLocation(name); loc >= 0 {
			glExt("glUniform3fv", uintptr(loc), 12, uintptr(unsafe.Pointer(&a[0])))
		}
	}
	sceneInt("uVehicleLightCount", count)
	sceneInt("uShadowFilter", beta.cfg.ShadowFilter)
	sceneInt("uStreetCount", 0)
	sceneVector("uStreetDir", transform(pos3{0, -1, 0}, 0))
	bindLightfield052()
}

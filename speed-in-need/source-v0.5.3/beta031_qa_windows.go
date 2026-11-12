package main

import (
	"math"
	"speedinneed/engine"
)

var qa031 struct {
	PedCount                                                       int
	SawWalker, SawRagdoll, SawNPCPush, SawRearLock, SawRearRelease bool
	SawStop                                                        bool
	InitialNPC                                                     engine.Vec
}

func beta031QAStage() {
	if qa.stage < 105 {
		return
	}
	w := app.world
	qa031.PedCount = len(w.Pedestrians)
	qa031.SawWalker = false
	qa031.SawRagdoll = false
	qa031.SawNPCPush = false
	qa031.SawRearLock = false
	qa031.SawRearRelease = false
	qa031.SawStop = false
	t := w.Remaster.Def.Tunnels[0]
	f := engine.HeadingVec(t.Heading)
	r := engine.Vec{X: f.Z, Z: -f.X}
	switch qa.stage {
	case 105, 106, 111:
		best, index := math.Inf(1), -1
		for i, p := range w.Pedestrians {
			if d := p.Pos.Dist(w.Remaster.Def.Spawn); d < best {
				best, index = d, i
			}
		}
		if index < 0 {
			qa.checks["031-sidewalk-population"] = false
			return
		}
		p := &w.Pedestrians[index]
		pf := engine.HeadingVec(p.Heading)
		if qa.stage == 106 {
			remasterQAPlace(p.Pos.Sub(pf.Mul(5.8)), p.Heading)
			w.Car.Speed = 12
		} else {
			_, _, route := w.RemasteredRoadAt(p.Pos)
			at, _ := w.Routes[route].ClosestDistance(p.Pos, 0)
			center, _ := w.Routes[route].AtDistance(at)
			remasterQAPlace(center.Sub(pf.Mul(8)), p.Heading)
		}
	case 107:
		ri := 0
		for i, road := range w.Remaster.Def.Roads {
			if road.Kind == "tunnel" {
				ri = i
				break
			}
		}
		route := w.Routes[ri]
		distance := route.Len * .30
		at, forward := route.AtDistance(distance)
		heading := math.Atan2(forward.X, forward.Z)
		remasterQAPlace(at, heading)
		w.Car.Speed = 18
		w.Traffic = []engine.Traffic{{Path: ri, T: distance + 4, Pos: at.Add(forward.Mul(4)), Heading: heading + math.Pi, Variant: 1, Speed: 0, Direction: -1, ImpactCooldown: 5}}
		qa031.InitialNPC = w.Traffic[0].Pos
	case 108:
		remasterQAPlace(t.Center.Sub(f.Mul(180)), t.Heading)
		w.Car.Speed = 35
	case 109:
		remasterQAPlace(t.Center.Add(r.Mul(66)).Add(f.Mul(-90)), t.Heading+math.Pi*.35)
		beta.cfg.CameraMode = 2
	case 110:
		remasterQAPlace(t.Center.Sub(f.Mul(t.Length*.5+22)), t.Heading)
	}
	qa.checks["031-sidewalk-population"] = qa031.PedCount > 50
	app.prevCar = w.Car
	app.renderCar = w.Car
	remaster.CameraReset = true
}
func beta031QAInput() engine.Input {
	w := app.world
	if qa.stage == 106 {
		for _, p := range w.Pedestrians {
			if p.Ragdoll {
				return engine.Input{Brake: .8}
			}
		}
		return engine.Input{}
	}
	if qa.stage == 107 {
		return engine.Input{}
	}
	if qa.stage == 108 {
		if qa031.SawStop {
			return engine.Input{}
		}
		if w.Elapsed < .2 {
			return engine.Input{}
		}
		if w.Elapsed < .4 {
			return engine.Input{Brake: 1, Handbrake: true}
		}
		return engine.Input{Brake: 1, Steer: .2}
	}
	return engine.Input{Brake: 1, Handbrake: true}
}
func beta031QAObserve() {
	if qa.stage < 105 {
		return
	}
	w := app.world
	qa031.SawWalker = qa031.SawWalker || pedestrianMesh.Drawn > 0
	qa.checks["032-four-imported-character-models"] = characters032.Models[0] != nil && characters032.Models[1] != nil && characters032.Models[2] != nil && characters032.Models[3] != nil
	if walker051.Active && beta.screen == screenDrive {
		qa.checks["051-on-foot-player-rendered"] = characters032.Models[4] != nil && len(characters032.Batches[4].Indices) > 0
	} else if !app.boolFPP && beta.screen == screenDrive {
		qa.checks["032-seated-player-driver-rendered"] = characters032.DriverDrawn
	}
	for _, p := range w.Pedestrians {
		qa031.SawRagdoll = qa031.SawRagdoll || p.Ragdoll
	}
	qa031.SawRearLock = qa031.SawRearLock || w.Car.RearLocked
	if qa031.SawRearLock && !w.Car.RearLocked {
		qa031.SawRearRelease = true
	}
	if qa.stage == 105 || qa.stage == 111 {
		qa.checks["031-visible-walking-pedestrians"] = qa031.SawWalker
		qa.checks["032-textured-pedestrian-skins-rendered"] = characters032.PedestriansDrawn > 0
	}
	if qa.stage == 106 {
		qa.checks["031-rendered-pedestrian-impact"] = qa031.SawRagdoll && qa031.SawWalker
	}
	if qa.stage == 107 {
		for _, v := range w.Traffic {
			qa031.SawNPCPush = qa031.SawNPCPush || v.Impulse.Length() > 1
		}
		qa.checks["031-npc-receives-impact-velocity"] = qa031.SawNPCPush
	}
	if qa.stage == 108 {
		qa031.SawStop = qa031.SawStop || w.Elapsed > .2 && math.Abs(w.Car.Speed) < 1
		qa.checks["031-service-brakes-stop-vehicle"] = qa031.SawStop
		qa.checks["031-rear-lock-releases"] = qa031.SawRearLock && qa031.SawRearRelease
	}
}

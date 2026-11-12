//go:build windows

package main

import (
	"math"
	"speedinneed/engine"
	"unsafe"
)

func activateSelectedMap(restart bool) {
	if remaster.Enabled {
		activateRemasteredWorld(restart)
		return
	}
	id := engine.ClampInt(beta.mapSelect, 0, 2)
	changed := app.world == nil || app.world.Map != id
	app.mapID = id
	if changed || restart {
		app.world = engine.NewWorld(id)
		app.world.CarClass = engine.ClampInt(beta.carSelect, 0, 1)
		app.world.DriveMode = engine.ClampInt(beta.cfg.DriveMode, 0, 2)
		app.world.InitializeChassis()
		app.prevCar = app.world.Car
		app.renderCar = app.world.Car
		app.renderAlpha = 1
		app.mouseSteer, app.camYaw, app.camPitch = 0, 0, 0
		app.mouseInit = false
		if changed {
			invalidateWorldGeometry()
		}
	}
}

func resetDriveCamera() {
	c := app.world.Car
	f := engine.HeadingVec(c.Heading)
	beta.cameraHeading = c.Heading
	app.cam = pos3{c.Pos.X - f.X*app.camDist, c.Body.Height + 3, c.Pos.Z - f.Z*app.camDist}
	app.prevCar = c
	app.renderCar = c
}

func drawMapPreview() {
	activateSelectedMap(false)
	if app.staticList == 0 && !gpuWorld.ready {
		buildStatic()
	}
	centre, _ := app.world.Routes[0].AtDistance(90)
	target := pos3{centre.X, engine.Terrain(centre.X, centre.Z) + 1, centre.Z}
	eye := pos3{centre.X + 82, target.y + 72, centre.Z - 108}
	app.renderCar = app.world.Car
	beginHDRScene()
	perspective(62)
	gl("glViewport", 0, 0, uintptr(app.width), uintptr(app.height))
	gl("glClear", GL_COLOR_BUFFER_BIT|GL_DEPTH_BUFFER_BIT)
	gl("glLoadIdentity")
	applyCamera(eye, target)
	useLighting(false)
	drawAtmosphere(eye)
	dis(GL_FOG)
	useLighting(app.lightingOn)
	updateSceneLights(app.world.Car)
	if !drawBufferedWorld(eye, target) {
		gl("glCallList", uintptr(app.staticList))
		renderVisibleWorldChunks(eye, target)
	}
	drawPBRBuildings()
	renderPoles()
	useLighting(false)
	finishHDRScene()
	drawBetaMenuUI()
}

func renderRamps() {
	en(GL_TEXTURE_2D)
	bind("asphalt.jpg")
	for _, r := range app.world.Ramps {
		const steps = 30
		for i := 0; i < steps; i++ {
			alongA, alongB := r.Length*float64(i)/steps, r.Length*float64(i+1)/steps
			a, fa := r.Sample(alongA)
			b, fb := r.Sample(alongB)
			right := engine.Vec{X: fa.Z, Z: -fa.X}
			rightB := engine.Vec{X: fb.Z, Z: -fb.X}
			pa, pb := a.Add(right.Mul(-r.Width*.5)), a.Add(right.Mul(r.Width*.5))
			pc, pd := b.Add(rightB.Mul(r.Width*.5)), b.Add(rightB.Mul(-r.Width*.5))
			ya := r.DeckHeight(pa, alongA)
			yb := r.DeckHeight(pb, alongA)
			yc := r.DeckHeight(pc, alongB)
			yd := r.DeckHeight(pd, alongB)
			col(.70, .75, .80, 1)
			begin(GL_QUADS)
			tq(pos3{pa.X, ya, pa.Z}, pos3{pb.X, yb, pb.Z}, pos3{pc.X, yc, pc.Z}, pos3{pd.X, yd, pd.Z}, 0, float64(i)*.4, 2, float64(i+1)*.4)
			end()
			// Bright edge stripes and side walls make approach and drop readable.
			dis(GL_TEXTURE_2D)
			for _, side := range []float64{-1, 1} {
				p := a.Add(right.Mul(side * (r.Width*.5 - .22)))
				q := b.Add(rightB.Mul(side * (r.Width*.5 - .22)))
				py := r.DeckHeight(p, alongA)
				qy := r.DeckHeight(q, alongB)
				if i%2 == 0 {
					col(1, .68, .08, 1)
				} else {
					col(.09, .12, .15, 1)
				}
				begin(GL_QUADS)
				tq(pos3{p.X - right.X*.15, py + .025, p.Z - right.Z*.15}, pos3{p.X + right.X*.15, py + .025, p.Z + right.Z*.15}, pos3{q.X + right.X*.15, qy + .025, q.Z + right.Z*.15}, pos3{q.X - right.X*.15, qy + .025, q.Z - right.Z*.15}, 0, 0, 1, 1)
				col(.18, .23, .28, 1)
				tq(pos3{p.X, engine.Terrain(p.X, p.Z), p.Z}, pos3{q.X, engine.Terrain(q.X, q.Z), q.Z}, pos3{q.X, qy, q.Z}, pos3{p.X, py, p.Z}, 0, 0, 1, 1)
				end()
			}
			en(GL_TEXTURE_2D)
			bind("asphalt.jpg")
		}
	}
}

func drawPlayerLamps(id int) {
	active := lightPipeline.active
	useLighting(false)
	dis(GL_TEXTURE_2D)
	z := 2.30
	if id == 1 {
		z = 2.08
	}
	for _, side := range []float64{-1, 1} {
		if app.renderCar.Speed < -.1 || app.renderCar.Gearbox.Gear < 0 {
			col(1, .98, .92, 1)
			box(side*.55-.08, .70, -2.045, side*.55+.08, .77, -2.015)
		}
	}
	if app.renderCar.Boosting {
		for _, side := range []float64{-1, 1} {
			col(.18, .62, 1, .86)
			box(side*.42-.07, .20, -z-.65, side*.42+.07, .31, -z)
		}
	}
	if active {
		useLighting(true)
		setVehicleGloss(true)
	}
}

func uiFontScale() float64 {
	return math.Max(.55, math.Min(float64(app.width)/1280, float64(app.height)/720)) * engine.Clamp(beta.cfg.UIScale, 1, 1.5)
}

func releaseSceneResources() {
	releaseTrafficVisual040()
	releaseBeta020Sky()
	releasePBRAssets()
	releaseBetaRenderTargets()
	for _, id := range app.tex {
		if id != 0 {
			gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&id)))
		}
	}
	if app.whiteTexture != 0 {
		gl("glDeleteTextures", 1, uintptr(unsafe.Pointer(&app.whiteTexture)))
	}
	for _, id := range beta.mesh {
		if id != 0 {
			gl("glDeleteLists", uintptr(id), 1)
		}
	}
	for _, id := range beta.trafficMesh {
		if id != 0 {
			gl("glDeleteLists", uintptr(id), 1)
		}
	}
	for _, car := range beta.paintMesh {
		for _, id := range car {
			if id != 0 {
				gl("glDeleteLists", uintptr(id), 1)
			}
		}
	}
	for _, car := range beta.wheelMesh {
		for _, id := range car {
			if id != 0 {
				gl("glDeleteLists", uintptr(id), 1)
			}
		}
	}
	if lightPipeline.program != 0 {
		glExt("glDeleteProgram", uintptr(lightPipeline.program))
	}
}

func renderVehicleLampMeshes() {
	if sunShadow.drawing {
		return
	}
	active := lightPipeline.active
	useLighting(false)
	dis(GL_TEXTURE_2D)
	for _, t := range app.world.Traffic {
		if t.Pos.Dist(app.renderCar.Pos) > 200 {
			continue
		}
		push()
		move(t.Pos.X, engine.Terrain(t.Pos.X, t.Pos.Z)+.25, t.Pos.Z)
		rotate(t.Heading*180/math.Pi, 0, 1, 0)
		z, y := 2.15, .56
		if t.Variant%4 == 3 {
			z = 2.8
			y = 1.1
		}
		for _, side := range []float64{-1, 1} {
			if engine.AutomaticLights(sceneHour()) {
				col(1, .93, .77, 1)
			} else {
				col(.25, .27, .28, 1)
			}
			box(side*.78-.14, y, z, side*.78+.14, y+.10, z+.03)
			if t.Speed < t.CruiseSpeed*.65 {
				col(1, .025, .008, 1)
			} else if engine.AutomaticLights(sceneHour()) {
				col(.65, .015, .006, 1)
			} else {
				col(.16, .02, .01, 1)
			}
			box(side*.78-.14, y, -z-.03, side*.78+.14, y+.10, -z)
		}
		pop()
	}
	if engine.AutomaticLights(sceneHour()) {
		for _, p := range app.world.Poles {
			if p.Falling {
				continue
			}
			head, y := p.LampHead()
			push()
			move(head.X, y-.12, head.Z)
			col(1, .83, .5, 1)
			box(-.30, -.025, -.20, .30, .025, .20)
			pop()
		}
	}
	if active {
		useLighting(true)
	}
}

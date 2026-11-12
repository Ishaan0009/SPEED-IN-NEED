package main

import (
	"math"
	"speedinneed/engine"
)

// A looping road diorama keeps title motion independent of saved gameplay.
// Meshes and materials are uploaded once during graphics initialization.
func drawBeta012RoadTitle(dt float64) {
	remaster.Idle += dt
	if remaster.Idle > 2.8 && !remaster.Drag {
		remaster.Orbit *= math.Exp(-dt * .55)
		remaster.Pitch *= math.Exp(-dt * .55)
	}
	beginHDRScene()
	perspective(43)
	gl("glViewport", 0, 0, uintptr(hdrRenderWidth()), uintptr(hdrRenderHeight()))
	gl("glClearColor", b32(.12), b32(.23), b32(.35), b32(1))
	gl("glClear", GL_COLOR_BUFFER_BIT|GL_DEPTH_BUFFER_BIT)
	eye, target := beta014PreviewCamera(dt)
	applyCamera(eye, target)
	app.cam = eye
	drawRemasterSky(eye)
	useLighting(true)
	updateSceneLights(engine.Car{})
	dis(GL_FOG)
	sceneInt("uFogEnabled", 0)
	sceneInt("uShadowEnabled", 0)
	// Repeat a 100 m section while the tires turn at the same road speed.
	travel := math.Mod(beta.menuTime*18, 100)
	push()
	move(0, 0, -travel)
	drawPBRAsset("b012_title_road", 0)
	pop()
	for i := 0; i < 22; i++ {
		z := float64(i)*15 - 70 - travel
		if z < -70 {
			z += 330
		}
		for _, side := range []float64{-1, 1} {
			push()
			move(side*(11+math.Mod(float64(i)*.618, 3)), 0, z)
			resize(.9, .9, .9)
			drawPBRAsset("b012_avenue_tree", 0)
			pop()
			if i%3 == 0 {
				push()
				move(side*19, 0, z+3)
				rotate(-side*90, 0, 1, 0)
				drawPBRAsset("b012_building00", 0)
				pop()
			}
		}
	}
	c := engine.Car{Speed: 18, Heading: 0}
	c.Body.Height = .74
	for i := range c.Rig.Wheels {
		c.Rig.Wheels[i].RollAngle = beta.menuTime * 18 / .34
	}
	// Soft contact occlusion grounds the presentation car on the road.
	useLighting(false)
	dis(GL_TEXTURE_2D)
	gl("glDepthMask", 0)
	begin(GL_TRIANGLES)
	for i := 0; i < 48; i++ {
		a, b := float64(i)*engine.Tau/48, float64(i+1)*engine.Tau/48
		col(.008, .012, .018, .46)
		vtx(0, .014, 0)
		col(.008, .012, .018, 0)
		vtx(math.Cos(a)*1.5, .014, math.Sin(a)*2.7)
		vtx(math.Cos(b)*1.5, .014, math.Sin(b)*2.7)
	}
	end()
	gl("glDepthMask", 1)
	useLighting(true)
	// The title uses the selected real car, not a detached source showroom scene.
	drawRemasterCar(c, beta.carSelect, true)
	useLighting(false)
	finishHDRScene()
}

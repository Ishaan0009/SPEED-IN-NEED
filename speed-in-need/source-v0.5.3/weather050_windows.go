package main

import (
	"math"
	"speedinneed/engine"
	"unsafe"
)

var weather050 struct {
	Rain   bool
	Amount float64
}

// Rain is a shared room setting. Geometry stays depth-tested, including tunnels.
func drawWeather050(dt float64) {
	target := 0.
	if weather050.Rain {
		target = 1
	}
	weather050.Amount += (target - weather050.Amount) * (1 - math.Exp(-engine.Clamp(dt, 0, .1)*1.5))
	sceneFloat("uWetness050", weather050.Amount)
	if app.world != nil && app.world.Remaster != nil {
		for i := range app.world.Remaster.Racers {
			c := &app.world.Remaster.Racers[i].Car
			c.Wetness050 = weather050.Amount
			if app.world.Remaster.RaceDifficulty040 == 2 {
				c.Tyre050 = 1
				if weather050.Amount > .35 {
					c.Tyre050 = 2
				}
			}
		}
	}
	if weather050.Amount < .01 || app.world == nil || (workshop051.Active || app.world.InsideTunnel(app.world.Car.Pos)) {
		return
	}
	useLighting(false)
	en(GL_BLEND)
	gl("glBlendFunc", GL_SRC_ALPHA, GL_ONE_MINUS_SRC_ALPHA)
	dis(GL_TEXTURE_2D)
	gl("glDepthMask", 0)
	en(GL_TEXTURE_2D)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(pbrTexture("b052_rain_streak.png")))
	count := []int{700, 1500, 2400, 3500}[engine.ClampInt(app.quality, 0, 3)]
	time := app.t
	var view [16]float32
	gl("glGetFloatv", 0x0BA6, uintptr(unsafe.Pointer(&view[0])))
	right := pos3{float64(view[0]), float64(view[4]), float64(view[8])}
	drawRainStreaks053(count, time, right)
	dis(GL_TEXTURE_2D)
	// Small ripples stay on nearby ground; no screen-space rain behind the HUD.
	if app.quality > 0 {
		for i := 0; i < 28; i++ {
			a := float64(i)
			phase := math.Mod(time*1.8+a*.319, 1)
			p := engine.Vec{X: app.world.Car.Pos.X + math.Sin(a*7.13)*18, Z: app.world.Car.Pos.Z + math.Cos(a*11.21)*18}
			y := app.world.GroundHeight(p) + .035
			col(.55, .68, .75, (1-phase)*.18*weather050.Amount)
			begin(GL_LINE_STRIP)
			for j := 0; j <= 12; j++ {
				angle := float64(j) * engine.Tau / 12
				vtx(p.X+math.Cos(angle)*phase*.32, y, p.Z+math.Sin(angle)*phase*.32)
			}
			end()
		}
	}
	gl("glDepthMask", 1)
	dis(GL_BLEND)
	useLighting(true)
}

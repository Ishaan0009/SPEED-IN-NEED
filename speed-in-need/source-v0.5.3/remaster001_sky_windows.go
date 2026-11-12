//go:build windows

package main

import (
	"math"
	"speedinneed/engine"
	"unsafe"
)

var remasterSky struct {
	VBO, Stars            uint32
	Count, StarCount, Key int
	HDR                   bool
	Clouds                uint32
	CloudCount            int
}

func releaseRemasterSky() {
	for _, p := range []*uint32{&remasterSky.VBO, &remasterSky.Stars, &remasterSky.Clouds} {
		if *p != 0 {
			glExt("glDeleteBuffers", 1, uintptr(unsafe.Pointer(p)))
			*p = 0
		}
	}
	remasterSky.Key = -1
}
func buildRemasterSky() {
	releaseRemasterSky()
	remasterSky.Key = remaster.Time
	remasterSky.HDR = hdrScene.active
	night := engine.NightAmount(sceneHour())
	sunset := engine.Clamp((sceneHour()-15)/3.5, 0, 1) * (1 - night)
	zenith, horizon := pos3{.16, .39, .72}, pos3{.70, .81, .89}
	zenith = add(scale(zenith, 1-sunset), scale(pos3{.09, .17, .32}, sunset))
	horizon = add(scale(horizon, 1-sunset), scale(pos3{.91, .46, .22}, sunset))
	zenith = add(scale(zenith, 1-night), scale(pos3{.026, .038, .088}, night))
	horizon = add(scale(horizon, 1-night), scale(pos3{.12, .16, .24}, night))
	var vertices [][7]float32
	vertex := func(p, c pos3, alpha float64) {
		if hdrScene.active {
			c = pos3{math.Pow(c.x, 2.2), math.Pow(c.y, 2.2), math.Pow(c.z, 2.2)}
		}
		vertices = append(vertices, [7]float32{float32(p.x), float32(p.y), float32(p.z), float32(c.x), float32(c.y), float32(c.z), float32(alpha)})
	}
	point := func(el, az float64) (pos3, pos3) {
		t := math.Pow(engine.Clamp(math.Sin(el), 0, 1), .42)
		c := add(scale(horizon, 1-t), scale(zenith, t))
		direction := pos3{math.Cos(el) * math.Cos(az), math.Sin(el), math.Cos(el) * math.Sin(az)}
		light := beta011CelestialDirection()
		cosine := engine.Clamp(direction.x*light.x+direction.y*light.y+direction.z*light.z, 0, 1)
		halo := math.Pow(cosine, 18)*(.12+.25*sunset)*(1-night) + math.Pow(cosine, 100)*.10
		c = add(c, scale(pos3{1, .71, .40}, halo))
		r := 2200 * math.Cos(el)
		return pos3{r * math.Cos(az), 2200 * math.Sin(el), r * math.Sin(az)}, c
	}
	for j := 0; j < 20; j++ {
		a, b := -.18+float64(j)/20*(math.Pi*.5+.18), -.18+float64(j+1)/20*(math.Pi*.5+.18)
		for k := 0; k < 96; k++ {
			u, v := float64(k)*engine.Tau/96, float64(k+1)*engine.Tau/96
			for _, q := range [][2]float64{{a, u}, {a, v}, {b, v}, {a, u}, {b, v}, {b, u}} {
				p, c := point(q[0], q[1])
				vertex(p, c, 1)
			}
		}
	}

	upload := func(v [][7]float32) *uint32 {
		p := new(uint32)
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(p)))
		glExt("glBindBuffer", glArrayBuffer, uintptr(*p))
		glExt("glBufferData", glArrayBuffer, uintptr(len(v)*28), uintptr(unsafe.Pointer(&v[0])), glStaticDraw)
		return p
	}
	remasterSky.VBO = *upload(vertices)
	remasterSky.Count = len(vertices)
	if night < .8 {
		var clouds [][9]float32
		color := add(scale(pos3{.93, .96, 1}, 1-sunset), scale(pos3{1, .75, .53}, sunset))
		if hdrScene.active {
			color = pos3{math.Pow(color.x, 2.2), math.Pow(color.y, 2.2), math.Pow(color.z, 2.2)}
		}
		for i := 0; i < 24; i++ {
			az := float64(i) * 2.39996
			el := .24 + math.Mod(float64(i)*.173, .43)
			dir := pos3{math.Cos(az) * math.Cos(el), math.Sin(el), math.Sin(az) * math.Cos(el)}
			center := scale(dir, 830)
			right := norm(cross(dir, pos3{0, 1, 0}))
			up := cross(right, dir)
			size := 90 + math.Mod(float64(i)*31, 75)
			for _, q := range [][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 0}, {1, 1}, {0, 1}} {
				p := add(center, add(scale(right, (q[0]-.5)*size*2.5), scale(up, (q[1]-.5)*size)))
				clouds = append(clouds, [9]float32{float32(p.x), float32(p.y), float32(p.z), float32(color.x), float32(color.y), float32(color.z), .55, float32(q[0]), float32(q[1])})
			}
		}
		glExt("glGenBuffers", 1, uintptr(unsafe.Pointer(&remasterSky.Clouds)))
		glExt("glBindBuffer", glArrayBuffer, uintptr(remasterSky.Clouds))
		glExt("glBufferData", glArrayBuffer, uintptr(len(clouds)*36), uintptr(unsafe.Pointer(&clouds[0])), glStaticDraw)
		remasterSky.CloudCount = len(clouds)
		pbrTexture("remaster_cloud_sprite.png")
	}
	if night > .5 {
		vertices = nil
		for i := 0; i < 420; i++ {
			az := math.Mod(float64(i)*2.39996323, engine.Tau)
			el := .13 + math.Mod(float64(i)*.173205, 1.35)
			r := 930 * math.Cos(el)
			bright := (.35 + .60*math.Mod(float64(i)*.618, 1)) * night
			vertex(pos3{r * math.Cos(az), 930 * math.Sin(el), r * math.Sin(az)}, pos3{bright * .85, bright * .9, bright}, 1)
		}
		remasterSky.Stars = *upload(vertices)
		remasterSky.StarCount = len(vertices)
	}
	glExt("glBindBuffer", glArrayBuffer, 0)
}
func drawRemasterSky(camera pos3) {
	if drawBeta020Sky() {
        // The high-resolution panorama already contains natural cloud cover.
		return
	}
	useLighting(false)
	if remasterSky.VBO == 0 || remasterSky.Key != remaster.Time || remasterSky.HDR != hdrScene.active {
		buildRemasterSky()
	}
	dis(GL_DEPTH_TEST)
	dis(GL_FOG)
	dis(GL_TEXTURE_2D)
	gl("glDepthMask", 0)
	glExt("glBindVertexArray", 0)
	push()
	move(camera.x, camera.y, camera.z)
	gl("glEnableClientState", glVertexArray)
	gl("glEnableClientState", glColorArray)
	draw := func(buffer uint32, count int, mode uintptr) {
		glExt("glBindBuffer", glArrayBuffer, uintptr(buffer))
		gl("glVertexPointer", 3, glFloat, 28, 0)
		gl("glColorPointer", 4, glFloat, 28, 12)
		gl("glDrawArrays", mode, 0, uintptr(count))
		gpuWorld.draws++
	}
	draw(remasterSky.VBO, remasterSky.Count, GL_TRIANGLES)
	if remasterSky.Clouds != 0 && app.quality > 0 {
		push()
		rotate(app.t*.08, 0, 1, 0)
		en(GL_TEXTURE_2D)
		en(GL_BLEND)
		gl("glBindTexture", GL_TEXTURE_2D, uintptr(pbrTexture("remaster_cloud_sprite.png")))
		glExt("glBindBuffer", glArrayBuffer, uintptr(remasterSky.Clouds))
		gl("glEnableClientState", 0x8078)
		gl("glVertexPointer", 3, glFloat, 36, 0)
		gl("glColorPointer", 4, glFloat, 36, 12)
		gl("glTexCoordPointer", 2, glFloat, 36, 28)
		gl("glDrawArrays", GL_TRIANGLES, 0, uintptr(remasterSky.CloudCount))
		gpuWorld.draws++
		gl("glDisableClientState", 0x8078)
		dis(GL_TEXTURE_2D)
		pop()
	}
	if remasterSky.Stars != 0 {
		gl("glPointSize", b32(1.3))
		draw(remasterSky.Stars, remasterSky.StarCount, 0)
	}
	gl("glDisableClientState", glVertexArray)
	gl("glDisableClientState", glColorArray)
	glExt("glBindBuffer", glArrayBuffer, 0)
	drawBeta011Celestial()
	pop()
	gl("glDepthMask", 1)
	en(GL_DEPTH_TEST)
}

//go:build windows

package main

import (
	"math"
	"speedinneed/engine"
)

func drawAtmosphere(camera pos3) {
	useLighting(false)
	dis(GL_TEXTURE_2D)
	dis(GL_DEPTH_TEST)
	dis(GL_FOG)
	gl("glDepthMask", 0)
	night := engine.NightAmount(sceneHour())
	sunset := engine.Clamp((sceneHour()-16)/2.5, 0, 1) * (1 - night)
	zenith := pos3{.12, .32, .59}
	horizon := pos3{.63, .75, .81}
	zenith = add(scale(zenith, 1-sunset), scale(pos3{.09, .17, .32}, sunset))
	horizon = add(scale(horizon, 1-sunset), scale(pos3{.91, .46, .22}, sunset))
	zenith = add(scale(zenith, 1-night), scale(pos3{.006, .012, .032}, night))
	horizon = add(scale(horizon, 1-night), scale(pos3{.055, .075, .12}, night))
	const rings = 20
	const slices = 96
	for j := 0; j < rings; j++ {
		a := -.18 + float64(j)/rings*(math.Pi*.5+.18)
		b := -.18 + float64(j+1)/rings*(math.Pi*.5+.18)
		begin(GL_QUADS)
		for k := 0; k < slices; k++ {
			u, v := float64(k)*engine.Tau/slices, float64(k+1)*engine.Tau/slices
			for _, p := range [][2]float64{{a, u}, {a, v}, {b, v}, {b, u}} {
				t := math.Pow(engine.Clamp(math.Sin(p[0]), 0, 1), .42)
				c := add(scale(horizon, 1-t), scale(zenith, t))
				if remaster.Enabled && hdrScene.active {
					c = pos3{math.Pow(c.x, 2.2), math.Pow(c.y, 2.2), math.Pow(c.z, 2.2)}
				}
				col(c.x, c.y, c.z, 1)
				radius := 970 * math.Cos(p[0])
				vtx(camera.x+radius*math.Cos(p[1]), camera.y+970*math.Sin(p[0]), camera.z+radius*math.Sin(p[1]))
			}
		}
		end()
	}
	if night > .5 {
		begin(0)
		for i := 0; i < 420; i++ {
			az := math.Mod(float64(i)*2.39996323, engine.Tau)
			el := .13 + math.Mod(float64(i)*.173205, 1.35)
			r := 930 * math.Cos(el)
			bright := (.35 + .60*math.Mod(float64(i)*.618, 1)) * night
			col(bright*.85, bright*.9, bright, 1)
			vtx(camera.x+r*math.Cos(az), camera.y+930*math.Sin(el), camera.z+r*math.Sin(az))
		}
		end()
	}
	// A small atmospheric sun/moon disk. Clouds remain distant and translucent.
	dir := norm(pos3{-.45, .28, .55})
	if night > .5 {
		dir = norm(pos3{.35, .72, .42})
	}
	centre := add(camera, scale(dir, 900))
	right := norm(cross(dir, pos3{0, 1, 0}))
	up := cross(right, dir)
	radius := 8.
	for ring := 3; ring >= 1; ring-- {
		alpha := .08
		if ring == 1 {
			alpha = 1
		}
		if night > .5 {
			col(.70, .78, .95, alpha)
		} else {
			col(1, .85, .56, alpha)
		}
		begin(GL_TRIANGLES)
		for i := 0; i < 48; i++ {
			a, b := float64(i)*engine.Tau/48, float64(i+1)*engine.Tau/48
			p := add(centre, add(scale(right, math.Cos(a)*radius*float64(ring)), scale(up, math.Sin(a)*radius*float64(ring))))
			q := add(centre, add(scale(right, math.Cos(b)*radius*float64(ring)), scale(up, math.Sin(b)*radius*float64(ring))))
			vtx(centre.x, centre.y, centre.z)
			vtx(p.x, p.y, p.z)
			vtx(q.x, q.y, q.z)
		}
		end()
	}
	gl("glDepthMask", 1)
	en(GL_DEPTH_TEST)
}

func renderRoundedHills() {
	if app.mapID == 1 {
		return
	} // City sightlines stay open around its avenues.
	en(GL_TEXTURE_2D)
	bind("rock.jpg")
	for _, hill := range app.world.Hills {
		j := hill.Seed
		x, z := hill.Pos.X, hill.Pos.Z
		radius, height := hill.Radius, hill.Height
		point := func(r, a float64) pos3 {
			noise := 1 + .09*math.Sin(a*3+float64(j)) + .045*math.Sin(a*7)
			xx, zz := x+math.Cos(a)*r*radius*noise, z+math.Sin(a)*r*radius*noise
			shape := math.Pow(math.Max(0, 1-r*r), 1.35)
			return pos3{xx, engine.Terrain(xx, zz) - .20 + height*shape, zz}
		}
		for ring := 0; ring < 12; ring++ {
			r1, r2 := float64(ring)/12, float64(ring+1)/12
			begin(GL_QUADS)
			for side := 0; side < 48; side++ {
				a, b := float64(side)*engine.Tau/48, float64(side+1)*engine.Tau/48
				shade := .76 + .06*math.Sin(a*4+float64(j))
				if app.mapID == 2 {
					col(.86*shade, .68*shade, .49*shade, 1)
				} else {
					col(.67*shade, .69*shade, .64*shade, 1)
				}
				tq(point(r1, a), point(r2, a), point(r2, b), point(r1, b), r1*math.Cos(a)*8, r1*math.Sin(a)*8, r2*math.Cos(b)*8, r2*math.Sin(b)*8)
			}
			end()
		}
	}
}

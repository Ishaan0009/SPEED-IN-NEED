package main

import (
	"math"
	"speedinneed/engine"
	"strings"
)

var miniArea035 struct {
	Definition *engine.WorldDefinition
	Position   engine.Vec
	Name       string
}

type mapProjection035 struct {
	center, forward, right engine.Vec
	scale                  float64
	headingUp              bool
}

func newMapProjection035(center engine.Vec, heading, scale float64, headingUp bool) mapProjection035 {
	p := mapProjection035{center: center, scale: scale, headingUp: headingUp}
	if headingUp {
		p.forward = engine.HeadingVec(heading)
		p.right = engine.Vec{X: -p.forward.Z, Z: p.forward.X}
	}
	return p
}

func (p mapProjection035) project(world engine.Vec) engine.Vec {
	d := world.Sub(p.center)
	if !p.headingUp {
		return d.Mul(p.scale)
	}
	return engine.Vec{X: (d.X*p.right.X + d.Z*p.right.Z) * p.scale, Z: -(d.X*p.forward.X + d.Z*p.forward.Z) * p.scale}
}

func (p mapProjection035) unproject(screen engine.Vec) engine.Vec {
	if !p.headingUp {
		return p.center.Add(screen.Mul(1 / p.scale))
	}
	return p.center.Add(p.right.Mul(screen.X / p.scale)).Add(p.forward.Mul(-screen.Z / p.scale))
}

// The camera uses cross(forward, worldUp) for screen right. Using the opposite
// perpendicular mirrors a heading-up map even when the atlas and route agree.
func navProject035(p, center engine.Vec, heading, scale float64, headingUp bool) engine.Vec {
	return newMapProjection035(center, heading, scale, headingUp).project(p)
}

func navUnproject035(p, center engine.Vec, heading, scale float64, headingUp bool) engine.Vec {
	return newMapProjection035(center, heading, scale, headingUp).unproject(p)
}

func navLabelBox035(px, py, width, height, x, y, size float64, occupied [][4]float64) ([4]float64, bool) {
	for _, offset := range [][2]float64{{0, -1}, {0, 1}, {-.6, -1}, {.6, -1}, {-1, 0}, {1, 0}, {-.6, 1}, {.6, 1}, {0, -2.1}, {0, 2.1}, {-1, -1.8}, {1, -1.8}} {
		box := [4]float64{px - width*.5 + offset[0]*(width*.5+9), py - height*.5 + offset[1]*(height*.5+8), width, height}
		box[0] = engine.Clamp(box[0], x+5, math.Max(x+5, x+size-width-5))
		box[1] = engine.Clamp(box[1], y+5, math.Max(y+5, y+size-height-5))
		clear := true
		for _, r := range occupied {
			if box[0] < r[0]+r[2]+3 && box[0]+width > r[0]-3 && box[1] < r[1]+r[3]+3 && box[1]+height > r[1]-3 {
				clear = false
				break
			}
		}
		if clear {
			return box, true
		}
	}
	return [4]float64{}, false
}

func drawMiniArea035(c engine.Car, x, y, size float64) {
	loadPlaces034()
	if miniArea035.Definition != nav033.Definition || c.Pos.Dist(miniArea035.Position) > 18 {
		miniArea035.Definition, miniArea035.Position = nav033.Definition, c.Pos
		roadName, nearest := "", math.Inf(1)
		for _, road := range nav033.Definition.Roads {
			for i := 0; i+1 < len(road.Points); i++ {
				a, b := road.Points[i], road.Points[i+1]
				d := b.Sub(a)
				den := d.X*d.X + d.Z*d.Z
				u := 0.0
				if den > .0001 {
					q := c.Pos.Sub(a)
					u = engine.Clamp((q.X*d.X+q.Z*d.Z)/den, 0, 1)
				}
				if distance := c.Pos.Dist(a.Add(d.Mul(u))); distance < nearest {
					roadName, nearest = road.Name, distance
				}
			}
		}
		miniArea035.Name = ""
		friendlyRoad := friendlyPlaceName035(roadName)
		for _, p := range places034 {
			if p.Mode == remaster.Mode && p.Kind == "district" && strings.HasPrefix(strings.ToLower(friendlyRoad), strings.ToLower(p.Name)) {
				miniArea035.Name = p.Name
				break
			}
		}
		if strings.HasPrefix(roadName, "THE TUNNEL") {
			miniArea035.Name = "The Tunnel"
		}
	}
	name := miniArea035.Name
	if name == "" {
		return
	}
	font := math.Min(13, size*.068)
	width := remasterTextWidth(name, font)
	if width > size-12 {
		font *= (size - 12) / width
		width = remasterTextWidth(name, font)
	}
	roundRect032(x, y+size+7, size, font+14, 3, [4]float64{.025, .035, .04, .72})
	hudRect(x+size*.35,y+size+7,size*.30,1,cyan[0],cyan[1],cyan[2],.75)
	remasterText(name, x+(size-width)*.5, y+size+13, font, white)
}

package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
	"time"
)

var nav033 struct {
	Definition                      *engine.WorldDefinition
	Graph                           *engine.RoadNavigator
	Path                            engine.NavPath033
	Active                          bool
	Target, Cursor, Center, LastCar engine.Vec
	Zoom, Span                      float64
	Updated, Braked                 time.Time
	Down, Dragged                   bool
	X, Y                            int
}
var navFrameDT033 = 1.0 / 60

var routeColor033 = [4]float64{.76, .36, 1, 1}

func initializeNavigation033(d *engine.WorldDefinition) {
	if nav033.Definition == d {
		return
	}
	nav033.Definition = d
	nav033.Graph = engine.NewRoadNavigator(d)
	nav033.Active = false
	nav033.Zoom = 1
	nav033.Span = 650
	nav033.Center = d.Bounds[0].Add(d.Bounds[1]).Mul(.5)
	nav033.Cursor = d.Spawn
}
func setWaypoint033(p engine.Vec) {
	if app.world == nil || app.world.Remaster == nil {
		return
	}
	initializeNavigation033(app.world.Remaster.Def)
	nav033.Target = p
	nav033.Active = true
	nav033.Updated = time.Time{}
	updateNavigation033(0)
	playAudio030("ui_confirm", .32)
}
func updateNavigation033(dt float64) {
	if dt > 0 {
		navFrameDT033 = dt
	}
	if app.world == nil || app.world.Remaster == nil {
		return
	}
	initializeNavigation033(app.world.Remaster.Def)
	c := app.world.Car
	if nav033.Active && (nav033.Updated.IsZero() || time.Since(nav033.Updated) > time.Second && c.Pos.Dist(nav033.LastCar) > 8) {
		nav033.Path = nav033.Graph.FindPath(c.Pos, nav033.Target, c.Body.Height-.74)
		nav033.Updated = time.Now()
		nav033.LastCar = c.Pos
		if c.Pos.Dist(nav033.Target) < 9 {
			nav033.Active = false
		}
	}
	if remaster.FullMap {
		step := navMapSpan033() * engine.Clamp(dt, 0, .05) * .35
		if key(0x25) {
			nav033.Cursor.X -= step
		}
		if key(0x27) {
			nav033.Cursor.X += step
		}
		if key(0x26) {
			nav033.Cursor.Z -= step
		}
		if key(0x28) {
			nav033.Cursor.Z += step
		}
		d := nav033.Definition
		nav033.Cursor.X = engine.Clamp(nav033.Cursor.X, d.Bounds[0].X, d.Bounds[1].X)
		nav033.Cursor.Z = engine.Clamp(nav033.Cursor.Z, d.Bounds[0].Z, d.Bounds[1].Z)
	}
}
func navMapRect033() (float64, float64, float64) {
	h, w := float64(app.height), float64(app.width)
	size := h * .84
	return (w - size) * .5, h * .085, size
}
func clampMap050() {
	d := nav033.Definition
	if d == nil {
		return
	}
	half := navMapSpan033() * .5
	axis := func(v, lo, hi float64) float64 {
		margin := math.Min(half*.30, (hi-lo)*.12)
		return engine.Clamp(v, lo+half-margin, hi-half+margin)
	}
	for _, a := range []struct {
		v      *float64
		lo, hi float64
	}{{&nav033.Center.X, d.Bounds[0].X, d.Bounds[1].X}, {&nav033.Center.Z, d.Bounds[0].Z, d.Bounds[1].Z}} {
		if half*2 >= a.hi-a.lo {
			*a.v = (a.lo + a.hi) * .5
		} else {
			*a.v = axis(*a.v, a.lo, a.hi)
		}
	}
}
func navMapSpan033() float64 {
	d := nav033.Definition
	if d == nil {
		return 3000
	}
	return math.Max(d.Bounds[1].X-d.Bounds[0].X, d.Bounds[1].Z-d.Bounds[0].Z) / math.Max(1, nav033.Zoom)
}
func navMapWorld033(x, y float64) engine.Vec {
	fx, fy, size := navMapRect033()
	return navUnproject035(engine.Vec{X: x - fx - size*.5, Z: y - fy - size*.5}, nav033.Center, 0, size/navMapSpan033(), false)
}
func navMapKey033(k uintptr) bool {
	defer clampMap050()
	if k == 'M' || k == VK_ESCAPE {
		remaster.FullMap = false
		beta020SyncPointer()
		return true
	}
	if k == VK_RETURN || k == VK_SPACE {
		setWaypoint033(nav033.Cursor)
	}
	if k == 8 || k == 0x2e {
		nav033.Active = false
		nav033.Path = engine.NavPath033{}
	}
	if k == 0xbb || k == 0x6b {
		nav033.Zoom = engine.Clamp(nav033.Zoom*1.2, 1, 14)
	}
	if k == 0xbd || k == 0x6d {
		nav033.Zoom = engine.Clamp(nav033.Zoom/1.2, 1, 14)
	}
	return true
}
func navMapPointer033(message uint32, w, l uintptr) bool {
	defer clampMap050()
	if !remaster.FullMap || beta.screen != screenDrive {
		return false
	}
	x, y := int(int16(l&65535)), int(int16(l>>16))
	fx, fy, size := navMapRect033()
	inside := float64(x) >= fx && float64(x) <= fx+size && float64(y) >= fy && float64(y) <= fy+size
	switch message {
	case WM_LBUTTONDOWN:
		if !inside {
			return true
		}
		win(user32, "SetCapture", app.hwnd)
		nav033.Down = true
		nav033.Dragged = false
		nav033.X, nav033.Y = x, y
		return true
	case WM_MOUSEMOVE:
		if nav033.Down && (math.Abs(float64(x-nav033.X))+math.Abs(float64(y-nav033.Y)) > 3) {
			nav033.Dragged = true
			nav033.Center = nav033.Center.Sub(engine.Vec{X: float64(x-nav033.X) * navMapSpan033() / size, Z: float64(y-nav033.Y) * navMapSpan033() / size})
		}
		nav033.X, nav033.Y = x, y
		if inside {
			nav033.Cursor = navMapWorld033(float64(x), float64(y))
		}
		return true
	case 0x202:
		if nav033.Down && !nav033.Dragged && inside {
			setWaypoint033(navMapWorld033(float64(x), float64(y)))
		}
		win(user32, "ReleaseCapture")
		nav033.Down = false
		return true
	case WM_RBUTTONDOWN:
		nav033.Active = false
		nav033.Path = engine.NavPath033{}
		return true
	case WM_RBUTTONUP:
		return true
	case 0x20a:
		// Wheel coordinates are in screen space; use the tracked client point.
		p := navMapWorld033(float64(nav033.X), float64(nav033.Y))
		delta := float64(int16(w>>16)) / 120
		nav033.Zoom = engine.Clamp(nav033.Zoom*math.Pow(1.2, delta), 1, 14)
		nav033.Center = nav033.Center.Add(p.Sub(navMapWorld033(float64(nav033.X), float64(nav033.Y))))
		return true
	}
	return false
}
func circle033(x, y, r float64, c [4]float64) {
	dis(GL_TEXTURE_2D)
	col(c[0], c[1], c[2], c[3])
	begin(0x0006)
	vtx(x, y, 0)
	for i := 0; i <= 48; i++ {
		a := float64(i) * engine.Tau / 48
		vtx(x+math.Cos(a)*r, y+math.Sin(a)*r, 0)
	}
	end()
}
func navDistance033(v float64) string {
	if v >= 1000 {
		return fmt.Sprintf("%.1f KM", v/1000)
	}
	return fmt.Sprintf("%.0f M", v)
}
func navCircleSegment033(a, b engine.Vec, r float64) (engine.Vec, engine.Vec, bool) {
	// Most island roads are wholly outside the small driving crop. Reject
	// their bounds before solving an intersection with the map circle.
	if a.X > r && b.X > r || a.X < -r && b.X < -r || a.Z > r && b.Z > r || a.Z < -r && b.Z < -r {
		return a, b, false
	}
	d := b.Sub(a)
	den := d.X*d.X + d.Z*d.Z
	if den < 1e-9 {
		return a, b, false
	}
	if a.X*a.X+a.Z*a.Z <= r*r && b.X*b.X+b.Z*b.Z <= r*r {
		return a, b, true
	}
	bb := 2 * (a.X*d.X + a.Z*d.Z)
	cc := a.X*a.X + a.Z*a.Z - r*r
	disc := bb*bb - 4*den*cc
	if disc < 0 {
		return a, b, false
	}
	root := math.Sqrt(disc)
	lo, hi := math.Max(0, (-bb-root)/(2*den)), math.Min(1, (-bb+root)/(2*den))
	return a.Add(d.Mul(lo)), a.Add(d.Mul(hi)), lo <= hi
}
func drawNavigationLine033(project func(engine.Vec) engine.Vec, cx, cy, r float64, circle bool) {
	if !nav033.Active || !nav033.Path.Reachable {
		return
	}
	points := nav033.Path.Points
	// Drop travelled geometry so the highlighted route begins at the car.
	if len(points) > 1 && app.world != nil {
		best, at := math.Inf(1), 0
		var first engine.Vec
		for i := 0; i+1 < len(points); i++ {
			a, b := points[i], points[i+1]
			d := b.Sub(a)
			den := d.X*d.X + d.Z*d.Z
			t := 0.0
			if den > 0 {
				q := app.world.Car.Pos.Sub(a)
				t = engine.Clamp((q.X*d.X+q.Z*d.Z)/den, 0, 1)
			}
			p := a.Add(d.Mul(t))
			if dist := p.Dist(app.world.Car.Pos); dist < best {
				best, at, first = dist, i, p
			}
		}
		points = append([]engine.Vec{first}, points[at+1:]...)
	}
	for pass := 0; pass < 2; pass++ {
		gl("glLineWidth", b32([]float64{7, 4}[pass]))
		if pass == 0 {
			col(.04, .015, .07, .95)
		} else {
			col(routeColor033[0], routeColor033[1], routeColor033[2], 1)
		}
		begin(0x0001)
		for i := 0; i+1 < len(points); i++ {
			a, b := project(points[i]), project(points[i+1])
			ok := true
			if circle {
				a, b, ok = navCircleSegment033(a, b, r)
			}
			if ok {
				vtx(cx+a.X, cy+a.Z, 0)
				vtx(cx+b.X, cy+b.Z, 0)
			}
		}
		end()
	}
	if nav033.Path.OffRoadMetres > 8 {
		a, b := project(nav033.Path.RoadEnd), project(nav033.Target)
		count := engine.ClampInt(int(a.Dist(b)/8), 1, 500)
		gl("glLineWidth", b32(2))
		col(.83, .63, 1, .8)
		begin(0x0001)
		for i := 0; i < count; i += 2 {
			p := a.Add(b.Sub(a).Mul(float64(i) / float64(count)))
			q := a.Add(b.Sub(a).Mul(math.Min(1, float64(i+1)/float64(count))))
			ok := true
			if circle {
				p, q, ok = navCircleSegment033(p, q, r)
			}
			if ok {
				vtx(cx+p.X, cy+p.Z, 0)
				vtx(cx+q.X, cy+q.Z, 0)
			}
		}
		end()
	}
	gl("glLineWidth", b32(1))
	p := project(nav033.Target)
	if circle && p.Length() > r-10 {
		p = p.Unit().Mul(r - 10)
	}
	circle033(cx+p.X, cy+p.Z, 7, [4]float64{.07, .02, .10, 1})
	circle033(cx+p.X, cy+p.Z, 4.5, routeColor033)
}
func mapTexture033() uint32 {
	key := "b038_map_free.png"
	if remaster.Mode == 1 {
		key = "b038_map_race.png"
	}
	if t := app.tex[key]; t != 0 {
		return t
	}
	return app.tex["beta020_atlas.png"]
}
func drawRoadMinimap033(c engine.Car) {
	x, y, size := beta011MiniRect()
	cx, cy, r := x+size*.5, y+size*.5, size*.49
	now := time.Now()
	if c.Braking && math.Abs(c.Speed) > 3 {
		nav033.Braked = now
	}
	base := 650.0
	want := base + engine.Clamp((math.Abs(c.Speed)*3.6-118)*3.8, 0, 550)
	if now.Sub(nav033.Braked) < 1800*time.Millisecond {
		want = math.Max(want, 1000)
	}
	if nav033.Span < 1 {
		nav033.Span = want
	}
	nav033.Span += (want - nav033.Span) * (1 - math.Exp(-engine.Clamp(navFrameDT033, 0, .1)*2.5))
	scale := size / nav033.Span
	headingUp := beta.cfg.MapOrientation == 0
	projection := newMapProjection035(c.Pos, c.Heading, scale, headingUp)
	project := projection.project
	loX, loZ, ex, ez := beta011AtlasDomain()
	circle033(cx, cy, r+3, [4]float64{.01, .017, .024, .90})
	en(GL_TEXTURE_2D)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(mapTexture033()))
	col(1, 1, 1, .98)
	begin(0x0006)
	for i := -1; i <= 64; i++ {
		sx, sy := 0.0, 0.0
		if i >= 0 {
			a := float64(i) * engine.Tau / 64
			sx, sy = math.Cos(a)*r, math.Sin(a)*r
		}
		p := projection.unproject(engine.Vec{X: sx, Z: sy})
		tex((p.X-loX)/ex, (p.Z-loZ)/ez)
		vtx(cx+sx, cy+sy, 0)
	}
	end()
	dis(GL_TEXTURE_2D)
	gl("glLineWidth", b32(1.5))
	col(.83, .84, .84, .76)
	begin(0x0001)
	for _, road := range nav033.Definition.Roads {
		for i := 0; i+1 < len(road.Points); i++ {
			a, b, ok := navCircleSegment033(project(road.Points[i]), project(road.Points[i+1]), r)
			if ok {
				vtx(cx+a.X, cy+a.Z, 0)
				vtx(cx+b.X, cy+b.Z, 0)
			}
		}
	}
	end()
	drawNavigationLine033(project, cx, cy, r, true)
	angle := 0.0
	if !headingUp {
		angle = math.Pi - c.Heading
	}
	push()
	move(cx, cy, 0)
	rotate(angle*180/math.Pi, 0, 0, 1)
	col(1, 1, 1, 1)
	begin(GL_TRIANGLES)
	vtx(0, -9, 0)
	vtx(-6, 7, 0)
	vtx(6, 7, 0)
	end()
	pop()
	north := project(c.Pos.Add(engine.Vec{Z: -100})).Unit().Mul(r * .84)
	remasterText("N", cx+north.X-4, cy+north.Z-6, 10, [4]float64{1, 1, 1, .7})
	miniFrame037(cx, cy, r)
	drawMiniArea035(c, x, y, size)
	if nav033.Active {
		roundRect032(x, y-28, size, 24, 5, [4]float64{.02, .025, .04, .85})
		remasterText(navDistance033(nav033.Path.Metres), x+9, y-22, 13, routeColor033)
	}
	gl("glLineWidth", b32(1))
}
func drawFullMap033() {
	if app.world == nil || app.world.Remaster == nil {
		return
	}
	initializeNavigation033(app.world.Remaster.Def)
	remasterUIBegin()
	w, h := float64(app.width), float64(app.height)
	x, y, size := navMapRect033()
	span := navMapSpan033()
	cx, cy := x+size*.5, y+size*.5
	beta011MapProgress = 1
	hudRect(0, 0, w, h, .012, .017, .024, .95)
	remasterText("MAP", w*.055, h*.045, h*.028, white)
	remasterText("SPEED IN NEED  /  0.5.1", w*.72, h*.05, h*.012, amber)
	en(GL_SCISSOR_TEST)
	gl("glScissor", uintptr(x), uintptr(h-y-size), uintptr(size), uintptr(size))
	loX, loZ, ex, ez := beta011AtlasDomain()
	scale := size / span
	project := func(p engine.Vec) engine.Vec { return navProject035(p, nav033.Center, 0, scale, false) }
	a, b := project(engine.Vec{X: loX, Z: loZ}), project(engine.Vec{X: loX + ex, Z: loZ + ez})
	hudRect(x, y, size, size, .045, .08, .09, 1)
	en(GL_TEXTURE_2D)
	gl("glBindTexture", GL_TEXTURE_2D, uintptr(mapTexture033()))
	col(1, 1, 1, 1)
	begin(GL_QUADS)
	tex(0, 0)
	vtx(cx+a.X, cy+a.Z, 0)
	tex(1, 0)
	vtx(cx+b.X, cy+a.Z, 0)
	tex(1, 1)
	vtx(cx+b.X, cy+b.Z, 0)
	tex(0, 1)
	vtx(cx+a.X, cy+b.Z, 0)
	end()
	dis(GL_TEXTURE_2D)
	drawMapDetail034(project, cx, cy, x, y, size, scale)
	drawNavigationLine033(project, cx, cy, 0, false)
	p := project(app.world.Car.Pos)
	circle033(cx+p.X, cy+p.Z, 9, [4]float64{.02, .07, .10, 1})
	push()
	move(cx+p.X, cy+p.Z, 0)
	rotate((math.Pi-app.world.Car.Heading)*180/math.Pi, 0, 0, 1)
	col(.8, .96, 1, 1)
	begin(GL_TRIANGLES)
	vtx(0, -8, 0)
	vtx(-5, 5, 0)
	vtx(5, 5, 0)
	end()
	pop()
	q := project(nav033.Cursor)
	gl("glLineWidth", b32(1.5))
	col(1, .65, .20, 1)
	begin(0x0001)
	vtx(cx+q.X-11, cy+q.Z, 0)
	vtx(cx+q.X+11, cy+q.Z, 0)
	vtx(cx+q.X, cy+q.Z-11, 0)
	vtx(cx+q.X, cy+q.Z+11, 0)
	end()
	gl("glLineWidth", b32(1))
	drawMapLabels034(project, cx, cy, x, y, size)
	dis(GL_SCISSOR_TEST)
	label := "SELECT A DESTINATION"
	if nav033.Active {
		label = "ROAD ROUTE  " + navDistance033(nav033.Path.Metres)
		if nav033.Path.OffRoadMetres > 8 {
			label += "  /  OFF ROAD " + navDistance033(nav033.Path.OffRoadMetres)
		}
	}
	remasterText(label, w*.055, h*.925, h*.014, routeColor033)
	remasterText("CLICK / ENTER / A: SET    RIGHT CLICK / X: CLEAR    DRAG: PAN    WHEEL: ZOOM    M / ESC / B: CLOSE", w*.055, h*.965, h*.010, white)
	en(GL_DEPTH_TEST)
}

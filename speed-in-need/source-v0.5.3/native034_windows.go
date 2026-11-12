package main

import (
	"encoding/json"
	"math"
	"speedinneed/engine"
	"strings"
)

var autopilot034 bool

// Occlude a complete static building cell only when an intervening terrain
// cross-section covers its upper bound across its full projected width.
func terrainOccluded034(g remasterInstance, eye pos3) bool {
	if app.world == nil || app.world.Remaster == nil || math.IsInf(g.TopY034, 0) {
		return false
	}
	dx, dz := g.X-eye.x, g.Z-eye.z
	distance := math.Hypot(dx, dz)
	if distance < g.Radius+220 {
		return false
	}
	grid := app.world.Remaster.Def.Grid
	right := engine.Vec{X: dz / distance, Z: -dx / distance}
	denom := distance - g.Radius
	if g.TopY034 < eye.y {
		denom = distance + g.Radius
	}
	for _, t := range []float64{.35, .55, .72} {
		center := engine.Vec{X: eye.x + dx*t, Z: eye.z + dz*t}
		radius := g.Radius*t + grid.Step*2
		ceiling := eye.y + (g.TopY034-eye.y)*distance*t/denom + 10
		blocked := true
		steps := int(math.Ceil(2 * radius / math.Max(1, grid.Step*.5)))
		for i := 0; i <= steps; i++ {
			p := center.Add(right.Mul(-radius + 2*radius*float64(i)/float64(steps)))
			if grid.Height(p) < ceiling {
				blocked = false
				break
			}
		}
		if blocked {
			return true
		}
	}
	return false
}

func fog034() ([1]float32, [1]float32, [4]float32) {
	night, sunset := engine.NightAmount(sceneHour()), sunset034()
	day := add(scale(pos3{.62, .74, .83}, 1-sunset), scale(pos3{.65, .43, .28}, sunset))
	color := add(scale(day, 1-night), scale(pos3{.018, .029, .055}, night))
	return [1]float32{float32(680*(1-night) + 190*night)}, [1]float32{float32(2400*(1-night) + 1500*night)}, [4]float32{float32(color.x), float32(color.y), float32(color.z), 1}
}

var cycle034 struct {
	Elapsed float64
	Mode    int
}

// Native OpenGL profile: no frame generation or vendor-specific upscaler.
func applyRTX3060Profile034() {
	applyAokawaPreset(1)
	c := &beta.cfg
	c.Resolution = 2
	c.TrafficDensity = 1
	c.RenderScale = 1
	c.TextureBudget = 2
	c.ViewDistance = 1.3
	c.Vegetation = .95
	c.FoliageDensity = .85
	c.GroundDetail = 2
	c.Shadows = 1
	c.ShadowFilter = 1
	c.Anisotropy = 16
	c.FXAA = true
	c.SSAO = true
	c.MotionBlur = true
	c.BlurStrength = .28
	c.BloomStrength = .70
	c.Exposure = 1
	c.Saturation = 1
	c.CameraShake = .25
	c.FrameCap = 60
	c.VSync = true
	c.AdaptiveQuality = true
}

func migratePad034() {
	if beta.cfg.PadMappingVersion < 34 {
		beta.cfg.GamepadButtons = defaultPadButtons033
		beta.cfg.GamepadSteerAxis = 0
		beta.cfg.GamepadThrottleAxis = 5
		beta.cfg.GamepadBrakeAxis = 4
		beta.cfg.GamepadTriggerMode = 0
		beta.cfg.GamepadSteerInvert = true
		beta.cfg.PadMappingVersion = 34
	}
}

// Optional full-resolution presentation preset. The RTX preset keeps adaptive
// quality available for sustained driving; this preset holds the effects fixed.
func applyCinematicShaderProfile034() {
	applyRTX3060Profile034()
	app.quality = 2
	c := &beta.cfg
	c.Quality = 2
	c.LightingDetail = 2
	c.Shadows = 2
	c.ShadowFilter = 2
	c.BloomStrength = .85
	c.AdaptiveQuality = false
}

func sonyPad034(name string, mid uint16) bool {
	name = strings.ToLower(name)
	return mid == 0x054c || strings.Contains(name, "wireless controller") || strings.Contains(name, "dualshock") || strings.Contains(name, "dualsense") || strings.Contains(name, "playstation")
}

// Centered sticks use the driver's calibrated midpoint. Reconnecting with a
// stick held must not make that position the new zero for the whole session.
func centeredJoy034(v float64) float64 { return engine.Clamp(v*2-1, -1, 1) }

func toggleAutopilot034() {
	if autopilot034 {
		autopilot034 = false
		return
	}
	if app.world != nil && nav033.Active && nav033.Path.Reachable && len(nav033.Path.Points) > 1 {
		autopilot034 = true
		playAudio030("ui_confirm", .3)
	}
}

// Consume the existing v0.5.3 path without changing its graph or routing rules.
// Manual input always takes control immediately. Stop at the road endpoint;
// the dashed off-road final leg is never driven automatically.
func mergeAutopilot034(in engine.Input) engine.Input {
	if !autopilot034 {
		return in
	}
	if math.Abs(in.Steer) > .18 || in.Throttle > .18 || in.Brake > .18 || in.Handbrake || in.Boost || app.world == nil || !nav033.Active || !nav033.Path.Reachable {
		autopilot034 = false
		return in
	}
	c := app.world.Car
	points := nav033.Path.Points
	if len(points) < 2 {
		autopilot034 = false
		return engine.Input{Brake: 1}
	}
	best, at := math.Inf(1), 0
	var nearest engine.Vec
	for i := 0; i+1 < len(points); i++ {
		a, b := points[i], points[i+1]
		d := b.Sub(a)
		den := d.X*d.X + d.Z*d.Z
		t := 0.
		if den > 1e-9 {
			q := c.Pos.Sub(a)
			t = engine.Clamp((q.X*d.X+q.Z*d.Z)/den, 0, 1)
		}
		p := a.Add(d.Mul(t))
		dist := p.Dist(c.Pos)
		if dist < best {
			best, at, nearest = dist, i, p
		}
	}
	// Losing the road or encountering an obstacle requires the driver to resume.
	if best > 22 {
		autopilot034 = false
		return engine.Input{Brake: 1}
	}
	look := 8 + math.Abs(c.Speed)*.60
	target := nearest
	remaining := 0.
	for j := at + 1; j < len(points); j++ {
		remaining += points[j].Dist(target)
		target = points[j]
	}
	target = nearest
	travel := look
	for j := at + 1; j < len(points); j++ {
		d := points[j].Dist(target)
		if d >= travel {
			target = target.Add(points[j].Sub(target).Mul(travel / math.Max(d, .001)))
			break
		}
		travel -= d
		target = points[j]
	}
	q := target.Sub(c.Pos)
	error := math.Remainder(math.Atan2(q.X, q.Z)-c.Heading, engine.Tau)
	steer := engine.Clamp(math.Atan2(2*2.7*math.Sin(error), math.Max(4, look))/.42, -1, 1)
	speed := math.Min(16, math.Sqrt(3.2/math.Max(.012, 2*math.Abs(math.Sin(error))/look)))
	speed = math.Min(speed, math.Sqrt(math.Max(0, remaining-2)*3.0))
	if math.Abs(error) > 1.2 {
		speed = math.Min(speed, 3)
	}
	// Autonomous drive starts the engine; manual takeover preserves pad ignition.
	out := engine.Input{Steer: steer, EngineGated: false}
	if remaining < 3 && math.Abs(c.Speed) < .6 {
		autopilot034 = false
		out.Brake = 1
		return out
	}
	if c.Speed > speed+.5 {
		out.Brake = engine.Clamp((c.Speed-speed)*.3, .1, 1)
	} else {
		out.Throttle = engine.Clamp((speed-c.Speed)*.22, 0, .65)
	}
	// Check the physical collision scene on the upcoming path, including props.
	f := engine.HeadingVec(c.Heading)
	end := c.Pos.Add(f.Mul(math.Max(7, c.Speed*1.2)))
	if app.world.RemasteredCameraFraction([3]float64{c.Pos.X, c.Body.Height + .3, c.Pos.Z}, [3]float64{end.X, c.Body.Height + .3, end.Z}) < .85 {
		out.Throttle = 0
		out.Brake = 1
	}
	return out
}

func advanceCycle034(dt float64) {
	if cycle034.Mode != remaster.Time {
		cycle034.Mode = remaster.Time
		cycle034.Elapsed = 0
	}
	if !qa.enabled && beta.screen == screenDrive && !remaster.FullMap && !app.paused {
		cycle034.Elapsed += engine.Clamp(dt, 0, .075)
	}
}
func cycleHour034() float64 {
	base := []float64{10, 17.8, 22}[engine.ClampInt(remaster.Time, 0, 2)]
	if qa.enabled {
		return []float64{16.2, 17.8, 22}[engine.ClampInt(remaster.Time, 0, 2)]
	}
	if !beta.cfg.RandomTime040 {return base}
	if math.IsNaN(randomHour040){startRandomTime040()}
	return math.Mod(randomHour040+cycle034.Elapsed/90, 24) // Random mode alone advances the day.
}
func skyBlend034() (int, int, float64) {
	if garage033 {
		return 0, 0, 0
	}
	h := sceneHour()
	switch {
	case h < 5.5:
		return 2, 2, 0
	case h < 7.5:
		return 2, 0, (h - 5.5) / 2
	case h < 16:
		return 0, 0, 0
	case h < 18:
		return 0, 1, (h - 16) / 2
	case h < 20:
		return 1, 2, (h - 18) / 2
	default:
		return 2, 2, 0
	}
}
func sunset034() float64 {
	h := sceneHour()
	return math.Max(0, 1-math.Abs(h-18)/3) * (1 - engine.NightAmount(h))
}

// Conservative horizontal frustum for cells with a supplied bounding radius.
func visibleCone034(x, z, r float64, eye, target pos3) bool {
	return visibleConeCache035.contains(x, z, r, eye, target, app.width, app.height)
}

type place034 struct {
	Name         string
	OriginalName string
	X, Z, Zoom   float64
	Kind, Source string
	Mode         int
}

var places034 []place034
var titleFont034 struct {
	Glyphs  map[string]remasterGlyph
	Kerning map[string]float64
}

func titleText034(label string, x, y, h, maxWidth float64) {
	if titleFont034.Glyphs == nil {
		if raw, err := assets.ReadFile("assets/b034_title_font.json"); err == nil {
			json.Unmarshal(raw, &titleFont034)
		}
	}
	glyphs, atlas := remaster.Glyphs, remaster.Atlas
	if titleFont034.Glyphs != nil {
		remaster.Glyphs = titleFont034.Glyphs
		remaster.Atlas = app.tex["b034_title_font.png"]
	}
	defer func() { remaster.Glyphs = glyphs; remaster.Atlas = atlas }()
	scale := beta.cfg.UIScale
	if scale <= 0 {
		scale = 1
	}
	width := remasterTextWidth(label, h)
	if width > maxWidth {
		h *= maxWidth / width
	}
	previous := ""
	for _, r := range label {
		ch := string(r)
		x += titleFont034.Kerning[previous+ch] * h * scale
		remasterText(ch, x+1, y+2, h, [4]float64{0, 0, 0, .7})
		remasterText(ch, x, y, h, white)
		x += remaster.Glyphs[ch].Advance * h * scale
		previous = ch
	}
}
func loadPlaces034() {
	if places034 != nil {
		return
	}
	places034 = []place034{}
	if raw, err := assets.ReadFile("assets/b037_places.json"); err == nil {
		json.Unmarshal(raw, &places034)
	} else if raw, err := assets.ReadFile("assets/b034_places.json"); err == nil {
		json.Unmarshal(raw, &places034)
	}
	for i := range places034 {
		if places034[i].OriginalName == "" {
			places034[i].OriginalName = places034[i].Name
		}
		places034[i].Name = friendlyPlaceName035(places034[i].Name)
	}
}
func drawMapDetail034(project func(engine.Vec) engine.Vec, cx, cy, x, y, size, scale float64) {
	en(0x0B20)
	for pass := 0; pass < 2; pass++ {
		for _, road := range nav033.Definition.Roads {
			width := engine.Clamp(road.Width*scale, 1.3, 18)
			if pass == 0 {
				gl("glLineWidth", b32(width+2))
				col(.055, .075, .085, .92)
			} else {
				gl("glLineWidth", b32(width))
				col(.83, .85, .80, 1)
			}
			if road.Kind == "tunnel" && pass == 1 {
				col(.65, .72, .76, 1)
			}
			begin(0x0003)
			for _, p := range road.Points {
				q := project(p)
				vtx(cx+q.X, cy+q.Z, 0)
			}
			end()
		}
	}
	gl("glLineWidth", b32(1))
	dis(0x0B20)
}
func drawMapLabels034(project func(engine.Vec) engine.Vec, cx, cy, x, y, size float64) {
	loadPlaces034()
	occupied := [][4]float64{}
	markers := []engine.Vec{app.world.Car.Pos, nav033.Cursor}
	if nav033.Active {
		markers = append(markers, nav033.Target)
	}
	for _, p := range markers {
		q := project(p)
		occupied = append(occupied, [4]float64{cx + q.X - 14, cy + q.Z - 14, 28, 28})
	}
	for _, p := range places034 {
		if nav033.Zoom < p.Zoom || p.Mode != remaster.Mode {
			continue
		}
		q := project(engine.Vec{X: p.X, Z: p.Z})
		px, py := cx+q.X, cy+q.Z
		if px < x+3 || px > x+size-3 || py < y+3 || py > y+size-3 {
			continue
		}
		font := engine.Clamp(float64(app.height)*.014, 12, 20)
		if p.Kind == "road" {
			font *= .82
		}
		width, height := remasterTextWidth(p.Name, font)+16, font+12
		box, placed := navLabelBox035(px, py, width, height, x, y, size, occupied)
		if !placed {
			// Always retain the actual place marker. Crowded secondary road
			// names appear as the player zooms further into their location.
			circle033(px, py, 2, [4]float64{.91, .79, .49, .9})
			continue
		}
		occupied = append(occupied, box)
		// A leader keeps displaced labels tied to the correct actual feature.
		gl("glLineWidth", b32(1))
		col(.86, .92, .83, .8)
		begin(0x0001)
		vtx(px, py, 0)
		vtx(engine.Clamp(px, box[0], box[0]+width), engine.Clamp(py, box[1], box[1]+height), 0)
		end()
		circle033(px, py, 2.5, [4]float64{.98, .84, .46, 1})
		background := [4]float64{.022, .055, .068, .94}
		if p.Kind == "district" {
			background = [4]float64{.055, .12, .13, .95}
		}
		roundRect032(box[0], box[1], width, height, 4, background)
		remasterText(p.Name, box[0]+8, box[1]+4, font, [4]float64{.98, .99, .94, 1})
	}
}

// A bounded, depth-tested ambient batch. No dynamic lights or per-frame uploads.
func drawFireflies034() {
	night := engine.NightAmount(sceneHour())
	if night < .1 || app.quality == 0 || reducedEffects026() {
		return
	}
	c := app.renderCar
	d := app.world.Remaster.Def
	useLighting(false)
	dis(GL_TEXTURE_2D)
	gl("glDepthMask", 0)
	en(GL_BLEND)
	gl("glPointSize", b32(3))
	en(0x0B10)
	begin(0x0000)
	for i := 0; i < 160; i++ {
		seed := float64(i)
		x := 650 + math.Sin(seed*2.39996)*520
		z := -700 + math.Cos(seed*1.618)*650
		if math.Hypot(x-c.Pos.X, z-c.Pos.Z) > 150 {
			continue
		}
		p := engine.Vec{X: x, Z: z}
		ground := d.Grid.Height(p)
		if d.WaterGrid.Height(p) > .1 {
			continue
		}
		a := (.25 + .75*math.Pow(.5+.5*math.Sin(app.t*1.6+seed*2.1), 3)) * night
		col(.40, .85, .18, a)
		vtx(x+math.Sin(app.t*.6+seed)*1.2, ground+1.5+math.Sin(app.t+seed)*.6, z+math.Cos(app.t*.4+seed))
	}
	end()
	dis(0x0B10)
	gl("glPointSize", b32(1))
	gl("glDepthMask", 1)
	useLighting(true)
}

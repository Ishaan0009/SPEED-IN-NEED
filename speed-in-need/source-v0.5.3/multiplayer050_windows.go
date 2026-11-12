package main

import (
	"crypto/sha256"
	"fmt"
	"math"
	"net"
	"speedinneed/engine"
	"speedinneed/lan"
	"strings"
	"time"
)

var multiplayer050 struct {
	Session               *lan.Session
	Open                  bool
	Address, Message, Key string
	Rooms                 []lan.Room
	Discovery             chan []lan.Room
	View                  lan.Snapshot
	Epoch                 uint64
	Remote                map[string]engine.Car
	Target, FOV           float64
	ViewTarget            pos3
	Race                  map[string]*racePeer050
}

type racePeer050 struct {
	Lap, Gate         int
	Crossed, Finished bool
	Time              float64
}

func worldKey050() string {
	if multiplayer050.Key != "" {
		return multiplayer050.Key
	}
	h := sha256.New()
	for _, n := range []string{"assets/b038_world_free.json", "assets/b038_world_race.json"} {
		b, e := assets.ReadFile(n)
		if e != nil {
			return ""
		}
		h.Write(b)
	}
	multiplayer050.Key = fmt.Sprintf("051-%x", h.Sum(nil)[:12])
	return multiplayer050.Key
}
func lanIdentity050() (string, string) {
	if p := activeProfile040(); p != nil {
		return p.Name, p.Avatar
	}
	return "Driver", ""
}
func leaveLAN050() {
	trafficSmoothing051 = map[int]lan.Car{}
	if multiplayer050.Session != nil {
		multiplayer050.Session.Close()
	}
	multiplayer050.Session = nil
	multiplayer050.Remote = nil
	multiplayer050.View = lan.Snapshot{}
	multiplayer050.Epoch = 0
	if app.world != nil {
		app.world.NetworkClient050 = false
		app.world.NetworkObservers050 = nil
	}
	multiplayer050.Message = "Left the LAN session"
}
func startLANWorld050(mode int, epoch uint64) {
	remaster.Mode = engine.ClampInt(mode, 0, 1)
	activateRemasteredWorld(true)
	if app.world == nil || app.world.Remaster == nil {
		return
	}
	multiplayer050.Epoch = epoch
	multiplayer050.Remote = map[string]engine.Car{}
	multiplayer050.Race = map[string]*racePeer050{}
	app.world.NetworkClient050 = multiplayer050.Session != nil && !multiplayer050.Session.IsHost()
	applySetupSpawnRefine()
	app.started = true
	app.paused = false
	selectScreen(screenDrive)
	app.lightingOn = true
	// Every participant receives its own grid slot. AI grid starts further ahead.
	if multiplayer050.Session != nil {
		slot := 0
		for i, p := range multiplayer050.View.Players {
			if p.ID == multiplayer050.Session.ID() {
				slot = i
			}
		}
		if mode == 1 && len(app.world.Routes) > 0 {
			back, lane := engine.LANGrid050(remaster.Opponents, slot)
			route := app.world.Routes[0]
			distance := route.Len - back
			p, f := route.AtDistance(distance)
			app.world.Car.Pos = p.Add(engine.Vec{X: f.Z, Z: -f.X}.Mul(lane))
			app.world.Car.Heading = math.Atan2(f.X, f.Z)
			app.world.Remaster.PlayerLastRouteDistance = distance
			app.world.Remaster.PlayerProgress = -back
			app.world.Remaster.NextGate = 0
			app.world.Car.Body.Initialized = false
		} else if slot > 0 {
			f := engine.HeadingVec(app.world.Car.Heading)
			right := engine.Vec{X: f.Z, Z: -f.X}
			app.world.Car.Pos = app.world.Car.Pos.Sub(f.Mul(float64(slot/2)*7 + 8)).Add(right.Mul(float64(slot%2)*4 - 2))
			app.world.Car.Body.Initialized = false
		}
		app.prevCar = app.world.Car
		app.renderCar = app.world.Car
		multiplayer050.Session.SetReady(true)
	}
}
func lanAction050(id string) {
	if offlineAudit053() {
		return
	}
	q := &multiplayer050
	name, avatar := lanIdentity050()
	key := worldKey050()
	if key == "" {
		q.Message = "World files could not be read"
		return
	}
	switch id {
	case "host":
		leaveLAN050()
		s, e := lan.NewHost(name, avatar, key, lan.DefaultPort)
		if e != nil {
			q.Message = e.Error()
			return
		}
		q.Session = s
		q.Message = "Hosting on UDP 42850"
		startLANWorld050(0, uint64(time.Now().UnixMilli()))
	case "join":
		address := strings.TrimSpace(q.Address)
		if address == "" {
			q.Message = "Enter the host IPv4 address, or discover a room"
			return
		}
		leaveLAN050()
		s, e := lan.NewClient(address, name, avatar, key)
		if e != nil {
			q.Message = e.Error()
			return
		}
		q.Session = s
		q.Message = "Join request sent"
	case "discover":
		if q.Discovery == nil {
			q.Discovery = make(chan []lan.Room, 1)
			ch := q.Discovery
			go func() { ch <- lan.Discover(lan.DefaultPort, 900*time.Millisecond) }()
			q.Message = "Finding nearby rooms..."
		}
	case "leave":
		leaveLAN050()
	case "free", "race":
		if q.Session != nil && q.Session.IsHost() {
			mode := 0
			if id == "race" {
				mode = 1
			}
			startLANWorld050(mode, uint64(time.Now().UnixMilli()))
		}
	case "ai-":
		remaster.Opponents = engine.ClampInt(remaster.Opponents-1, 0, 7)
	case "ai+":
		remaster.Opponents = engine.ClampInt(remaster.Opponents+1, 0, 7)
	case "difficulty":
		beta.cfg.RaceDifficulty040 = (beta.cfg.RaceDifficulty040 + 1) % 3
	default:
		if strings.HasPrefix(id, "add:") && q.Session != nil {
			if e := q.Session.Approve(strings.TrimPrefix(id, "add:")); e != nil {
				q.Message = e.Error()
			}
		}
		if strings.HasPrefix(id, "kick:") && q.Session != nil {
			q.Session.Remove(strings.TrimPrefix(id, "kick:"))
		}
		if strings.HasPrefix(id, "room:") {
			var i int
			fmt.Sscanf(id, "room:%d", &i)
			if i >= 0 && i < len(q.Rooms) {
				q.Address = q.Rooms[i].Address
				lanAction050("join")
			}
		}
	}
}
func lanWndProc050(hwnd uintptr, m uint32, w, l uintptr) bool {
	if m == WM_KEYDOWN && w == 0x76 && beta.screen == screenDrive && (multiplayer050.Session == nil || multiplayer050.Session.IsHost()) {
		weather050.Rain = !weather050.Rain
		return true
	}
	if m == WM_KEYDOWN && w == 0x77 && (l>>30)&1 == 0 {
		multiplayer050.Open = !multiplayer050.Open
		beta020SyncPointer()
		return true
	}
	if !multiplayer050.Open {
		return false
	}
	switch m {
	case WM_KEYDOWN:
		if w == VK_ESCAPE {
			multiplayer050.Open = false
			beta020SyncPointer()
		} else if w == VK_RETURN {
			lanAction050("join")
		}
		return true
	case 0x102:
		if w == 8 {
			if len(multiplayer050.Address) > 0 {
				multiplayer050.Address = multiplayer050.Address[:len(multiplayer050.Address)-1]
			}
		} else if (w >= '0' && w <= '9' || w == '.' || w == ':') && len(multiplayer050.Address) < 48 {
			multiplayer050.Address += string(rune(w))
		}
		return true
	case WM_LBUTTONDOWN:
		for _, b := range lanButtons050 {
			x, y := float64(int16(l&65535)), float64(int16(l>>16))
			if x >= b.X && x <= b.X+b.W && y >= b.Y && y <= b.Y+b.H {
				lanAction050(b.ID)
				return true
			}
		}
		return true
	case WM_MOUSEMOVE, 0x202, WM_RBUTTONDOWN, WM_RBUTTONUP, 0x20a:
		return true
	}
	return false
}
func carToLAN050(c engine.Car, class, paint int) lan.Car {
	return lan.Car{Model: c.Variant050, Tyre: c.Tyre050, Rim: c.Rim050, Livery: c.Livery050, X: c.Pos.X, Y: c.Body.Height, Z: c.Pos.Z, H: c.Heading, S: c.Speed, L: c.Lateral, Steer: c.VisualSteer, Wheel: c.WheelSpin, Nitro: c.Nitro, Pitch: c.Rig.Pitch, Roll: c.Rig.Roll, Vehicle: class, Paint: paint, Boost: c.Boosting, Brake: c.Braking}
}
func applyLANCar050(c *engine.Car, p lan.Car) {
	c.Variant050 = p.Model
	c.Tyre050 = p.Tyre
	c.Rim050 = p.Rim
	c.Livery050 = p.Livery
	c.Pos = engine.Vec{X: p.X, Z: p.Z}
	c.Body.Height = p.Y
	c.Heading = p.H
	c.Speed = p.S
	c.Lateral = p.L
	c.VisualSteer = p.Steer
	c.WheelSpin = p.Wheel
	c.Nitro = p.Nitro
	c.Rig.Pitch = p.Pitch
	c.Rig.Roll = p.Roll
	c.Body.Initialized = true
	c.Body.Grounded = true
	c.OnRoad = true
	c.Boosting = p.Boost
	c.Braking = p.Brake
	for i := range c.Rig.Wheels {
		c.Rig.Wheels[i].RollAngle = p.Wheel
		c.Rig.Wheels[i].SteerAngle = 0
		if i < 2 {
			c.Rig.Wheels[i].SteerAngle = p.Steer
		}
	}
}
func updateLAN050(dt float64) {
	if offlineAudit053() {
		return
	}
	q := &multiplayer050
	if q.Discovery != nil {
		select {
		case rooms := <-q.Discovery:
			q.Rooms = rooms
			q.Discovery = nil
			q.Message = fmt.Sprintf("Found %d rooms; click one, or enter the host IP", len(rooms))
		default:
		}
	}
	s := q.Session
	if s == nil {
		return
	}
	v := s.View()
	age051 := v.Age051 + .025
	for _, p := range v.Players {
		if p.ID == s.ID() {
			age051 += math.Min(.05, float64(p.Ping)*.0005)
		}
	}
	q.View = v
	if !s.IsHost() && s.Connected() && v.World.Epoch != 0 && v.World.Epoch != q.Epoch {
		remaster.Time = v.World.Time
		remaster.Laps = v.World.Laps
		remaster.Opponents = v.World.Opponents
		beta.cfg.RaceDifficulty040 = v.World.Difficulty
		startLANWorld050(v.World.Mode, v.World.Epoch)
	}
	if app.world == nil || app.world.Remaster == nil || q.Epoch == 0 {
		return
	}
	world := app.world
	r := world.Remaster
	world.NetworkObservers050 = world.NetworkObservers050[:0]
	for _, p := range v.Players {
		if p.ID != s.ID() && p.Ready {
			world.NetworkObservers050 = append(world.NetworkObservers050, engine.Vec{X: p.State.X, Z: p.State.Z})
		}
	}
	if !s.IsHost() && s.Connected() {
		r.Countdown = math.Max(0, v.World.Countdown-v.Age051)
		r.RaceTime = v.World.RaceTime
		if r.Mode == 1 && r.Countdown <= 0 {
			r.RaceTime += v.Age051
		}
		r.Started = r.Countdown <= 0
		weather050.Rain = v.World.Rain
		for i, c := range v.AI {
			c = predictLANState051(c, age051)
			if i < len(r.Racers) {
				old := r.Racers[i].Car
				applyLANCar050(&r.Racers[i].Car, c)
				if old.Pos.Dist(r.Racers[i].Car.Pos) < 25 {
					k := 1 - math.Exp(-engine.Clamp(dt, 0, .1)*22)
					r.Racers[i].Car.Pos = old.Pos.Add(r.Racers[i].Car.Pos.Sub(old.Pos).Mul(k))
					r.Racers[i].Car.Heading = old.Heading + math.Atan2(math.Sin(c.H-old.Heading), math.Cos(c.H-old.Heading))*k
				}
				r.Racers[i].Class = c.Vehicle
				r.Racers[i].Lap = c.Lap
				r.Racers[i].Finished = c.Finished
			}
		}
		// The host sends traffic transforms. No independent client traffic simulation.
		world.Traffic = world.Traffic[:0]
		for _, c := range v.Traffic {
			c = predictLANState051(c, age051)
			c = smoothTraffic051(c, dt)
			world.Traffic = append(world.Traffic, engine.Traffic{Pos: engine.Vec{X: c.X, Z: c.Z}, Heading: c.H, Speed: c.S, WheelSpin: c.Wheel, SteerAngle: c.Steer, Variant: c.Vehicle, Braking: c.Brake})
		}
		r.TrafficSnapshot = append(r.TrafficSnapshot[:0], world.Traffic...)
		for _, p := range v.Players {
			if p.ID == s.ID() && r.Mode == 1 {
				r.Lap = p.State.Lap
				r.NextGate = p.State.Gate
				r.Finished = p.State.Finished
				r.FinishTime = p.State.RaceTime
				world.Finished = r.Finished
			}
		}
	}
	if q.Remote == nil {
		q.Remote = map[string]engine.Car{}
	}
	live := map[string]bool{}
	for _, p := range v.Players {
		if p.ID == s.ID() {
			continue
		}
		live[p.ID] = true
		p.State = predictLANState051(p.State, age051+p.Age051)
		c, exists := q.Remote[p.ID]
		if !exists || c.Pos.Dist(engine.Vec{X: p.State.X, Z: p.State.Z}) > 35 {
			applyLANCar050(&c, p.State)
		} else {
			old := c
			applyLANCar050(&c, p.State)
			k := 1 - math.Exp(-engine.Clamp(dt, 0, .1)*18)
			c.Pos = old.Pos.Add(c.Pos.Sub(old.Pos).Mul(k))
			c.Heading = old.Heading + math.Atan2(math.Sin(c.Heading-old.Heading), math.Cos(c.Heading-old.Heading))*k
		}
		q.Remote[p.ID] = c
	}
	for id := range q.Remote {
		if !live[id] {
			delete(q.Remote, id)
		}
	}
	if s.IsHost() && r.Mode == 1 {
		allReady := true
		for _, p := range v.Players {
			if !p.Ready {
				allReady = false
			}
		}
		if !allReady {
			r.Countdown = 5
			r.Started = false
		}
	}
}
func publishLAN050() {
	if offlineAudit053() {
		return
	}
	q := &multiplayer050
	s := q.Session
	if s == nil || app.world == nil || app.world.Remaster == nil || q.Epoch == 0 {
		return
	}
	w := app.world
	r := w.Remaster
	c := carToLAN050(w.Car, w.CarClass, beta.paint)
	c.Lap = r.Lap
	c.Gate = r.NextGate
	c.Finished = r.Finished
	c.RaceTime = r.FinishTime
	meta := lan.World{Mode: r.Mode, Time: r.Time, Laps: r.Laps, Opponents: r.Opponents, Difficulty: r.RaceDifficulty040, Epoch: q.Epoch, Countdown: r.Countdown, RaceTime: r.RaceTime, Hour: sceneHour(), Rain: weather050.Rain}
	var ai, traffic []lan.Car
	if s.IsHost() {
		for _, a := range r.Racers {
			c := carToLAN050(a.Car, a.Class, (a.Class*3+2)%len(paintPalette))
			c.Lap = a.Lap
			c.Gate = a.NextGate
			c.Finished = a.Finished
			ai = append(ai, c)
		}
		for _, i := range selectTraffic051(w, q.View.Players) {
			a := w.Traffic[i]
			traffic = append(traffic, lan.Car{Entity051: i + 1, X: a.Pos.X, Y: w.GroundHeight(a.Pos) + .74, Z: a.Pos.Z, H: a.Heading, S: a.Speed, Steer: a.SteerAngle, Wheel: a.WheelSpin, Vehicle: a.Variant, Brake: a.Braking})
		}
	}
	c.Model = w.Car.Variant050
	c.Tyre = w.Car.Tyre050
	c.Rim = setup050().Wheel
	c.Livery = setup050().Livery
	s.Publish(c, meta, ai, traffic)
}
func drawLANPlayers050(shadow bool, radius float64) {
	q := &multiplayer050
	if q.Session == nil {
		return
	}
	for i, p := range q.View.Players {
		if p.ID == q.Session.ID() {
			continue
		}
		c, ok := q.Remote[p.ID]
		if !ok || !p.Ready || c.Pos.Dist(app.renderCar.Pos) > radius {
			continue
		}
		old := beta.paint
		beta.paint = p.State.Paint % len(paintPalette)
		temporal038.Entity = 128 + i
		drawRemasterCar(c, engine.ClampInt(p.State.Vehicle, 0, 3), false)
		temporal038.Entity = 0
		beta.paint = old
	}
}

var lanButtons050 []button020

func drawLANOverlay050() {
	if qa.enabled && qa.stage != 203 || remaster.Loading || !remaster.Ready {
		return
	}
	q := &multiplayer050
	if !q.Open && q.Session == nil {
		return
	}
	remasterUIBegin()
	defer en(GL_DEPTH_TEST)
	w, h := float64(app.width), float64(app.height)
	if !q.Open {
		remasterText(fmt.Sprintf("LAN / %d PLAYERS / F8 LOBBY", len(q.View.Players)), w*.37, h*.025, h*.010, cyan)
		vp := motion034.ViewProjection
		for _, p := range q.View.Players {
			if p.ID == q.Session.ID() || !p.Ready {
				continue
			}
			c, ok := q.Remote[p.ID]
			if !ok {
				continue
			}
			x, y, z := float32(c.Pos.X), float32(c.Body.Height+1.8), float32(c.Pos.Z)
			cw := vp[3]*x + vp[7]*y + vp[11]*z + vp[15]
			if cw <= 0 {
				continue
			}
			sx := (float64((vp[0]*x+vp[4]*y+vp[8]*z+vp[12])/cw)*.5 + .5) * w
			sy := (.5 - float64((vp[1]*x+vp[5]*y+vp[9]*z+vp[13])/cw)*.5) * h
			if sx > 0 && sx < w && sy > 0 && sy < h {
				label := lanPlayerLabel053(p, false)
				remasterText(label, sx-remasterTextWidth(label, h*.012)/2, sy, h*.012, cyan)
			}
		}
		return
	}
	lanButtons050 = nil
	hudRect(w*.08, h*.07, w*.84, h*.86, .015, .025, .032, .98)
	remasterText("LOCAL MULTIPLAYER / 0.5.3", w*.11, h*.105, h*.023, white)
	headline035("F8 OR ESC: CLOSE LOBBY / YOUR PROFILE NAME IS SHOWN TO FRIENDS", w*.11, h*.16, h*.010, w*.77, mutedRefine)
	button := func(id, label string, x, y, bw float64) {
		lanButtons050 = append(lanButtons050, button020{id, w * x, h * y, w * bw, h * .045})
		roundRect032(w*x, h*y, w*bw, h*.045, 3, [4]float64{.12, .18, .19, 1})
		headline035(label, w*(x+.008), h*(y+.014), h*.011, w*(bw-.016), cyan)
	}
	button("host", "CREATE ROOM", .11, .21, .20)
	button("discover", "FIND ROOMS", .33, .21, .20)
	button("leave", "LEAVE ROOM", .55, .21, .20)
	remasterText("HOST IP: "+q.Address+"_", w*.11, h*.285, h*.014, white)
	button("join", "JOIN", .76, .27, .12)
	row := .345
	if q.Session != nil {
		headline035(q.Session.Status(), w*.11, h*.33, h*.010, w*.77, cyan)
		row = .38
		for _, p := range q.View.Players {
			label := lanPlayerLabel053(p, q.Session.IsHost() && p.ID == q.Session.ID())
			if p.ID == q.Session.ID() {
				label += " / YOU"
			}
			if !p.Ready {
				label += " / LOADING"
			}
			headline035(label, w*.12, h*row, h*.013, w*.55, white)
			if q.Session.IsHost() && p.ID != q.Session.ID() {
				button("kick:"+p.ID, "REMOVE", .73, row-.009, .14)
			}
			row += .052
		}
		if q.Session.IsHost() {
			for _, p := range q.Session.Pending() {
				headline035(p.Name+" WANTS TO JOIN", w*.12, h*row, h*.010, w*.47, mutedRefine)
				button("add:"+p.ID, "ADD PLAYER", .58, row-.009, .16)
				button("kick:"+p.ID, "DECLINE", .76, row-.009, .12)
				row += .052
				if row > .70 {
					break
				}
			}
			button("free", "START FREE ROAM", .11, .73, .24)
			button("race", "START RACE", .37, .73, .22)
			button("ai-", "- AI", .61, .73, .09)
			button("ai+", "+ AI", .72, .73, .09)
			button("difficulty", []string{"EASY", "NORMAL", "HARD"}[engine.ClampInt(beta.cfg.RaceDifficulty040, 0, 2)], .61, .79, .20)
			remasterText(fmt.Sprintf("AI OPPONENTS: %d", remaster.Opponents), w*.12, h*.80, h*.012, white)
		}
	} else {
		for i, r := range q.Rooms {
			button(fmt.Sprintf("room:%d", i), r.Name+" / "+r.Address, .12, row, .74)
			row += .055
			if row > .70 {
				break
			}
		}
	}
	headline035(q.Message, w*.11, h*.86, h*.010, w*.77, mutedRefine)
	if q.Session != nil && q.Session.IsHost() {
		ips, _ := net.InterfaceAddrs()
		var labels []string
		for _, a := range ips {
			if ip, ok := a.(*net.IPNet); ok && ip.IP.To4() != nil && !ip.IP.IsLoopback() {
				labels = append(labels, ip.IP.String())
			}
		}
		headline035("HOST IP / "+strings.Join(labels, " / "), w*.11, h*.89, h*.009, w*.77, mutedRefine)
	}
}

func resolveLANContacts050() {
	q := &multiplayer050
	if q.Session == nil || app.world == nil {
		return
	}
	for _, p := range q.View.Players {
		if p.ID == q.Session.ID() || !p.Ready {
			continue
		}
		if c, ok := q.Remote[p.ID]; ok {
			app.world.ResolveNetworkContact050(c, p.State.Vehicle)
		}
	}
}
func updateLANRace050(dt float64) {
	q := &multiplayer050
	if q.Session == nil || !q.Session.IsHost() || app.world == nil || app.world.Remaster.Mode != 1 {
		return
	}
	r := app.world.Remaster
	if r.Countdown > 0 {
		return
	}
	if r.Finished {
		r.RaceTime += dt
	}
	if q.Race == nil {
		q.Race = map[string]*racePeer050{}
	}
	for _, p := range q.View.Players {
		if p.ID == q.Session.ID() || !p.Ready {
			continue
		}
		a := q.Race[p.ID]
		if a == nil {
			a = &racePeer050{Lap: 1}
			q.Race[p.ID] = a
		}
		if !a.Finished && app.world.NetworkGatePassed050(engine.Vec{X: p.State.X, Z: p.State.Z}, a.Gate) {
			if a.Gate == 0 {
				if a.Crossed {
					a.Lap++
					if a.Lap > r.Laps {
						a.Finished = true
						a.Time = r.RaceTime
					}
				}
				a.Crossed = true
			}
			a.Gate = (a.Gate + 1) % len(r.Gates)
		}
		q.Session.SetRace(p.ID, a.Lap, a.Gate, a.Finished, a.Time)
	}
}

// Extrapolate only a short, bounded window from the latest received transform.
// The sample age advances locally between packets, removing 20 Hz stair steps.
func predictLANState051(c lan.Car, age float64) lan.Car {
	age = engine.Clamp(age, 0, .18)
	f := engine.HeadingVec(c.H)
	right := engine.Vec{X: f.Z, Z: -f.X}
	c.X += (f.X*c.S + right.X*c.L) * age
	c.Z += (f.Z*c.S + right.Z*c.L) * age
	return c
}

func lanPlayerLabel053(p lan.Player, localHost bool) string {
	if localHost {
		return p.Name + "  HOST"
	}
	if p.Ping <= 0 {
		return p.Name + "  PING UNAVAILABLE"
	}
	return fmt.Sprintf("%s  %d MS RTT", p.Name, p.Ping)
}

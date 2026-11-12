//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"speedinneed/engine"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

type waveFormat030 struct {
	Tag, Channels      uint16
	Rate, Bytes        uint32
	Align, Bits, Extra uint16
}
type waveHeader030 struct {
	Data             uintptr
	Length, Recorded uint32
	User             uintptr
	Flags, Loops     uint32
	Next, Reserved   uintptr
}
type audioEvent030 struct {
	Name string
	Gain float64
}

var audio030 struct {
	mu                  sync.Mutex
	Frame               audioFrame030
	Events              chan audioEvent030
	Stop                chan struct{}
	Done                chan struct{}
	Started             bool
	DeviceResult        uint32
	Samples             atomic.Uint64
	PreviousGear        int
	PreviousDamage      float64
	PreviousBoost       bool
	AmbienceAt          time.Time
	Coast, City, Nature float64
}
var winmm030 = syscall.NewLazyDLL("winmm.dll")

func playAudio030(name string, gain float64) {
	if !audio030.Started {
		return
	}
	select {
	case audio030.Events <- audioEvent030{name, gain}:
	default:
	}
}
func cycleRadio030() {
	beta.cfg.RadioStation = (beta.cfg.RadioStation + 1) % len(radioNames030)
	playAudio030("radio_tune", .7)
	saveCfg()
}
func startAudio030() {
	if audio030.Started || qa.enabled && os.Getenv("SPEED_QA_AUDIO") != "1" {
		return
	}
	mixer := &audioMixer030{Clips: map[string][]float32{}}
	for _, name := range []string{"engine0", "engine1", "engine2", "engine3", "idle0", "idle1", "idle2", "idle3", "skid", "surf", "city", "nature", "gear", "impact", "nitro", "ui_move", "ui_confirm", "radio_tune", "radio0", "radio1", "menu"} {
		folder, file := "assets/audio032/", name
		if name == "radio0" || name == "radio1" || name == "menu" {
			folder = "assets/audio041/"
		}
		if name == "ui_move" || name == "ui_confirm" {
			folder, file = "assets/audio041/", "click"
		}
		raw, err := assets.ReadFile(folder + file + ".wav")
		if err != nil {
			pbrErrors = append(pbrErrors, "audio: "+name)
			continue
		}
		samples, err := decodeWAV030(raw)
		if err != nil {
			pbrErrors = append(pbrErrors, "audio: "+err.Error())
			continue
		}
		mixer.Clips[name] = samples
	}
	mixer.Clips["horn"] = hornSamples036()
	audio030.Events = make(chan audioEvent030, 32)
	audio030.Stop = make(chan struct{})
	audio030.Done = make(chan struct{})
	audio030.Started = true
	go runAudio030(mixer)
}
func stopAudio030() {
	if !audio030.Started {
		return
	}
	close(audio030.Stop)
	<-audio030.Done
	audio030.Started = false
}
func runAudio030(mixer *audioMixer030) {
	defer close(audio030.Done)
	var device uintptr
	format := waveFormat030{Tag: 1, Channels: 2, Rate: audioRate030, Bytes: audioRate030 * 4, Align: 4, Bits: 16}
	result, _, _ := winmm030.NewProc("waveOutOpen").Call(uintptr(unsafe.Pointer(&device)), uintptr(^uint32(0)), uintptr(unsafe.Pointer(&format)), 0, 0, 0)
	atomic.StoreUint32(&audio030.DeviceResult, uint32(result))
	if result != 0 {
		return
	} // Audio-less machines still run the game.
	var buffers [4][1024]int16
	var headers [4]waveHeader030
	var pins runtime.Pinner
	defer pins.Unpin()
	for i := range headers {
		pins.Pin(&buffers[i][0])
		pins.Pin(&headers[i])
		headers[i].Data = uintptr(unsafe.Pointer(&buffers[i][0]))
		headers[i].Length = uint32(len(buffers[i]) * 2)
		r, _, _ := winmm030.NewProc("waveOutPrepareHeader").Call(device, uintptr(unsafe.Pointer(&headers[i])), unsafe.Sizeof(headers[i]))
		if r != 0 {
			atomic.StoreUint32(&audio030.DeviceResult, uint32(r))
			winmm030.NewProc("waveOutClose").Call(device)
			return
		}
		headers[i].Flags |= 1
	}
	defer func() {
		winmm030.NewProc("waveOutReset").Call(device)
		for i := range headers {
			winmm030.NewProc("waveOutUnprepareHeader").Call(device, uintptr(unsafe.Pointer(&headers[i])), unsafe.Sizeof(headers[i]))
		}
		winmm030.NewProc("waveOutClose").Call(device)
	}()
	ticker := time.NewTicker(4 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-audio030.Stop:
			return
		case <-ticker.C:
		}
		audio030.mu.Lock()
		frame := audio030.Frame
		audio030.mu.Unlock()
		for len(audio030.Events) > 0 {
			e := <-audio030.Events
			if len(mixer.Voices) < 16 {
				mixer.Voices = append(mixer.Voices, audioVoice030{Name: e.Name, Gain: e.Gain})
			}
		}
		for i := range headers {
			if atomic.LoadUint32(&headers[i].Flags)&1 == 0 {
				continue
			}
			mixer.render(buffers[i][:], frame)
			r, _, _ := winmm030.NewProc("waveOutWrite").Call(device, uintptr(unsafe.Pointer(&headers[i])), unsafe.Sizeof(headers[i]))
			if r != 0 {
				atomic.StoreUint32(&audio030.DeviceResult, uint32(r))
				return
			}
			audio030.Samples.Add(uint64(len(buffers[i]) / 2))
		}
	}
}
func updateAudio030() {
	if !audio030.Started {
		return
	}
	cfg := beta.cfg
	frame := audioFrame030{Master: cfg.MasterVolume, Effects: cfg.EffectsVolume, Music: cfg.MusicVolume, Ambience: cfg.AmbienceVolume, Muted: cfg.AudioMuted, Station: cfg.RadioStation, Car: beta.carSelect}
	if app.world != nil && app.world.Remaster != nil && app.started {
		c := app.world.Car
		frame.Driving = true
		frame.Paused = beta.screen != screenDrive || remaster.FullMap
		frame.RPM = c.Gearbox.RPM
		frame.EngineOff = !c.EngineRunning
		if pressed(0) && !frame.Paused {
			frame.Throttle = 1
		}
		if !frame.Paused && beta.cfg.GamepadEnabled && pad033.State.Connected {
			padInput := mergeGamepad033(engine.Input{})
			frame.Throttle = math.Max(frame.Throttle, math.Max(padInput.Throttle, padInput.Brake))
		}
		frame.Speed = math.Abs(c.Speed)
		frame.Car = app.world.CarClass
		frame.Gravel = c.OffroadBlend
		frame.Slip = math.Min(1, math.Abs(c.Lateral)/7) * math.Min(1, frame.Speed/8)
		if c.DriftActive {
			frame.Slip = math.Max(frame.Slip, .65)
		}
		if c.Burnout {
			frame.Slip = math.Max(frame.Slip, .8)
		}
		if c.Boosting {
			frame.Boost = 1
		}
		if time.Since(audio030.AmbienceAt) > 180*time.Millisecond {
			audio030.Coast, audio030.City, audio030.Nature = ambienceGainsRemastered(c.Pos, app.world.Remaster.Def)
			audio030.AmbienceAt = time.Now()
		}
		frame.Coast, frame.City, frame.Nature = audio030.Coast, audio030.City, audio030.Nature
		if app.world.InsideTunnel(c.Pos) && c.Body.Height < app.world.Remaster.Def.Grid.Height(c.Pos)-1 {
			frame.Tunnel = 1
		}
		if !frame.Paused {
			if audio030.PreviousGear != 0 && c.Gearbox.Gear != audio030.PreviousGear {
				playAudio030("gear", .7)
			}
			if c.Damage > audio030.PreviousDamage+.06 {
				playAudio030("impact", math.Min(1, .3+(c.Damage-audio030.PreviousDamage)*.1))
			}
			if c.Boosting && !audio030.PreviousBoost {
				playAudio030("nitro", .7)
			}
		}
		audio030.PreviousGear = c.Gearbox.Gear
		audio030.PreviousDamage = c.Damage
		audio030.PreviousBoost = c.Boosting
		nearest := 50.
		for _, car := range app.world.Traffic {
			dist := car.Pos.Dist(c.Pos)
			if dist >= nearest {
				continue
			}
			nearest = dist
			frame.NPCGain = .15 * (1 - dist/50)
			frame.NPCPitch = .65 + math.Abs(car.Speed)*.045
			frame.NPCPan = math.Max(-1, math.Min(1, (car.Pos.X-c.Pos.X)/math.Max(1, dist)))
		}
	}
	audio030.mu.Lock()
	audio030.Frame = frame
	audio030.mu.Unlock()
}

// Coast is sampled from actual surrounding water, city follows authored town
// roads, and countryside fills the remaining space. Weights crossfade gradually.
func ambienceGainsRemastered(p engine.Vec, d *engine.WorldDefinition) (coast, city, nature float64) {
	for _, distance := range []float64{18, 42, 75, 110} {
		for i := 0; i < 8; i++ {
			a := float64(i) * math.Pi / 4
			q := p.Add(engine.Vec{X: math.Cos(a) * distance, Z: math.Sin(a) * distance})
			if d.WaterGrid.Height(q) > .5 || q.X < -1490 || q.X > 1490 || q.Z < -1490 || q.Z > 1490 {
				coast = math.Max(coast, 1-distance/140)
			}
		}
	}
	nearest := 150.
	for _, road := range d.Roads {
		if road.Kind != "town" {
			continue
		}
		for _, q := range road.Points {
			nearest = math.Min(nearest, p.Dist(q))
		}
	}
	city = engine.Clamp(1-nearest/110, 0, 1) * (1 - coast*.6)
	nature = math.Max(.08, 1-math.Max(city, coast)) * .65
	return
}
func writeAudioReport030() {
	if !qa.enabled {
		return
	}
	result := map[string]any{"device_result": atomic.LoadUint32(&audio030.DeviceResult), "frames_submitted": audio030.Samples.Load(), "original_radio_tracks": 4, "enabled": audio030.Started}
	raw, _ := json.MarshalIndent(result, "", "  ")
	path := filepath.Join(qa.dir, "AUDIO_030.json")
	if err := os.WriteFile(path, raw, 0644); err != nil {
		fmt.Println(err)
	}
}

func drawRadio030() {
	w, h := float64(app.width), float64(app.height)
	station := max(0, min(len(radioNames030)-1, beta.cfg.RadioStation))
	beta021Panel(w*.027, h*.035, w*.30, h*.091, .74)
	remasterText("RADIO", w*.04, h*.049, h*.012, cyan)
	remasterText(radioNames030[station], w*.04, h*.077, h*.014, white)
	remasterText("F  NEXT STATION / OFF", w*.04, h*.106, h*.009, white)
}

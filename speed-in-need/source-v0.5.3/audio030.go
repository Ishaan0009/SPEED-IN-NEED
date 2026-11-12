package main

import (
	"encoding/binary"
	"errors"
	"math"
)

const audioRate030 = 22050

var radioNames030 = [...]string{"RADIO OFF", "01 / MANUEL - GAS GAS GAS", "02 / RUNNING IN THE 90S"}

type audioFrame030 struct {
	Master, Effects, Music, Ambience                float64
	RPM, Speed, Slip, Gravel, Boost, Forest, Tunnel float64
	Coast, City, Nature, Throttle                   float64
	NPCGain, NPCPitch, NPCPan                       float64
	Car, Station                                    int
	Driving, Paused, Muted                          bool
	EngineOff                                       bool
}
type audioVoice030 struct {
	Name string
	At   int
	Gain float64
}
type audioMixer030 struct {
	Clips                   map[string][]float32
	Voices                  []audioVoice030
	Phase                   [10]float64
	MusicPhase              float64
	Station                 int
	Gain                    float64
	Delay                   [4096]float64
	DelayAt                 int
	EnginePitch, EngineLoad float64
}

func decodeWAV030(raw []byte) ([]float32, error) {
	if len(raw) < 44 || string(raw[:4]) != "RIFF" || string(raw[8:12]) != "WAVE" {
		return nil, errors.New("invalid WAV header")
	}
	format, channels, rate, bits := uint16(0), uint16(0), uint32(0), uint16(0)
	var pcm []byte
	for off := 12; off+8 <= len(raw); {
		n := int(binary.LittleEndian.Uint32(raw[off+4:]))
		start := off + 8
		if n < 0 || n > len(raw)-start {
			return nil, errors.New("truncated WAV chunk")
		}
		switch string(raw[off : off+4]) {
		case "fmt ":
			if n < 16 {
				return nil, errors.New("invalid PCM format")
			}
			format = binary.LittleEndian.Uint16(raw[start:])
			channels = binary.LittleEndian.Uint16(raw[start+2:])
			rate = binary.LittleEndian.Uint32(raw[start+4:])
			bits = binary.LittleEndian.Uint16(raw[start+14:])
		case "data":
			pcm = raw[start : start+n]
		}
		off = start + n + (n & 1)
	}
	if format != 1 || channels != 1 || rate != audioRate030 || bits != 16 || len(pcm) < 4 || len(pcm)%2 != 0 {
		return nil, errors.New("expected mono 22050 Hz signed 16-bit PCM")
	}
	samples := make([]float32, len(pcm)/2)
	for i := range samples {
		samples[i] = float32(int16(binary.LittleEndian.Uint16(pcm[i*2:]))) / 32768
	}
	return samples, nil
}
func (m *audioMixer030) loop(name string, phase *float64, pitch float64) float64 {
	samples := m.Clips[name]
	if len(samples) < 2 {
		return 0
	}
	*phase = math.Mod(*phase, float64(len(samples)))
	i := int(*phase)
	f := *phase - float64(i)
	j := (i + 1) % len(samples)
	value := float64(samples[i])*(1-f) + float64(samples[j])*f
	*phase += math.Max(.05, pitch)
	return value
}
func (m *audioMixer030) render(out []int16, frame audioFrame030) {
	station := frame.Station
	if !frame.Driving {
		station = -1
	}
	if station < -1 || frame.Driving && station < 0 || station >= len(radioNames030) {
		station = 0
	}
	if station != m.Station {
		m.Station = station
		m.MusicPhase = 0
	}
	engineName := [4]string{"engine0", "engine1", "engine2", "engine3"}[max(0, min(3, frame.Car))]
	musicName := "menu"
	if frame.Driving {
		musicName = [3]string{"", "radio0", "radio1"}[station]
	}
	target := frame.Master
	if frame.Muted {
		target = 0
	}
	effects := frame.Effects
	if frame.Paused {
		effects *= .20
	}
	drive := 0.
	if frame.Driving {
		drive = 1
	}
	for i := 0; i+1 < len(out); i += 2 {
		m.Gain += (target - m.Gain) * .002
		load := math.Max(0, math.Min(1, frame.Throttle))
		rpm := math.Max(0, math.Min(1, (frame.RPM-900)/7000))
		m.EnginePitch += (.78 + .58*rpm - m.EnginePitch) * .00065
		m.EngineLoad += (load - m.EngineLoad) * .001
		idleName := [4]string{"idle0", "idle1", "idle2", "idle3"}[max(0, min(3, frame.Car))]
		idle := m.loop(idleName, &m.Phase[1], .88+.20*rpm)
		engine := (m.loop(engineName, &m.Phase[0], m.EnginePitch)*(.12+.29*m.EngineLoad)*(.25+.75*rpm) + idle*.17*(1-rpm)) * drive
		if frame.EngineOff {
			engine = 0
		}
		road := m.loop("skid", &m.Phase[3], .85+math.Min(.5, frame.Speed*.008)) * frame.Slip * drive * .35
		ambient := m.loop("nature", &m.Phase[4], 1) * frame.Nature
		ambient += m.loop("surf", &m.Phase[5], 1) * frame.Coast
		ambient += m.loop("city", &m.Phase[7], 1) * frame.City
		ambient *= (1 - .85*frame.Tunnel) * drive * .65
		npc := m.loop("engine1", &m.Phase[6], .82) * frame.NPCGain * drive * .25
		sound := (engine+road)*effects + ambient*frame.Ambience
		music := m.loop(musicName, &m.MusicPhase, 1) * frame.Music
		if !frame.Driving {
			music *= .36
		} else {
			music *= .22
		}
		music *= 1 - .35*frame.Boost
		for j := range m.Voices {
			voice := &m.Voices[j]
			samples := m.Clips[voice.Name]
			if voice.At < len(samples) {
				sound += float64(samples[voice.At]) * voice.Gain * effects * .65
				voice.At++
			}
		}
		delayed := m.Delay[m.DelayAt]
		m.Delay[m.DelayAt] = sound + delayed*.25*frame.Tunnel
		m.DelayAt = (m.DelayAt + 1) % len(m.Delay)
		sound += delayed * .26 * frame.Tunnel
		left := (sound + music + npc*effects*(1-frame.NPCPan*.65)) * m.Gain
		right := (sound + music + npc*effects*(1+frame.NPCPan*.65)) * m.Gain
		out[i] = int16(math.Tanh(left) * 30000)
		out[i+1] = int16(math.Tanh(right) * 30000)
	}
	kept := m.Voices[:0]
	for _, v := range m.Voices {
		if v.At < len(m.Clips[v.Name]) {
			kept = append(kept, v)
		}
	}
	m.Voices = kept
}

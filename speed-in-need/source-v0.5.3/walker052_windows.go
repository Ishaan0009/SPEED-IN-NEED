package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
)

var walk052 struct {
	Tried                     bool
	Vertices, Indices, Frames int
	Stride                    float64
	UV, Data                  []float32
	Index                     []uint32
}

func loadWalk052() {
	if walk052.Tried {
		return
	}
	walk052.Tried = true
	raw, err := assets.ReadFile("assets/characters/walker052.walk")
	fail := func() { pbrErrors = append(pbrErrors, "0.5.3 walking clip could not be loaded") }
	if err != nil || len(raw) < 8 || string(raw[:4]) != "WAL3" {
		fail()
		return
	}
	size := int(binary.LittleEndian.Uint32(raw[4:]))
	if size < 2 || 8+size > len(raw) {
		fail()
		return
	}
	var m struct {
		Vertices, Indices, Frames int
		Stride                    float64
	}
	if json.Unmarshal(raw[8:8+size], &m) != nil || m.Vertices < 3 || m.Vertices > 100000 || m.Indices < 3 || m.Indices > 600000 || m.Frames < 2 || m.Frames > 120 || m.Stride < .2 {
		fail()
		return
	}
	if 8+size+(m.Vertices*2+m.Indices+m.Vertices*6*m.Frames)*4 != len(raw) {
		fail()
		return
	}
	at := 8 + size
	floats := func(n int) []float32 {
		out := make([]float32, n)
		for i := range out {
			out[i] = math.Float32frombits(binary.LittleEndian.Uint32(raw[at:]))
			at += 4
		}
		return out
	}
	walk052.UV = floats(m.Vertices * 2)
	walk052.Index = make([]uint32, m.Indices)
	for i := range walk052.Index {
		id := binary.LittleEndian.Uint32(raw[at:])
		at += 4
		if int(id) >= m.Vertices {
			fail()
			walk052.Index = nil
			return
		}
		walk052.Index[i] = id
	}
	walk052.Data = floats(m.Vertices * 6 * m.Frames)
	walk052.Vertices = m.Vertices
	walk052.Indices = m.Indices
	walk052.Frames = m.Frames
	walk052.Stride = m.Stride
}
func drawWalkClip052() bool {
	loadWalk052()
	if len(walk052.Data) == 0 {
		return false
	}
	p := walker051.Ped
	frame := 0.
	if p.Moving {
		frame = math.Mod(p.Phase/5.5/walk052.Stride, 1) * float64(walk052.Frames-1)
	}
	a, b, blend := 0, 0, float32(0)
	if p.Moving {
		a, b, blend = walkFrameIndices052(frame, walk052.Frames-1)
	}
	ca, sa := float32(math.Cos(p.Heading+math.Pi)), float32(math.Sin(p.Heading+math.Pi))
	batch := &characters032.Batches[4]
	batch.Vertices = batch.Vertices[:0]
	batch.Indices = append(batch.Indices[:0], walk052.Index...)
	batch.Previous = batch.Previous[:0]
	for i := 0; i < walk052.Vertices; i++ {
		aa, bb := (a*walk052.Vertices+i)*6, (b*walk052.Vertices+i)*6
		var v [6]float32
		for k := range v {
			v[k] = walk052.Data[aa+k]*(1-blend) + walk052.Data[bb+k]*blend
			if p.Moving {
				weight := float32(math.Min(1, p.Speed/.8))
				v[k] = walk052.Data[i*6+k]*(1-weight) + v[k]*weight
			}
		}
		batch.Vertices = append(batch.Vertices, [12]float32{float32(p.Pos.X) + ca*v[0] + sa*v[2], float32(p.Height) + v[1], float32(p.Pos.Z) - sa*v[0] + ca*v[2], ca*v[3] + sa*v[5], v[4], -sa*v[3] + ca*v[5], walk052.UV[i*2], walk052.UV[i*2+1], 1, 0, 0, 1})
	}
	if hdrScene.motion != 0 && hdrScene.active && !sunShadow.drawing {
		batch.MotionScratch = batch.MotionScratch[:0]
		for _, v := range batch.Vertices {
			batch.MotionScratch = append(batch.MotionScratch, [3]float32{v[0], v[1], v[2]})
		}
		batch.Previous = append(batch.Previous, previousSkin038(fmt.Sprintf("walker052/%d", walk052.Vertices), batch.MotionScratch)...)
	}
	flushCharacter032(4)
	return true
}

// Frame zero is a separate relaxed idle; walking frames occupy 1..count.
func walkFrameIndices052(frame float64, count int) (int, int, float32) {
	f := math.Mod(frame, float64(count))
	if f < 0 {
		f += float64(count)
	}
	whole := int(math.Floor(f))
	return 1 + whole, 1 + (whole+1)%count, float32(f - float64(whole))
}

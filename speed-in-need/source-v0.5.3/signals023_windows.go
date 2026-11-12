//go:build windows

package main

import (
	"math"
	"speedinneed/engine"
)

var signalDefinition023 *engine.WorldDefinition
var signalPositions023 []engine.Vec
var signalIndices023 map[[3]float64]int

func signal023State(o engine.PlacedObject) int {
	if o.Kind != "signal" {
		return 0
	}
	d := app.world.Remaster.Def
	if signalDefinition023 != d {
		signalDefinition023 = d
		signalIndices023 = map[[3]float64]int{}
		signalPositions023 = nil
		for _, s := range d.Signals {
			_, width, _ := app.world.RemasteredRoadAt(s.Pos)
			f := engine.HeadingVec(s.Heading)
			signalPositions023 = append(signalPositions023, s.Pos.Add(engine.Vec{X: f.Z, Z: -f.X}.Mul(width*.5+3.5)))
		}
	}
	best, index := math.Inf(1), -1
	if cached, ok := signalIndices023[o.Pos]; ok {
		index = cached
	} else {
		for i, p := range signalPositions023 {
			distance := p.Dist(engine.Vec{X: o.Pos[0], Z: o.Pos[2]})
			if distance < best {
				best, index = distance, i
			}
		}
		signalIndices023[o.Pos] = index
	}
	if index >= 0 && app.world.RemasteredSignalRed(app.world.Remaster.Def.Signals[index]) {
		return 1
	}
	return 2
}

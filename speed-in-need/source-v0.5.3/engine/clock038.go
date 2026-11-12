package engine

import "math"

// FixedClock038 advances simulation time independently of rendered/presented
// frames. A short hitch catches up at 120 Hz; suspend-sized gaps are bounded.
// It never advances physics for an interpolated or generated display frame.
type FixedClock038 struct {
	Accumulator, DroppedSeconds float64
	Ticks                       uint64
}

const PhysicsStep038 = 1. / 120

func (c *FixedClock038) Advance(elapsed float64, step func(float64)) int {
	if elapsed <= 0 || math.IsNaN(elapsed) || math.IsInf(elapsed, 0) {
		return 0
	}
	if elapsed > .25 {
		c.DroppedSeconds += elapsed - .25
		elapsed = .25
	}
	c.Accumulator += elapsed
	n := 0
	for c.Accumulator+1e-12 >= PhysicsStep038 && n < 32 {
		step(PhysicsStep038)
		c.Accumulator = math.Max(0, c.Accumulator-PhysicsStep038)
		c.Ticks++
		n++
	}
	return n
}

func (c *FixedClock038) Alpha() float64 { return Clamp(c.Accumulator/PhysicsStep038, 0, 1) }
func (c *FixedClock038) Pause()         { c.Accumulator = 0 }

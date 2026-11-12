package main

import (
	"math"
	"time"
)

// No GPU model names or vendor-specific APIs enter the quality policy.
// Runtime factors are separate from the user's saved quality ceiling.
type qualityGovernor026 struct {
	Scale, Foliage, Scene              float64
	warmup, window, gpu, cpu, recovery float64
	frames, overloaded                 int
}

// Without GPU queries, use unslept whole-frame work as a conservative
// fallback. With queries, keep submission cost and GPU execution separate.
func qualityCosts026(timed bool, gpu, cpu, wholeFrame float64) (float64, float64) {
	if timed {
		return gpu, cpu
	}
	return math.Max(cpu, wholeFrame), 0
}

func (q *qualityGovernor026) Reset() { *q = qualityGovernor026{Scale: 1, Foliage: 1, Scene: 1} }

func (q *qualityGovernor026) Sample(dt, gpuMS, cpuMS, targetMS float64) {
	if dt <= 0 || dt > .25 || targetMS <= 0 || math.IsNaN(dt) || math.IsNaN(gpuMS) || math.IsNaN(cpuMS) || math.IsInf(gpuMS, 0) || math.IsInf(cpuMS, 0) {
		return
	}
	if q.Scale == 0 {
		q.Reset()
	}
	q.warmup += dt
	if q.warmup < 3 {
		return
	}
	q.window += dt
	q.gpu += math.Max(0, gpuMS)
	q.cpu += math.Max(0, cpuMS)
	q.frames++
	if q.window < 1 || q.frames < 12 {
		return
	}
	gpu, cpu := q.gpu/float64(q.frames), q.cpu/float64(q.frames)
	span := q.window
	q.window, q.gpu, q.cpu, q.frames = 0, 0, 0, 0
	if math.Max(gpu, cpu) > targetMS*1.08 {
		q.overloaded++
		q.recovery = 0
		if q.overloaded < 2 {
			return
		}
		q.overloaded = 0
		if gpu > targetMS*.9 && gpu >= cpu*.85 && q.Scale > .65 {
			q.Scale = math.Max(.65, math.Round((q.Scale-.05)*20)/20)
		} else if q.Foliage > .55 {
			q.Foliage = math.Max(.55, math.Round((q.Foliage-.1)*20)/20)
		} else {
			q.Scene = math.Max(.65, math.Round((q.Scene-.1)*20)/20)
		}
		return
	}
	q.overloaded = 0
	if math.Max(gpu, cpu) < targetMS*.72 {
		q.recovery += span
		if q.recovery >= 10 {
			q.recovery = 0
			if q.Scene < 1 {
				q.Scene = math.Min(1, q.Scene+.05)
			} else if q.Foliage < 1 {
				q.Foliage = math.Min(1, q.Foliage+.05)
			} else {
				q.Scale = math.Min(1, q.Scale+.05)
			}
		}
	} else {
		q.recovery = 0
	}
}

type memoryBudget026 struct{ HeapMB, TextureCap int }

func memoryPolicy026(total, available uint64) memoryBudget026 {
	if total == 0 {
		return memoryBudget026{768, 8192}
	}
	b := memoryBudget026{1024, 8192}
	if total <= 4<<30 {
		b = memoryBudget026{384, 1024}
	} else if total <= 8<<30 {
		b = memoryBudget026{512, 1024}
	} else if total <= 16<<30 {
		b = memoryBudget026{768, 2048}
	}
	if available > 0 {
		share := int(available / (3 << 20))
		if share < b.HeapMB {
			b.HeapMB = max(256, share)
		}
		if available < 1536<<20 {
			b.TextureCap = min(b.TextureCap, 1024)
		}
	}
	return b
}

// Absolute deadlines keep sleep overshoot from accumulating across frames.
// A hitch starts a new deadline; there is no busy-spin or burst of catch-up frames.
type framePacer026 struct {
	deadline time.Time
	fps      int
}

func (p *framePacer026) Next(start time.Time, fps int) time.Time {
	if fps <= 0 {
		*p = framePacer026{}
		return start
	}
	period := time.Second / time.Duration(max(1, fps))
	if p.fps != fps || p.deadline.IsZero() {
		p.deadline = start.Add(period)
	} else {
		p.deadline = p.deadline.Add(period)
		if p.deadline.Before(start) {
			p.deadline = start.Add(period)
		}
	}
	p.fps = fps
	return p.deadline
}

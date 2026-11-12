package main

import (
	"math"
	"math/rand"
	"testing"
)

func referenceCone035(x, z, r float64, eye, target pos3, width, height int) bool {
	dx, dz := x-eye.x, z-eye.z
	distance := math.Hypot(dx, dz)
	if distance < r+40 {
		return true
	}
	fx, fz := target.x-eye.x, target.z-eye.z
	length := math.Hypot(fx, fz)
	if length < .001 {
		return true
	}
	half := math.Atan(math.Tan(72*math.Pi/360)*float64(width)/math.Max(1, float64(height))) + .12
	margin := math.Asin(math.Min(1, (r+20)/distance))
	return (dx*fx+dz*fz)/(distance*length) > math.Cos(math.Min(math.Pi, half+margin))
}

func TestVisibility035MatchesOriginalCulling(t *testing.T) {
	rng := rand.New(rand.NewSource(350035))
	var cache cone035
	for _, resolution := range [][2]int{{1280, 720}, {1920, 1080}, {2560, 1080}, {7680, 1080}, {10000, 1}, {600, 1200}} {
		for i := 0; i < 20000; i++ {
			eye := pos3{rng.Float64()*3000 - 1500, 24, rng.Float64()*3000 - 1500}
			target := add(eye, pos3{rng.Float64()*400 - 200, -5, rng.Float64()*400 - 200})
			if i%300 == 0 {
				target = eye
			}
			x, z, radius := rng.Float64()*7000-3500, rng.Float64()*7000-3500, rng.Float64()*400
			want := referenceCone035(x, z, radius, eye, target, resolution[0], resolution[1])
			if got := cache.contains(x, z, radius, eye, target, resolution[0], resolution[1]); got != want {
				t.Fatalf("culling changed at resolution%v point(%v,%v) r%v eye%v target%v: got%v want%v", resolution, x, z, radius, eye, target, got, want)
			}
		}
	}
}

var benchmarkConeResult035 bool

func BenchmarkVisibility035(b *testing.B) {
	for _, fast := range []bool{false, true} {
		name := "Original"
		if fast {
			name = "CachedAlgebra"
		}
		b.Run(name, func(b *testing.B) {
			var cache cone035
			eye, target := pos3{100, 45, -300}, pos3{400, 30, -60}
			for i := 0; i < b.N; i++ {
				x, z := float64(i%300)*8-1200, float64(i%271)*8-1084
				if fast {
					benchmarkConeResult035 = cache.contains(x, z, 35, eye, target, 1920, 1080)
				} else {
					benchmarkConeResult035 = referenceCone035(x, z, 35, eye, target, 1920, 1080)
				}
			}
		})
	}
}

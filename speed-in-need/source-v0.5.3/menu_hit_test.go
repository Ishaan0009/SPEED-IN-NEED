package main

import "testing"

func TestEveryMapClickable(t *testing.T) {
	for _, size := range [][2]int{{1280, 720}, {1920, 1080}, {800, 600}} {
		for i := 0; i < 3; i++ {
			x := int(.2 * float64(size[0]))
			y := int((.405 + float64(i)*.135) * float64(size[1]))
			got, ok := nativeMenuHit("maps", x, y, size[0], size[1])
			if !ok || got != i {
				t.Fatalf("map %d at %v: got %d, %v", i, size, got, ok)
			}
		}
	}
	if _, ok := nativeMenuHit("maps", 1200, 200, 1280, 720); ok {
		t.Fatal("menu accepted out-of-panel click")
	}
}

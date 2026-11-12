package main

type normalizedRect struct{ X, Y, W, H float64 }

func (r normalizedRect) Contains(x, y float64) bool {
	return x >= r.X && x <= r.X+r.W && y >= r.Y && y <= r.Y+r.H
}
func settingRect(index int) normalizedRect {
	left := []int{0, 1, 2, 3, 4, 7, 8, 9, 10, 11, 12}
	for row, i := range left {
		if i == index {
			return normalizedRect{.05, .415 + float64(row)*.042, .43, .039}
		}
	}
	for row, i := range []int{5, 6, 13, 14} {
		if i == index {
			return normalizedRect{.545, .415 + float64(row)*.09, .37, .07}
		}
	}
	return normalizedRect{}
}

// nativeMenuHit translates mouse clicks into item indexes using the same
// percentages as the native OpenGL menu, independent of window size.
func nativeMenuHit(screen string, x, y, width, height int) (int, bool) {
	if width <= 0 || height <= 0 || x < 0 || y < 0 {
		return 0, false
	}
	fx, fy := float64(x)/float64(width), float64(y)/float64(height)
	switch screen {
	case "maps":
		if fx < .05 || fx > .46 {
			return 0, false
		}
		for i := 0; i < 3; i++ {
			top := .385 + float64(i)*.135 - .016
			if fy >= top && fy <= top+.10 {
				return i, true
			}
		}
	}
	return 0, false
}

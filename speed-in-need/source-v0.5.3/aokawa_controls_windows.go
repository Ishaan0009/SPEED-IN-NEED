package main

// Look-at uses cross(forward, up): screen left is positive engine yaw.
func aokawaKeyboardSteer(left, right bool) float64 {
	value := 0.
	if left {
		value++
	}
	if right {
		value--
	}
	return value
}

package main

import "math"

// Reuse the horizontal camera basis for all terrain, buildings and vegetation
// in a render pass. This changes neither the original FOV nor its safety margin.
type cone035 struct {
	width, height       int
	fx, fz              float64
	nx, nz              float64
	cosHalf, sinHalf    float64
	ready, validForward bool
}

var visibleConeCache035 cone035

func (c *cone035) contains(x, z, radius float64, eye, target pos3, width, height int) bool {
	dx, dz := x-eye.x, z-eye.z
	distanceSquared := dx*dx + dz*dz
	if distanceSquared < (radius+40)*(radius+40) {
		return true
	}
	fx, fz := target.x-eye.x, target.z-eye.z
	if !c.ready || c.width != width || c.height != height {
		half := math.Atan(math.Tan(72*math.Pi/360)*float64(width)/math.Max(1, float64(height))) + .12
		c.sinHalf, c.cosHalf = math.Sincos(half)
		c.width, c.height = width, height
	}
	if !c.ready || c.fx != fx || c.fz != fz {
		length := math.Hypot(fx, fz)
		c.validForward = length >= .001
		if c.validForward {
			c.nx, c.nz = fx/length, fz/length
		}
		c.fx, c.fz = fx, fz
	}
	c.ready = true
	if !c.validForward {
		return true
	}
	margin := radius + 20
	// d*cos(half+asin(margin/d)) simplifies to the expression below.
	// It avoids per-cell trigonometry and a second distance normalization.
	dot := dx*c.nx + dz*c.nz
	if c.cosHalf < 0 && margin*margin > distanceSquared*c.sinHalf*c.sinHalf {
		return dot > -math.Sqrt(distanceSquared) // original angle clamp at PI
	}
	threshold := c.cosHalf*math.Sqrt(math.Max(0, distanceSquared-margin*margin)) - c.sinHalf*margin
	return dot > threshold
}

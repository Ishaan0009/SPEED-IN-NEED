package main

import (
	"math"
	"speedinneed/engine"
)

// The camera sits at the skinned driver's eyes inside the real car cabin.
func beta021CockpitEye(c engine.Car) (pos3, pos3) {
	id := engine.ClampInt(beta.carSelect, 0, 3)
	if app.world != nil {
		id = engine.ClampInt(app.world.CarClass, 0, 3)
	}
	pose := seatedDriverPose053(id, c.Variant050, c.VisualSteer)
	head := pose[2]
	eye := aokawaToWorld(c, pos3{head.X, head.Y + .10, head.Z + .08})
	yaw := c.Heading + app.camYaw + lookBackYaw036()
	pitch := engine.Clamp(c.Rig.Pitch, -.45, .45)
	target := add(eye, pos3{math.Sin(yaw) * math.Cos(pitch) * 25, -math.Sin(pitch)*25 - app.camPitch*8 - .10, math.Cos(yaw) * math.Cos(pitch) * 25})
	return eye, target
}
func drawRemasterCockpit(c engine.Car) {
	if c.Variant050 > 0 {
		return
	}
	id := engine.ClampInt(app.world.CarClass, 0, 3)
	key := "b033_cabin_" + fmtSpeed021(float64(id))
	if pbrAssets[key] == nil {
		return
	}
	pose := seatedDriverPose032(id, 0)
	x, y, z := pose[0].X, (pose[5].Y+pose[8].Y)*.5, pose[5].Z
	push()
	move(c.Pos.X, c.Body.Height-.74, c.Pos.Z)
	rotate(c.Heading*180/math.Pi, 0, 1, 0)
	rotate(c.Rig.Pitch*180/math.Pi, 1, 0, 0)
	rotate(c.Rig.Roll*180/math.Pi, 0, 0, 1)
	move(x, y, z)
	rotate(steeringWheelAngle035(c.VisualSteer)*180/math.Pi, 0, .25, 1)
	move(-x, -y, -z)
	drawPBRAsset(key, 9)
	pop()
}

package engine

import "math"

// NitroDriveForce038 adds engine torque through the engaged gear and final
// drive. Its power budget falls with road speed; the tire solver then limits
// the combined normal-engine/nitro demand through each wheel's friction ellipse.
func NitroDriveForce038(spec RemasteredVehicle, rpm, ratio, speed, shiftTime float64) float64 {
	p := spec.P
	if rpm <= 0 || rpm > p.RedlineRPM*1.05 || ratio <= 0 {
		return 0
	}
	torque := engineTorque(rpm, p.PeakTorque, p.RedlineRPM) * .70
	force := torque * ratio * p.FinalDrive * .88 / p.WheelRadius
	powerLimit := spec.PowerKW * .55 * 1000 * .88 / math.Max(7, math.Abs(speed))
	force = math.Min(force, math.Min(powerLimit, spec.BoostForce*1.35))
	if shiftTime > 0 {
		force *= .22
	}
	return math.Max(0, force)
}

package engine

import "testing"

func Test050WetTyresAndVariants(t *testing.T) {
	if TyreGrip050(2, 1, 0) <= TyreGrip050(1, 1, 0) {
		t.Fatal("wet tyres must beat dry performance tyres in rain")
	}
	if TyreGrip050(3, 0, 1) <= TyreGrip050(1, 0, 1) {
		t.Fatal("all terrain tyres must improve loose ground")
	}
	for v := 1; v <= 3; v++ {
		s := VehicleVariantSpec050(0, v)
		if s.P.Mass < 1000 || s.PowerKW < 300 || s.TopSpeed < 75 || s.HalfLength < 2 {
			t.Fatal("invalid imported vehicle physics")
		}
	}
}

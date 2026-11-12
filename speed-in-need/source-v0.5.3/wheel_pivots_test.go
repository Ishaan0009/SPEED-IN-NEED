package main

import (
	"math"
	"os"
	"testing"
)

func TestAlpha051PorscheWheelPivotIsNotOrigin(t *testing.T) {
	data, err := os.ReadFile("models/porsche911_wheel_FL.s3d")
	if err != nil {
		t.Fatal(err)
	}
	pivot, ok := s3dWheelPivot(data)
	if !ok {
		t.Fatal("valid wheel rejected")
	}
	if math.Abs(pivot[0]) < .04 || math.Abs(pivot[2]) < .05 {
		t.Fatalf("regressed offset detector: %v", pivot)
	}
	data, err = os.ReadFile("models/challenge296_wheel_FL.s3d")
	if err != nil {
		t.Fatal(err)
	}
	pivot, ok = s3dWheelPivot(data)
	if !ok {
		t.Fatal("ferrari wheel rejected")
	}
	if math.Abs(pivot[0]) > .02 || math.Abs(pivot[1]) > .02 || math.Abs(pivot[2]) > .02 {
		t.Fatalf("already centered wheel should remain centered: %v", pivot)
	}
}
func TestAlpha051PivotBadAssetRejected(t *testing.T) {
	if _, ok := s3dWheelPivot([]byte("BAD!")); ok {
		t.Fatal("accepted bad geometry")
	}
}

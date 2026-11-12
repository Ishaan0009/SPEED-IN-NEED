//go:build windows

package main

import (
	"speedinneed/engine"
	"testing"
)

func Test040FinalMenuInitialIshaanAndCarArrows(t *testing.T) {
	timeProgressFixture040(t)
	p := activeProfile040()
	if p == nil || p.ID != "ishaan" || p.Name != "Ishaan" || p.Level040() != engine.IshaanLevel040 || !p.AllUnlocked040() {
		t.Fatal("initial memory-only QA profile is not the protected Ishaan profile", p)
	}
	for _, id := range []string{"car0", "car1", "car2", "car3", "car+", "car-"} {
		if !allowProgressAction040(id) {
			t.Errorf("protected profile incorrectly blocks %q: %s", id, progress040.Gate)
		}
	}
	oldGarage := garage050
	defer func() { garage050 = oldGarage }()
	garage050.Loaded = true
	garage050.Setups = map[string]garageSetup050{}
	beta.carSelect, beta.paint = 3, 0
	garage050.Setups[garageKey050()] = garageSetup050{Model: 1}
	beta020Action("car+")
	if selectedCar051() != 1 || setup050().Model != 2 {
		t.Errorf("right arrow did not select EB110: %+v", setup050())
	}
	beta020Action("car-")
	if selectedCar051() != 0 || setup050().Model != 1 {
		t.Errorf("left arrow did not return to M4: %+v", setup050())
	}

}

func Test040FinalMenuCarArrowsRespectFreshProfileUnlocks(t *testing.T) {
	timeProgressFixture040(t)
	p, err := profiles040.Store.Create040("Menu progression driver", "custom_001")
	if err != nil {
		t.Fatal(err)
	}
	applyProfileSettings040()
	for _, level := range []int{1, 4, 7, 10} {
		p.XP = engine.XPForLevel040(level)
		for _, id := range []string{"car+", "car-"} {
			if !allowProgressAction040(id) {
				t.Errorf("level %d incorrectly gates cycle control %s", level, id)
			}
		}
		for class := 0; class < 4; class++ {
			if allowProgressAction040("car"+string(rune('0'+class))) != p.CanUseCar040(class) {
				t.Errorf("level %d direct car %d gate does not match unlock state", level, class)
			}
		}
		beta.carSelect, beta.paint = 3, 0
		expected := nextProgressCar040(3, 1)
		beta020Action("car+")
		if beta.carSelect != expected || !p.CanUseCar040(beta.carSelect) {
			t.Errorf("level %d cycle selected %d instead of safe next car %d", level, beta.carSelect, expected)
		}
	}
}

func Test040FinalMenuSelectingIshaanImmediatelyRestoresGates(t *testing.T) {
	timeProgressFixture040(t)
	fresh, err := profiles040.Store.Create040("QA isolation driver", "custom_001")
	if err != nil {
		t.Fatal(err)
	}
	applyProfileSettings040()
	if activeProfile040() != fresh || beta.carSelect != 3 || allowProgressAction040("car0") || allowProgressAction040("mode1") {
		t.Fatal("fresh fixture did not enter locked state")
	}
	// This is the stage isolation operation. Gates read the active store, not a
	// cached configuration, so selecting Ishaan alone restores unlock checks.
	if err = profiles040.Store.Select040("ishaan"); err != nil {
		t.Fatal(err)
	}
	if p := activeProfile040(); p == nil || !p.AllUnlocked040() || !allowProgressAction040("car0") || !allowProgressAction040("mode1") {
		t.Fatal("stage selection left old profile unlock gates active")
	}
	// Selection does not promise to reload the previous profile's settings.
	if beta.carSelect != 3 {
		t.Fatal("store selection unexpectedly changed live configuration")
	}
	if next := nextProgressCar040(0, 1); next != 1 {
		t.Fatal("correct active profile cannot cycle all four cars", next)
	}
}

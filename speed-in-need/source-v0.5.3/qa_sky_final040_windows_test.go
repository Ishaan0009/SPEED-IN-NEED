//go:build windows

package main

import (
	"testing"
	"time"
)

func Test040QASkyOverrideAppliesAfterPresetsOnlyDuringQA(t *testing.T) {
	previousQA, previousConfig, previousState := qa, beta.cfg, qaSkyFinal040
	defer func() { qa = previousQA; beta.cfg = previousConfig; qaSkyFinal040 = previousState }()
	for _, raw := range []string{"", "-1", "5", "4K", "1.5"} {
		if _, valid := parseQASkyDetail040(raw); valid {
			t.Fatal("invalid override accepted", raw)
		}
	}
	for detail, raw := range []string{"0", "1", "2", "3", "4"} {
		if value, valid := parseQASkyDetail040(raw); !valid || value != detail {
			t.Fatal("valid override rejected", raw, value, valid)
		}
	}
	t.Setenv("SPEED_QA_SKY_DETAIL040", "4")
	qa.enabled = false
	beta.cfg.SkyDetail040 = 1
	applyQASkyDetail040()
	if beta.cfg.SkyDetail040 != 1 {
		t.Fatal("QA override changed normal gameplay")
	}
	qa.enabled = true
	applyQASkyDetail040()
	if beta.cfg.SkyDetail040 != 4 || !qaSkyFinal040.Applied {
		t.Fatal("16K override not applied after preset")
	}
	// Simulate another stage profile resetting the setting; the final hook wins.
	beta.cfg.SkyDetail040 = 2
	applyQASkyDetail040()
	if beta.cfg.SkyDetail040 != 4 {
		t.Fatal("preset overwrote final QA sky override")
	}
	t.Setenv("SPEED_QA_SKY_DETAIL040", "0")
	applyQASkyDetail040()
	if beta.cfg.SkyDetail040 != 0 || !qaSkyFinal040.Applied {
		t.Fatal("Auto was confused with an absent override")
	}
	t.Setenv("SPEED_QA_SKY_DETAIL040", "5")
	applyQASkyDetail040()
	if qaSkyFinal040.Applied || qaSkyFinal040.Problem == "" || beta.cfg.SkyDetail040 != 0 {
		t.Fatal("bad environment value silently enabled", qaSkyFinal040)
	}
}

func Test040QASkyProofRequiresActualMeaningfulSources(t *testing.T) {
	sizes := [3][2]int{{16384, 8192}, {2048, 1024}, {2048, 1024}}
	hdr := [3]bool{true, true, true}
	textures := [3]uint32{1, 2, 3}
	sources, complete := requiredSkySourcesQA040(0, 1, .01, 16384, [3]bool{}, sizes, hdr, textures)
	if !complete || !sources[0].True16K || sources[1].Required {
		t.Fatal("inactive 2K sources blocked a valid selected-source 16K capture", sources, complete)
	}
	_, complete = requiredSkySourcesQA040(0, 1, .5, 16384, [3]bool{}, sizes, hdr, textures)
	if complete {
		t.Fatal("half-finished blend was reported complete")
	}
	sizes[1] = [2]int{16384, 8192}
	_, complete = requiredSkySourcesQA040(0, 1, .5, 16384, [3]bool{}, sizes, hdr, textures)
	if !complete {
		t.Fatal("two complete 16K blend sources were rejected")
	}
	hdr[1] = false
	_, complete = requiredSkySourcesQA040(0, 1, .5, 16384, [3]bool{}, sizes, hdr, textures)
	if complete {
		t.Fatal("legacy half-float/JPEG sky counted as genuine source HDR")
	}
	hdr[1] = true
	textures[1] = 0
	_, complete = requiredSkySourcesQA040(0, 1, .5, 16384, [3]bool{}, sizes, hdr, textures)
	if complete {
		t.Fatal("missing GPU texture counted as complete")
	}
	if requestedSkyWidthQA040(1, 4) != 16384 || requestedSkyWidthQA040(2, 0) != 8192 || requestedSkyWidthQA040(0, 0) != 4096 {
		t.Fatal("requested tier evidence was derived from the capped size")
	}
}

func Test040QASkyCaptureWaitIsBoundedAndHonest(t *testing.T) {
	previousQA, previousConfig, previousState := qa, beta.cfg, qaSkyFinal040
	previousSky, previousTexture, previousGPU, previousGarage, previousQuality := sky040, sky020, gpu038, garage033, app.quality
	defer func() {
		qa = previousQA
		beta.cfg = previousConfig
		qaSkyFinal040 = previousState
		sky040 = previousSky
		sky020 = previousTexture
		gpu038 = previousGPU
		garage033 = previousGarage
		app.quality = previousQuality
	}()
	previousRemaster := remaster.Enabled
	defer func() { remaster.Enabled = previousRemaster }()
	remaster.Enabled = false
	beta.cfg.TimeOfDay = 1
	qa.enabled = true
	garage033 = false
	app.quality = 3
	beta.cfg.SkyDetail040 = 4
	gpu038.Dedicated = 12 << 30
	sky040.MaxTexture = 16384
	sky040.TrueHDR = [3]bool{true, true, true}
	sky040.Failed = map[[2]int]bool{}
	sky020.Sizes = [3][2]int{{2048, 1024}, {2048, 1024}, {2048, 1024}}
	sky020.Textures = [3]uint32{1, 2, 3}
	qaSkyFinal040.Applied = true
	qaSkyFinal040.Problem = ""
	qaSkyFinal040.WaitStart = time.Time{}
	if !waitQASkyDetail040() {
		t.Fatal("capture did not allow ordinary rendering to finish the sky")
	}
	qaSkyFinal040.WaitStart = time.Now().Add(-qaSkyWaitLimit040 - time.Second)
	if waitQASkyDetail040() || !qaSkyFinal040.TimedOut || qaSkyFinal040.Problem == "" {
		t.Fatal("stalled upload was not bounded and reported", qaSkyFinal040)
	}
	qaSkyFinal040.WaitStart = time.Time{}
	qaSkyFinal040.TimedOut = false
	qaSkyFinal040.Problem = ""
	sky040.Failed[[2]int{0, 16384}] = true
	if waitQASkyDetail040() || qaSkyFinal040.Problem == "" {
		t.Fatal("failed upgrade waited forever or claimed completion")
	}
	delete(sky040.Failed, [2]int{0, 16384})
	qaSkyFinal040.Problem = ""
	sky020.Sizes[0] = [2]int{16384, 8192}
	if waitQASkyDetail040() || qaSkyFinal040.Problem != "" {
		t.Fatal("ready 16K sky could not be captured")
	}
}

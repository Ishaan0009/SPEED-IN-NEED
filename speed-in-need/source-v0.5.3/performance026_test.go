package main

import (
	"math"
	"speedinneed/engine"
	"testing"
	"time"
)

func feedQuality026(q *qualityGovernor026, seconds, gpu, cpu float64) {
	for i := 0; i < int(seconds*60); i++ {
		q.Sample(1./60, gpu, cpu, 1000./60)
	}
}

func TestQuality026SeparatesGPUAndCPUPressure(t *testing.T) {
	var gpu, cpu qualityGovernor026
	gpu.Reset()
	cpu.Reset()
	feedQuality026(&gpu, 12, 32, 6)
	feedQuality026(&cpu, 12, 3, 32)
	if gpu.Scale >= .9 || gpu.Scale < .65 {
		t.Fatalf("GPU pressure did not produce bounded resolution scaling: %+v", gpu)
	}
	if cpu.Scale != 1 || cpu.Foliage >= 1 || cpu.Foliage < .55 {
		t.Fatalf("CPU pressure reduced the wrong budget: %+v", cpu)
	}
	feedQuality026(&gpu, 100, 5, 5)
	feedQuality026(&cpu, 100, 5, 5)
	if math.Abs(gpu.Scale-1) > 1e-6 || math.Abs(cpu.Foliage-1) > 1e-6 {
		t.Fatal("quality does not recover after sustained headroom")
	}
}

func TestQuality026IgnoresLoadingSpikesAndSingleSlowWindow(t *testing.T) {
	var q qualityGovernor026
	q.Reset()
	feedQuality026(&q, 4, 5, 5)
	q.Sample(3, 5000, 5000, 16.7)
	q.Sample(.01, math.NaN(), 20, 16.7)
	feedQuality026(&q, 1, 50, 50)
	feedQuality026(&q, 2, 5, 5)
	if q.Scale != 1 || q.Foliage != 1 {
		t.Fatal("a loading hitch or short spike changed quality")
	}
}

func TestQuality026PersistentPressureBoundsSceneAndRecovers(t *testing.T) {
	var q qualityGovernor026
	q.Reset()
	feedQuality026(&q, 60, 4, 38)
	if q.Scale != 1 || q.Scene != .65 || q.Foliage != .55 {
		t.Fatalf("persistent CPU pressure did not lower bounded scene work: %+v", q)
	}
	feedQuality026(&q, 240, 3, 3)
	if q.Scene != 1 || q.Foliage != 1 || q.Scale != 1 {
		t.Fatalf("quality did not recover: %+v", q)
	}
}

func TestIntegrated026BudgetsPreserveSavedCeilings(t *testing.T) {
	old, governor, enabled, adaptiveQA := beta.cfg, adaptive026, qa.enabled, qaAdaptive026
	oldTab := remaster.SettingsTab
	t.Cleanup(func() {
		beta.cfg, adaptive026, qa.enabled, qaAdaptive026 = old, governor, enabled, adaptiveQA
		remaster.SettingsTab = oldTab
	})
	beta.cfg = defaultCfg()
	qa.enabled, qaAdaptive026 = false, false
	adaptive026.Reset()
	beta.cfg.Lightweight = true
	ceiling := beta.cfg
	if renderScale026() != .6 || viewDistance026() != .5 || shadowQuality026() != 0 || lightingDetail026() != 0 || !reducedEffects026() {
		t.Fatal("integrated mode did not bound expensive rendering")
	}
	adaptive026.Scale, adaptive026.Scene = .65, .65
	if renderScale026() != .4 || math.Abs(viewDistance026()-.325) > 1e-6 || beta.cfg != ceiling {
		t.Fatal("effective budgets changed saved settings or exceeded their floors")
	}
	beta.cfg.Lightweight = false
	beta.cfg.AdaptiveQuality = false
	if renderScale026() != ceiling.RenderScale || viewDistance026() != ceiling.ViewDistance || shadowQuality026() != ceiling.Shadows || reducedEffects026() {
		t.Fatal("disabling performance controls failed to restore saved rendering")
	}
	remaster.SettingsTab = 0
	if len(aokawaSettingValues()) != len(settingNames[0]) {
		t.Fatal("Display settings names and values are misaligned")
	}
}

func TestQuality026FallbackAccountsForDriverWaitWithoutQueries(t *testing.T) {
	gpu, cpu := qualityCosts026(false, 0, 4, 30)
	var q qualityGovernor026
	q.Reset()
	feedQuality026(&q, 10, gpu, cpu)
	if q.Scale >= 1 {
		t.Fatal("a renderer without GPU queries never adapts")
	}
	gpu, cpu = qualityCosts026(true, 5, 7, 30)
	if gpu != 5 || cpu != 7 {
		t.Fatal("valid GPU timing was replaced by display wait")
	}
}

func TestMemory026ScalesWithRAMAndAvailableHeadroom(t *testing.T) {
	for _, tc := range []struct {
		ram, available uint64
		heap, texture  int
	}{
		{4 << 30, 2 << 30, 384, 1024}, {8 << 30, 4 << 30, 512, 1024},
		{16 << 30, 8 << 30, 768, 2048}, {24 << 30, 12 << 30, 1024, 8192},
		{32 << 30, 512 << 20, 256, 1024}, {0, 0, 768, 8192},
	} {
		b := memoryPolicy026(tc.ram, tc.available)
		if b.HeapMB != tc.heap || b.TextureCap != tc.texture {
			t.Fatalf("RAM policy: %+v -> %+v", tc, b)
		}
	}
}

func TestFramePacer026DoesNotAccumulateOversleepOrBurstAfterHitch(t *testing.T) {
	var p framePacer026
	start := time.Unix(1000, 0)
	deadline := p.Next(start, 60)
	period := time.Second / 60
	for i := 1; i < 120; i++ {
		deadline = p.Next(deadline.Add(100*time.Microsecond), 60)
		if deadline != start.Add(time.Duration(i+1)*period) {
			t.Fatal("oversleep accumulated")
		}
	}
	hitch := deadline.Add(time.Second)
	if p.Next(hitch, 60) != hitch.Add(period) {
		t.Fatal("a hitch caused catch-up pacing")
	}
	if p.Next(hitch, 30) != hitch.Add(time.Second/30) {
		t.Fatal("new cap did not reset deadline")
	}
	if p.Next(hitch, 0) != hitch {
		t.Fatal("unlimited mode sleeps")
	}
}

func TestBuilding026CacheTracksVisibilityOrderAndCount(t *testing.T) {
	cells := []buildingCell026{{1, 2}, {2, 3}}
	groups := []remasterInstance{{VBO: 1, Count: 2}, {VBO: 2, Count: 3}}
	if !sameBuildingCells026(cells, groups) {
		t.Fatal("unchanged cells were invalidated")
	}
	groups[1].Count = 4
	if sameBuildingCells026(cells, groups) {
		t.Fatal("changed count was reused")
	}
	groups[1].Count = 3
	groups[0], groups[1] = groups[1], groups[0]
	if sameBuildingCells026(cells, groups) || sameBuildingCells026(cells, groups[:1]) {
		t.Fatal("visibility change was reused")
	}
}

func TestBuilding026WarmCacheHasNoPerFrameAllocation(t *testing.T) {
	old, pass, stats := buildingCache026, buildingPass026, buildingStats026
	t.Cleanup(func() { buildingCache026, buildingPass026, buildingStats026 = old, pass, stats })
	groups := []remasterInstance{{Key: "fixture", Kind: "building", VBO: 1, Count: 2}, {Key: "fixture", Kind: "building", VBO: 2, Count: 3}}
	buildingPass026 = 0
	buildingCache026 = map[buildingKey026]*buildingBatch026{{"fixture", 0}: {VBO: 3, Cells: []buildingCell026{{1, 2}, {2, 3}}, Output: [1]remasterInstance{{VBO: 3, Count: 5}}}}
	allocs := testing.AllocsPerRun(1000, func() {
		if beta012MergeBuildings(groups)[0].Count != 5 {
			panic("cache lost instances")
		}
	})
	if allocs != 0 {
		t.Fatalf("warm building cache allocates %v times per call", allocs)
	}
}

func TestAdaptive026HonorsManualCeilingAndFixedQA(t *testing.T) {
	config, state, enabled, adaptiveQA := beta.cfg, adaptive026, qa.enabled, qaAdaptive026
	t.Cleanup(func() { beta.cfg, adaptive026, qa.enabled, qaAdaptive026 = config, state, enabled, adaptiveQA })
	qaAdaptive026 = false
	beta.cfg = defaultCfg()
	beta.cfg.RenderScale = .8
	qa.enabled = false
	adaptive026.Reset()
	adaptive026.Scale = .65
	if s := renderScale026(); math.Abs(s-.52) > 1e-6 {
		t.Fatalf("wrong scaled ceiling: %v", s)
	}
	beta.cfg.RenderScale = .5
	if renderScale026() != .5 {
		t.Fatal("render scale fell below safe floor")
	}
	qa.enabled = true
	if renderScale026() != .5 || foliageFactor026() != 1 {
		t.Fatal("fixed QA is adaptive")
	}
	qaAdaptive026 = true
	beta.cfg.RenderScale = .8
	if math.Abs(renderScale026()-.52) > 1e-6 {
		t.Fatal("explicit adaptive verification does not exercise the governor")
	}
	beta.cfg.RenderScale = .8
	beta.cfg.AdaptiveQuality = false
	qa.enabled = false
	if renderScale026() != .8 || foliageFactor026() != 1 {
		t.Fatal("disabled quality still changes rendering")
	}
}

func TestDrift026FixtureRecoversWithoutAnyCollisionStop(t *testing.T) {
	oldWorld, oldQA := app.world, qa
	oldPrev, oldRender, oldAlpha, oldCamera, oldDrift := app.prevCar, app.renderCar, app.renderAlpha, remaster.CameraReset, driftQA026
	t.Cleanup(func() {
		app.world, qa = oldWorld, oldQA
		app.prevCar, app.renderCar, app.renderAlpha, remaster.CameraReset, driftQA026 = oldPrev, oldRender, oldAlpha, oldCamera, oldDrift
	})
	app.world = engine.NewRemasteredWorld(shippedRemasterDefinition(t, 0), 0, 0, 0, 1, 0)
	qa.stage, qa.checks = 93, map[string]bool{}
	beta026QAStage()
	if !qa.checks["drift-fixture-clear-of-solid-scenery"] {
		t.Fatalf("no starting area: %.2f m", driftQA026.FixtureClearance)
	}
	for app.world.Elapsed < 3.2 {
		app.world.Step(remasterQAInput(), 1./120)
		observeDriftQA026()
	}
	c := app.world.Car
	if driftQA026.PeakContact >= 1 || math.Abs(c.DriftAngle) >= 8 || math.Abs(c.YawRate) >= .2 || c.DriftActive || c.Speed <= 15 {
		t.Fatalf("fixture used collision or did not recover: peak_contact=%f speed=%f slip=%f yaw=%f", driftQA026.PeakContact, c.Speed, c.DriftAngle, c.YawRate)
	}
	t.Logf("fixture margin %.2f m, peak contact %.5f, peak slip %.2f degrees, recovered speed %.2f m/s", driftQA026.FixtureClearance, driftQA026.PeakContact, driftQA026.PeakSlip, c.Speed)
}

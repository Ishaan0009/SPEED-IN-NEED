package main

import (
	"math"
	"sort"
)

type instanceSpan050 struct {
	Start, End int
	Kind       string
	Radius     float64
}

var instanceIndex050 []instanceSpan050
var instanceSpansScratch050 []instanceSpan050

// The source buffer is sorted by asset then X/Z. Retain indices, not copied
// objects, so barrier counts and other mutable group state remain current.
func indexInstances050(groups []remasterInstance) []instanceSpan050 {
	spans := make([]instanceSpan050, 0, 128)
	for i := 0; i < len(groups); {
		end := i + 1
		radius := groups[i].Radius
		for end < len(groups) && groups[end].Key == groups[i].Key && groups[end].Kind == groups[i].Kind {
			radius = math.Max(radius, groups[end].Radius)
			end++
		}
		spans = append(spans, instanceSpan050{i, end, groups[i].Kind, radius})
		i = end
	}
	return spans
}
func nearbyInstanceSpans050(groups []remasterInstance, index []instanceSpan050, eyeX, grassLimit float64, shadow bool, out []instanceSpan050) []instanceSpan050 {
	out = out[:0]
	for _, span := range index {
		if shadow && (span.Kind == "grass" || span.Kind == "bush" || span.Kind == "pickup") {
			continue
		}
		if span.Kind == "grass" {
			source := groups[span.Start:span.End]
			reach := grassLimit + span.Radius
			first := sort.Search(len(source), func(i int) bool { return source[i].X >= eyeX-reach })
			end := sort.Search(len(source), func(i int) bool { return source[i].X > eyeX+reach })
			if first == end {
				continue
			}
			span.End = span.Start + end
			span.Start += first
		}
		out = append(out, span)
	}
	return out
}
func visibleInstanceSpans050(eye pos3, shadow bool) []instanceSpan050 {
	if len(instanceIndex050) == 0 && len(remaster.Instances) > 0 {
		instanceIndex050 = indexInstances050(remaster.Instances)
	}
	limit := []float64{110, 240, 420, 650}[app.quality] * beta.cfg.Vegetation * foliageDistance026()
	instanceSpansScratch050 = nearbyInstanceSpans050(remaster.Instances, instanceIndex050, eye.x, limit, shadow, instanceSpansScratch050)
	return instanceSpansScratch050
}

package engine

import (
	"os"
	"testing"
)

// Use the actual shipped road graph and civilian fleet, outside the timed setup.
func BenchmarkTraffic038(b *testing.B) {
	raw, err := os.ReadFile("../assets/b037_world_free.json")
	if err != nil {
		b.Fatal(err)
	}
	d, err := DecodeWorldDefinition(raw)
	if err != nil {
		b.Fatal(err)
	}
	for _, density := range []int{1, 2} {
		name := "256"
		if density == 2 {
			name = "1032"
		}
		b.Run(name, func(b *testing.B) {
			w := NewRemasteredWorld(d, 0, 0, 0, 1, 0)
			w.SetTrafficDensity(density)
			for tick := 0; tick < 120; tick++ {
				w.Elapsed += 1. / 120
				w.stepRemasteredTraffic(1. / 120)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for tick := 0; tick < b.N; tick++ {
				w.Elapsed += 1. / 120
				w.stepRemasteredTraffic(1. / 120)
			}
		})
	}
}

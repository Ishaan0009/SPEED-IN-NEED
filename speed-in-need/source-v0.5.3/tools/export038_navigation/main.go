package main

import (
	"encoding/json"
	"fmt"
	"os"
	"speedinneed/engine"
)

func main() {
	raw, err := os.ReadFile("assets/b037_world_free.json")
	if err != nil {
		panic(err)
	}
	d, err := engine.DecodeWorldDefinition(raw)
	if err != nil {
		panic(err)
	}
	g := engine.NewRoadNavigator(d)
	w := engine.NewRemasteredWorld(d, 0, 0, 0, 1, 0)
	out := map[string]any{"nodes": g.Nodes, "segments": g.Segments, "traffic_core": w.BuildTrafficCore035()}
	raw, err = json.Marshal(out)
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile("NAVIGATION_INPUT_038.json", raw, 0644); err != nil {
		panic(err)
	}
	fmt.Println("Navigation export", len(g.Nodes), "nodes", len(g.Segments), "segments")
}

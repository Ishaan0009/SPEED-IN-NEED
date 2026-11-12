package main

// Geometry cost and an actual distant mesh determine the policy. Versioned
// asset names must not silently bypass tree LOD or shadow LOD selection.
func detailedTreeLOD038(key string) bool {
	a := pbrAssets[key]
	if a == nil || pbrAssets[key+"_lod"] == nil {
		return false
	}
	indices := 0
	for _, batch := range a.Batches {
		indices += batch.Indices
	}
	return indices >= 15000
}

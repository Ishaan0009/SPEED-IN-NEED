package main

import "testing"

func TestTreeLOD038DoesNotDependOnAssetVersionPrefix(t *testing.T) {
	old := pbrAssets
	defer func() { pbrAssets = old }()
	pbrAssets = map[string]*pbrAsset{}
	for _, key := range []string{"b037_pine", "b038_b037_pine", "future_imported_conifer"} {
		pbrAssets[key] = &pbrAsset{Batches: []pbrPart{{Indices: 24000}}}
		pbrAssets[key+"_lod"] = &pbrAsset{}
		if !detailedTreeLOD038(key) {
			t.Fatal("detailed tree lost its LOD policy", key)
		}
	}
	pbrAssets["simple"] = &pbrAsset{Batches: []pbrPart{{Indices: 900}}}
	pbrAssets["simple_lod"] = &pbrAsset{}
	if detailedTreeLOD038("simple") {
		t.Fatal("simple foliage treated as a heavy imported crown")
	}
}

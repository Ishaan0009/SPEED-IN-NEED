package main

import (
	"testing"
	"unsafe"
)

func TestTemporal038NativeABI(t *testing.T) {
	if unsafe.Sizeof(temporalNativeFrame038{}) != 224 {
		t.Fatal("native frame ABI changed")
	}
}
func TestSkinHistory038FollowsEntityAcrossBatchReordering(t *testing.T) {
	oldHistory, oldFrame := skinHistory038, temporal038.Frame
	defer func() { skinHistory038 = oldHistory; temporal038.Frame = oldFrame }()
	skinHistory038 = map[string]*skinHistoryFrame038{}
	temporal038.Frame = 100
	a, b := [][3]float32{{1, 2, 3}, {4, 5, 6}}, [][3]float32{{20, 0, 0}, {21, 0, 0}}
	previousSkin038("A", a)
	previousSkin038("B", b)
	temporal038.Frame++
	pB := previousSkin038("B", [][3]float32{{30, 0, 0}, {31, 0, 0}})
	pA := previousSkin038("A", [][3]float32{{2, 3, 4}, {5, 6, 7}})
	if pB[0] != b[0] || pA[1] != a[1] {
		t.Fatal("motion history followed batch order rather than actor")
	}
	temporal038.Frame += 4
	returned := [][3]float32{{80, 0, 0}}
	p := previousSkin038("A", returned)
	if len(p) != 1 || p[0] != returned[0] {
		t.Fatal("new LOD or returning actor streaks to stale geometry")
	}
	returned[0][0] = 999
	if skinHistory038["A"].Current[0][0] != 80 {
		t.Fatal("history aliases the mutable skin buffer")
	}
}

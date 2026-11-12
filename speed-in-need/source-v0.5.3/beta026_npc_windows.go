package main

import (
	"fmt"
	"math"
	"speedinneed/engine"
)

func drawBeta026NPC(t engine.Traffic) {
	key := fmt.Sprintf("rm_npc%d", t.Variant%4)
	oldPaint := beta.paint
	defer func() { beta.paint = oldPaint }()
	if t.Variant >= 4 {
		key = fmt.Sprintf("b011_npc%d", (t.Variant-4)%10)
		beta.paint = (t.Variant * 7) % len(paintPalette)
	}
	if t.Variant == 0 {
		resize(1.38, 1.59, 1.588)
	}
	a := pbrAssets[key]
	if a == nil {
		return
	}
	drawPBRAsset(key, 0)
	wheelKeys := [...]string{"1", "2", "3", "4"}
	for i := 0; i < 4; i++ {
		anchor, ok := a.Anchors[wheelKeys[i]]
		if !ok {
			continue
		}
		push()
		move(anchor[0], anchor[1], anchor[2])
		if i < 2 {
			rotate(t.SteerAngle*180/math.Pi, 0, 1, 0)
		}
		radius := a.WheelRadii[wheelKeys[i]]
		spin := t.WheelSpin
		if radius > .1 {
			spin *= .34 / radius
		}
		rotate(spin*180/math.Pi, 1, 0, 0)
		drawPBRAsset(key, i+1)
		pop()
	}
}

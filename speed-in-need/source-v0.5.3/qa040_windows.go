package main

import("fmt";"math";"speedinneed/engine")

func registerQA040(){
 if len(remasterQANames)>148{return}
 for _,car:=range []int{0,2}{for paint:=0;paint<18;paint++{remasterQANames=append(remasterQANames,fmt.Sprintf("040-paint-car%d-%02d",car,paint))}}
 remasterQANames=append(remasterQANames,"040-map-spawn","040-rooted-meadow","040-mode-difficulty")
}
func stageQA040(){
 if qa.stage<148{return}
 if qa.stage<184{
  car:=0;if qa.stage>=166{car=2};beta.carSelect=car;beta.paint=(qa.stage-148)%18
  selectScreen(screenMaps);ui020.Stage=2;app.started=false;preview014.ready=false
  return
 }
 if qa.stage==184 {selectScreen(screenMaps);ui020.Stage=1;app.started=false;return}
 if qa.stage==186 {selectScreen(screenMaps);ui020.Stage=0;remaster.Mode=1;app.started=false;return}
 for _,o:=range app.world.Remaster.Def.Objects{
  if o.Kind=="tree"&&o.Pos[1]>60&&o.Pos[1]<100{
   p:=engine.Vec{X:o.Pos[0]+6,Z:o.Pos[2]+6};remasterQAPlace(p,math.Pi);app.camDist=4.8;break
  }
 }
}

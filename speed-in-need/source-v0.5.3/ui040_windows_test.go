package main

import("testing";"math")

func Test040SettingsGroupsCoverEveryOptionOnce(t *testing.T){
 seen:=map[settingRefine]bool{}
 for group:=range groupsRefine{for _,row:=range groupRowsRefine(group){if seen[row]{t.Fatalf("duplicate setting %+v",row)};seen[row]=true}}
 for tab,rows:=range settingNames{for row:=range rows{if !seen[settingRefine{tab,row}]{t.Fatalf("missing setting %d/%d",tab,row)}}}
}
func Test040MouseAndStickLookUseSameDirections(t *testing.T){
 cfg,pad,yaw,pitch,invert:=beta.cfg,pad033,app.camYaw,app.camPitch,app.invertMouseX
 defer func(){beta.cfg,pad033,app.camYaw,app.camPitch,app.invertMouseX=cfg,pad,yaw,pitch,invert}()
 beta.cfg=defaultCfg();app.invertMouseX=false;app.camYaw,app.camPitch=0,0
 beta020Look(20,-20);mx,my:=app.camYaw,app.camPitch
 app.camYaw,app.camPitch=0,0;pad033.State=padSample033{Connected:true,Axes:[6]float64{0,0,1,1,0,0}}
	 drivePad036(pad033.State,0,.016)
 if mx*app.camYaw<=0||my*app.camPitch<=0{t.Fatalf("opposite look direction mouse %.3f/%.3f pad %.3f/%.3f",mx,my,app.camYaw,app.camPitch)}
 beta.cfg.InvertLookYRefine=true;app.camPitch=0;beta020Look(0,-20)
 if math.Signbit(my)==math.Signbit(app.camPitch){t.Fatal("vertical inversion did not invert")}
}
func Test040SetupMapClickSelectsSpawn(t *testing.T){
 screen,stage,mode,width,height,spawn:=beta.screen,ui020.Stage,remaster.Mode,app.width,app.height,setupSpawnRefine
 defer func(){beta.screen,ui020.Stage,remaster.Mode,app.width,app.height,setupSpawnRefine=screen,stage,mode,width,height,spawn}()
 beta.screen=screenMaps;ui020.Stage=1;remaster.Mode=0;app.width,app.height=1920,1080
 if !pickSetupSpawnRefine(1100,650)||!setupSpawnRefine.Selected{t.Fatal("map click did not select a spawn")}
 if pickSetupSpawnRefine(20,20){t.Fatal("outside map accepted as spawn")}
 remaster.Mode=1;if pickSetupSpawnRefine(1100,650){t.Fatal("race grid changed by map click")}
}

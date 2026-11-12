package main

import("os";"path/filepath";"runtime";"runtime/pprof";"time")

// Explicit diagnostic mode only. It never changes normal startup or timing.
func beginStartupProfile041(){
 if os.Getenv("SPEED_STARTUP_PROFILE041")!="1"{return}
 exe,_:=os.Executable();folder:=filepath.Join(filepath.Dir(exe),"Reports","STARTUP041");os.MkdirAll(folder,0755)
 f,e:=os.Create(filepath.Join(folder,"cpu.pprof"));if e!=nil{return}
 if pprof.StartCPUProfile(f)!=nil{f.Close();return}
 go func(){time.Sleep(35*time.Second);b:=make([]byte,4<<20);n:=runtime.Stack(b,true);os.WriteFile(filepath.Join(folder,"stacks.txt"),b[:n],0644);pprof.StopCPUProfile();f.Close()}()
}

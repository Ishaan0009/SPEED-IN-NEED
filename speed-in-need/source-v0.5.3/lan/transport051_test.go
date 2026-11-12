package lan

import (
	"bytes"
	"math/rand"
	"testing"
	"time"
)

func Test051MTUReassemblyReorderingAndLoss(t *testing.T) {
	raw := make([]byte, 15000)
	rand.New(rand.NewSource(51)).Read(raw)
	frames := frames051(raw, 51)
	if len(frames) < 2 {
		t.Fatal("test did not fragment")
	}
	a := assembler051{}
	now := time.Now()
	for i := len(frames) - 1; i > 0; i-- {
		if len(frames[i]) > datagram051 {
			t.Fatal("MTU exceeded")
		}
		if _, ok := a.push("peer", frames[i], now); ok {
			t.Fatal("partial message delivered")
		}
	}
	got, ok := a.push("peer", frames[0], now)
	if !ok || !bytes.Equal(raw, got) {
		t.Fatal("reordered message corrupted")
	}
	a.push("peer", frames[0], now)
	a.push("peer", frames[1], now.Add(time.Second))
	if len(a) != 1 {
		t.Fatal("expired incomplete message retained")
	}
}
func Test051CompressedPacketAndBoundedDecode(t *testing.T) {
	raw := bytes.Repeat([]byte("racing-car-state"), 800)
	f := frames051(raw, 2)
	if len(f) != 1 {
		t.Fatal("repetitive snapshot was not compressed")
	}
	got, ok := (assembler051{}).push("p", f[0], time.Now())
	if !ok || !bytes.Equal(raw, got) {
		t.Fatal("compression round trip failed")
	}
	if _, ok := decode051([]byte("SZ51invalid")); ok {
		t.Fatal("invalid zlib accepted")
	}
	if ping051(0, time.Now().Add(-20*time.Millisecond).UnixMilli(), time.Now()) < 19 {
		t.Fatal("RTT missing")
	}
	if ping051(20, time.Now().Add(time.Hour).UnixMilli(), time.Now()) != 20 {
		t.Fatal("future timestamp accepted")
	}
}

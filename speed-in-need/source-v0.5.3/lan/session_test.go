package lan

import (
	"math"
	"net"
	"testing"
	"time"
)

func eventually(t *testing.T, f func() bool) {
	t.Helper()
	end := time.Now().Add(3 * time.Second)
	for time.Now().Before(end) {
		if f() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("LAN condition timed out")
}
func TestFourPlayersApprovalReplicationAndRemoval(t *testing.T) {
	checkUDP(t)
	h, e := NewHost("Host", "avatar", "test-world", 0)
	if e != nil {
		t.Fatal(e)
	}
	defer h.Close()
	h.Publish(Car{X: 10, Vehicle: 2}, World{Epoch: 1, Mode: 1, Laps: 3}, []Car{{X: 90, S: 35}}, nil)
	clients := []*Session{}
	for i := 0; i < 3; i++ {
		c, e := NewClient(net.JoinHostPort("127.0.0.1", fmtPort(h.Port())), "Friend", "avatar", "test-world")
		if e != nil {
			t.Fatal(e)
		}
		defer c.Close()
		clients = append(clients, c)
	}
	defer func() {
		for _, c := range clients {
			t.Log(c.Status())
		}
		t.Log(h.Status())
	}()
	eventually(t, func() bool { return len(h.Pending()) == 3 })
	if len(h.View().Players) != 1 {
		t.Fatal("unapproved peer entered room")
	}
	for _, p := range h.Pending() {
		if e := h.Approve(p.ID); e != nil {
			t.Fatal(e)
		}
	}
	for i, c := range clients {
		eventually(t, c.Connected)
		c.Publish(Car{X: float64(i + 20), S: 33, Vehicle: i}, World{Epoch: 1}, nil, nil)
	}
	eventually(t, func() bool {
		v := h.View()
		if len(v.Players) != 4 {
			return false
		}
		for _, p := range v.Players[1:] {
			if p.State.X < 20 || !p.Ready {
				return false
			}
		}
		return true
	})
	eventually(t, func() bool {
		v := clients[0].View()
		return len(v.Players) == 4 && len(v.AI) == 1 && v.World.Mode == 1 && v.World.Epoch == 1
	})
	removed := clients[1]
	h.Remove(removed.ID())
	eventually(t, func() bool { return !removed.Connected() && len(h.View().Players) == 3 })
	h.Publish(Car{}, World{Epoch: 2}, nil, nil)
	eventually(t, func() bool {
		for _, p := range h.View().Players[1:] {
			if p.Ready {
				return false
			}
		}
		return true
	})
}
func fmtPort(p int) string {
	const digits = "0123456789"
	if p == 0 {
		return "0"
	}
	b := []byte{}
	for p > 0 {
		b = append([]byte{digits[p%10]}, b...)
		p /= 10
	}
	return string(b)
}
func TestVersionMismatchAndInvalidCar(t *testing.T) {
	checkUDP(t)
	h, e := NewHost("Host", "", "world-A", 0)
	if e != nil {
		t.Fatal(e)
	}
	defer h.Close()
	c, e := NewClient(net.JoinHostPort("127.0.0.1", fmtPort(h.Port())), "Friend", "", "world-B")
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	eventually(t, func() bool { return c.Status() == "Game/world version differs: use the same 0.5.1 Remastered build" })
	if len(h.Pending()) != 0 || c.Connected() {
		t.Fatal("mismatched world admitted")
	}
	if valid(Car{X: math.NaN()}) || valid(Car{S: 1000}) || valid(Car{Vehicle: 99}) {
		t.Fatal("invalid network physics accepted")
	}
}
func TestDiscovery(t *testing.T) {
	checkUDP(t)
	h, e := NewHost("Discoverable", "", "test", 0)
	if e != nil {
		t.Fatal(e)
	}
	defer h.Close()
	rooms := Discover(h.Port(), 250*time.Millisecond)
	for _, r := range rooms {
		if r.Name == "Discoverable" {
			return
		}
	}
	t.Fatal("host not discovered over loopback")
}

func checkUDP(t *testing.T) {
	t.Helper()
	c, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if e != nil {
		t.Skip("UDP unavailable: ", e)
	}
	defer c.Close()
	c.SetReadDeadline(time.Now().Add(150 * time.Millisecond))
	c.WriteToUDP([]byte("probe"), c.LocalAddr().(*net.UDPAddr))
	b := make([]byte, 10)
	if _, _, e = c.ReadFromUDP(b); e != nil {
		t.Skip("Execution environment blocks UDP loopback; run these integration tests on the host PCs: ", e)
	}
}

func TestProtocolAdmissionAuthenticationAndOrdering(t *testing.T) {
	c, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero})
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	h := newSession(c, "Host", "", "world")
	h.host = true
	h.world.Epoch = 9
	a := &net.UDPAddr{IP: net.IPv4(192, 168, 1, 12), Port: 50001}
	id := nonce()
	h.receive(a, wire{Kind: "hello", ID: id, Name: " Friend ", WorldKey: "world"})
	if len(h.Pending()) != 1 || len(h.View().Players) != 1 {
		t.Fatal("pending peer must not be admitted")
	}
	if e := h.Approve(id); e != nil {
		t.Fatal(e)
	}
	token := h.peers[id].Token
	h.receive(a, wire{Kind: "state", ID: id, Token: "wrong", Seq: 1, State: Car{X: 100}})
	if h.peers[id].Seq != 0 {
		t.Fatal("unauthenticated packet changed state")
	}
	h.receive(a, wire{Kind: "state", ID: id, WorldKey: "world", Token: token, Seq: 2, State: Car{X: 25}, Ready: true, Epoch: 9})
	h.receive(a, wire{Kind: "state", ID: id, WorldKey: "world", Token: token, Seq: 1, State: Car{X: 10}, Ready: true, Epoch: 9})
	if h.peers[id].Player.State.X != 25 || !h.peers[id].Player.Ready {
		t.Fatal("out-of-order packet replaced newest state")
	}
	h.receive(a, wire{Kind: "state", ID: id, Token: token, Seq: 3, State: Car{X: math.NaN()}})
	if h.peers[id].Seq != 2 {
		t.Fatal("nonfinite physics accepted")
	}
	h.Remove(id)
	h.receive(a, wire{Kind: "hello", ID: id, Name: "Friend", WorldKey: "world"})
	if len(h.Pending()) != 0 || len(h.View().Players) != 1 {
		t.Fatal("removed peer rejoined without a new session")
	}
}

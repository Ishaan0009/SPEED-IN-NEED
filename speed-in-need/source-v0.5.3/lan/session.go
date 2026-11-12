// Package lan implements the versioned, four-player local-network protocol.
// It owns sockets and lobby state; rendering and physics stay on the game thread.
package lan

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

const Protocol = "SIN-LAN-052-1"
const DefaultPort = 42850
const MaxPlayers = 4
const maxPacket = 16384

type Car struct {
	Entity051                                          int
	Model, Tyre, Rim, Livery                           int
	X, Y, Z, H, S, L, Steer, Wheel, Nitro, Pitch, Roll float64
	Vehicle, Paint, Lap, Gate                          int
	Boost, Brake, Finished                             bool
	RaceTime                                           float64
}
type Player struct {
	Age051           float64
	ID, Name, Avatar string
	State            Car
	Ping             int
	Ready            bool
}
type World struct {
	Mode, Time, Laps, Opponents, Difficulty int
	Epoch                                   uint64
	Countdown, RaceTime, Hour               float64
	Rain                                    bool
}
type Snapshot struct {
	Players     []Player
	AI, Traffic []Car
	World       World
	Seq         uint64
	Age051      float64 `json:"-"`
}
type Pending struct{ ID, Name, Avatar string }
type wire struct {
	Protocol, Kind, ID, Name, Avatar, WorldKey, Token, Reason string
	Seq, Epoch                                                uint64
	State                                                     Car
	Snapshot                                                  Snapshot
	Ready                                                     bool
	Sent, Echo                                                int64
}
type peer struct {
	Player Player
	Addr   *net.UDPAddr
	Token  string
	Seq    uint64
	Seen   time.Time
}
type candidate struct {
	Pending
	Addr *net.UDPAddr
	Seen time.Time
}
type Session struct {
	mu                 sync.Mutex
	conn               *net.UDPConn
	host               bool
	remote             *net.UDPAddr
	self               Player
	key, token, status string
	peers              map[string]*peer
	pending            map[string]candidate
	blocked            map[string]bool
	latest             Snapshot
	world              World
	ai, traffic        []Car
	seq, receiveSeq    uint64
	lastHost           time.Time
	stopped            chan struct{}
	closeOnce          sync.Once
	datagramID051      uint64
	pingSent051        int64
}

func nonce() string {
	b := make([]byte, 12)
	if _, e := rand.Read(b); e != nil {
		panic(e)
	}
	return hex.EncodeToString(b)
}
func cleanName(s string) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) > 24 {
		r = r[:24]
	}
	s = string(r)
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
	if s == "" {
		s = "Driver"
	}
	return s
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func valid(c Car) bool {
	for _, v := range []float64{c.X, c.Y, c.Z, c.H, c.S, c.L, c.Steer, c.Wheel, c.Nitro, c.Pitch, c.Roll, c.RaceTime} {
		if !finite(v) {
			return false
		}
	}
	return math.Abs(c.X) <= 20000 && math.Abs(c.Z) <= 20000 && math.Abs(c.Y) <= 10000 && math.Abs(c.S) <= 400 && math.Abs(c.L) <= 200 && c.Vehicle >= 0 && c.Vehicle < 16 && c.Paint >= 0 && c.Paint < 128 && c.Lap >= 0 && c.Lap < 100 && c.Gate >= 0 && c.Gate < 10000
}
func NewHost(name, avatar, key string, port int) (*Session, error) {
	c, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: port})
	if e != nil {
		return nil, e
	}
	s := newSession(c, name, avatar, key)
	s.host = true
	s.status = "Hosting: approve joining players"
	s.lastHost = time.Now()
	s.world.Laps = 3
	go s.read()
	go s.run()
	return s, nil
}
func NewClient(address, name, avatar, key string) (*Session, error) {
	if !strings.Contains(address, ":") {
		address = fmt.Sprintf("%s:%d", address, DefaultPort)
	}
	remote, e := net.ResolveUDPAddr("udp4", address)
	if e != nil {
		return nil, e
	}
	if remote.Port < 1 || remote.IP == nil {
		return nil, errors.New("enter the host's IPv4 address")
	}
	c, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero})
	if e != nil {
		return nil, e
	}
	s := newSession(c, name, avatar, key)
	s.remote = remote
	s.status = "Request sent: waiting for host approval"
	s.lastHost = time.Now()
	go s.read()
	go s.run()
	return s, nil
}
func newSession(c *net.UDPConn, name, avatar, key string) *Session {
	return &Session{conn: c, self: Player{ID: nonce(), Name: cleanName(name), Avatar: avatar, Ready: true}, key: key, peers: map[string]*peer{}, pending: map[string]candidate{}, blocked: map[string]bool{}, stopped: make(chan struct{})}
}
func (s *Session) Close() {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		if !s.host && s.remote != nil {
			s.send(s.remote, wire{Kind: "leave", Token: s.token})
		}
		s.mu.Unlock()
		close(s.stopped)
		s.conn.Close()
	})
}
func (s *Session) IsHost() bool   { return s.host }
func (s *Session) ID() string     { return s.self.ID }
func (s *Session) Port() int      { return s.conn.LocalAddr().(*net.UDPAddr).Port }
func (s *Session) Status() string { s.mu.Lock(); defer s.mu.Unlock(); return s.status }
func (s *Session) Connected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.host || s.token != "" && time.Since(s.lastHost) < 12*time.Second
}
func (s *Session) Publish(c Car, w World, ai, traffic []Car) {
	if !valid(c) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.self.State = c
	if s.host {
		if w.Epoch != s.world.Epoch {
			for _, p := range s.peers {
				p.Player.Ready = false
			}
		}
		s.world = w
		s.ai = append(s.ai[:0], ai...)
		s.traffic = append(s.traffic[:0], traffic...)
		if len(s.ai) > 7 {
			s.ai = s.ai[:7]
		}
		if len(s.traffic) > 12 {
			s.traffic = s.traffic[:12]
		}
	} else {
		s.world = w
	}
}
func (s *Session) SetReady(ready bool) { s.mu.Lock(); s.self.Ready = ready; s.mu.Unlock() }
func (s *Session) Pending() []Pending {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []Pending{}
	for _, p := range s.pending {
		out = append(out, p.Pending)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
func (s *Session) Approve(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.host {
		return errors.New("only the host can add players")
	}
	p, ok := s.pending[id]
	if !ok {
		return errors.New("request expired")
	}
	if len(s.peers)+1 >= MaxPlayers {
		return errors.New("room is full")
	}
	s.peers[id] = &peer{Player: Player{ID: id, Name: p.Name, Avatar: p.Avatar, Ready: true}, Addr: p.Addr, Token: nonce(), Seen: time.Now()}
	delete(s.pending, id)
	s.status = p.Name + " joined"
	return nil
}
func (s *Session) Remove(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.host {
		return
	}
	if p, ok := s.peers[id]; ok {
		s.send(p.Addr, wire{Kind: "removed", Token: p.Token, Reason: "Removed by the host"})
		delete(s.peers, id)
		s.blocked[id] = true
	}
	if p, ok := s.pending[id]; ok {
		s.send(p.Addr, wire{Kind: "removed", Reason: "Join request declined"})
		delete(s.pending, id)
		s.blocked[id] = true
	}
}
func (s *Session) View() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.host {
		v := clone(s.latest)
		v.Age051 = math.Min(.2, time.Since(s.lastHost).Seconds())
		for i := range v.Players {
			if v.Players[i].ID == s.self.ID {
				v.Players[i].Ping = s.self.Ping
			}
		}
		return v
	}
	return s.snapshot()
}
func clone(v Snapshot) Snapshot {
	v.Players = append([]Player(nil), v.Players...)
	v.AI = append([]Car(nil), v.AI...)
	v.Traffic = append([]Car(nil), v.Traffic...)
	return v
}
func (s *Session) snapshot() Snapshot {
	v := Snapshot{Players: []Player{s.self}, World: s.world, AI: append([]Car(nil), s.ai...), Traffic: append([]Car(nil), s.traffic...), Seq: s.seq}
	for _, p := range s.peers {
		player := p.Player
		player.Age051 = math.Min(.2, time.Since(p.Seen).Seconds())
		v.Players = append(v.Players, player)
	}
	sort.Slice(v.Players[1:], func(i, j int) bool { return v.Players[i+1].ID < v.Players[j+1].ID })
	return v
}
func (s *Session) send(addr *net.UDPAddr, m wire) {
	m.Protocol = Protocol
	m.ID = s.self.ID
	m.WorldKey = s.key
	if m.Sent == 0 {
		m.Sent = time.Now().UnixMilli()
	}
	raw, e := json.Marshal(m)
	if e == nil && len(raw) <= maxPacket {
		s.datagramID051++
		for _, frame := range frames051(raw, s.datagramID051) {
			if _, err := s.conn.WriteToUDP(frame, addr); err != nil {
				s.status = "LAN send failed: " + err.Error()
				break
			}
		}
	}
}
func (s *Session) run() {
	t := time.NewTicker(50 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-s.stopped:
			return
		case <-t.C:
			s.mu.Lock()
			s.seq++
			if s.seq%10 == 0 {
				s.pingSent051 = time.Now().UnixMilli()
				if s.host {
					for _, p := range s.peers {
						s.send(p.Addr, wire{Kind: "ping", Token: p.Token, Sent: s.pingSent051})
					}
				} else if s.remote != nil && s.token != "" {
					s.send(s.remote, wire{Kind: "ping", Token: s.token, Sent: s.pingSent051})
				}
			}
			if s.host {
				for id, p := range s.pending {
					if time.Since(p.Seen) > 8*time.Second {
						delete(s.pending, id)
					}
				}
				for id, p := range s.peers {
					if time.Since(p.Seen) > 30*time.Second {
						delete(s.peers, id)
						s.status = p.Player.Name + " disconnected"
					}
				}
				v := s.snapshot()
				for _, p := range s.peers {
					s.send(p.Addr, wire{Kind: "snapshot", Token: p.Token, Snapshot: v, Seq: s.seq})
				}
			} else if s.remote != nil {
				if s.token == "" {
					if s.seq%10 == 1 {
						s.send(s.remote, wire{Kind: "hello", Name: s.self.Name, Avatar: s.self.Avatar})
					}
				}
				if s.token != "" {
					s.send(s.remote, wire{Kind: "state", Token: s.token, State: s.self.State, Ready: s.self.Ready, Seq: s.seq, Epoch: s.world.Epoch})
				}
				if time.Since(s.lastHost) > 12*time.Second {
					s.status = "Connection interrupted: retrying the host automatically"
				}
			}
			s.mu.Unlock()
		}
	}
}
func (s *Session) read() {
	b := make([]byte, datagram051+1)
	chunks := assembler051{}
	for {
		n, addr, e := s.conn.ReadFromUDP(b)
		if e != nil {
			return
		}
		if n > datagram051 {
			continue
		}
		raw, complete := chunks.push(addr.String(), b[:n], time.Now())
		if !complete {
			continue
		}
		var m wire
		if json.Unmarshal(raw, &m) != nil || m.Protocol != Protocol {
			continue
		}
		s.mu.Lock()
		s.receive(addr, m)
		s.mu.Unlock()
	}
}
func (s *Session) receive(addr *net.UDPAddr, m wire) {
	if s.host {
		if m.Kind == "discover" {
			s.send(addr, wire{Kind: "room", Name: s.self.Name})
			return
		}
		if m.WorldKey != s.key {
			if m.Kind == "hello" {
				s.send(addr, wire{Kind: "incompatible", Reason: "Game/world version differs: use the same 0.5.1 Remastered build"})
			}
			return
		}
		if m.Kind == "hello" {
			if len(m.ID) != 24 || s.blocked[m.ID] {
				return
			}
			if _, ok := s.peers[m.ID]; ok {
				return
			}
			if len(s.peers)+1 >= MaxPlayers {
				s.send(addr, wire{Kind: "removed", Reason: "Room is full"})
				return
			}
			if len(s.pending) < 16 {
				s.pending[m.ID] = candidate{Pending: Pending{ID: m.ID, Name: cleanName(m.Name), Avatar: m.Avatar}, Addr: addr, Seen: time.Now()}
			}
			return
		}
		p, ok := s.peers[m.ID]
		if !ok || p.Addr.String() != addr.String() || m.Token != p.Token {
			return
		}
		if m.Kind == "ping" {
			s.send(addr, wire{Kind: "pong", Token: p.Token, Echo: m.Sent})
			return
		}
		if m.Kind == "pong" {
			if m.Echo == s.pingSent051 {
				p.Player.Ping = ping051(p.Player.Ping, m.Echo, time.Now())
			}
			return
		}
		if m.Kind == "leave" {
			delete(s.peers, m.ID)
			return
		}
		if m.Kind != "state" || m.Seq <= p.Seq || !valid(m.State) {
			return
		}
		p.Seq = m.Seq
		p.Seen = time.Now()
		if s.world.Mode == 1 {
			m.State.Lap = p.Player.State.Lap
			m.State.Gate = p.Player.State.Gate
			m.State.Finished = p.Player.State.Finished
			m.State.RaceTime = p.Player.State.RaceTime
		}
		p.Player.State = m.State
		p.Player.Ready = m.Ready && m.Epoch == s.world.Epoch
		return
	}
	if s.remote == nil || addr.String() != s.remote.String() {
		return
	}
	if m.Kind == "incompatible" || m.Kind == "removed" {
		s.status = m.Reason
		s.token = ""
		s.remote = nil
		return
	}
	if s.token != "" && m.Token == s.token && m.WorldKey == s.key {
		if m.Kind == "ping" {
			s.send(addr, wire{Kind: "pong", Token: s.token, Echo: m.Sent})
			return
		}
		if m.Kind == "pong" {
			if m.Echo == s.pingSent051 {
				s.self.Ping = ping051(s.self.Ping, m.Echo, time.Now())
			}
			return
		}
	}
	if m.Kind != "snapshot" || m.WorldKey != s.key || m.Token == "" || s.token != "" && m.Token != s.token || m.Seq <= s.receiveSeq {
		return
	}
	if len(m.Snapshot.Players) > MaxPlayers || len(m.Snapshot.AI) > 7 || len(m.Snapshot.Traffic) > 12 {
		return
	}
	found := false
	for _, p := range m.Snapshot.Players {
		if !valid(p.State) {
			return
		}
		if p.ID == s.self.ID {
			found = true
		}
	}
	for _, c := range append(append([]Car(nil), m.Snapshot.AI...), m.Snapshot.Traffic...) {
		if !valid(c) {
			return
		}
	}
	if !found {
		return
	}
	s.token = m.Token
	s.receiveSeq = m.Seq
	s.latest = m.Snapshot
	s.lastHost = time.Now()
	s.status = "Connected to " + m.Snapshot.Players[0].Name
}

type Room struct{ Address, Name string }

// Discover listens on an ephemeral port, so clients can discover on a host PC too.
func Discover(port int, timeout time.Duration) []Room {
	c, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero})
	if e != nil {
		return nil
	}
	defer c.Close()
	c.SetReadDeadline(time.Now().Add(timeout))
	raw, _ := json.Marshal(wire{Protocol: Protocol, Kind: "discover"})
	targets := []net.IP{net.IPv4bcast, net.IPv4(127, 0, 0, 1)}
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		if p, ok := a.(*net.IPNet); ok && p.IP.To4() != nil && !p.IP.IsLoopback() {
			ip := p.IP.To4()
			mask := p.Mask
			if len(mask) == 4 {
				targets = append(targets, net.IPv4(ip[0]|^mask[0], ip[1]|^mask[1], ip[2]|^mask[2], ip[3]|^mask[3]))
			}
		}
	}
	for _, ip := range targets {
		c.WriteToUDP(raw, &net.UDPAddr{IP: ip, Port: port})
	}
	rooms := map[string]Room{}
	b := make([]byte, 2048)
	for {
		n, a, e := c.ReadFromUDP(b)
		if e != nil {
			break
		}
		var m wire
		if json.Unmarshal(b[:n], &m) == nil && m.Protocol == Protocol && m.Kind == "room" {
			rooms[a.String()] = Room{Address: a.String(), Name: m.Name}
		}
	}
	out := []Room{}
	for _, r := range rooms {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Address < out[j].Address })
	return out
}

func (s *Session) SetRace(id string, lap, gate int, finished bool, seconds float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.host {
		return
	}
	if p, ok := s.peers[id]; ok {
		p.Player.State.Lap = lap
		p.Player.State.Gate = gate
		p.Player.State.Finished = finished
		p.Player.State.RaceTime = seconds
	}
}

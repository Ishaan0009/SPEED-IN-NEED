package engine

import (
	"container/heap"
	"math"
	"sort"
)

// RoadNavigator splits the authored road polylines at real, at-grade junctions.
// Overpasses retain their elevation and cannot become imaginary turns.
type RoadNavigator struct {
	Nodes      []NavNode033
	Segments   []NavSegment033
	Components int
	links      [][]navLink033
}
type NavNode033 struct {
	Pos       Vec
	Y         float64
	Component int
}
type NavSegment033 struct{ A, B, Road int }
type navLink033 struct {
	To     int
	Metres float64
}
type NavPath033 struct {
	Points                      []Vec
	Start, RoadEnd, Destination Vec
	Metres, OffRoadMetres       float64
	Reachable                   bool
}
type navSource033 struct {
	a, b        Vec
	ya, yb      float64
	road        int
	first, last bool
	cuts        []float64
}
type navConnector033 struct {
	a, b   Vec
	ya, yb float64
}

func navDot033(a, b Vec) float64   { return a.X*b.X + a.Z*b.Z }
func navCross033(a, b Vec) float64 { return a.X*b.Z - a.Z*b.X }
func navProject033(p, a, b Vec) (Vec, float64) {
	d := b.Sub(a)
	den := navDot033(d, d)
	t := 0.0
	if den > 1e-12 {
		t = Clamp(navDot033(p.Sub(a), d)/den, 0, 1)
	}
	return a.Add(d.Mul(t)), t
}
func navFinite033(p Vec) bool {
	return !math.IsNaN(p.X) && !math.IsNaN(p.Z) && !math.IsInf(p.X, 0) && !math.IsInf(p.Z, 0)
}

func NewRoadNavigator(d *WorldDefinition) *RoadNavigator {
	g := &RoadNavigator{}
	if d == nil {
		return g
	}
	sources := []navSource033{}
	for ri, r := range d.Roads {
		if r.Kind == "stunt" {
			continue
		}
		for i := 0; i+1 < len(r.Points); i++ {
			a, b := r.Points[i], r.Points[i+1]
			if !navFinite033(a) || !navFinite033(b) || a.Dist(b) < .001 {
				continue
			}
			ya, yb := d.Grid.Height(a), d.Grid.Height(b)
			if len(r.Elevations) == len(r.Points) {
				ya, yb = r.Elevations[i], r.Elevations[i+1]
			}
			physicalHeight := func(p Vec, y float64) float64 {
				if r.Kind == "tunnel" {
					for _, t := range d.Tunnels {
						if floor, ok := t.Floor(p); ok {
							return floor
						}
					}
				}
				for _, bridge := range d.BridgeSpans {
					if bridge.Route == ri {
						q, t := navProject033(p, bridge.Start, bridge.End)
						if t > 0 && t < 1 && q.Dist(p) < bridge.Width*.5+1 {
							return bridge.Y
						}
					}
				}
				return y
			}
			ya, yb = physicalHeight(a, ya), physicalHeight(b, yb)
			sources = append(sources, navSource033{a: a, b: b, ya: ya, yb: yb, road: ri, first: i == 0, last: i+2 == len(r.Points), cuts: []float64{0, 1}})
		}
	}
	grid := map[[2]int][]int{}
	connectors := []navConnector033{}
	for i := range sources {
		s := &sources[i]
		seen := map[int]bool{}
		loX, hiX := int(math.Floor((math.Min(s.a.X, s.b.X)-2.1)/64)), int(math.Floor((math.Max(s.a.X, s.b.X)+2.1)/64))
		loZ, hiZ := int(math.Floor((math.Min(s.a.Z, s.b.Z)-2.1)/64)), int(math.Floor((math.Max(s.a.Z, s.b.Z)+2.1)/64))
		for x := loX; x <= hiX; x++ {
			for z := loZ; z <= hiZ; z++ {
				for _, j := range grid[[2]int{x, z}] {
					if seen[j] {
						continue
					}
					seen[j] = true
					t := &sources[j]
					u, v := s.b.Sub(s.a), t.b.Sub(t.a)
					den := navCross033(u, v)
					if math.Abs(den) > 1e-10 {
						delta := t.a.Sub(s.a)
						a, b := navCross033(delta, v)/den, navCross033(delta, u)/den
						if a >= -1e-8 && a <= 1+1e-8 && b >= -1e-8 && b <= 1+1e-8 && math.Abs(Lerp(s.ya, s.yb, a)-Lerp(t.ya, t.yb, b)) < 2.5 {
							s.cuts = append(s.cuts, Clamp(a, 0, 1))
							t.cuts = append(t.cuts, Clamp(b, 0, 1))
						}
					}
					// A T-junction may end two metres from the other centreline while
					// both paved surfaces meet. Join only whole-road endpoints, never
					// adjacent parallel lanes or elevated crossings.
					near := func(a, b *navSource033) {
						for _, e := range []struct {
							p   Vec
							y   float64
							end bool
						}{{a.a, a.ya, a.first}, {a.b, a.yb, a.last}} {
							if !e.end {
								continue
							}
							p, f := navProject033(e.p, b.a, b.b)
							y := Lerp(b.ya, b.yb, f)
							if p.Dist(e.p) <= 2.0 && math.Abs(e.y-y) < 2.5 {
								b.cuts = append(b.cuts, f)
								connectors = append(connectors, navConnector033{e.p, p, e.y, y})
							}
						}
					}
					near(s, t)
					near(t, s)
				}
			}
		}
		for x := loX; x <= hiX; x++ {
			for z := loZ; z <= hiZ; z++ {
				grid[[2]int{x, z}] = append(grid[[2]int{x, z}], i)
			}
		}
	}
	positions := map[[2]int][]int{}
	node := func(p Vec, y float64) int {
		k := [2]int{int(math.Floor(p.X)), int(math.Floor(p.Z))}
		for x := k[0] - 1; x <= k[0]+1; x++ {
			for z := k[1] - 1; z <= k[1]+1; z++ {
				for _, id := range positions[[2]int{x, z}] {
					n := g.Nodes[id]
					if p.Dist(n.Pos) < .08 && math.Abs(y-n.Y) < 2.5 {
						return id
					}
				}
			}
		}
		id := len(g.Nodes)
		g.Nodes = append(g.Nodes, NavNode033{Pos: p, Y: y, Component: -1})
		g.links = append(g.links, nil)
		positions[k] = append(positions[k], id)
		return id
	}
	link := func(a, b, road int) {
		if a == b {
			return
		}
		length := math.Hypot(g.Nodes[a].Pos.Dist(g.Nodes[b].Pos), g.Nodes[a].Y-g.Nodes[b].Y)
		if length < .001 {
			return
		}
		g.links[a] = append(g.links[a], navLink033{b, length})
		g.links[b] = append(g.links[b], navLink033{a, length})
		g.Segments = append(g.Segments, NavSegment033{a, b, road})
	}
	for _, s := range sources {
		sort.Float64s(s.cuts)
		last := -1
		previous := -1.0
		for _, f := range s.cuts {
			if f-previous < 1e-7 {
				continue
			}
			previous = f
			id := node(s.a.Add(s.b.Sub(s.a).Mul(f)), Lerp(s.ya, s.yb, f))
			if last >= 0 {
				link(last, id, s.road)
			}
			last = id
		}
	}
	for _, c := range connectors {
		link(node(c.a, c.ya), node(c.b, c.yb), -1)
	}
	for i := range g.Nodes {
		if g.Nodes[i].Component >= 0 {
			continue
		}
		component := g.Components
		g.Components++
		queue := []int{i}
		g.Nodes[i].Component = component
		for at := 0; at < len(queue); at++ {
			for _, edge := range g.links[queue[at]] {
				if g.Nodes[edge.To].Component < 0 {
					g.Nodes[edge.To].Component = component
					queue = append(queue, edge.To)
				}
			}
		}
	}
	return g
}

type navSnap033 struct {
	edge     int
	p        Vec
	t, score float64
}

func (g *RoadNavigator) closest(p Vec, y float64, component int) navSnap033 {
	best := navSnap033{edge: -1, score: math.Inf(1)}
	for i, s := range g.Segments {
		a, b := g.Nodes[s.A], g.Nodes[s.B]
		if component >= 0 && a.Component != component {
			continue
		}
		q, t := navProject033(p, a.Pos, b.Pos)
		distance := q.Dist(p)
		score := distance * distance
		if !math.IsNaN(y) {
			height := math.Max(0, math.Abs(Lerp(a.Y, b.Y, t)-y)-2.5)
			score += height * height * 4
		}
		if score < best.score {
			best = navSnap033{i, q, t, score}
		}
	}
	return best
}

type navQueueItem033 struct {
	id             int
	cost, priority float64
}
type navQueue033 []navQueueItem033

func (q navQueue033) Len() int           { return len(q) }
func (q navQueue033) Less(i, j int) bool { return q[i].priority < q[j].priority }
func (q navQueue033) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *navQueue033) Push(x any)        { *q = append(*q, x.(navQueueItem033)) }
func (q *navQueue033) Pop() any          { a := *q; x := a[len(a)-1]; *q = a[:len(a)-1]; return x }

// FindPath snaps the destination to the nearest road reachable from the
// player's current road, then runs A*. No off-road shortcut is called a route.
func (g *RoadNavigator) FindPath(start, target Vec, startY float64) NavPath033 {
	out := NavPath033{Destination: target}
	if g == nil || !navFinite033(start) || !navFinite033(target) {
		return out
	}
	s := g.closest(start, startY, -1)
	if s.edge < 0 {
		return out
	}
	a := g.Segments[s.edge]
	t := g.closest(target, math.NaN(), g.Nodes[a.A].Component)
	if t.edge < 0 {
		return out
	}
	b := g.Segments[t.edge]
	out.Start, out.RoadEnd, out.OffRoadMetres = s.p, t.p, t.p.Dist(target)
	if s.edge == t.edge {
		out.Points = []Vec{s.p, t.p}
		out.Metres = s.p.Dist(t.p)
		out.Reachable = true
		return out
	}
	distance := make([]float64, len(g.Nodes))
	parent := make([]int, len(g.Nodes))
	for i := range distance {
		distance[i] = math.Inf(1)
		parent[i] = -1
	}
	queue := &navQueue033{}
	heap.Init(queue)
	for _, id := range []int{a.A, a.B} {
		cost := s.p.Dist(g.Nodes[id].Pos)
		distance[id] = cost
		heap.Push(queue, navQueueItem033{id, cost, cost + g.Nodes[id].Pos.Dist(t.p)})
	}
	best, end := math.Inf(1), -1
	for queue.Len() > 0 {
		item := heap.Pop(queue).(navQueueItem033)
		if item.cost > distance[item.id]+1e-7 {
			continue
		}
		if item.priority >= best {
			break
		}
		if item.id == b.A || item.id == b.B {
			cost := item.cost + g.Nodes[item.id].Pos.Dist(t.p)
			if cost < best {
				best, end = cost, item.id
			}
		}
		for _, e := range g.links[item.id] {
			cost := item.cost + e.Metres
			if cost+1e-7 < distance[e.To] {
				distance[e.To] = cost
				parent[e.To] = item.id
				heap.Push(queue, navQueueItem033{e.To, cost, cost + g.Nodes[e.To].Pos.Dist(t.p)})
			}
		}
	}
	if end < 0 {
		return out
	}
	back := []Vec{t.p}
	for id := end; id >= 0; id = parent[id] {
		back = append(back, g.Nodes[id].Pos)
	}
	back = append(back, s.p)
	for i := len(back) - 1; i >= 0; i-- {
		if len(out.Points) == 0 || out.Points[len(out.Points)-1].Dist(back[i]) > .01 {
			out.Points = append(out.Points, back[i])
		}
	}
	out.Metres, out.Reachable = best, true
	return out
}

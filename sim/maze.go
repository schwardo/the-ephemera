// Computes shortest paths between every pair of openable rooms in the 4x4x4
// hotel, scoring each path as (moves) + 0.5 * (direction changes), then groups
// the pairs into Easy/Medium/Hard tertiles by score and emits JSON to stdout.
package main

import (
	"container/heap"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var openableRe = regexp.MustCompile(`^([1-4](12|21|34|43)|(12|21|34|43)[1-4])$`)

func canOpen(room int) bool {
	return openableRe.MatchString(strconv.Itoa(room))
}

type direction struct{ dx, dy, dz int }

var directions = []direction{
	{1, 0, 0}, {-1, 0, 0},
	{0, 1, 0}, {0, -1, 0},
	{0, 0, 1}, {0, 0, -1},
}

type neighbor struct {
	room, dir int
}

func neighborsOf(room int) []neighbor {
	x := room / 100
	y := (room / 10) % 10
	z := room % 10
	out := make([]neighbor, 0, 6)
	for i, d := range directions {
		nx, ny, nz := x+d.dx, y+d.dy, z+d.dz
		if nx < 1 || nx > 4 || ny < 1 || ny > 4 || nz < 1 || nz > 4 {
			continue
		}
		nr := 100*nx + 10*ny + nz
		if !canOpen(nr) {
			continue
		}
		out = append(out, neighbor{nr, i})
	}
	return out
}

type state struct {
	room, dir int
}

type item struct {
	cost float64
	s    state
}

type pq []*item

func (p pq) Len() int            { return len(p) }
func (p pq) Less(i, j int) bool  { return p[i].cost < p[j].cost }
func (p pq) Swap(i, j int)       { p[i], p[j] = p[j], p[i] }
func (p *pq) Push(x interface{}) { *p = append(*p, x.(*item)) }
func (p *pq) Pop() interface{}   { o := *p; n := len(o); x := o[n-1]; *p = o[:n-1]; return x }

// shortestScore returns moves + 0.5*directionChanges along the optimal route
// from start to dest. Adjacent moves through openable rooms only.
func shortestScore(start, dest int) float64 {
	if start == dest {
		return 0
	}
	dist := map[state]float64{}
	open := &pq{}
	heap.Init(open)
	initial := state{start, -1}
	dist[initial] = 0
	heap.Push(open, &item{0, initial})
	for open.Len() > 0 {
		it := heap.Pop(open).(*item)
		if d, ok := dist[it.s]; ok && it.cost > d {
			continue
		}
		if it.s.room == dest {
			return it.cost
		}
		for _, n := range neighborsOf(it.s.room) {
			extra := 1.0
			if it.s.dir != -1 && it.s.dir != n.dir {
				extra += 0.5
			}
			ns := state{n.room, n.dir}
			nc := it.cost + extra
			if d, ok := dist[ns]; !ok || nc < d {
				dist[ns] = nc
				heap.Push(open, &item{nc, ns})
			}
		}
	}
	return -1
}

func main() {
	var rooms []int
	for x := 1; x <= 4; x++ {
		for y := 1; y <= 4; y++ {
			for z := 1; z <= 4; z++ {
				r := 100*x + 10*y + z
				if canOpen(r) {
					rooms = append(rooms, r)
				}
			}
		}
	}

	type entry struct {
		start, dest int
		score       float64
	}
	var entries []entry
	for _, s := range rooms {
		for _, d := range rooms {
			if s == d {
				continue
			}
			entries = append(entries, entry{s, d, shortestScore(s, d)})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].score != entries[j].score {
			return entries[i].score < entries[j].score
		}
		if entries[i].start != entries[j].start {
			return entries[i].start < entries[j].start
		}
		return entries[i].dest < entries[j].dest
	})

	n := len(entries)
	// Bucket by score so equal-score pairs land together. Pick the two cut
	// scores closest to the 1/3 and 2/3 points.
	cut1 := entries[n/3].score
	cut2 := entries[2*n/3].score
	if cut2 == cut1 {
		// Find the next score above cut1 if the 1/3 and 2/3 marks land on
		// the same value (small graph quirk).
		for _, e := range entries {
			if e.score > cut1 {
				cut2 = e.score
				break
			}
		}
	}

	out := map[string][][]int{"easy": {}, "medium": {}, "hard": {}}
	for _, e := range entries {
		pair := []int{e.start, e.dest}
		switch {
		case e.score < cut1:
			out["easy"] = append(out["easy"], pair)
		case e.score < cut2:
			out["medium"] = append(out["medium"], pair)
		default:
			out["hard"] = append(out["hard"], pair)
		}
	}

	// Compact JSON, one inner pair per element, all pairs on a single line
	// per bucket — easier to embed as a JS literal without giant whitespace.
	fmt.Println("{")
	keys := []string{"easy", "medium", "hard"}
	for ki, k := range keys {
		pairs := out[k]
		parts := make([]string, len(pairs))
		for i, p := range pairs {
			parts[i] = fmt.Sprintf("[%d,%d]", p[0], p[1])
		}
		comma := ","
		if ki == len(keys)-1 {
			comma = ""
		}
		fmt.Printf("  %q: [%s]%s\n", k, strings.Join(parts, ","), comma)
	}
	fmt.Println("}")

	fmt.Fprintf(os.Stderr, "openable rooms: %d, total pairs: %d\n", len(rooms), n)
	fmt.Fprintf(os.Stderr, "easy:   %4d pairs (score < %.1f)\n", len(out["easy"]), cut1)
	fmt.Fprintf(os.Stderr, "medium: %4d pairs (%.1f <= score < %.1f)\n", len(out["medium"]), cut1, cut2)
	fmt.Fprintf(os.Stderr, "hard:   %4d pairs (score >= %.1f, max %.1f)\n", len(out["hard"]), cut2, entries[n-1].score)
}

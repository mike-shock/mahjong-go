// Package game implements the rules of Mahjong solitaire independently of the GUI.
package game

import "math/rand"

type Layout int

const (
	Classic Layout = iota
	Pyramid
	Fortress
	layoutCount
)

var layoutNames = [...]string{"Классика", "Пирамида", "Крепость"}

type rectangle struct{ x, y, width, height int }

// A layer can contain disjoint rectangles, allowing courtyards and towers.
type layer []rectangle

var layouts = [...][]layer{
	Classic: {
		{{x: 0, y: 0, width: 12, height: 8}},
		{{x: 1, y: 2, width: 10, height: 4}},
		{{x: 4, y: 3, width: 4, height: 2}},
	},
	Pyramid: {
		{{x: 1, y: 0, width: 10, height: 8}},
		{{x: 2, y: 1, width: 8, height: 6}},
		{{x: 4, y: 2, width: 4, height: 4}},
	},
	Fortress: {
		// Each wall level has 12*8 - 8*4 = 64 tiles.
		{
			{x: 0, y: 0, width: 12, height: 2},
			{x: 0, y: 6, width: 12, height: 2},
			{x: 0, y: 2, width: 2, height: 4},
			{x: 10, y: 2, width: 2, height: 4},
		},
		{
			{x: 0, y: 0, width: 12, height: 2},
			{x: 0, y: 6, width: 12, height: 2},
			{x: 0, y: 2, width: 2, height: 4},
			{x: 10, y: 2, width: 2, height: 4},
		},
		{
			{x: 0, y: 0, width: 2, height: 2},
			{x: 10, y: 0, width: 2, height: 2},
			{x: 0, y: 6, width: 2, height: 2},
			{x: 10, y: 6, width: 2, height: 2},
		},
	},
}

func LayoutNames() []string { return append([]string(nil), layoutNames[:]...) }

func (l Layout) Name() string {
	if l < 0 || l >= layoutCount {
		return "Неизвестная"
	}
	return layoutNames[l]
}

type Tile struct {
	X, Y, Z, Kind int
	Removed       bool
}

type Game struct {
	Tiles   []Tile
	Layout  Layout
	history [][2]int
	rng     *rand.Rand
}

func New(seed int64) *Game {
	rng := rand.New(rand.NewSource(seed))
	return newWithRNG(rng, Layout(rng.Intn(int(layoutCount))))
}

func NewWithLayout(seed int64, layout Layout) *Game {
	if layout < 0 || layout >= layoutCount {
		layout = Classic
	}
	return newWithRNG(rand.New(rand.NewSource(seed)), layout)
}

func newWithRNG(rng *rand.Rand, layout Layout) *Game {
	g := &Game{Layout: layout, rng: rng}
	for z, currentLayer := range layouts[layout] {
		for _, area := range currentLayer {
			for y := area.y; y < area.y+area.height; y++ {
				for x := area.x; x < area.x+area.width; x++ {
					g.Tiles = append(g.Tiles, Tile{X: x, Y: y, Z: z})
				}
			}
		}
	}
	g.Shuffle()
	return g
}

// Free means uncovered, with at least one open horizontal side.
func (g *Game) Free(i int) bool {
	if i < 0 || i >= len(g.Tiles) || g.Tiles[i].Removed {
		return false
	}
	t := g.Tiles[i]
	left, right := false, false
	for j, o := range g.Tiles {
		if j == i || o.Removed {
			continue
		}
		if o.X == t.X && o.Y == t.Y && o.Z > t.Z {
			return false
		}
		if o.Z == t.Z && o.Y == t.Y {
			left = left || o.X == t.X-1
			right = right || o.X == t.X+1
		}
	}
	return !left || !right
}

func (g *Game) Remaining() int {
	n := 0
	for _, t := range g.Tiles {
		if !t.Removed {
			n++
		}
	}
	return n
}

func (g *Game) Match(a, b int) bool {
	if a == b || !g.Free(a) || !g.Free(b) || g.Tiles[a].Kind != g.Tiles[b].Kind {
		return false
	}
	g.Tiles[a].Removed = true
	g.Tiles[b].Removed = true
	g.history = append(g.history, [2]int{a, b})
	return true
}

func (g *Game) Hint() (int, int, bool) {
	for a := range g.Tiles {
		if !g.Free(a) {
			continue
		}
		for b := a + 1; b < len(g.Tiles); b++ {
			if g.Free(b) && g.Tiles[a].Kind == g.Tiles[b].Kind {
				return a, b, true
			}
		}
	}
	return 0, 0, false
}

func (g *Game) Undo() bool {
	if len(g.history) == 0 {
		return false
	}
	p := g.history[len(g.history)-1]
	g.history = g.history[:len(g.history)-1]
	g.Tiles[p[0]].Removed = false
	g.Tiles[p[1]].Removed = false
	return true
}

// Shuffle assigns pairs along a legal removal order, guaranteeing a solution.
// Existing removed tiles remain removed; undo history is cleared.
func (g *Game) Shuffle() {
	g.shuffle()
}

// shuffle returns the solution used for the deal so tests can replay each move.
func (g *Game) shuffle() [][2]int {
	remaining := []int{}
	kinds := []int{}
	counts := map[int]int{}
	for i, t := range g.Tiles {
		if !t.Removed {
			remaining = append(remaining, i)
			counts[t.Kind]++
		}
	}
	if len(remaining) == 0 {
		return nil
	}
	// The first deal contains four copies of each of 36 designs.
	if len(remaining) == 144 && len(counts) == 1 {
		for k := 0; k < 36; k++ {
			kinds = append(kinds, k, k)
		}
	} else {
		for k := 0; k < 36; k++ {
			for n := 0; n < counts[k]/2; n++ {
				kinds = append(kinds, k)
			}
		}
	}
	g.rng.Shuffle(len(kinds), func(i, j int) { kinds[i], kinds[j] = kinds[j], kinds[i] })
	pairs := make([][2]int, 0, len(kinds))
	if !g.findRemovalPairs(len(remaining), &pairs) {
		panic("mahjong: layout has no complete removal sequence")
	}
	for pair, k := range kinds {
		g.Tiles[pairs[pair][0]].Kind = k
		g.Tiles[pairs[pair][1]].Kind = k
	}
	for _, i := range remaining {
		g.Tiles[i].Removed = false
	}
	g.history = nil
	return pairs
}

// findRemovalPairs searches for a complete sequence before kinds are assigned.
// This prevents a random choice from stranding a final tile.
func (g *Game) findRemovalPairs(remaining int, pairs *[][2]int) bool {
	if remaining == 0 {
		return true
	}
	free := make([]int, 0, remaining)
	for i := range g.Tiles {
		if g.Free(i) {
			free = append(free, i)
		}
	}
	if len(free) < 2 {
		return false
	}
	g.rng.Shuffle(len(free), func(i, j int) { free[i], free[j] = free[j], free[i] })
	a := free[0]
	for _, b := range free[1:] {
		g.Tiles[a].Removed = true
		g.Tiles[b].Removed = true
		*pairs = append(*pairs, [2]int{a, b})
		if g.findRemovalPairs(remaining-2, pairs) {
			return true
		}
		*pairs = (*pairs)[:len(*pairs)-1]
		g.Tiles[a].Removed = false
		g.Tiles[b].Removed = false
	}
	return false
}

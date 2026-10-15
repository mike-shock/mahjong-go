package game

import "testing"

func TestLayoutCoordinates(t *testing.T) {
	inside := func(x, y, left, top, width, height int) bool {
		return x >= left && x < left+width && y >= top && y < top+height
	}
	expected := []func(int, int, int) bool{
		Classic: func(x, y, z int) bool {
			return z == 0 && inside(x, y, 0, 0, 12, 8) ||
				z == 1 && inside(x, y, 1, 2, 10, 4) ||
				z == 2 && inside(x, y, 4, 3, 4, 2)
		},
		Pyramid: func(x, y, z int) bool {
			return z == 0 && inside(x, y, 0, 0, 11, 7) ||
				z == 1 && inside(x, y, 1, 1, 9, 5) ||
				z == 2 && inside(x, y, 2, 2, 7, 3) ||
				z == 3 && x == 5 && y == 3
		},
		Fortress: func(x, y, z int) bool {
			wall := inside(x, y, 0, 0, 12, 8) && !inside(x, y, 2, 2, 8, 4)
			tower := (x < 2 || x >= 10) && (y < 2 || y >= 6)
			return wall && (z == 0 || z == 1 || z == 2 && tower)
		},
	}
	positions := make([]map[[3]int]bool, layoutCount)
	for layout := Classic; layout < layoutCount; layout++ {
		t.Run(layout.Name(), func(t *testing.T) {
			g := NewWithLayout(0, layout)
			layerCount := len(layouts[layout])
			positions[layout] = make(map[[3]int]bool)
			for _, tile := range g.Tiles {
				p := [3]int{tile.X, tile.Y, tile.Z}
				if tile.X < 0 || tile.X >= 12 || tile.Y < 0 || tile.Y >= 8 || tile.Z < 0 || tile.Z >= layerCount {
					t.Fatalf("position outside board: %v", p)
				}
				if positions[layout][p] {
					t.Fatalf("duplicate position: %v", p)
				}
				positions[layout][p] = true
			}
			if len(positions[layout]) != 144 {
				t.Fatalf("got %d positions, want 144", len(positions[layout]))
			}
			for z := 0; z < layerCount; z++ {
				for y := 0; y < 8; y++ {
					for x := 0; x < 12; x++ {
						p := [3]int{x, y, z}
						if positions[layout][p] != expected[layout](x, y, z) {
							t.Errorf("unexpected occupancy at %v", p)
						}
						if z > 0 && positions[layout][p] && !positions[layout][[3]int{x, y, z - 1}] {
							t.Errorf("unsupported tile at %v", p)
						}
					}
				}
			}
		})
	}
	for a := Classic; a < layoutCount; a++ {
		for b := a + 1; b < layoutCount; b++ {
			different := false
			for p := range positions[a] {
				if !positions[b][p] {
					different = true
					break
				}
			}
			if !different {
				t.Errorf("%s and %s have identical coordinates", a.Name(), b.Name())
			}
		}
	}
}

func TestPyramidHasCenteredSummitAndShrinkingTiers(t *testing.T) {
	g := NewWithLayout(42, Pyramid)
	positions := make(map[[3]int]bool)
	var counts [4]int
	for _, tile := range g.Tiles {
		if tile.Z < 0 || tile.Z >= len(counts) {
			t.Fatalf("unexpected pyramid tier %d", tile.Z)
		}
		positions[[3]int{tile.X, tile.Y, tile.Z}] = true
		counts[tile.Z]++
	}
	if counts != [4]int{77, 45, 21, 1} {
		t.Fatalf("got tier sizes %v, want [77 45 21 1]", counts)
	}
	if !positions[[3]int{5, 3, 3}] {
		t.Fatal("summit is not at the center (5, 3)")
	}
	for p := range positions {
		// Both mirror axes must pass through the summit on every tier.
		if !positions[[3]int{10 - p[0], p[1], p[2]}] || !positions[[3]int{p[0], 6 - p[1], p[2]}] {
			t.Errorf("tier is not symmetric around the summit at %v", p)
		}
		if p[2] > 0 {
			// Each upper tier must be inset on all four sides, not merely
			// repeat the tier beneath it or create an unsupported overhang.
			for _, offset := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
				if !positions[[3]int{p[0] + offset[0], p[1] + offset[1], p[2] - 1}] {
					t.Errorf("tier is not inset at %v toward %v", p, offset)
				}
			}
		}
	}
}

func TestLayoutsHaveCompleteMatchingSolution(t *testing.T) {
	for layout := Classic; layout < layoutCount; layout++ {
		for seed := int64(0); seed < 50; seed++ {
			g := NewWithLayout(seed, layout)
			pairs := g.shuffle()
			if len(pairs) != 72 {
				t.Fatalf("%s seed %d: got %d solution pairs, want 72", layout.Name(), seed, len(pairs))
			}
			for move, pair := range pairs {
				if !g.Match(pair[0], pair[1]) {
					t.Fatalf("%s seed %d: solution move %d is not a legal matching pair", layout.Name(), seed, move)
				}
				// Also verify a shuffle after some tiles have been removed.
				if move == 9 {
					partial := *g
					partial.Tiles = append([]Tile(nil), g.Tiles...)
					partial.history = nil
					for _, next := range partial.shuffle() {
						if !partial.Match(next[0], next[1]) {
							t.Fatalf("%s seed %d: shuffled partial deal has an illegal solution move", layout.Name(), seed)
						}
					}
					if partial.Remaining() != 0 {
						t.Fatalf("%s seed %d: shuffled partial deal was not cleared", layout.Name(), seed)
					}
				}
			}
			if g.Remaining() != 0 {
				t.Fatalf("%s seed %d: solution did not clear the board", layout.Name(), seed)
			}
		}
	}
}

func TestLayoutsCreateCompleteValidDeals(t *testing.T) {
	for layout := Classic; layout < layoutCount; layout++ {
		for seed := int64(0); seed < 50; seed++ {
			g := NewWithLayout(seed, layout)
			if len(g.Tiles) != 144 {
				t.Fatalf("%s seed %d: got %d tiles, want 144", layout.Name(), seed, len(g.Tiles))
			}
			positions := make(map[[3]int]bool, len(g.Tiles))
			counts := make(map[int]int, 36)
			for _, tile := range g.Tiles {
				position := [3]int{tile.X, tile.Y, tile.Z}
				if positions[position] {
					t.Fatalf("%s seed %d: duplicate position %v", layout.Name(), seed, position)
				}
				positions[position] = true
				counts[tile.Kind]++
			}
			for kind := 0; kind < 36; kind++ {
				if counts[kind] != 4 {
					t.Fatalf("%s seed %d: kind %d appears %d times, want 4", layout.Name(), seed, kind, counts[kind])
				}
			}
			if _, _, ok := g.Hint(); !ok {
				t.Fatalf("%s seed %d: initial deal has no available pair", layout.Name(), seed)
			}
		}
	}
}

func TestNewChoosesDifferentLayouts(t *testing.T) {
	seen := make(map[Layout]bool)
	for seed := int64(0); seed < 100; seed++ {
		seen[New(seed).Layout] = true
	}
	if len(seen) != int(layoutCount) {
		t.Fatalf("random new games used %d layouts, want %d", len(seen), layoutCount)
	}
}

func TestShufflePreservesRemovedTilesAndRemainingKinds(t *testing.T) {
	for layout := Classic; layout < layoutCount; layout++ {
		g := NewWithLayout(42, layout)
		a, b, ok := g.Hint()
		if !ok || !g.Match(a, b) {
			t.Fatalf("%s: could not remove initial hinted pair", layout.Name())
		}
		removed := map[int]bool{a: true, b: true}
		countsBefore := make(map[int]int)
		for _, tile := range g.Tiles {
			if !tile.Removed {
				countsBefore[tile.Kind]++
			}
		}
		g.Shuffle()
		for index, tile := range g.Tiles {
			if tile.Removed != removed[index] {
				t.Fatalf("%s: shuffle changed removed state at tile %d", layout.Name(), index)
			}
		}
		countsAfter := make(map[int]int)
		for _, tile := range g.Tiles {
			if !tile.Removed {
				countsAfter[tile.Kind]++
			}
		}
		for kind := 0; kind < 36; kind++ {
			if countsAfter[kind] != countsBefore[kind] {
				t.Fatalf("%s: kind %d count changed from %d to %d", layout.Name(), kind, countsBefore[kind], countsAfter[kind])
			}
		}
		if len(g.history) != 0 {
			t.Fatalf("%s: shuffle did not clear undo history", layout.Name())
		}
	}
}

func TestHintedPairsDisappearImmediately(t *testing.T) {
	for layout := Classic; layout < layoutCount; layout++ {
		for seed := int64(0); seed < 20; seed++ {
			g := NewWithLayout(seed, layout)
			for {
				a, b, ok := g.Hint()
				if !ok {
					break
				}
				before := g.Remaining()
				if !g.Match(a, b) {
					t.Fatalf("%s seed %d: hinted pair %d,%d did not match", layout.Name(), seed, a, b)
				}
				if !g.Tiles[a].Removed || !g.Tiles[b].Removed {
					t.Fatalf("%s seed %d: matched pair %d,%d remained visible", layout.Name(), seed, a, b)
				}
				if after := g.Remaining(); after != before-2 {
					t.Fatalf("%s seed %d: remaining count changed from %d to %d", layout.Name(), seed, before, after)
				}
			}
		}
	}
}

func TestLayoutNamesReturnsCopy(t *testing.T) {
	names := LayoutNames()
	names[0] = "changed"
	if Classic.Name() == "changed" {
		t.Fatal("LayoutNames exposed mutable internal state")
	}
}

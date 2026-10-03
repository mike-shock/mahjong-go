package game

import "testing"

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

package board

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"mahjong/internal/game"
)

func TestPyramidProjectionAndHit(t *testing.T) {
	g := game.NewWithLayout(42, game.Pyramid)
	peak := -1
	for i, tile := range g.Tiles {
		x, y := Origin(tile)
		if x < 0 || y < 0 || x+80 > Width || y+48+thickness(tile) > Height {
			t.Fatalf("tile %d is outside the board", i)
		}
		if tile.Z == 3 {
			peak = i
		}
		if g.Free(i) {
			visible := false
			for dy := 2; dy < 48 && !visible; dy += 3 {
				for dx := 2; dx < 80; dx += 3 {
					if Hit(g.Tiles, float32(x+dx), float32(y+dy)) == i {
						visible = true
						break
					}
				}
			}
			if !visible {
				t.Errorf("free tile %d cannot be clicked", i)
			}
		}
	}
	if peak < 0 {
		t.Fatal("no summit")
	}
	x, y := Origin(g.Tiles[peak])
	if got := Hit(g.Tiles, float32(x+40), float32(y+24)); got != peak {
		t.Fatalf("summit click selected %d, want %d", got, peak)
	}
	if got := Hit(g.Tiles, 0, 0); got != -1 {
		t.Fatalf("background selected tile %d", got)
	}
	g.Tiles[peak].Removed = true
	if got := Hit(g.Tiles, float32(x+40), float32(y+24)); got == peak {
		t.Fatal("removed summit still intercepts clicks")
	}
}

func TestRenderPyramid(t *testing.T) {
	artwork := make([]image.Image, 36)
	for kind := range artwork {
		path := filepath.Join("..", "..", "assets", "tiles", fmt.Sprintf("%02d.png", kind))
		file, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		artwork[kind], err = png.Decode(file)
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	g := game.NewWithLayout(42, game.Pyramid)
	preview := Render(g, artwork, -1, -1, -1)
	if preview.Bounds() != image.Rect(0, 0, Width, Height) {
		t.Fatal("incorrect preview bounds")
	}
	// Keep the preview in the test log for visual inspection during development.
	if path := os.Getenv("MAHJONG_PREVIEW_PATH"); path != "" {
		file, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(file, preview)
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}

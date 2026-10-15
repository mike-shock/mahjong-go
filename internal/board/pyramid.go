// Package board projects the pyramid onto an inclined plane without changing
// the logical tile coordinates used by the game rules.
package board

import (
	"image"
	"image/color"
	"image/draw"
	"sort"

	"mahjong/internal/game"
)

const Width, Height = 840, 680

// Origin gives the top-left of a tile's projected bounding box. Raising a tier
// moves it vertically, rather than sliding it sideways across the board.
func Origin(t game.Tile) (int, int) {
	return 300 + (t.X-t.Y)*42, 140 + (t.X+t.Y)*24 - t.Z*38
}

// Order draws distant tiles first within each tier, then all higher tiers.
func Order(tiles []game.Tile) []int {
	indices := make([]int, 0, len(tiles))
	for i, tile := range tiles {
		if !tile.Removed {
			indices = append(indices, i)
		}
	}
	sort.SliceStable(indices, func(a, b int) bool {
		x, y := tiles[indices[a]], tiles[indices[b]]
		if x.Z != y.Z {
			return x.Z < y.Z
		}
		return x.X+x.Y < y.X+y.Y
	})
	return indices
}

// face maps a diamond back to the rectangular artwork's coordinates.
func face(x, y float32) (u, v float32, inside bool) {
	u = (x-40)/80 + y/48
	v = -(x-40)/80 + y/48
	return u, v, u >= 0 && u < 1 && v >= 0 && v < 1
}

func thickness(tile game.Tile) int {
	if tile.Z > 0 {
		return 38 // Join each raised tier to the tier beneath it.
	}
	return 8
}

// Hit tests the exact visible face or side, front to back.
func Hit(tiles []game.Tile, x, y float32) int {
	order := Order(tiles)
	for n := len(order) - 1; n >= 0; n-- {
		i := order[n]
		ox, oy := Origin(tiles[i])
		for depth := 0; depth <= thickness(tiles[i]); depth++ {
			if _, _, ok := face(x-float32(ox), y-float32(oy+depth)); ok {
				return i
			}
		}
	}
	return -1
}

// Render keeps artwork on the inclined top faces and shades the exposed sides.
func Render(g *game.Game, artwork []image.Image, selected, hintA, hintB int) image.Image {
	out := image.NewNRGBA(image.Rect(0, 0, Width, Height))
	draw.Draw(out, out.Bounds(), image.NewUniform(color.NRGBA{R: 24, G: 77, B: 66, A: 255}), image.Point{}, draw.Src)
	for _, i := range Order(g.Tiles) {
		tile := g.Tiles[i]
		ox, oy := Origin(tile)
		free := g.Free(i)
		depth := thickness(tile)
		art := artwork[tile.Kind]
		bounds := art.Bounds()
		for y := 0; y < 48+depth; y++ {
			for x := 0; x < 80; x++ {
				u, v, top := face(float32(x)+0.5, float32(y)+0.5)
				var c color.NRGBA
				if top {
					c = color.NRGBAModel.Convert(art.At(bounds.Min.X+int(u*float32(bounds.Dx())), bounds.Min.Y+int(v*float32(bounds.Dy())))).(color.NRGBA)
					// Composite transparent artwork on an ivory tile face.
					alpha := uint32(c.A)
					c = color.NRGBA{R: uint8((uint32(c.R)*alpha + 244*(255-alpha)) / 255), G: uint8((uint32(c.G)*alpha + 235*(255-alpha)) / 255), B: uint8((uint32(c.B)*alpha + 210*(255-alpha)) / 255), A: 255}
					if !free {
						c.R = uint8(uint32(c.R) * 4 / 5)
						c.G = uint8(uint32(c.G) * 4 / 5)
						c.B = uint8(uint32(c.B) * 4 / 5)
					}
					if u < .035 || v < .035 || u > .965 || v > .965 {
						c = color.NRGBA{R: 215, G: 198, B: 155, A: 255}
						if i == selected {
							c = color.NRGBA{R: 255, G: 174, A: 255}
						} else if i == hintA || i == hintB {
							c = color.NRGBA{R: 20, G: 205, B: 125, A: 255}
						}
					}
				} else if _, _, side := face(float32(x)+.5, float32(y-depth)+.5); side || y >= 24 && y < 24+depth {
					c = color.NRGBA{R: 166, G: 143, B: 97, A: 255}
					if x >= 40 {
						c = color.NRGBA{R: 194, G: 173, B: 127, A: 255}
					}
				} else {
					continue
				}
				out.SetNRGBA(ox+x, oy+y, c)
			}
		}
	}
	return out
}

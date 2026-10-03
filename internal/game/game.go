// Package game implements the rules of Mahjong solitaire independently of the GUI.
package game

import "math/rand"

type Tile struct {
 X, Y, Z, Kind int
 Removed bool
}

type Game struct {
 Tiles []Tile
 history [][2]int
 rng *rand.Rand
}

func New(seed int64) *Game {
 g := &Game{rng: rand.New(rand.NewSource(seed))}
 for z, size := range [][4]int{{0,0,12,8},{1,2,10,4},{4,3,4,2}} {
  for y:=size[1]; y<size[1]+size[3]; y++ {
   for x:=size[0]; x<size[0]+size[2]; x++ { g.Tiles=append(g.Tiles,Tile{X:x,Y:y,Z:z}) }
  }
 }
 g.Shuffle()
 return g
}

// Free means uncovered, with at least one open horizontal side.
func (g *Game) Free(i int) bool {
 if i<0 || i>=len(g.Tiles) || g.Tiles[i].Removed { return false }
 t:=g.Tiles[i]
 left,right:=false,false
 for j,o:=range g.Tiles {
  if j==i || o.Removed {continue}
  if o.X==t.X && o.Y==t.Y && o.Z>t.Z {return false}
  if o.Z==t.Z && o.Y==t.Y {
   left=left || o.X==t.X-1
   right=right || o.X==t.X+1
  }
 }
 return !left || !right
}

func (g *Game) Remaining() int {
 n:=0
 for _,t:=range g.Tiles {if !t.Removed {n++}}
 return n
}

func (g *Game) Match(a,b int) bool {
 if a==b || !g.Free(a) || !g.Free(b) || g.Tiles[a].Kind!=g.Tiles[b].Kind {return false}
 g.Tiles[a].Removed=true
 g.Tiles[b].Removed=true
 g.history=append(g.history,[2]int{a,b})
 return true
}

func (g *Game) Hint() (int,int,bool) {
 for a:=range g.Tiles {
  if !g.Free(a) {continue}
  for b:=a+1;b<len(g.Tiles);b++ {
   if g.Free(b) && g.Tiles[a].Kind==g.Tiles[b].Kind {return a,b,true}
  }
 }
 return 0,0,false
}

func (g *Game) Undo() bool {
 if len(g.history)==0 {return false}
 p:=g.history[len(g.history)-1]
 g.history=g.history[:len(g.history)-1]
 g.Tiles[p[0]].Removed=false
 g.Tiles[p[1]].Removed=false
 return true
}

// Shuffle assigns pairs along a legal removal order, guaranteeing a solution.
// Existing removed tiles remain removed; undo history is cleared.
func (g *Game) Shuffle() {
 remaining:=[]int{}
 kinds:=[]int{}
 counts:=map[int]int{}
 for i,t:=range g.Tiles {if !t.Removed {remaining=append(remaining,i);counts[t.Kind]++}}
 if len(remaining)==0 {return}
 // The first deal contains four copies of each of 36 designs.
 if len(remaining)==144 && len(counts)==1 {
  for k:=0;k<36;k++ {kinds=append(kinds,k,k)}
 } else {
  for k:=0;k<36;k++ {for n:=0;n<counts[k]/2;n++ {kinds=append(kinds,k)}}
 }
 g.rng.Shuffle(len(kinds),func(i,j int){kinds[i],kinds[j]=kinds[j],kinds[i]})
 for pair,k:=range kinds {
  free:=[]int{}
  for _,i:=range remaining {if g.Free(i) {free=append(free,i)}}
  // Aligned rectangular layers always expose at least two tiles.
  g.rng.Shuffle(len(free),func(i,j int){free[i],free[j]=free[j],free[i]})
  for _,i:=range free[:2] {g.Tiles[i].Kind=k;g.Tiles[i].Removed=true}
  _=pair
 }
 for _,i:=range remaining {g.Tiles[i].Removed=false}
 g.history=nil
}

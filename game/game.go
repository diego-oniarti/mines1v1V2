package game

import (
	"math/rand"

	"github.com/diego-oniarti/minesV2/utils"
)

type cell struct {
	x      int
	y      int
	bomb   bool
	flag   bool
	num    int
	hidden bool
}

type Game struct {
	turnA bool
	cells [][]cell

	utils.GameOptions
}

type Update struct {
	x int
	y int
	flag bool
	num int
}

func NewGame(opt utils.GameOptions) Game {
	cells := make([][]cell, opt.Height)
	W := opt.Width
	H := opt.Height

	for y := range(H) {
		cells[y] = make([]cell, W)
		for x := range(W) {
			cells[y][x] = cell {
				x: x,
				y: y,
				bomb: false,
				flag: false,
				num: 0,
				hidden: true,
			}
		}
	}

	r := rand.New(rand.NewSource(int64(opt.Seed)))
	added := 0
	for added<opt.Bombs {
		y := r.Intn(H)
		x := r.Intn(W)
		if cells[y][x].bomb { continue }
		cells[y][x].bomb = true
		for i:=-1; i<=1; i++ {
			yy := y+i
			if yy<0 || yy>=H { continue }
			for j:=-1; j<=1; j++ {
				xx := x+j
				if xx<0 || xx>=W { continue }
				cells[yy][xx].num++
			}
		}
		added++
	}

	g := Game {
		cells: cells,
		turnA: true,
		GameOptions: opt,
	}
	return g
}

func (g *Game) Start() []Update {
	var x, y int
	for {
		x = rand.Intn(g.Width)
		y = rand.Intn(g.Height)
		if g.cells[y][x].num == 0 {
			break
		}
	}
	updates, _ := g.Click(x, y)
	return updates
}

type coords struct{x,y int}
func (g *Game) Click(x, y int) ([]Update, bool) {
	ret := []Update{}
	queue := []coords{{x, y}}

	for len(queue) > 0 {
		var top coords
		top, queue = queue[0], queue[1:]
		x,y := top.x, top.y
		cell := &g.cells[y][x]

		// checks
		if cell.bomb { return nil, true }
		if !cell.hidden { continue }

		// update
		cell.hidden = false
		ret = append(ret, Update{
			x: x,
			y: y,
			num: cell.num,
			flag: false,
		})
		
		// recursion
		if cell.num==0 {
			for i:=-1; i<=1; i++ {
				if y+i<0 || y+i>=g.Height { continue }
				for j:=-1; j<=1; j++ {
					if x+j<0 || x+j>=g.Width { continue }
					queue = append(queue, coords{x+j, y+i})
				}
			}
		}
	}

	return ret, false
}

func (g *Game) Flag(x, y int) (Update, bool) {
	if !g.cells[y][x].hidden { return Update{}, false }
	g.cells[y][x].flag = !g.cells[y][x].flag
	return Update{
		x: x,
		y: y,
		num: 0,
		flag: true,
	}, true
}

func (g *Game) Won() bool {
	for _, row := range g.cells {
		for _, cell := range row {
			if cell.hidden && !cell.bomb {
				return false
			}
		}
	}
	return true
}

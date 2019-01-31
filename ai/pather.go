package ai

import (
	"fmt"
	"github.com/battlesnakeio/starter-snake-go/api"
	"github.com/beefsack/go-astar"
	"log"
	"math"
)

type Pather struct {
	*api.Board
	X, Y  int
	Cost  float64
	Neigh []*Pather
}

func (p *Pather) CostOfTile() float64 {
	for _, s := range p.Board.Snakes {
		for _, b := range s.Body {
			if b.X == p.X && b.Y == p.Y {
				return 100
			}
		}
	}

	return 1
}

func NewPather(parent *Pather, x, y int) *Pather {
	path := &Pather{parent.Board, x, y, 0, make([]*Pather, 0, 4)}
	return path
}

func (p *Pather) Up() *Pather {
	x, y := p.X, p.Y-1
	if y < 0 {
		return nil
	}

	return NewPather(p, x, y)
}

func (p *Pather) Down() *Pather {
	x, y := p.X, p.Y+1
	if y >= p.Board.Height {
		return nil
	}

	return NewPather(p, x, y)
}

func (p *Pather) Right() *Pather {
	x, y := p.X+1, p.Y
	if x >= p.Board.Width {
		return nil
	}

	return NewPather(p, x, y)
}

func (p *Pather) Left() *Pather {
	x, y := p.X-1, p.Y
	if x < 0 {
		return nil
	}

	return NewPather(p, x, y)
}

func (p Pather) PathNeighbors() []astar.Pather {
	out := make([]astar.Pather, 4)
	for i, child := range p.Neigh {
		out[i] = child
	}
	return out
}

func (t Pather) PathNeighborCost(to astar.Pather) float64 {
	out := 1000.0
	if to == nil {
		return out
	}
	if tc, ok := to.(*Pather); ok {
		if tc == nil {
			return out
		}

		out = tc.CostOfTile()
	}
	if tc, ok := to.(Pather); ok {
		//log.Print("ADF")
		//log.Print("Cost", tc.Cost)
		out = tc.CostOfTile()
	}

	log.Print("Neighbour Cost: ", out)
	return out
}

func (t Pather) PathEstimatedCost(to astar.Pather) float64 {
	out := 1000.0
	if to == nil {
		return out
	}
	if tp, ok := to.(*Pather); ok {
		if tp == nil {
			return 100
		}
		//log.Print("JDF")
		out = math.Abs(float64(t.X*tp.X + t.Y*tp.Y))
	}
	//log.Print("C EST")
	if tp, ok := to.(Pather); ok {
		out = math.Abs(float64(t.X*tp.X + t.Y*tp.Y))
	}

	//log.Print("PathEstimatedCost: ", out)

	return out
}

func GetStart(board *api.Board, me *api.Snake) *Pather {
	m := me.Body[0]
	return &Pather{board, m.X, m.Y, 0, make([]*Pather, 0, 4)}
}

func DescribeMove(board *api.Board, from *api.Snake, tx, ty int) (string, error) {
	f := from.Body[0]
	x, y := tx-f.X, ty-f.Y
	log.Print(x, y)

	if (x != 0 && y != 0) || (x == 0 && y == 0) {
		return "", fmt.Errorf("Could not describe move.. x=%d y=%d", x, y)
	}

	if x != 0 {
		if x > 0 {
			return "right", nil
		}
		return "left", nil
	}

	if y != 0 {
		if y > 0 {
			return "down", nil
		}
	}
	return "up", nil
}

func (t *Pather) Ensure(b *Pather) {
	exist := false
	for _, n := range t.Neigh {
		if n.X == b.X && n.Y == b.Y {
			exist = true
			break
		}
	}

	if !exist {
		t.Neigh = append(t.Neigh, b)
	}
}

func Ensure(a, b *Pather) {
	a.Ensure(b)
	b.Ensure(a)
}

func (t *Pather) Generate(game *map[int]map[int]*Pather) {
	paths := []*Pather{
		t.Up(),
		t.Right(),
		t.Down(),
		t.Left(),
	}

	for _, p := range paths {
		if p != nil {
			if x, ok := (*game)[p.X]; ok {
				if _, ok := x[p.Y]; !ok {
					x[p.Y] = p
					Ensure(t, p)
					p.Generate(game)
				} else {
					Ensure(t, p)
				}
			} else {
				(*game)[p.X] = make(map[int]*Pather)
				(*game)[p.X][p.Y] = p
				Ensure(t, p)
				p.Generate(game)
			}

		}
	}
}

func NextMove(req *api.SnakeRequest) (string, error) {
	board := req.Board
	me := req.You
	log.Print(req.Turn)
	log.Print(req.You)
	//return "right", nil
	game := make(map[int]map[int]*Pather)
	start := GetStart(&board, &me)
	game[start.X] = make(map[int]*Pather)
	game[start.X][start.Y] = start

	start.Generate(&game)

	nei := start.PathNeighbors()
	for _, nn := range nei {
		log.Println(nn)
	}

	found := false
	var best []astar.Pather
	var dist float64 = math.MaxFloat64

	for _, f := range board.Food {
		log.Print("Food", f.X, f.Y)
		food := game[f.X][f.Y]

		path, d, fnd := astar.Path(start, food)
		if found && d < dist {
			best = path
			dist = d
			found = fnd
		}
	}
	log.Print("Found: ", found)
	if !found {
		return "", fmt.Errorf("No path found")
	}

	if bst, ok := best[0].(*Pather); ok {
		log.Print("Describing")
		return DescribeMove(&board, &me, bst.X, bst.Y)
	}
	return "", fmt.Errorf("NextMove: Can't convert to Pather")
}

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
	X, Y int
	Cost float64
}

func (p *Pather) CostOfTile(x, y int) float64 {
	for _, s := range p.Board.Snakes {
		for _, b := range s.Body {
			if x == b.X && y == b.Y {
				return 10
			}
		}
	}

	return 0
}

func (p *Pather) Up() *Pather {
	x, y := p.X, p.Y-1
	if y < 0 {
		return nil
	}

	return &Pather{p.Board, x, y, p.CostOfTile(x, y)}
}

func (p *Pather) Down() *Pather {
	x, y := p.X, p.Y+1
	if y >= p.Board.Height {
		return nil
	}

	return &Pather{p.Board, x, y, p.CostOfTile(x, y)}
}

func (p *Pather) Right() *Pather {
	x, y := p.X+1, p.Y
	if x >= p.Board.Width {
		return nil
	}

	return &Pather{p.Board, x, y, p.CostOfTile(x, y)}
}

func (p *Pather) Left() *Pather {
	x, y := p.X-1, p.Y
	if x < 0 {
		return nil
	}

	return &Pather{p.Board, x, y, p.CostOfTile(x, y)}
}

func (p Pather) PathNeighbors() []astar.Pather {
	paths := []astar.Pather{
		p.Up(),
		p.Right(),
		p.Down(),
		p.Left(),
	}

	var out []astar.Pather
	for _, p := range paths {
		if p != nil {
			out = append(out, p)
		}
	}

	return out
}

func (t Pather) PathNeighborCost(to astar.Pather) float64 {
	if tc, ok := to.(*Pather); ok {
		if tc == nil {
			return 100
		}
		return tc.Cost
	}
	if tc, ok := to.(Pather); ok {
		return tc.Cost
	}
	log.Fatalf("PathNeighborCost got bad Pather: %T", to)
	return 100
}

func (t Pather) PathEstimatedCost(to astar.Pather) float64 {
	// TaxiCab distance
	if tp, ok := to.(*Pather); ok {
		if tp == nil {
			return 100
		}
		return float64(t.X*tp.X + t.Y*tp.Y)
	}
	if tp, ok := to.(Pather); ok {
		return float64(t.X*tp.X + t.Y*tp.Y)
	}

	log.Fatalf("PathEstimatedCost got bad Pather: %T", to)
	return 100
}

func GetPather(board *api.Board, x, y int) Pather {
	return Pather{board, x, y, 0}
}

func GetStart(board *api.Board, me *api.Snake) Pather {
	m := me.Body[0]
	return GetPather(board, m.X, m.Y)
}

func DescribeMove(board *api.Board, from *api.Snake, tx, ty int) (string, error) {
	f := from.Body[0]
	x, y := tx-f.X, ty-f.Y
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

func NextMove(board *api.Board, me *api.Snake) (string, error) {
	start := GetStart(board, me)

	found := false
	var best []astar.Pather
	var dist float64 = math.MaxFloat64

	for _, f := range board.Food {
		food := GetPather(board, f.X, f.Y)

		path, d, fnd := astar.Path(start, food)
		if found && d < dist {
			best = path
			dist = d
			found = fnd
		}
	}
	if !found {
		return "", fmt.Errorf("No path found")
	}

	if bst, ok := best[0].(*Pather); ok {
		return DescribeMove(board, me, bst.X, bst.Y)
	}
	return "", fmt.Errorf("NextMove: Can't convert to Pather")
}

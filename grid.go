package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Status uint8

const (
	Unselected Status = iota
	Obstacle
	Start
	End
)

type Cell struct {
	status Status
}

type Grid struct {
	Cells         [][]Cell
	SelectionMode Status
}

func NewGrid(size int) *Grid {
	grid := Grid{
		Cells:         make([][]Cell, size),
		SelectionMode: Obstacle,
	}

	for i := range grid.Cells {
		grid.Cells[i] = make([]Cell, size)
		for j := range grid.Cells[i] {
			grid.Cells[i][j] = Cell{
				status: Unselected,
			}
		}
	}

	return &grid
}

func (g *Grid) Reset() {
	for i := range g.Cells {
		for j := range g.Cells[i] {
			g.Cells[i][j].status = Unselected
		}
	}
}

func (g *Grid) UpdateSubset(size int32) {
	const (
		globalMargin = 50
		blockMargin  = 10
	)

	var space int32 = (screenWidth - 182) / size
	gridSize := int32(len(g.Cells))

	switch rl.GetKeyPressed() {
	case rl.KeyQ:
		g.SelectionMode = Obstacle
	case rl.KeyW:
		g.SelectionMode = Start
	case rl.KeyE:
		g.SelectionMode = End
	}

	// Getting mouse position relative to grid

	mx := rl.GetMouseX()
	my := rl.GetMouseY()

	if mx < globalMargin || my < globalMargin || mx > screenWidth-globalMargin || my > sliderRect.ToInt32().Y {
		return
	}

	x := (mx - 50) / (space + blockMargin)
	y := (my - 50) / (space + blockMargin)

	if x >= gridSize || y >= gridSize || x < 0 || y < 0 {
		return
	}

	if rl.IsMouseButtonDown(rl.MouseLeftButton) {
		if g.Cells[x][y].status != g.SelectionMode {
			g.Cells[x][y].status = g.SelectionMode
		}
	} else if rl.IsMouseButtonDown(rl.MouseRightButton) {
		if g.Cells[x][y].status != Unselected {
			g.Cells[x][y].status = Unselected
		}
	}
}

func (g *Grid) DrawSubset(size int) {
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			var color rl.Color

			switch g.Cells[i][j].status {
			case Unselected:
				color = rl.White
			case Obstacle:
				color = rl.DarkGray
			case Start:
				color = rl.Green
			case End:
				color = rl.Red
			}

			var space int32 = (screenWidth - 182) / int32(size)

			i32 := int32(i)
			j32 := int32(j)

			x := i32*space + 50 + i32*10
			y := j32*space + 50 + j32*10

			if space+x > screenWidth-50 || space+y > int32(sliderRect.Y)-10 {
				continue
			}

			rl.DrawRectangle(
				x,
				y,
				space,
				space,
				color,
			)
		}
	}
}

package main

import (
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Status uint8
type Mode uint8

const (
	Unselected Status = iota
	Obstacle
	Visited
	Frontier
	Start
	End
	Path
)

const (
	PaintMode Mode = iota
	PlayMode
)

type Cell struct {
	status         Status
	score          int
	heuristicScore int
	parent         *Cell
	onScreen       bool
	x              int
	y              int
}

type Grid struct {
	Cells         [][]Cell
	GridMode      Mode
	SelectionMode Status
	Ticks         int
	StartCell     *Cell
	EndCell       *Cell
	FrontierCells []*Cell
}

func NewGrid(size int) *Grid {
	grid := Grid{
		Cells:         make([][]Cell, size),
		SelectionMode: Obstacle,
		GridMode:      PaintMode,
	}

	for i := range grid.Cells {
		grid.Cells[i] = make([]Cell, size)
		for j := range grid.Cells[i] {
			grid.Cells[i][j] = Cell{
				status: Unselected,
				x:      i,
				y:      j,
			}
		}
	}

	return &grid
}

func (g *Grid) GetNeighbors(x, y, size int) []*Cell {
	var neighbors []*Cell

	if x+1 < len(g.Cells) && g.Cells[x+1][y].status != Obstacle && x+1 < size && g.Cells[x+1][y].onScreen {
		neighbors = append(neighbors, &g.Cells[x+1][y])
	}

	if x-1 >= 0 && g.Cells[x-1][y].status != Obstacle {
		neighbors = append(neighbors, &g.Cells[x-1][y])
	}

	if y+1 < len(g.Cells) && g.Cells[x][y+1].status != Obstacle && g.Cells[x][y+1].onScreen {
		neighbors = append(neighbors, &g.Cells[x][y+1])
	}

	if y-1 >= 0 && g.Cells[x][y-1].status != Obstacle {
		neighbors = append(neighbors, &g.Cells[x][y-1])
	}

	return neighbors
}

func (g *Grid) Reset() {
	g.StartCell = nil
	g.EndCell = nil
	g.FrontierCells = nil
	g.GridMode = PaintMode
	g.Ticks = 0
	for i := range g.Cells {
		for j := range g.Cells[i] {
			g.Cells[i][j].status = Unselected
			g.Cells[i][j].score = 0
			g.Cells[i][j].heuristicScore = 0
			g.Cells[i][j].parent = nil
		}
	}
}

func buildPath(cell *Cell) {
	if cell == nil {
		return
	}

	cell.status = Path

	buildPath(cell.parent)
}

func (g *Grid) paintMode(size int32) {
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

	x := (mx - 50) / (space + (int32((100 - size) / 10)))
	y := (my - 50) / (space + (int32((100 - size) / 10)))

	if x >= gridSize || y >= gridSize || x < 0 || y < 0 {
		return
	}

	if rl.IsMouseButtonDown(rl.MouseLeftButton) {
		if g.Cells[x][y].status != g.SelectionMode {
			if g.SelectionMode == Start {
				if g.StartCell != nil {
					g.StartCell.status = Unselected
				}

				g.StartCell = &g.Cells[x][y]

			} else if g.SelectionMode == End {
				if g.EndCell != nil {
					g.EndCell.status = Unselected
				}
				g.EndCell = &g.Cells[x][y]
			}
			g.Cells[x][y].status = g.SelectionMode
		}
	} else if rl.IsMouseButtonDown(rl.MouseRightButton) {
		if g.Cells[x][y].status != Unselected {
			g.Cells[x][y].status = Unselected
		}
	}
}

func popSmallestScore(cells []*Cell) (*Cell, []*Cell) {
	if len(cells) == 0 {
		return nil, nil
	}
	smallest := cells[0]
	smallestIndex := 0
	for i, cell := range cells {
		if cell.score < smallest.score {
			smallest = cell
			smallestIndex = i
		}
	}

	cells = append(cells[:smallestIndex], cells[smallestIndex+1:]...)
	return smallest, cells
}

func (g *Grid) playMode(size, speed int32) bool {
	if g.StartCell == nil || g.EndCell == nil {
		return false
	}

	g.Ticks += int(speed)
	if g.Ticks < 60 {
		return false
	}

	g.Ticks = 0

	currentCell, newCells := popSmallestScore(g.FrontierCells)
	g.FrontierCells = newCells
	if currentCell == nil {
		return false
	}

	neighbors := g.GetNeighbors(currentCell.x, currentCell.y, int(size))

	for _, neighbor := range neighbors {
		if neighbor.status == Start {
			continue
		}
		if neighbor.status == End {
			g.GridMode = PaintMode
			neighbor.parent = currentCell
			buildPath(neighbor)
			return true
		}
		if neighbor.status == Unselected {
			newScore := currentCell.score + 1
			neighbor.score = newScore
			neighbor.status = Frontier
			neighbor.parent = currentCell
			g.FrontierCells = append(g.FrontierCells, neighbor)
		}
		if neighbor.status == Visited {
			newScore := currentCell.score + 1
			if newScore < neighbor.score {
				neighbor.score = newScore
				neighbor.parent = currentCell
			}
		}
	}
	if currentCell == g.EndCell {
		buildPath(currentCell)
		g.GridMode = PaintMode
		return true
	} else {
		if currentCell.status != Start {
			currentCell.status = Visited
		}
	}

	return false
}

func (g *Grid) UpdateSubset(size, speed int32, play bool) bool {

	if play && g.GridMode != PlayMode {
		g.StartCell.score = 0
		g.StartCell.heuristicScore = 0
		g.FrontierCells = append(g.FrontierCells, g.StartCell)
		g.GridMode = PlayMode
	}

	if !play && g.GridMode == PaintMode {
		g.GridMode = PaintMode
	}

	switch g.GridMode {
	case PaintMode:
		g.paintMode(size)
	case PlayMode:
		return g.playMode(size, speed)
	}

	return false
}

func (g *Grid) DrawSubset(size int) {
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			var color rl.Color

			switch g.Cells[i][j].status {
			case Unselected:
				color = rl.White
			case Obstacle:
				color = rl.Black
			case Start:
				color = rl.Green
			case End:
				color = rl.Red
			case Visited:
				color = rl.DarkGray
			case Frontier:
				color = rl.Blue
			case Path:
				color = rl.Purple
			}

			var space int32 = (screenWidth - 182) / int32(size)

			i32 := int32(i)
			j32 := int32(j)

			x := i32*space + 50 + i32*(int32((100-size)/10))
			y := j32*space + 50 + j32*(int32((100-size)/10))

			if space+x > screenWidth-50 || space+y > int32(sliderRect.Y)-10 {
				g.Cells[i][j].onScreen = false
				continue
			}

			g.Cells[i][j].onScreen = true

			rl.DrawRectangle(
				x,
				y,
				space,
				space,
				color,
			)
		}
	}
	if g.StartCell != nil {
		rl.DrawText("Start Cords: "+strconv.FormatInt(int64(g.StartCell.x), 10)+", "+strconv.FormatInt(int64(g.StartCell.y), 10), 100, 10, 20, rl.White)
	}
	if g.EndCell != nil {
		rl.DrawText("End Cords: "+strconv.FormatInt(int64(g.EndCell.x), 10)+", "+strconv.FormatInt(int64(g.EndCell.y), 10), 300, 10, 20, rl.White)
	}
}

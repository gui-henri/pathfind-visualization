package main

import (
	"math"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Status uint8
type Mode uint8
type AvaliationMethod uint8

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

const (
	BFS AvaliationMethod = iota
	DFS
	AStar
)

var bestScore float64 = 9999999

type Cell struct {
	status         Status
	score          float64
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
				score:  9999999,
			}
		}
	}

	return &grid
}

func getEuclidianDistance(s, d *Cell) float64 {
	return math.Sqrt(math.Pow(float64(s.x-d.x), 2) + math.Pow(float64(s.y-d.y), 2))
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
			g.Cells[i][j].score = 9999999
			g.Cells[i][j].heuristicScore = 0
			g.Cells[i][j].parent = nil
		}
	}
}

func (g *Grid) SoftReset() {
	g.FrontierCells = nil
	g.GridMode = PaintMode
	g.Ticks = 0
	for i := range g.Cells {
		for j := range g.Cells[i] {
			if g.Cells[i][j].status != Obstacle && g.Cells[i][j].status != Start && g.Cells[i][j].status != End {
				g.Cells[i][j].status = Unselected
			}
			g.Cells[i][j].score = 9999999
			g.Cells[i][j].heuristicScore = 0
			g.Cells[i][j].parent = nil
		}
	}
}

func buildPath(cell *Cell) {
	if cell == nil {
		return
	}

	if cell.status != Start && cell.status != End {
		cell.status = Path
	}

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

func (g *Grid) calculateScore(c *Cell, m AvaliationMethod) float64 {
	switch m {
	case DFS:
		return getEuclidianDistance(c, g.EndCell)
	case AStar:
		return getEuclidianDistance(c, g.StartCell) + getEuclidianDistance(c, g.EndCell)
	default:
		return c.score + 1
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

func decideVolume(newScore float64, m AvaliationMethod, fx rl.Sound) {
	if newScore < bestScore {
		switch m {
		case BFS:
			rl.SetAudioStreamPitch(fx.Stream, float32(0.3+(newScore/10)))
		case DFS:
			rl.SetAudioStreamPitch(fx.Stream, 5-float32(0.3+(newScore/10)))
		case AStar:
			rl.SetAudioStreamPitch(fx.Stream, 5-float32(0.3+(newScore/10)))
		}
	}
}

func (g *Grid) playMode(size, speed int32, m AvaliationMethod, fx rl.Sound) bool {
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
		target := neighbor
		if BFS == m {
			target = currentCell
		}
		if neighbor.status == End {
			newScore := g.calculateScore(target, avaliationMethod)
			if newScore < neighbor.score {
				neighbor.score = newScore
				neighbor.parent = currentCell
				decideVolume(newScore, m, fx)
			}
			g.GridMode = PaintMode
			buildPath(neighbor)
			return true
		}
		if neighbor.status == Unselected {
			newScore := g.calculateScore(target, avaliationMethod)
			if newScore < neighbor.score {
				neighbor.score = newScore
				neighbor.parent = currentCell
				decideVolume(newScore, m, fx)
			}
			neighbor.status = Frontier
			neighbor.parent = currentCell
			g.FrontierCells = append(g.FrontierCells, neighbor)
		}
		if neighbor.status == Visited {
			newScore := g.calculateScore(target, avaliationMethod)
			if newScore < neighbor.score {
				neighbor.score = newScore
				neighbor.parent = currentCell
				decideVolume(newScore, m, fx)
			}
		}
	}

	if currentCell == g.EndCell {
		g.GridMode = PaintMode
		buildPath(currentCell)
		return true
	} else {
		if currentCell.status != Start {
			currentCell.status = Visited
		}
	}
	rl.PlayAudioStream(fx.Stream)
	return false
}

func (g *Grid) UpdateSubset(size, speed int32, play bool, m AvaliationMethod, fx rl.Sound) bool {
	if play && g.GridMode != PlayMode {
		g.StartCell.score = 0
		g.StartCell.heuristicScore = 0
		g.FrontierCells = append(g.FrontierCells, g.StartCell)
		g.GridMode = PlayMode
	} else if !play && g.GridMode == PlayMode {
		g.GridMode = PaintMode
	}

	if !play && g.GridMode == PaintMode {
		g.GridMode = PaintMode
	}

	switch g.GridMode {
	case PaintMode:
		g.paintMode(size)
	case PlayMode:
		return g.playMode(size, speed, avaliationMethod, fx)
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

			rl.DrawText(strconv.FormatInt(int64(g.Cells[i][j].score), 10), x+5, y+5, 10, rl.Black)
			if g.Cells[i][j].parent != nil {
				parent := g.Cells[i][j].parent
				rl.DrawText(strconv.FormatInt(int64(parent.x), 10)+", "+strconv.FormatInt(int64(parent.y), 10), x+15, y+15, 10, rl.Black)
			}
		}
	}
	if g.StartCell != nil {
		rl.DrawText("Start Cords: "+strconv.FormatInt(int64(g.StartCell.x), 10)+", "+strconv.FormatInt(int64(g.StartCell.y), 10), 100, 10, 20, rl.White)
	}
	if g.EndCell != nil {
		rl.DrawText("End Cords: "+strconv.FormatInt(int64(g.EndCell.x), 10)+", "+strconv.FormatInt(int64(g.EndCell.y), 10), 300, 10, 20, rl.White)
	}
}

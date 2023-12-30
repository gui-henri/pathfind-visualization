package main

import (
	"strconv"

	rg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth      int32   = 1280
	screenHeight     int32   = 720
	playButtonWidth  float32 = 40
	playButtonHeight float32 = 40
	sliderWidth      float32 = 720
	sliderHeight     float32 = 40
	padding          float32 = 50
	SOFT_RED                 = 0xfc5f8bff
	RED                      = 0xff0048ff
)

var (
	gridSubsetSize float32 = 0
	play           bool    = false
	sliderRect             = rl.NewRectangle(
		(float32(screenWidth)/2)-(sliderWidth/2),
		(float32(screenHeight)/8)*7,
		sliderWidth,
		sliderHeight,
	)
)

func main() {

	// INITIALIZATION

	grid := NewGrid(100)

	rl.InitWindow(screenWidth, screenHeight, "Test Window")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	// GUI STYLING

	rg.SetStyle(rg.DEFAULT, rg.TEXT_SIZE, 20)
	rg.SetStyle(rg.DEFAULT, rg.BORDER_WIDTH, 4)
	rg.SetStyle(rg.DEFAULT, rg.TEXT_COLOR_FOCUSED, SOFT_RED)
	rg.SetStyle(rg.DEFAULT, rg.TEXT_COLOR_PRESSED, RED)
	rg.SetStyle(rg.DEFAULT, rg.BORDER_COLOR_FOCUSED, SOFT_RED)
	rg.SetStyle(rg.DEFAULT, rg.BORDER_COLOR_PRESSED, RED)

	// MAIN LOOP

	for !rl.WindowShouldClose() {

		// USER INPUT

		if rl.IsKeyPressed(rl.KeyR) {
			grid.Reset()
		}

		gridSubsetSize = rg.Slider(
			sliderRect,
			"2",
			"100",
			gridSubsetSize,
			2,
			100,
		)

		var icon string

		if play {
			icon = "#133#"
		} else {
			icon = "#131#"
		}

		buttonClick := rg.Button(
			rl.NewRectangle(
				(float32(screenWidth))-(playButtonWidth*5),
				(float32(screenHeight)/8)*7,
				playButtonWidth,
				playButtonHeight,
			),
			icon,
		)

		if buttonClick {
			play = !play
		}

		// UPDATE

		finded := grid.UpdateSubset(int32(gridSubsetSize), 60, play)

		if finded {
			play = false
		}

		// DRAW
		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.Black)

			grid.DrawSubset(int(gridSubsetSize))

			rl.DrawText(strconv.FormatInt(int64(gridSubsetSize), 10), (screenWidth / 2), (screenHeight/9)*8, 20, rl.DarkGray)
			rl.DrawFPS(10, 10)
		}

		rl.EndDrawing()
	}
}

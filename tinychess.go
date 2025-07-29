package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

func main() {
	tinyChess := app.New()
	window := tinyChess.NewWindow("TinyChess")

	var light_square_color = color.RGBA{150, 74, 34, 255}
	var dark_square_color = color.RGBA{238, 220, 151, 255}

	var squares []fyne.CanvasObject
	for i := range 64 {
		var square_color color.Color
		if i%2 == (i/8)%2 {
			square_color = light_square_color
		} else {
			square_color = dark_square_color
		}
		squares = append(squares, canvas.NewRectangle(square_color))
	}

	grid := container.NewGridWithRows(8, squares...)

	window.SetContent(grid)
	window.Resize(fyne.NewSize(600, 600))
	window.ShowAndRun()
}

package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

func main() {
	tinyChess := app.New()
	window := tinyChess.NewWindow("TinyChess")

	image := canvas.NewImageFromFile("images/chessboard.svg")
	image.FillMode = canvas.ImageFillOriginal

	var squares []fyne.CanvasObject
	for range 64 {
		squares = append(squares, canvas.NewImageFromFile("images/pawn_black.svg"))
	}

	grid := container.NewGridWithRows(8, squares...)

	window.SetContent(container.New(layout.NewStackLayout(), image, grid))
	window.Resize(fyne.NewSize(600, 600))
	window.ShowAndRun()
}

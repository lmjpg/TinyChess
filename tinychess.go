package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

const (
	Pawn   = iota
	Knight = iota
	Bishop = iota
	Rook   = iota
	Queen  = iota
	King   = iota
)

const (
	Black = iota
	White = iota
)

type Piece struct {
	Type   int
	Colour int
	X, Y   int
}

func getInitialBoard() []Piece {
	var pieces []Piece
	for c := range 2 {
		for i := range 8 {
			pieces = append(pieces, Piece{Pawn, c, i, 1 + (c*7 - c*2)})
		}
		for i, v := range []int{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook} {
			pieces = append(pieces, Piece{v, c, i, 0 + c*7})
		}
	}
	return pieces
}

func getEmptySquares() []fyne.CanvasObject {
	var squares []fyne.CanvasObject
	var empty *canvas.Image = canvas.NewImageFromFile("images/empty.svg")
	for range 64 {
		squares = append(squares, empty)
	}
	return squares
}

func getPieceImage(piece Piece, images []string) *canvas.Image {
	var colour string
	if piece.Colour == Black {
		colour = "black"
	} else {
		colour = "white"
	}
	return canvas.NewImageFromFile("images/" + images[piece.Type] + "_" + colour + ".svg")
}

func updateSquaresFromBoard(squares []fyne.CanvasObject, pieces []Piece, images []string) []fyne.CanvasObject {
	for _, piece := range pieces {
		squares[piece.X+piece.Y*8] = getPieceImage(piece, images)
	}
	return squares
}

func main() {
	tinyChess := app.New()
	window := tinyChess.NewWindow("TinyChess")

	images := []string{"pawn", "knight", "bishop", "rook", "queen", "king"}

	image := canvas.NewImageFromFile("images/chessboard.svg")
	image.FillMode = canvas.ImageFillOriginal

	squares := getEmptySquares()
	board := getInitialBoard()
	squares = updateSquaresFromBoard(squares, board, images)

	grid := container.NewGridWithColumns(8, squares...)

	window.SetContent(container.New(layout.NewStackLayout(), image, grid))
	window.Resize(fyne.NewSize(600, 600))
	window.ShowAndRun()
}

package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
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
	Type, Colour, X, Y int
	Selected           bool
}

type pieceWidget struct {
	widget.Icon
	Board     []Piece
	X, Y      int
	Resources []fyne.Resource
	Window    fyne.Window
}

func newPieceWidget(res fyne.Resource, board []Piece, x int, y int, resources []fyne.Resource, window fyne.Window) *pieceWidget {
	widget := &pieceWidget{Board: board, X: x, Y: y, Resources: resources, Window: window}
	widget.ExtendBaseWidget(widget)
	widget.SetResource(res)

	return widget
}

func (t *pieceWidget) Tapped(_ *fyne.PointEvent) {
	var prevSelectedN int = -1
	var foundPiece = false

	for n, piece := range t.Board {
		if piece.X == t.X && piece.Y == t.Y {
			piece.Selected = true
			t.Board[n] = piece
			foundPiece = true
			fmt.Printf("Clicked %v at %v %v\n", piece.Type, piece.X, piece.Y)
		} else if piece.Selected {
			piece.Selected = false
			t.Board[n] = piece
			prevSelectedN = n
		}
	}

	if !foundPiece && prevSelectedN != -1 {
		prevX, prevY := t.Board[prevSelectedN].X, t.Board[prevSelectedN].Y
		fmt.Printf("Moved from %v %v to %v %v\n", prevX, prevY, t.X, t.Y)
		piece := t.Board[prevSelectedN]
		piece.X, piece.Y = t.X, t.Y
		t.Board[prevSelectedN] = piece

		updateWindowFromBoard(t.Board, t.Resources, t.Window)
	}
}

func getInitialBoard() []Piece {
	var board []Piece = make([]Piece, 32)
	var n = 0
	for c := range 2 {
		for i := range 8 {
			board[n] = Piece{Pawn, c, i, 1 + (c * 5), false}
			n++
		}
		for i, v := range []int{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook} {
			board[n] = Piece{v, c, i, c * 7, false}
			n++
		}
	}
	return board
}

func getPieceImage(piece Piece, board []Piece, resources []fyne.Resource, window fyne.Window) *pieceWidget {
	res := resources[piece.Type+piece.Colour*6]
	return newPieceWidget(res, board, piece.X, piece.Y, resources, window)
}

func updateWindowFromBoard(board []Piece, resources []fyne.Resource, window fyne.Window) { //[]fyne.CanvasObject {
	var squares []fyne.CanvasObject
	for i := range 64 {
		squares = append(squares, newPieceWidget(resources[len(resources)-1], board, i%8, i/8, resources, window))
	}

	for _, piece := range board {
		squares[piece.X+piece.Y*8] = getPieceImage(piece, board, resources, window)
	}

	grid := container.NewGridWithColumns(8, squares...)
	chessboard := canvas.NewImageFromFile("images/chessboard.svg")
	chessboard.FillMode = canvas.ImageFillOriginal

	window.SetContent(container.New(layout.NewStackLayout(), chessboard, grid))
	// return squares
}

func main() {
	tinyChess := app.New()
	window := tinyChess.NewWindow("TinyChess")

	var resources []fyne.Resource
	for _, colour := range []string{"black", "white"} {
		for _, filename := range []string{"pawn", "knight", "bishop", "rook", "queen", "king"} {
			path := "images/" + filename + "_" + colour + ".svg"
			res, err := fyne.LoadResourceFromPath("images/" + filename + "_" + colour + ".svg")
			if err != nil {
				log.Fatal(path + " couldn't be loaded")
			}
			resources = append(resources, res)
		}
	}

	empty_res, err := fyne.LoadResourceFromPath("images/empty.svg")
	if err != nil {
		log.Fatal("images/empty.svg couldn't be loaded")
	}
	resources = append(resources, empty_res)

	board := getInitialBoard()
	updateWindowFromBoard(board, resources, window)

	window.Resize(fyne.NewSize(600, 600))
	window.ShowAndRun()
}

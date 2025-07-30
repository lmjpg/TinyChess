package main

import (
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
	Squares   []*pieceWidget
	X, Y      int
	Resources []fyne.Resource
}

func newPieceWidget(res fyne.Resource, board []Piece, squares []*pieceWidget, x int, y int, resources []fyne.Resource) *pieceWidget {
	widget := &pieceWidget{Board: board, Squares: squares, X: x, Y: y, Resources: resources}
	widget.ExtendBaseWidget(widget)
	widget.SetResource(res)

	return widget
}

func (t *pieceWidget) Tapped(_ *fyne.PointEvent) {
	var clickedN = -1
	var prevN = -1

	for n, piece := range t.Board {
		if t.X == piece.X && t.Y == piece.Y {
			clickedN = n
		}

		if piece.Selected {
			prevN = n
		}
	}

	if clickedN == prevN && clickedN != -1 { // clicked already clicked, unselect
		piece := t.Board[clickedN]
		piece.Selected = false
		t.Board[clickedN] = piece

	} else if clickedN != -1 && prevN == -1 { // clicked with none selected, just select
		piece := t.Board[clickedN]
		piece.Selected = true
		t.Board[clickedN] = piece

	} else if clickedN == -1 && prevN != -1 { // clicked empty square with piece selected, move
		piece := t.Board[prevN]
		piece.Selected = false
		t.Squares[piece.X+piece.Y*8].SetResource(t.Resources[len(t.Resources)-2])
		piece.X = t.X
		piece.Y = t.Y
		t.Squares[piece.X+piece.Y*8].SetResource(getPieceResource(piece, t.Resources))
		t.Board[prevN] = piece

	} else if clickedN != -1 && prevN != -1 && clickedN != prevN { // taking piece
		clickedPiece := t.Board[clickedN]
		prevPiece := t.Board[prevN]
		clickedPiece.Selected = false
		prevPiece.Selected = false
		clickedPiece.X = -1
		clickedPiece.Y = -1

		t.Squares[prevPiece.X+prevPiece.Y*8].SetResource(t.Resources[len(t.Resources)-2])
		prevPiece.X = t.X
		prevPiece.Y = t.Y
		t.Squares[prevPiece.X+prevPiece.Y*8].SetResource(getPieceResource(prevPiece, t.Resources))

		t.Board[clickedN] = clickedPiece
		t.Board[prevN] = prevPiece
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

func getPieceResource(piece Piece, resources []fyne.Resource) fyne.Resource {
	return resources[piece.Type+piece.Colour*6]
}

func updateWindowFromBoard(board []Piece, resources []fyne.Resource, window fyne.Window) { //[]fyne.CanvasObject {
	var squares = make([]*pieceWidget, 64)

	for i := range 64 {
		squares[i] = newPieceWidget(resources[len(resources)-2], board, squares, i%8, i/8, resources)
	}

	for _, piece := range board {
		squares[piece.X+piece.Y*8] = newPieceWidget(getPieceResource(piece, resources), board, squares, piece.X, piece.Y, resources)
	}

	// Why can't this be type cast instead?
	var squares2 = make([]fyne.CanvasObject, len(squares))
	for i, v := range squares {
		squares2[i] = v
	}

	grid := container.NewGridWithColumns(8, squares2...)
	chessboard := canvas.NewImageFromResource(resources[len(resources)-1])
	chessboard.FillMode = canvas.ImageFillOriginal

	window.SetContent(container.New(layout.NewStackLayout(), chessboard, grid))
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

	for _, filename := range []string{"empty", "chessboard"} {
		new_res, err := fyne.LoadResourceFromPath("images/" + filename + ".svg")
		if err != nil {
			log.Fatal("images/" + filename + ".svg couldn't be loaded")
		}
		resources = append(resources, new_res)
	}

	board := getInitialBoard()
	updateWindowFromBoard(board, resources, window)

	window.Resize(fyne.NewSize(600, 600))
	window.ShowAndRun()
}

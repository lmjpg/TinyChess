package main

import (
	"log"
	"math"

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
	Type, Colour, X, Y            int
	Selected, HasMoved, LastMoved bool
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
	} else if prevN != -1 {
		movePiece(t.X, t.Y, t.Board, prevN, clickedN, t.Squares, t.Resources)
	}
}

func getSquareIndex(x int, y int) int {
	return x + y*8
}

func movePiece(x int, y int, board []Piece, pieceN int, takenPieceN int, squares []*pieceWidget, resources []fyne.Resource) {
	piece := board[pieceN]
	piece.Selected = false

	if !isValidMove(board, piece, x, y, takenPieceN != -1) {
		board[pieceN] = piece
		return
	}

	if takenPieceN != -1 {
		removePiece(board, takenPieceN, squares)
	}

	squares[getSquareIndex(piece.X, piece.Y)].SetResource(resources[len(resources)-2])
	piece.X = x
	piece.Y = y
	piece.HasMoved = true
	piece.LastMoved = true
	squares[getSquareIndex(piece.X, piece.Y)].SetResource(getPieceResource(piece, resources))

	board[pieceN] = piece
}

func removePiece(board []Piece, pieceN int, squares []*pieceWidget) {
	piece := board[pieceN]
	i := getSquareIndex(piece.X, piece.Y)
	resources := squares[i].Resources
	squares[i].SetResource(resources[len(resources)-2])
	piece.X = -1
	piece.Y = -1
	piece.Selected = false
	board[pieceN] = piece
}

func isValidMove(board []Piece, movingPiece Piece, x int, y int, isTaking bool) bool {
	firstTurn := true
	var lastMovedN = -1

	ax, ay, bx, by := movingPiece.X, movingPiece.Y, x, y
	slopeAB := float64(by-ay) / float64(bx-ax)
	distanceAB := math.Sqrt(math.Pow(float64(bx-ax), 2) + math.Pow(float64(by-ay), 2))

	for n, piece := range board {
		if piece.X == x && piece.Y == y && piece.Colour == movingPiece.Colour {
			return false // can't take your own pieces
		}

		cx, cy := piece.X, piece.Y
		if ax != bx && ax != cx {
			slopeAC := float64(cy-ay) / float64(cx-ax)
			distanceAC := math.Sqrt(math.Pow(float64(cx-ax), 2) + math.Pow(float64(cy-ay), 2))
			if slopeAB == slopeAC && distanceAB > distanceAC {
				return false // piece in the way
			}
		} else if movingPiece.X == x && movingPiece.X == piece.X && ((movingPiece.Y < piece.Y && piece.Y < y) || (y < piece.Y && piece.Y < movingPiece.Y)) {
			return false // piece in the way (vertical)
		}

		if piece.LastMoved {
			firstTurn = false
			if movingPiece.Colour == piece.Colour {
				return false
			}
			lastMovedN = n
		}
	}

	if firstTurn && movingPiece.Colour != White {
		return false
	}

	dx, dy := movingPiece.X-x, movingPiece.Y-y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}

	// rules missing: castling, en passant, check, checkmate, stalement, three move repetition, fifty moves with no pawn moves
	pawn_valid_not_taking := !isTaking && dx == 0 && ((dy == 1) || (dy == 2 && !movingPiece.HasMoved))
	pawn_valid_taking := isTaking && dx == 1 && dy == 1 // en passant?
	knight_valid := ((dx == 1 && dy == 2) || (dx == 2 && dy == 1))
	bishop_valid := math.Abs(float64(movingPiece.Y-y)) == math.Abs(float64(movingPiece.X-x))
	rook_valid := movingPiece.X == x || movingPiece.Y == y
	king_valid := dx <= 1 && dy <= 1 // still needs to handle castling

	switch movingPiece.Type {
	case Pawn:
		if !(pawn_valid_not_taking || pawn_valid_taking) {
			return false
		}

	case Knight:
		if !knight_valid {
			return false
		}

	case Bishop:
		if !bishop_valid {
			return false
		}

	case Rook:
		if !rook_valid {
			return false
		}

	case Queen:
		if !bishop_valid && !rook_valid {
			return false
		}

	case King:
		if !king_valid {
			return false
		}

	default:
		log.Fatal("Invalid piece")
	}

	if lastMovedN != -1 {
		piece := board[lastMovedN]
		piece.LastMoved = false
		board[lastMovedN] = piece
	}

	return true
}

func getInitialBoard() []Piece {
	var board []Piece = make([]Piece, 32)
	var n = 0
	for c := range 2 {
		for i := range 8 {
			board[n] = Piece{Pawn, c, i, 1 + (c * 5), false, false, false}
			n++
		}
		for i, v := range []int{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook} {
			board[n] = Piece{v, c, i, c * 7, false, false, false}
			n++
		}
	}
	return board
}

func getPieceResource(piece Piece, resources []fyne.Resource) fyne.Resource {
	return resources[piece.Type+piece.Colour*6]
}

func updateWindowFromBoard(board []Piece, resources []fyne.Resource, window fyne.Window) {
	var squares = make([]*pieceWidget, 64)

	for i := range 64 {
		squares[i] = newPieceWidget(resources[len(resources)-2], board, squares, i%8, i/8, resources)
	}

	for _, piece := range board {
		squares[getSquareIndex(piece.X, piece.Y)] = newPieceWidget(getPieceResource(piece, resources), board, squares, piece.X, piece.Y, resources)
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

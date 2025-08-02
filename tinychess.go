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

type TinyChess struct {
	Game        *Game
	Squares     []*PieceWidget
	Overlay     []*widget.Icon
	Resources   Resources
	SelectedPos Position
}

type Resources struct {
	Pieces                                []fyne.Resource
	Empty, ChessBoard, Circle, CircleHole fyne.Resource
}

type PieceWidget struct {
	widget.Icon
	Session *TinyChess
	Pos     Position
}

func newPieceWidget(session *TinyChess, x int, y int, res fyne.Resource) *PieceWidget {
	widget := &PieceWidget{Session: session, Pos: Position{x, y}}
	widget.ExtendBaseWidget(widget)
	widget.SetResource(res)

	return widget
}

func (t *PieceWidget) Tapped(_ *fyne.PointEvent) {
	if t.Pos == t.Session.SelectedPos { // clicked already clicked, unselect
		t.Session.SelectedPos = invalidPosition()

	} else if t.Session.SelectedPos == invalidPosition() { // clicked with none selected, just select
		_, ok := t.Session.Game.Board[t.Pos]
		if ok && len(getLegalMoves(t.Session.Game, t.Pos)) > 0 {
			t.Session.SelectedPos = t.Pos
		}

	} else if t.Session.SelectedPos != invalidPosition() {
		moveSuccessful := movePiece(t.Session.Game, t.Session.SelectedPos, t.Pos, invalidPosition(), true)
		t.Session.SelectedPos = invalidPosition()
		_, ok := t.Session.Game.Board[t.Pos]
		if !moveSuccessful && ok {
			t.Session.SelectedPos = t.Pos
		}
	}

	updateSquares(t.Session)
}

func getPieceResource(piece Piece, resources Resources) fyne.Resource {
	return resources.Pieces[piece.Type+piece.Colour*6]
}

func getSquareIndexFromPosition(pos Position) int {
	return pos.X + pos.Y*8
}

func createWindowFromBoard(tinychess *TinyChess, w fyne.Window) ([]*PieceWidget, []*widget.Icon) {
	var squares = make([]*PieceWidget, 64)
	var overlay = make([]*widget.Icon, 64)

	for i := range 64 {
		squares[i] = newPieceWidget(tinychess, i%8, i/8, tinychess.Resources.Empty)
		overlay[i] = widget.NewIcon(tinychess.Resources.Empty)
	}

	for pos, piece := range tinychess.Game.Board {
		squares[getSquareIndexFromPosition(pos)].SetResource(getPieceResource(piece, tinychess.Resources))
	}

	// Why can't this be type cast instead?
	var squares2 = make([]fyne.CanvasObject, len(squares))
	var overlay2 = make([]fyne.CanvasObject, len(overlay))
	for i := range len(squares) {
		squares2[i] = squares[i]
		overlay2[i] = overlay[i]
	}

	grid := container.NewGridWithColumns(8, squares2...)
	gridOverlay := container.NewGridWithColumns(8, overlay2...)
	chessboard := canvas.NewImageFromResource(tinychess.Resources.ChessBoard)
	chessboard.FillMode = canvas.ImageFillOriginal

	w.SetContent(container.New(layout.NewStackLayout(), chessboard, grid, gridOverlay))

	return squares, overlay
}

func updateSquares(tinychess *TinyChess) {
	for _, square := range tinychess.Squares {
		piece, ok := tinychess.Game.Board[square.Pos]
		if ok {
			square.SetResource(getPieceResource(piece, tinychess.Resources))
		} else {
			square.SetResource(tinychess.Resources.Empty)
		}
	}

	for _, overlay := range tinychess.Overlay {
		overlay.SetResource(tinychess.Resources.Empty)
	}

	_, ok := tinychess.Game.Board[tinychess.SelectedPos]
	if ok {
		for _, move := range getLegalMoves(tinychess.Game, tinychess.SelectedPos) {
			res := tinychess.Resources.Circle
			_, isTakingPiece := tinychess.Game.Board[move.Pos]
			if isTakingPiece {
				res = tinychess.Resources.CircleHole
			}
			tinychess.Overlay[getSquareIndexFromPosition(move.Pos)].SetResource(res)
		}
	}
}

func main() {
	a := app.New()
	w := a.NewWindow("TinyChess")

	var pieceResources []fyne.Resource
	for _, colour := range []string{"black", "white"} {
		for _, filename := range []string{"pawn", "knight", "bishop", "rook", "queen", "king"} {
			path := "images/" + filename + "_" + colour + ".svg"
			res, err := fyne.LoadResourceFromPath("images/" + filename + "_" + colour + ".svg")
			if err != nil {
				log.Fatal(path + " couldn't be loaded")
			}
			pieceResources = append(pieceResources, res)
		}
	}

	var resources Resources
	resources.Pieces = pieceResources

	for _, filename := range []string{"empty", "chessboard", "circle", "circle_hole"} {
		new_res, err := fyne.LoadResourceFromPath("images/" + filename + ".svg")
		if err != nil {
			log.Fatal("images/" + filename + ".svg couldn't be loaded")
		}
		switch filename {
		case "empty":
			resources.Empty = new_res
		case "chessboard":
			resources.ChessBoard = new_res
		case "circle":
			resources.Circle = new_res
		case "circle_hole":
			resources.CircleHole = new_res
		}
	}

	tinychess := TinyChess{Game: getInitialGame(), Squares: nil, Overlay: nil, Resources: resources, SelectedPos: invalidPosition()}

	tinychess.Squares, tinychess.Overlay = createWindowFromBoard(&tinychess, w)

	w.Resize(fyne.NewSize(600, 600))
	w.ShowAndRun()
}

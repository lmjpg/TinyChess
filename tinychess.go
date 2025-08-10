package main

import (
	"embed"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

//go:embed images/*
var content embed.FS

type TinyChess struct {
	Game             *Game
	Window           *fyne.Window
	Squares          []*PieceWidget
	Overlay          []*widget.Icon
	PromotionOverlay *fyne.Container
	Resources        Resources
	SelectedPos      *Position
	PromotionMove    *Move
}

type Resources struct {
	Pieces                                                       []fyne.Resource
	Empty, ChessBoard, Circle, CircleHole, CircleRed, CircleGrey fyne.Resource
}

type PieceWidget struct {
	widget.Icon
	Session *TinyChess
	Pos     Position
}

type PromotionWidget struct {
	widget.Icon
	Session *TinyChess
	Piece   int
}

func newPieceWidget(session *TinyChess, x int, y int, res fyne.Resource) *PieceWidget {
	widget := &PieceWidget{Session: session, Pos: Position{x, y}}
	widget.ExtendBaseWidget(widget)
	widget.SetResource(res)

	return widget
}

func (t *PieceWidget) Tapped(_ *fyne.PointEvent) {
	if t.Session.PromotionMove != nil {
		return
	}

	clickedPos := t.Pos
	if t.Session.Game.Turn == Black {
		clickedPos.Y = 7 - clickedPos.Y
	}

	if t.Session.SelectedPos != nil && clickedPos == *t.Session.SelectedPos { // clicked already clicked, unselect
		t.Session.SelectedPos = nil

	} else if t.Session.SelectedPos == nil { // clicked with none selected, just select
		_, ok := t.Session.Game.Board[clickedPos]
		if ok && len(getLegalMoves(t.Session.Game, clickedPos)) > 0 {
			t.Session.SelectedPos = &clickedPos
		}

	} else if t.Session.SelectedPos != nil {
		piece, ok := t.Session.Game.Board[*t.Session.SelectedPos]
		isPromoting := false
		if ok && piece.Type == Pawn && (clickedPos.Y == 0 || clickedPos.Y == 7) {
			var promotionMove *Move
			isPromoting, promotionMove = isValidMove(t.Session.Game, *t.Session.SelectedPos, clickedPos)
			if isPromoting {
				t.Session.PromotionMove = promotionMove
				t.Session.PromotionOverlay.Show()
			}
		}

		if !isPromoting {
			moveSuccessful := movePiece(t.Session.Game, *t.Session.SelectedPos, clickedPos, nil, true)
			t.Session.SelectedPos = nil
			_, ok = t.Session.Game.Board[clickedPos]
			if !moveSuccessful && ok {
				t.Session.SelectedPos = &clickedPos
			}
		}
	}

	updateSquares(t.Session)
}

func newPromotionWidget(session *TinyChess, piece int) *PromotionWidget {
	widget := &PromotionWidget{Session: session, Piece: piece}
	widget.ExtendBaseWidget(widget)
	widget.SetResource(getPieceResource(Piece{Type: piece, Colour: session.Game.Turn}, session.Resources))
	return widget
}

func (p *PromotionWidget) Tapped(e *fyne.PointEvent) {
	ok := movePiece(p.Session.Game, *p.Session.SelectedPos, p.Session.PromotionMove.Pos, p.Session.PromotionMove, false)
	if !ok {
		log.Fatalf("Move from %v to %v was invalid despite being valid when checked earlier (PromotionWidget.Tapped)\n", p.Session.SelectedPos, p.Session.PromotionMove.Pos)
	}
	p.Session.PromotionOverlay.Hide()

	piece := p.Session.Game.Board[p.Session.PromotionMove.Pos]
	piece.Type = p.Piece
	p.Session.Game.Board[p.Session.PromotionMove.Pos] = piece

	p.Session.SelectedPos = nil
	p.Session.PromotionMove = nil

	updateSquares(p.Session)
}

func getPieceResource(piece Piece, resources Resources) fyne.Resource {
	return resources.Pieces[piece.Type+piece.Colour*6]
}

func getSquareIndexFromPosition(pos Position, turn int) int {
	if turn == Black {
		pos.Y = 7 - pos.Y
	}
	return pos.X + pos.Y*8
}

func createWindowFromBoard(tinychess *TinyChess, w fyne.Window) ([]*PieceWidget, []*widget.Icon, *fyne.Container) {
	var squares = make([]*PieceWidget, 64)
	var overlay = make([]*widget.Icon, 64)

	for i := range 64 {
		squares[i] = newPieceWidget(tinychess, i%8, i/8, tinychess.Resources.Empty)
		overlay[i] = widget.NewIcon(tinychess.Resources.Empty)
	}

	for pos, piece := range tinychess.Game.Board {
		squares[getSquareIndexFromPosition(pos, tinychess.Game.Turn)].SetResource(getPieceResource(piece, tinychess.Resources))
	}

	var promotionOptions []fyne.CanvasObject
	for _, piece := range []int{Queen, Rook, Bishop, Knight} {
		promotionOptions = append(promotionOptions, newPromotionWidget(tinychess, piece))
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
	promotionOverlay := container.NewGridWithColumns(4, promotionOptions...)
	chessboard := canvas.NewImageFromResource(tinychess.Resources.ChessBoard)
	chessboard.FillMode = canvas.ImageFillOriginal

	promotionOverlay.Hide()

	w.SetContent(container.New(layout.NewStackLayout(), chessboard, grid, gridOverlay, promotionOverlay))

	return squares, overlay, promotionOverlay
}

func updateSquares(tinychess *TinyChess) {
	for _, overlay := range tinychess.Overlay {
		overlay.SetResource(tinychess.Resources.Empty)
	}

	for _, square := range tinychess.Squares {
		piecePos := square.Pos
		if tinychess.Game.Turn == Black {
			piecePos.Y = 7 - piecePos.Y
		}
		piece, ok := tinychess.Game.Board[piecePos]
		if ok {
			square.SetResource(getPieceResource(piece, tinychess.Resources))
		} else {
			square.SetResource(tinychess.Resources.Empty)
		}

		if ok && tinychess.Game.Checkmate && piece.Type == King && piece.Colour != tinychess.Game.Turn {
			tinychess.Overlay[getSquareIndexFromPosition(square.Pos, tinychess.Game.Turn)].SetResource(tinychess.Resources.CircleRed)
		} else if ok && tinychess.Game.Draw && piece.Type == King {
			tinychess.Overlay[getSquareIndexFromPosition(square.Pos, tinychess.Game.Turn)].SetResource(tinychess.Resources.CircleGrey)
		}
	}

	if tinychess.SelectedPos != nil {
		_, ok := tinychess.Game.Board[*tinychess.SelectedPos]
		if ok {
			for _, move := range getLegalMoves(tinychess.Game, *tinychess.SelectedPos) {
				res := tinychess.Resources.Circle
				_, isTakingPiece := tinychess.Game.Board[move.Pos]
				if isTakingPiece {
					res = tinychess.Resources.CircleHole
				}
				tinychess.Overlay[getSquareIndexFromPosition(move.Pos, tinychess.Game.Turn)].SetResource(res)
			}
		}
	}
}

func main() {
	a := app.New()
	w := a.NewWindow("TinyChess")

	var pieceResources []fyne.Resource
	for _, colour := range []string{"black", "white"} {
		for _, filename := range []string{"pawn", "knight", "bishop", "rook", "queen", "king"} {
			name := filename + "_" + colour
			path := "images/" + name + ".svg"
			fileContents, err := embed.FS.ReadFile(content, path)
			if err != nil {
				log.Fatal(path + " couldn't be loaded")
			}
			res := fyne.NewStaticResource(name, fileContents)
			pieceResources = append(pieceResources, res)
		}
	}

	var resources Resources
	resources.Pieces = pieceResources

	for _, filename := range []string{"empty", "chessboard", "circle", "circle_hole", "circle_red", "circle_grey"} {
		fileContents, err := embed.FS.ReadFile(content, "images/"+filename+".svg")
		if err != nil {
			log.Fatal("images/" + filename + ".svg couldn't be loaded")
		}
		new_res := fyne.NewStaticResource(filename, fileContents)
		switch filename {
		case "empty":
			resources.Empty = new_res
		case "chessboard":
			resources.ChessBoard = new_res
		case "circle":
			resources.Circle = new_res
		case "circle_hole":
			resources.CircleHole = new_res
		case "circle_red":
			resources.CircleRed = new_res
		case "circle_grey":
			resources.CircleGrey = new_res
		}
	}

	tinychess := TinyChess{Game: getInitialGame(), Window: &w, Squares: nil, Overlay: nil, PromotionOverlay: nil, Resources: resources, SelectedPos: nil, PromotionMove: nil}

	tinychess.Squares, tinychess.Overlay, tinychess.PromotionOverlay = createWindowFromBoard(&tinychess, w)

	w.Resize(fyne.NewSize(600, 600))
	w.ShowAndRun()
}

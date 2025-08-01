package main

import (
	"log"
	"math"
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

type Game struct {
	Board      []Piece
	Turn       int
	LastMovedN int
}

type Position struct {
	X, Y int
}

type Piece struct {
	Type, Colour int
	Pos          Position
	HasMoved     bool
}

func getInitialGame() *Game {
	var board []Piece = make([]Piece, 32)
	var n = 0
	for c := range 2 {
		for i := range 8 {
			board[n] = Piece{Pawn, c, Position{i, 1 + (c * 5)}, false}
			n++
		}
		for i, v := range []int{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook} {
			board[n] = Piece{v, c, Position{i, c * 7}, false}
			n++
		}
	}

	game := Game{Board: board, Turn: White, LastMovedN: -1}
	return &game
}

func movePiece(game *Game, pos Position, pieceN int, takenPieceN int) {
	piece := game.Board[pieceN]

	if !isValidMove(game, piece, pos, takenPieceN != -1) {
		return
	}

	if takenPieceN != -1 {
		removePiece(game, takenPieceN)
	}

	piece.Pos = pos
	piece.HasMoved = true

	if game.Turn == White {
		game.Turn = Black
	} else {
		game.Turn = White
	}

	game.Board[pieceN] = piece
}

func removePiece(game *Game, pieceN int) {
	piece := game.Board[pieceN]
	piece.Pos = Position{X: -1, Y: -1}
	game.Board[pieceN] = piece
}

func isValidMove(game *Game, movingPiece Piece, pos Position, isTaking bool) bool {
	ax, ay, bx, by := movingPiece.Pos.X, movingPiece.Pos.Y, pos.X, pos.Y
	slopeAB := float64(by-ay) / float64(bx-ax)
	distanceAB := math.Sqrt(math.Pow(float64(bx-ax), 2) + math.Pow(float64(by-ay), 2))

	for _, piece := range game.Board {
		if piece.Pos == pos && piece.Colour == movingPiece.Colour {
			return false // can't take your own pieces
		}

		cx, cy := piece.Pos.X, piece.Pos.Y
		if ax != bx && ax != cx {
			slopeAC := float64(cy-ay) / float64(cx-ax)
			distanceAC := math.Sqrt(math.Pow(float64(cx-ax), 2) + math.Pow(float64(cy-ay), 2))
			if slopeAB == slopeAC && distanceAB > distanceAC {
				return false // piece in the way
			}
		} else if movingPiece.Pos.X == pos.X && movingPiece.Pos.X == piece.Pos.X && ((movingPiece.Pos.Y < piece.Pos.Y && piece.Pos.Y < pos.Y) || (pos.Y < piece.Pos.Y && piece.Pos.Y < movingPiece.Pos.Y)) {
			return false // piece in the way (vertical)
		}
	}

	if game.Turn != movingPiece.Colour {
		return false // wrong colour moving
	}

	dx, dy := movingPiece.Pos.X-pos.X, movingPiece.Pos.Y-pos.Y
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
	bishop_valid := dx == dy
	rook_valid := movingPiece.Pos.X == pos.X || movingPiece.Pos.Y == pos.Y
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

	return true
}

package main

import (
	"log"
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
	Board     []Piece
	Turn      int
	LastMoved *Piece
}

type Position struct {
	X, Y int
}

type Move struct {
	Pos     Position
	TakingN int
}

type Piece struct {
	Type, Colour    int
	Pos             Position
	HasMoved        bool
	PawnDoubleMoved bool
}

func getInitialGame() *Game {
	var board []Piece = make([]Piece, 32)
	var n = 0
	for c := range 2 {
		for i := range 8 {
			board[n] = Piece{Pawn, c, Position{i, 1 + (c * 5)}, false, false}
			n++
		}
		for i, v := range []int{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook} {
			board[n] = Piece{v, c, Position{i, c * 7}, false, false}
			n++
		}
	}

	game := Game{Board: board, Turn: White, LastMoved: nil}
	return &game
}

func movePiece(game *Game, pos Position, pieceN int) {
	piece := game.Board[pieceN]

	isValid, takingN := isValidMove(game, piece, pos)
	if !isValid {
		return
	}

	if takingN != -1 {
		removePiece(game, takingN)
	}

	if piece.Type == Pawn && piece.Pos.Y+2 == pos.Y {
		piece.PawnDoubleMoved = true
	} else {
		piece.PawnDoubleMoved = false
	}
	piece.Pos = pos
	piece.HasMoved = true

	if game.Turn == White {
		game.Turn = Black
	} else {
		game.Turn = White
	}

	game.Board[pieceN] = piece
	game.LastMoved = &piece
}

func removePiece(game *Game, pieceN int) {
	piece := game.Board[pieceN]
	piece.Pos = Position{X: -1, Y: -1}
	game.Board[pieceN] = piece
}

func getForward(colour int, amount int) int {
	if colour == White {
		return -amount
	} else {
		return amount
	}
}

func abs(n int) int {
	if n >= 0 {
		return n
	} else {
		return -n
	}
}

func isValidMove(game *Game, movingPiece Piece, pos Position) (bool, int) {
	validMoves := getValidMoves(game, movingPiece)
	for _, validPos := range validMoves {
		if pos == validPos.Pos {
			return true, validPos.TakingN
		}
	}
	return false, -1
}

func getValidMoves(game *Game, movingPiece Piece) []Move {
	// prevent moves that put the current player into check:
	// AFTER all other stuff, for each move in validMoves create a clone of the current game with that move applied
	// then getValidMoves() for each cloned game (but with this check disabled to prevent infinite recursion)
	// check the TakingN on each of the cloned games Moves, if any of them match the current players king, the non-clone move that led to that position is illegal

	var validMoves []Move

	if game.Turn != movingPiece.Colour {
		return validMoves
	}

	switch movingPiece.Type {
	case Pawn:
		forward1Free := true
		forward2Free := true
		for n, targetPiece := range game.Board {
			if movingPiece.Pos.X == targetPiece.Pos.X {
				if movingPiece.Pos.Y+getForward(movingPiece.Colour, 1) == targetPiece.Pos.Y {
					forward1Free = false // space in front blocked
				} else if movingPiece.Pos.Y+getForward(movingPiece.Colour, 2) == targetPiece.Pos.Y {
					forward2Free = false // space 2 in front blocked
				}
			} else if movingPiece.Colour != targetPiece.Colour && movingPiece.Pos.X+1 == targetPiece.Pos.X || movingPiece.Pos.X-1 == targetPiece.Pos.X {
				if movingPiece.Pos.Y+getForward(movingPiece.Colour, 1) == targetPiece.Pos.Y {
					validMoves = appendIfInBounds(validMoves, Move{Pos: targetPiece.Pos, TakingN: n}) // can take targetPiece
				} else if movingPiece.Pos.Y == targetPiece.Pos.Y && game.LastMoved != nil && *game.LastMoved == targetPiece && targetPiece.PawnDoubleMoved {
					nx, ny := targetPiece.Pos.X, targetPiece.Pos.Y+getForward(movingPiece.Colour, 1)
					validMoves = appendIfInBounds(validMoves, Move{Pos: Position{X: nx, Y: ny}, TakingN: n}) // en passant take
				}
			}
		}

		if forward1Free {
			nx, ny := movingPiece.Pos.X, movingPiece.Pos.Y+getForward(movingPiece.Colour, 1)
			validMoves = appendIfInBounds(validMoves, Move{Pos: Position{X: nx, Y: ny}, TakingN: -1})
		}
		if forward1Free && forward2Free && !movingPiece.HasMoved {
			nx, ny := movingPiece.Pos.X, movingPiece.Pos.Y+getForward(movingPiece.Colour, 2)
			validMoves = appendIfInBounds(validMoves, Move{Pos: Position{X: nx, Y: ny}, TakingN: -1})
		}

	case Knight:
		directions := []int{1, 2, -1, -2}
		validMoves = appendMoves(validMoves, game, movingPiece, directions, true, false)

	case Bishop:
		directions := []int{1, -1}
		validMoves = appendMoves(validMoves, game, movingPiece, directions, false, true)

	case Rook:
		directions := []int{1, -1}
		validMoves = appendMoves(validMoves, game, movingPiece, directions, false, true)

	case Queen:
		directions := []int{1, 0, -1}
		validMoves = appendMoves(validMoves, game, movingPiece, directions, false, true)

	case King:
		directions := []int{1, 0, -1}
		validMoves = appendMoves(validMoves, game, movingPiece, directions, false, false)

	default:
		log.Fatal("Invalid piece")
	}
	return validMoves
}

func appendMoves(validMoves []Move, game *Game, movingPiece Piece, directions []int, isKnight bool, continuous bool) []Move {
	for _, i := range directions {
		for _, j := range directions {
			if !isKnight || abs(i) != abs(j) {
				if continuous {
					validMoves = appendDirectionMoveCont(validMoves, game, movingPiece, i, j)
				} else {
					validMoves = appendDirectionMove(validMoves, game, movingPiece, i, j)
				}
			}
		}
	}
	return validMoves
}

func appendIfInBounds(arr []Move, el Move) []Move {
	if el.Pos.X >= 0 && el.Pos.X <= 7 && el.Pos.Y >= 0 && el.Pos.Y <= 7 {
		return append(arr, el)
	}
	return arr
}

func appendDirectionMove(validMoves []Move, game *Game, movingPiece Piece, dx int, dy int) []Move {
	pos := Position{X: movingPiece.Pos.X + dx, Y: movingPiece.Pos.Y + dy}
	isValid := true
	takingN := -1
	for n, piece := range game.Board {
		if pos == piece.Pos {
			if movingPiece.Colour == piece.Colour {
				isValid = false
			} else {
				takingN = n
			}
		}
	}

	if isValid && pos != movingPiece.Pos {
		return appendIfInBounds(validMoves, Move{Pos: pos, TakingN: takingN})
	} else {
		return validMoves
	}
}

func appendDirectionMoveCont(validMoves []Move, game *Game, movingPiece Piece, dx int, dy int) []Move {
	x, y := 0, 0
	oldLen := -1

	// continue if: the last attempt found a new move AND the last move found was not a take
	for oldLen < len(validMoves) && !(len(validMoves) > 0 && x != 0 && validMoves[len(validMoves)-1].TakingN != -1) {
		x += dx
		y += dy

		oldLen = len(validMoves)
		validMoves = appendDirectionMove(validMoves, game, movingPiece, x, y)
	}
	return validMoves
}

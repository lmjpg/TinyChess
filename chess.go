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
	Board     map[Position]Piece
	Turn      int
	LastMoved *Piece
}

type Position struct {
	X, Y int
}

type Move struct {
	Pos       Position
	TakingPos Position
}

type Piece struct {
	Type, Colour    int
	HasMoved        bool
	PawnDoubleMoved bool
}

func invalidPosition() Position {
	return Position{X: -1, Y: -1}
}

func getInitialGame() *Game {
	board := make(map[Position]Piece)
	for c := range 2 {
		for i := range 8 {
			board[Position{i, 1 + (c * 5)}] = Piece{Pawn, c, false, false}
		}
		for i, v := range []int{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook} {
			board[Position{i, c * 7}] = Piece{v, c, false, false}
		}
	}

	game := Game{Board: board, Turn: White, LastMoved: nil}
	return &game
}

func movePiece(game *Game, movingPiecePos Position, newPos Position) {
	piece, ok := game.Board[movingPiecePos]
	if !ok {
		log.Fatalf("No piece at selected position (movePiece)\n\nPos: %v\n", movingPiecePos)
	}

	isValid, takingPos := isValidMove(game, movingPiecePos, newPos)
	if !isValid {
		return
	}

	if takingPos != invalidPosition() {
		delete(game.Board, takingPos)
	}

	if piece.Type == Pawn && movingPiecePos.Y+2 == newPos.Y {
		piece.PawnDoubleMoved = true
	} else {
		piece.PawnDoubleMoved = false
	}

	if game.Turn == White {
		game.Turn = Black
	} else {
		game.Turn = White
	}

	piece.HasMoved = true
	game.Board[newPos] = piece
	delete(game.Board, movingPiecePos)
	game.LastMoved = &piece
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

func isValidMove(game *Game, movingPiecePos Position, destinationPos Position) (bool, Position) {
	validMoves := getValidMoves(game, movingPiecePos)
	for _, validMove := range validMoves {
		if destinationPos == validMove.Pos {
			return true, validMove.TakingPos
		}
	}
	return false, invalidPosition()
}

func getValidMoves(game *Game, movingPiecePos Position) []Move {
	// prevent moves that put the current player into check:
	// AFTER all other stuff, for each move in validMoves create a clone of the current game with that move applied
	// then getValidMoves() for each cloned game (but with this check disabled to prevent infinite recursion)
	// check the TakingN on each of the cloned games Moves, if any of them match the current players king, the non-clone move that led to that position is illegal

	var validMoves []Move
	movingPiece, ok := game.Board[movingPiecePos]
	if !ok {
		log.Fatalf("No piece at position (getValidMoves)\nPos: %v\n", movingPiecePos)
	}

	if game.Turn != movingPiece.Colour {
		return validMoves
	}

	switch movingPiece.Type {
	case Pawn:
		// moving forward
		forwardPos := Position{X: movingPiecePos.X, Y: movingPiecePos.Y + getForward(movingPiece.Colour, 1)}
		_, occupied := game.Board[forwardPos]
		if !occupied {
			validMoves = appendIfInBounds(validMoves, Move{Pos: forwardPos, TakingPos: invalidPosition()})
			forwardPos.Y += getForward(movingPiece.Colour, 1)
			_, occupied := game.Board[forwardPos]
			if !occupied {
				validMoves = appendIfInBounds(validMoves, Move{Pos: forwardPos, TakingPos: invalidPosition()})
			}
		}

		// taking
		for _, i := range []int{1, -1} {
			takingPos := Position{X: movingPiecePos.X + i, Y: movingPiecePos.Y + getForward(movingPiece.Colour, 1)}
			_, occupied := game.Board[takingPos]
			if occupied {
				validMoves = appendIfInBounds(validMoves, Move{Pos: takingPos, TakingPos: takingPos})
			}

			// taking (en passant)
			takingPosEnPassant := Position{X: movingPiecePos.X + i, Y: movingPiecePos.Y}
			_, occupied = game.Board[takingPos]
			if occupied {
				validMoves = appendIfInBounds(validMoves, Move{Pos: takingPos, TakingPos: takingPosEnPassant})
			}
		}

	case Knight:
		directions := []int{1, 2, -1, -2}
		validMoves = appendMoves(validMoves, game, movingPiecePos, directions, movingPiece.Type, false)

	case Bishop:
		directions := []int{1, -1}
		validMoves = appendMoves(validMoves, game, movingPiecePos, directions, movingPiece.Type, true)

	case Rook:
		directions := []int{1, 0, -1}
		validMoves = appendMoves(validMoves, game, movingPiecePos, directions, movingPiece.Type, true)

	case Queen:
		directions := []int{1, 0, -1}
		validMoves = appendMoves(validMoves, game, movingPiecePos, directions, movingPiece.Type, true)

	case King:
		directions := []int{1, 0, -1}
		validMoves = appendMoves(validMoves, game, movingPiecePos, directions, movingPiece.Type, false)

	default:
		log.Fatalf("Invalid piece (getValidMoves)\nType: %v, Pos %v\n", movingPiece.Type, movingPiecePos)
	}
	return validMoves
}

func appendMoves(validMoves []Move, game *Game, movingPiecePos Position, directions []int, pieceType int, continuous bool) []Move {
	for _, i := range directions {
		for _, j := range directions {
			if (pieceType == Knight && abs(i) != abs(j)) || (pieceType == Rook && (i == 0 || j == 0)) || (pieceType != Knight && pieceType != Rook) {
				if continuous {
					validMoves = appendDirectionMoveCont(validMoves, game, movingPiecePos, i, j)
				} else {
					validMoves = appendDirectionMove(validMoves, game, movingPiecePos, i, j)
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

func appendDirectionMove(validMoves []Move, game *Game, movingPiecePos Position, dx int, dy int) []Move {
	pos := Position{X: movingPiecePos.X + dx, Y: movingPiecePos.Y + dy}

	takingPiece, ok := game.Board[pos]
	takingPos := invalidPosition()
	isValid := pos != movingPiecePos
	if ok {
		takingPos = pos
		movingPiece, ok := game.Board[movingPiecePos]
		if !ok {
			log.Fatalf("No piece at position (appendDirectionMove)\nPos: %v\n", movingPiecePos)
		}
		colour := movingPiece.Colour
		isValid = isValid && colour != takingPiece.Colour
	}

	if isValid {
		return appendIfInBounds(validMoves, Move{Pos: pos, TakingPos: takingPos})
	} else {
		return validMoves
	}
}

func appendDirectionMoveCont(validMoves []Move, game *Game, movingPiecePos Position, dx int, dy int) []Move {
	x, y := 0, 0
	oldLen := -1

	// continue if: the last attempt found a new move AND the last move found was not a take
	for oldLen < len(validMoves) && !(len(validMoves) > 0 && x != 0 && validMoves[len(validMoves)-1].TakingPos != invalidPosition()) {
		x += dx
		y += dy

		oldLen = len(validMoves)
		validMoves = appendDirectionMove(validMoves, game, movingPiecePos, x, y)
	}
	return validMoves
}

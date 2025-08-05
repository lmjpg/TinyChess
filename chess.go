package main

import (
	"log"
	"maps"
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
	Board           map[Position]Piece
	Turn            int
	LastMoved       *Position
	Checkmate       bool
	Draw            bool
	NoPawnMoveCount int
	DrawMoveHistory []map[Position]Piece
}

type Position struct {
	X, Y int
}

type Move struct {
	Pos            Position
	TakingPos      *Position
	CastleStartPos *Position
	CastleEndPos   *Position
}

type Piece struct {
	Type, Colour    int
	HasMoved        bool
	PawnDoubleMoved bool
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

	game := Game{Board: board, Turn: White, LastMoved: nil, Checkmate: false, Draw: false, NoPawnMoveCount: 0, DrawMoveHistory: make([]map[Position]Piece, 0)}
	return &game
}

func movePiece(game *Game, movingPiecePos Position, newPos Position, move *Move, doLegalCheck bool) bool {
	if move != nil && doLegalCheck {
		log.Fatalf("movePiece should only recieve a Move if it has already been checked for legality, otherwise use Positions.\n%v", move)
	}

	piece, ok := game.Board[movingPiecePos]
	if !ok {
		log.Fatalf("No piece at selected position (movePiece)\n\nPos: %v\n", movingPiecePos)
	}

	if doLegalCheck {
		if game.Checkmate {
			return false
		}
		var isValid bool
		isValid, move = isValidMove(game, movingPiecePos, newPos)
		if !isValid {
			return false
		}
	}

	if move != nil && move.TakingPos != nil {
		delete(game.Board, *move.TakingPos)
	}

	if piece.Type == Pawn && movingPiecePos.Y+getForward(piece.Colour, 2) == newPos.Y {
		piece.PawnDoubleMoved = true
	} else {
		piece.PawnDoubleMoved = false
	}

	piece.HasMoved = true
	game.Board[newPos] = piece
	delete(game.Board, movingPiecePos)

	if move != nil && move.CastleStartPos != nil && move.CastleEndPos != nil {
		movePiece(game, *move.CastleStartPos, *move.CastleEndPos, nil, false)
		changeTurn(game)
	}

	game.LastMoved = &newPos

	isInCheck := false
	if doLegalCheck {
		isInCheck = isKingAttacked(game)
	}

	changeTurn(game)

	if doLegalCheck {
		gameIsOver := true
		for pos := range game.Board {
			if len(getLegalMoves(game, pos)) > 0 {
				gameIsOver = false
			}
		}

		if gameIsOver {
			game.Checkmate = isInCheck
			game.Draw = !isInCheck
		} else {
			// fifty-move rule
			if piece.Type == Pawn {
				game.NoPawnMoveCount = 0
			} else {
				game.NoPawnMoveCount++

				// the 50 move rule refers to full moves (both sides make a move)
				// but this is counting half moves (only one side moves) so 100 is used instead
				if game.NoPawnMoveCount == 100 {
					game.Draw = true
				}
			}

			// three-move repetition
			boardCount := 1
			for _, boardToCheck := range game.DrawMoveHistory {
				if maps.Equal(game.Board, boardToCheck) {
					boardCount++
				}
				if boardCount == 3 {
					game.Draw = true
					break
				}
			}

			if move.TakingPos != nil || piece.Type == Pawn {
				game.DrawMoveHistory = make([]map[Position]Piece, 0) // taking/moving a pawn is not reversible, so the board history can be reset
			} else {
				game.DrawMoveHistory = append(game.DrawMoveHistory, maps.Clone(game.Board))
			}
		}
	}

	return true
}

func changeTurn(game *Game) {
	if game.Turn == White {
		game.Turn = Black
	} else {
		game.Turn = White
	}
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

func cloneGame(gameOriginal *Game) *Game {
	gameClone := *gameOriginal
	gameClone.Board = maps.Clone(gameClone.Board)
	return &gameClone
}

func isValidMove(game *Game, movingPiecePos Position, destinationPos Position) (bool, *Move) {
	validMoves := getLegalMoves(game, movingPiecePos)
	for _, validMove := range validMoves {
		if destinationPos == validMove.Pos {
			return true, &validMove
		}
	}
	return false, nil
}

func isKingAttacked(game *Game) bool {
	for pos := range game.Board {
		for _, moveToCheck := range getPseudoLegalMoves(game, pos) {
			if moveToCheck.TakingPos != nil {
				maybeKing, maybeTaking := game.Board[*moveToCheck.TakingPos]
				if maybeTaking && maybeKing.Type == King {
					return true
				}
			}
		}
	}
	return false
}

func getLegalMoves(gameOriginal *Game, movingPiecePos Position) []Move {
	if gameOriginal.Checkmate || gameOriginal.Draw {
		var move []Move
		return move
	}

	moves := getPseudoLegalMoves(gameOriginal, movingPiecePos)

	i := 0
	for i < len(moves) {
		moveToCheck := moves[i]

		gameClone := cloneGame(gameOriginal)
		movePiece(gameClone, movingPiecePos, moveToCheck.Pos, &moveToCheck, false)

		// if castling, check if the king is moving out of check or through check
		invalidCastle := false
		if moveToCheck.CastleEndPos != nil {
			gameClone2 := cloneGame(gameOriginal)
			gameClone2.Board[*moveToCheck.CastleEndPos] = gameClone2.Board[movingPiecePos]
			if gameClone2.Board[*moveToCheck.CastleEndPos].Type != King {
				log.Fatalf("A non-king piece is trying to castle: %v %v %v\n", movingPiecePos, gameClone2.Board[movingPiecePos], moveToCheck)
			}
			changeTurn(gameClone2)
			invalidCastle = isKingAttacked(gameClone2)
		}

		if invalidCastle || isKingAttacked(gameClone) {
			moves[i] = moves[len(moves)-1]
			moves = moves[:len(moves)-1]
		} else {
			i++
		}
	}

	return moves
}

func getPseudoLegalMoves(game *Game, movingPiecePos Position) []Move {
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
			validMoves = appendIfInBounds(validMoves, Move{Pos: forwardPos, TakingPos: nil})
			forwardPos.Y += getForward(movingPiece.Colour, 1)
			_, occupied := game.Board[forwardPos]
			if !occupied && !movingPiece.HasMoved {
				validMoves = appendIfInBounds(validMoves, Move{Pos: forwardPos, TakingPos: nil})
			}
		}

		// taking
		for _, i := range []int{1, -1} {
			pos := Position{X: movingPiecePos.X + i, Y: movingPiecePos.Y}
			destPos := pos
			destPos.Y += getForward(movingPiece.Colour, 1)
			for _, takingPos := range []Position{destPos, pos} {
				// if destPos != takingPos, en passant
				takingPiece, occupied := game.Board[takingPos]
				if occupied && movingPiece.Colour != takingPiece.Colour && (destPos == takingPos || (takingPiece.PawnDoubleMoved && takingPos == *game.LastMoved)) {
					validMoves = appendIfInBounds(validMoves, Move{Pos: destPos, TakingPos: &takingPos})
				}
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

		// castling
		if !movingPiece.HasMoved {
			for _, dir := range []int{1, -1} {
				rookPos := Position{X: ((dir + 1) / 2) * 7, Y: movingPiecePos.Y}
				rook, ok := game.Board[rookPos]
				if ok && rook.Type == Rook && !rook.HasMoved {
					canCastle := true
					checkPos := Position{X: movingPiecePos.X + dir, Y: movingPiecePos.Y}
					for checkPos.X != rookPos.X {
						_, occupied := game.Board[checkPos]
						if occupied {
							canCastle = false
							break
						}
						checkPos.X += dir
						if !inBounds(checkPos) {
							log.Fatalf("%v is out of bounds, King Position is %v, Rook Position is %v\n", checkPos, movingPiecePos, rookPos)
						}
					}
					if canCastle {
						kingNewPos := Position{X: movingPiecePos.X + dir*2, Y: movingPiecePos.Y}
						rookNewPos := Position{X: movingPiecePos.X + dir*1, Y: movingPiecePos.Y}
						validMoves = appendIfInBounds(validMoves, Move{Pos: kingNewPos, TakingPos: nil, CastleStartPos: &rookPos, CastleEndPos: &rookNewPos})
					}
				}
			}
		}

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

func inBounds(pos Position) bool {
	return pos.X >= 0 && pos.X <= 7 && pos.Y >= 0 && pos.Y <= 7
}

func appendIfInBounds(arr []Move, el Move) []Move {
	if inBounds(el.Pos) {
		return append(arr, el)
	}
	return arr
}

func appendDirectionMove(validMoves []Move, game *Game, movingPiecePos Position, dx int, dy int) []Move {
	pos := Position{X: movingPiecePos.X + dx, Y: movingPiecePos.Y + dy}

	takingPiece, ok := game.Board[pos]
	var takingPos *Position = nil
	isValid := pos != movingPiecePos
	if ok {
		takingPos = &pos
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
	for oldLen < len(validMoves) && !(len(validMoves) > 0 && (x != 0 || y != 0) && validMoves[len(validMoves)-1].TakingPos != nil) {
		x += dx
		y += dy

		oldLen = len(validMoves)
		validMoves = appendDirectionMove(validMoves, game, movingPiecePos, x, y)
	}
	return validMoves
}

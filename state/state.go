package state

import (
	"math"
	"fmt"
	"strconv"
	"github.com/jinzhu/copier"
	"github.com/Pedraamy/Golang-RL-Chess-AI/pieces"
)

var (
	posTable = PosTable()
)

type State struct {
	White uint8
	WK uint64
	WQ uint64
	WR uint64
	WB uint64
	WN uint64
	WP uint64
	BK uint64
	BQ uint64
	BR uint64
	BB uint64
	BN uint64
	BP uint64
	CastleWK uint8
	CastleWQ uint8
	CastleBK uint8
	CastleBQ uint8
	JustCastled uint8
}

type Move struct {
	Name uint8
	Piece uint64
	Start uint64
	End uint64
	Castle uint8
}

func NewBoard() *State {
	var (
		White uint8 = 1
		WK uint64 = 1<<3
		WQ uint64 = 1<<4
		WR uint64 = 1 | 1<<7
		WB uint64 = 1<<2 | 1<<5
		WN uint64 = 1<<1 | 1<<6
		WP uint64 = 1<<8|1<<9|1<<10|1<<11|1<<12|1<<13|1<<14|1<<15
		BK uint64 = 1<<59
		BQ uint64 = 1<<60
		BR uint64 = 1<<56 | 1<<63
		BB uint64 = 1<<58 | 1<<61
		BN uint64 = 1<<57 | 1<<62
		BP uint64 = 1<<48|1<<49|1<<50|1<<51|1<<52|1<<53|1<<54|1<<55
	)
	return &State{White, WK, WQ, WR, WB, WN, WP, BK, BQ, BR, BB, BN, BP, 1, 1, 1, 1, 0}
}

func NewMove(name uint8, piece uint64, start uint64, end uint64, castle uint8) *Move {
	return &Move{name, piece, start, end, castle}
}


func (st *State) AllWhitePieces() uint64 {
	return st.WK|st.WQ|st.WR|st.WB|st.WN|st.WP
}

func (st *State) AllBlackPieces() uint64 {
	return st.BK|st.BQ|st.BR|st.BB|st.BN|st.BP
}

func (st *State) StateFromMove(mv *Move) *State {
	if st.White == 1 {
		return st.StateFromMoveWhite(mv)
	} else {
		return st.StateFromMoveBlack(mv)
	}
}

func (st *State) StateFromMoveWhite(mv *Move) *State {
	ns := &State{}
	copier.Copy(ns, st)
	ns.White ^= 1
	if mv.Castle != 0 {
		if mv.Castle == 1 {
			ns.WK = 1<<1
			ns.WR &= ^uint64(1)
			ns.WR |= 1<<2
			ns.JustCastled = 1
		} else {
			ns.WK = 1<<5
			ns.WR &= ^uint64(1<<7)
			ns.WR |= 1<<4
			ns.JustCastled = 2
		}
		ns.CastleWK = 0
		ns.CastleWQ = 0
	} else {
		np := mv.Piece&(^mv.Start)
		np |= mv.End
		ns.JustCastled = 0
		if mv.Name == 6 {
			ns.WK = np
			ns.CastleWK = 0
			ns.CastleWQ = 0
		} else if mv.Name == 5 {
			ns.WQ = np
		} else if mv.Name == 4 {
			ns.WR = np
			if mv.Start == 1 {
				ns.CastleWK = 0
			}
			if mv.Start == 1<<7 {
				ns.CastleWQ = 0
			}
		} else if mv.Name == 3 {
			ns.WB = np
		} else if mv.Name == 2 {
			ns.WN = np
		} else if mv.Name == 1{
			ns.WP = np
		}
		ns.BK &= ^mv.End
		ns.BQ &= ^mv.End
		ns.BR &= ^mv.End
		ns.BB &= ^mv.End
		ns.BN &= ^mv.End
		ns.BP &= ^mv.End
	}
	return ns
}


func (st *State) StateFromMoveBlack(mv *Move) *State {
	ns := &State{}
	copier.Copy(ns, st)
	ns.White ^= 1

	if mv.Castle != 0 {
		if mv.Castle == 1 {
			ns.BK = 1<<57
			ns.BR &= ^uint64(1<<56)
			ns.BR |= 1<<58
			ns.JustCastled = 1
		} else {
			ns.BK = 1<<61
			ns.WR &= ^uint64(1<<63)
			ns.WR |= 1<<60
			ns.JustCastled = 2
		}
		ns.CastleBK = 0
		ns.CastleBQ = 0
	} else {
		np := mv.Piece&(^mv.Start)
		np |= mv.End
		ns.JustCastled = 0
		if mv.Name == 6 {
			ns.BK = np
			ns.CastleBK = 0
			ns.CastleBQ = 0
		} else if mv.Name == 5 {
			ns.BQ = np
		} else if mv.Name == 4 {
			ns.BR = np
			if mv.Start == 1<<56 {
				ns.CastleBK = 0
			}
			if mv.Start == 1<<63 {
				ns.CastleBQ = 0
			}
		} else if mv.Name == 3 {
			ns.BB = np
		} else if mv.Name == 2 {
			ns.BN = np
		} else if mv.Name == 1{
			ns.BP = np
		}
		ns.WK &= ^mv.End
		ns.WQ &= ^mv.End
		ns.WR &= ^mv.End
		ns.WB &= ^mv.End
		ns.WN &= ^mv.End
		ns.WP &= ^mv.End
	}
	return ns
}

func (st *State) GetAllMoves() ([]*Move, []*Move) {
	if st.White == 1 {
		return st.GetAllMovesWhite()
	} else {
		return st.GetAllMovesBlack()
	}
}

func (st *State) GetAllMovesWhite() ([]*Move, []*Move) {
	moves := []*Move{}
	captures := []*Move{}
	white := st.AllWhitePieces()
	black := st.AllBlackPieces()

	//Castles
	if st.CanCastleKingWhite() {
		moves = append(moves, NewMove(0, 0, 0, 0, 1))
	}
	if st.CanCastleQueenWhite() {
		moves = append(moves, NewMove(0, 0, 0, 0, 2))
	}
	//King
	kings := pieces.GetPiecesFromBoard(st.WK)
	for _, k := range kings {
		caps, movs := pieces.KingMoves(k, white, black)
		for _, c := range caps {
			captures = append(captures, NewMove(6, st.WK, k, c, 0))
		}
		for _, m := range movs {
			moves = append(moves, NewMove(6, st.WK, k, m, 0))
		}
	}
	//Queens
	queens := pieces.GetPiecesFromBoard(st.WQ)
	for _, q := range queens {
		caps, movs := pieces.QueenMoves(q, white, black)
		for _, c := range caps {
			captures = append(captures, NewMove(5, st.WQ, q, c, 0))
		}
		for _, m := range movs {
			moves = append(moves, NewMove(5, st.WQ, q, m, 0))
		}
	}
	//Rooks
	rooks := pieces.GetPiecesFromBoard(st.WR)
	for _, r := range rooks {
		caps, movs := pieces.RookMoves(r, white, black)
		for _, c := range caps {
			captures = append(captures, NewMove(4, st.WR, r, c, 0))
		}
		for _, m := range movs {
			moves = append(moves, NewMove(4, st.WR, r, m, 0))
		}
	}
	//Bishops
	bishops := pieces.GetPiecesFromBoard(st.WB)
	for _, b := range bishops {
		caps, movs := pieces.BishopMoves(b, white, black)
		for _, c := range caps {
			captures = append(captures, NewMove(3, st.WB, b, c, 0))
		}
		for _, m := range movs {
			moves = append(moves, NewMove(3, st.WB, b, m, 0))
		}
	}
	//Knights
	knights := pieces.GetPiecesFromBoard(st.WN)
	for _, n := range knights {
		caps, movs := pieces.KnightMoves(n, white, black)
		for _, c := range caps {
			captures = append(captures, NewMove(2, st.WN, n, c, 0))
		}
		for _, m := range movs {
			moves = append(moves, NewMove(2, st.WN, n, m, 0))
		}
	}
	//Pawns
	pawns := pieces.GetPiecesFromBoard(st.WP)
	for _, p := range pawns {
		caps, movs := pieces.PawnMoves(p, white, black, 1)
		for _, c := range caps {
			captures = append(captures, NewMove(1, st.WP, p, c, 0))
		}
		for _, m := range movs {
			moves = append(moves, NewMove(1, st.WP, p, m, 0))
		}
	}
	return captures, moves
}

func (st *State) GetAllMovesBlack() ([]*Move, []*Move) {
	moves := []*Move{}
	captures := []*Move{}
	white := st.AllWhitePieces()
	black := st.AllBlackPieces()
	//Castles
	if st.CanCastleKingBlack() {
		moves = append(moves, NewMove(0, 0, 0, 0, 1))
	}
	if st.CanCastleQueenBlack() {
		moves = append(moves, NewMove(0, 0, 0, 0, 2))
	}
	//King
	kings := pieces.GetPiecesFromBoard(st.BK)
	for _, k := range kings {
		caps, movs := pieces.KingMoves(k, black, white)
		for _, c := range caps {
			captures = append(captures, NewMove(6, st.BK, k, c, 0))
		}
		for _, m := range movs {
			moves = append(moves, NewMove(6, st.BK, k, m, 0))
		}
	}
	//Queens
	queens := pieces.GetPiecesFromBoard(st.BQ)
	for _, q := range queens {
		caps, movs := pieces.QueenMoves(q, black, white)
		for _, c := range caps {
			captures = append(captures, NewMove(5, st.BQ, q, c, 0))
		}
		for _, m := range movs {
			moves = append(moves, NewMove(5, st.BQ, q, m, 0))
		}
	}
	//Rooks
	rooks := pieces.GetPiecesFromBoard(st.BR)
	for _, r := range rooks {
		caps, movs := pieces.RookMoves(r, black, white)
		for _, c := range caps {
			captures = append(captures, NewMove(4, st.BR, r, c, 0))
		}
		for _, m := range movs {
			moves = append(moves, NewMove(4, st.BR, r, m, 0))
		}
	}
	//Bishops
	bishops := pieces.GetPiecesFromBoard(st.BB)
	for _, b := range bishops {
		caps, movs := pieces.BishopMoves(b, black, white)
		for _, c := range caps {
			captures = append(captures, NewMove(3, st.BB, b, c, 0))
		}
		for _, m := range movs {
			moves = append(moves, NewMove(3, st.BB, b, m, 0))
		}
	}
	//Knights
	knights := pieces.GetPiecesFromBoard(st.BN)
	for _, n := range knights {
		caps, movs := pieces.KnightMoves(n, black, white)
		for _, c := range caps {
			captures = append(captures, NewMove(2, st.BN, n, c, 0))
		}
		for _, m := range movs {
			moves = append(moves, NewMove(2, st.BN, n, m, 0))
		}
	}
	//Pawns
	pawns := pieces.GetPiecesFromBoard(st.BP)
	for _, p := range pawns {
		caps, movs := pieces.PawnMoves(p, black, white, 0)
		for _, c := range caps {
			captures = append(captures, NewMove(1, st.BP, p, c, 0))
		}
		for _, m := range movs {
			moves = append(moves, NewMove(1, st.BP, p, m, 0))
		}
	}

	return captures, moves
}


func GetPositionsFromBoard(piece uint64) []uint64 {
	res := []uint64{}
	var curr, np uint64
	for piece != 0 {
		np = piece&(piece-1)
		curr = np^piece
		fmt.Println(strconv.FormatInt(int64(curr), 2))
		res = append(res, curr)
		piece = np
	}
	return res
}

func (st *State) CanCastleKingWhite() bool {
	if st.CastleWK == 0 {
		return false
	}
	white := st.AllWhitePieces()
	black := st.AllBlackPieces()
	both := white|black
	var need uint64 = 1<<1|1<<2
	if need&both != 0 || 1&black != 0 {
		return false
	}
	return true
}

func (st *State) CanCastleQueenWhite() bool {
	if st.CastleWQ == 0 {
		return false
	}
	white := st.AllWhitePieces()
	black := st.AllBlackPieces()
	both := white|black
	var need uint64 = 1<<4|1<<5|1<<6
	if need&both != 0 || 1<<7&black != 0 {
		return false
	}
	return true
}

func (st *State) CanCastleKingBlack() bool {
	if st.CastleBK == 0 {
		return false
	}
	white := st.AllWhitePieces()
	black := st.AllBlackPieces()
	both := white|black
	var need uint64 = 1<<57|1<<58
	if need&both != 0 || 1<<56&white != 0 {
		return false
	}
	return true
}

func (st *State) CanCastleQueenBlack() bool {
	if st.CastleBQ == 0 {
		return false
	}
	white := st.AllWhitePieces()
	black := st.AllBlackPieces()
	both := white|black
	var need uint64 = 1<<60|1<<61|1<<62
	if need&both != 0 || 1<<63&white != 0 {
		return false
	}
	return true
}

func PawnMoves(bin uint64, same uint64, opp uint64, color uint8) ([]uint64, []uint64) {
	moves := []uint64{}
	captures := []uint64{}
	row, col := GetRowCol(bin)
	both := same|opp
	var curr uint64

	if color == 1 {
		curr = GetBinPos(row+1, col)
		if curr&both == 0 {
			moves = append(moves, curr)
			if row == 1 {
				curr = GetBinPos(row+2, col)
				if curr&both == 0 {
					moves = append(moves, curr)
				}
			}
		}
		if col > 0 {
			curr = GetBinPos(row+1, col-1)
			if curr&opp != 0 {
				captures = append(captures, curr)
			}
		}
		if col < 7 {
			curr = GetBinPos(row+1, col+1)
			if curr&opp != 0 {
				captures = append(captures, curr)
			}
		}
	} else {
		curr = GetBinPos(row-1, col)
		if curr&both == 0 {
			moves = append(moves, curr)
			if row == 6 {
				curr = GetBinPos(row-2, col)
				if curr&both == 0 {
					moves = append(moves, curr)
				}
			}
		}
		if col > 0 {
			curr = GetBinPos(row-1, col-1)
			if curr&opp != 0 {
				captures = append(captures, curr)
			}
		}
		if col < 7 {
			curr = GetBinPos(row-1, col+1)
			if curr&opp != 0 {
				captures = append(captures, curr)
			}
		}
	}
	return captures, moves
}

func KingMoves(bin uint64, same uint64, opp uint64) ([]uint64, []uint64) {
	moves := []uint64{}
	captures := []uint64{}
	row, col := GetRowCol(bin)
	var curr uint64
	curr = GetBinPos(row+1, col+1)
	if curr&opp != 0 {
		captures = append(captures, curr)
	} else if curr&same == 0{
		moves = append(moves, curr)
	}
	curr = GetBinPos(row-1, col-1)
	if curr&opp != 0 {
		captures = append(captures, curr)
	} else if curr&same == 0{
		moves = append(moves, curr)
	}
	curr = GetBinPos(row+1, col-1)
	if curr&opp != 0 {
		captures = append(captures, curr)
	} else if curr&same == 0{
		moves = append(moves, curr)
	}
	curr = GetBinPos(row-1, col+1)
	if curr&opp != 0 {
		captures = append(captures, curr)
	} else if curr&same == 0{
		moves = append(moves, curr)
	}
	curr = GetBinPos(row+1, col)
	if curr&opp != 0 {
		captures = append(captures, curr)
	} else if curr&same == 0{
		moves = append(moves, curr)
	}
	curr = GetBinPos(row, col+1)
	if curr&opp != 0 {
		captures = append(captures, curr)
	} else if curr&same == 0{
		moves = append(moves, curr)
	}
	curr = GetBinPos(row-1, col)
	if curr&opp != 0 {
		captures = append(captures, curr)
	} else if curr&same == 0{
		moves = append(moves, curr)
	}
	curr = GetBinPos(row, col-1)
	if curr&opp != 0 {
		captures = append(captures, curr)
	} else if curr&same == 0{
		moves = append(moves, curr)
	}
	return captures, moves
}

func QueenMoves(bin uint64, same uint64, opp uint64) ([]uint64, []uint64) {
	captures, moves := RookMoves(bin, same, opp)
	cap2, mov2 := BishopMoves(bin, same, opp)
	captures = append(captures, cap2...)
	moves = append(moves, mov2...)
	return captures, moves
}

func RookMoves(bin uint64, same uint64, opp uint64) ([]uint64, []uint64) {
	moves := []uint64{}
	captures := []uint64{}
	row, col := GetRowCol(bin)
	var curr, nr, nc uint64
	nr = row+1
	for nr < 8 {
		curr = GetBinPos(nr, col)
		if curr&opp != 0{
			captures = append(captures, curr)
			break
		} else if curr&same != 0{
			break
		} else {
			moves = append(moves, curr)
		}
		nr += 1
	}
	nr = row-1
	for nr >= 0 {
		curr = GetBinPos(nr, col)
		if curr&opp != 0{
			captures = append(captures, curr)
			break
		} else if curr&same != 0{
			break
		} else {
			moves = append(moves, curr)
		}
		nr -= 1
	}
	nc = col+1
	for nc < 8 {
		curr = GetBinPos(row, nc)
		if curr&opp != 0{
			captures = append(captures, curr)
			break
		} else if curr&same != 0{
			break
		} else {
			moves = append(moves, curr)
		}
		nc += 1
	}
	nc = col-1
	for nc >= 0 {
		curr = GetBinPos(row, nc)
		if curr&opp != 0{
			captures = append(captures, curr)
			break
		} else if curr&same != 0{
			break
		} else {
			moves = append(moves, curr)
		}
		nc -= 1
	}
	return captures, moves
}

func BishopMoves(bin uint64, same uint64, opp uint64) ([]uint64, []uint64) {
	moves := []uint64{}
	captures := []uint64{}
	row, col := GetRowCol(bin)
	var curr, nr, nc uint64
	nr = row+1
	nc = col+1
	for nr < 8 && nc < 8 {
		curr = GetBinPos(nr, nc)
		if curr&opp != 0{
			captures = append(captures, curr)
			break
		} else if curr&same != 0{
			break
		} else {
			moves = append(moves, curr)
		}
		nr += 1
		nc += 1
	}
	nr = row-1
	nc = col-1
	for nr >= 0 && nc >= 0 {
		curr = GetBinPos(nr, nc)
		if curr&opp != 0{
			captures = append(captures, curr)
			break
		} else if curr&same != 0{
			break
		} else {
			moves = append(moves, curr)
		}
		nr += 1
		nc += 1
	}
	nr = row+1
	nc = col-1
	for nr < 8 && nc >= 0 {
		curr = GetBinPos(nr, nc)
		if curr&opp != 0{
			captures = append(captures, curr)
			break
		} else if curr&same != 0{
			break
		} else {
			moves = append(moves, curr)
		}
		nr += 1
		nc -= 1
	}
	nr = row-1
	nc = col+1
	for nr >= 0 && nc < 8 {
		curr = GetBinPos(nr, nc)
		if curr&opp != 0{
			captures = append(captures, curr)
			break
		} else if curr&same != 0{
			break
		} else {
			moves = append(moves, curr)
		}
		nr += 1
		nc += 1
	}
	return captures, moves

}
 
func KnightMoves(bin uint64, same uint64, opp uint64) ([]uint64, []uint64) {
	moves := []uint64{}
	captures := []uint64{}
	row, col := GetRowCol(bin)
	var curr, nr, nc uint64
	nr = row+2
	if nr < 8 {
		if col-1 >= 0 {
			curr = GetBinPos(nr, col-1)
			if curr&opp != 0 {
				captures = append(captures, curr)
			} else if curr&same == 0 {
				moves = append(moves, curr)
			}
		}
		if col+1 < 8 {
			curr = GetBinPos(nr, col+1)
			if curr&opp != 0 {
				captures = append(captures, curr)
			} else if curr&same == 0 {
				moves = append(moves, curr)
			}
		}
	}
	nr = row-2
	if nr < 8 {
		if col-1 >= 0 {
			curr = GetBinPos(nr, col-1)
			if curr&opp != 0 {
				captures = append(captures, curr)
			} else if curr&same == 0 {
				moves = append(moves, curr)
			}
		}
		if col+1 < 8 {
			curr = GetBinPos(nr, col+1)
			if curr&opp != 0 {
				captures = append(captures, curr)
			} else if curr&same == 0 {
				moves = append(moves, curr)
			}
		}
	}
	nc = col+2
	if nc < 8 {
		if row-1 >= 0 {
			curr = GetBinPos(row-1, nc)
			if curr&opp != 0 {
				captures = append(captures, curr)
			} else if curr&same == 0 {
				moves = append(moves, curr)
			}
		}
		if row+1 < 8 {
			curr = GetBinPos(row+1, nc)
			if curr&opp != 0 {
				captures = append(captures, curr)
			} else if curr&same == 0 {
				moves = append(moves, curr)
			}
		}
	}
	nc = col-2
	if nc >= 0 {
		if row-1 >= 0 {
			curr = GetBinPos(row-1, nc)
			if curr&opp != 0 {
				captures = append(captures, curr)
			} else if curr&same == 0 {
				moves = append(moves, curr)
			}
		}
		if row+1 < 8 {
			curr = GetBinPos(row+1, nc)
			if curr&opp != 0 {
				captures = append(captures, curr)
			} else if curr&same == 0 {
				moves = append(moves, curr)
			}
		}
	}
	return captures, moves
}

func GetRowCol(bin uint64) (uint64, uint64) {
	pos := posTable[bin]
	row := pos/8
	col := pos%8
	return row, col
}

func GetBinPos(row uint64, col uint64) uint64 {
	return uint64(math.Pow(2, float64(8*row+col)))
}



func PosTable() map[uint64]uint64 {
	table := make(map[uint64]uint64)
	var curr uint64
	for i := 0; i<64; i++ {
		curr = uint64(math.Pow(2, float64(i)))
		table[curr] = uint64(i)
	}
	return table
}
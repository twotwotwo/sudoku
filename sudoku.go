package main

import (
	"bytes"
	"fmt"
	"time"
)

var WeirdPuzzle = NewPuzzle(`
__53_____
8______2_
_7__1_5__
4____53__
_1__7___6
__32___8_
_6_5____9
__4____3_
_____97__
`)

var NormalPuzzle = NewPuzzle(`
_3__7__6_
__79_81__
8___4___9
1__2_4__5
__5___2__
2__3_6__7
3___8___2
__85_97__
_9__2__5_`)

type Puzzle []byte
type SolutionList []Puzzle

func BlankPuzzle() Puzzle {
	return make(Puzzle, 81)
}

func NewPuzzle(board string) (puz Puzzle) {
	puz = make(Puzzle, 0, 81)
	for _, c := range board {
		if c == '_' {
			c = 0
		} else if c >= '0' && c <= '9' {
			c -= '0'
		} else {
			// whitespace
			continue
		}
		puz = append(puz, byte(c))
	}
	return
}

var squareStarts = []int{0, 3, 6, 27, 30, 33, 54, 57, 60}
var squareOffsets = []int{0, 1, 2, 9, 10, 11, 18, 19, 20}

func (p Puzzle) String() string {
	s := ""
	for i := 0; i < 9; i++ {
		s += fmt.Sprint([]byte(p[i*9:(i+1)*9])) + "\n"
	}
	return s[:len(s)-1] // no \n
}

func (sl SolutionList) String() string {
	s := "[\n"
	for _, p := range sl {
		s += p.String() + "\n\n"
	}
	return s[:len(s)-1] + "]"
}

func (p Puzzle) OK() bool {
	for i := 0; i < 9; i++ {
		var seenRow, seenCol, seenSquare int
		for j := 0; j < 9; j++ {
			rowChar := p[i*9+j]
			if rowChar > 0 {
				if seenRow&(1<<rowChar) != 0 {
					return false
				}
				seenRow |= 1 << rowChar
			}

			colChar := p[i+j*9]
			if colChar > 0 {
				if seenCol&(1<<colChar) != 0 {
					return false
				}
				seenCol |= 1 << colChar
			}

			squareChar := p[squareStarts[i]+squareOffsets[j]]
			if squareChar > 0 {
				if seenSquare&(1<<squareChar) != 0 {
					return false
				}
				seenSquare |= 1 << squareChar
			}

		}
	}

	return true
}

type state struct {
	i, row, col, square byte
	fixed, tried        bool
}

type Mask uint16

// Solve, but keep track of what digits are taken for each row/col/square
// as you go, to make checking possibilities easy.
func (p Puzzle) Solve2(stopAt int) (solutions int) {
	var rows, cols, squares [9]Mask
	var stackBuf [81]state
	stack := stackBuf[:0]

	for i, c := range p {
		row := i / 9
		col := i % 9
		square := (i/27)*3 + (i%9)/3
		if c != 0 {
			mask := Mask(1) << c
			rows[row] |= mask
			cols[col] |= mask
			squares[square] |= mask
			continue
		}
		stack = append(stack, state{
			i:      byte(i),
			row:    byte(row),
			col:    byte(col),
			square: byte(square),
			fixed:  false,
			tried:  false,
		})
	}
	maxDepth := len(stack)

	var s state
	var row, col, square, i byte
	var takenDigits Mask
	depth := 0
	loadState := true

	//TRY:
	for {
		if loadState || true {
			s = stack[depth]
			row, col, square, i = s.row, s.col, s.square, s.i
			takenDigits = rows[row] | cols[col] | squares[square]
			loadState = false
		}
		p[i]++
		if p[i] == 10 || s.fixed { // out of options!
			stack[depth].fixed = false
			p[i] = 0
			depth--
			if depth < 0 { // and we're done!
				return
			}

			// fully backtrack--zero bits in masks set for the previous guess
			s = stack[depth]
			row, col, square, i = s.row, s.col, s.square, s.i
			takenDigits = rows[row] | cols[col] | squares[square]
			mask := Mask(1) << p[i]
			rows[row] &= ^mask
			cols[col] &= ^mask
			squares[square] &= ^mask
			continue
		}
		mask := Mask(1) << p[i]
		for takenDigits&mask != 0 && p[i] < 9 {
			mask <<= 1
			p[i]++
		}
		if takenDigits&mask != 0 { // darn, failed, have to backtrack
			continue
		}
		rows[row] |= mask
		cols[col] |= mask
		squares[square] |= mask
		depth++
		if s.fixed {
			s.tried = true
		}
		loadState = true
		if depth == maxDepth { // we filled in all digits; solved!
			solutions++
			// can stop at 1 solution or 2 or 5000
			if solutions == stopAt {
				return
			}
			// keep on truckin' by incrementing last free digit
			depth--
			loadState = false
			mask := Mask(1) << p[i]
			rows[row] &= ^mask
			cols[col] &= ^mask
			squares[square] &= ^mask
		} else {
			// let's find the square with the fewest choices (0 is few!),
			// and swap it into the next position
			rest := stack[depth:]
			restCnt := len(rest)
			mostTaken := -1
			mostPos := -1
			for i := 0; i < restCnt; i++ {
				s := rest[i]
				taken := 0
				takenDigits := rows[s.row] | cols[s.col] | squares[s.square]
				for takenDigits > 0 {
					if takenDigits&1 != 0 {
						taken++
					}
					takenDigits >>= 1
				}
				if taken > mostTaken {
					mostTaken = taken
					mostPos = i
				}
				// Bail if there's only one choice (or, as occasionally
				// happens, no choice)
				if taken > 7 {
					break
				}
			}
			// doing whoever has most digits taken next!
			t := stack[depth]
			stack[depth] = rest[mostPos]
			rest[mostPos] = t
			/*if mostTaken < 8 && depth < maxDepth-1 {
				var digitPlacesBySquare, digitPlacesByRow, digitPlacesByCol [9][9]byte
				for i := byte(0); i < 81; i++ {
					if p[i] != 0 {
						continue
					}
					row := i / 9
					col := i % 9
					square := (i/27)*3 + (i%9)/3
					takenDigits := rows[row] | cols[col] | squares[square]
					// note the 'digit' var is 0-8 here, not 1-9
					for digit := 0; digit < 9; digit++ {
						takenDigits >>= 1
						if takenDigits&1 == 0 { // not taken! can be here
							if digitPlacesBySquare[digit][square] == 0 {
								digitPlacesBySquare[digit][square] = i
							} else {
								digitPlacesBySquare[digit][square] = 255
							}
							if digitPlacesByRow[digit][row] == 0 {
								digitPlacesByRow[digit][row] = i
							} else {
								digitPlacesByRow[digit][row] = 255
							}
							if digitPlacesByCol[digit][col] == 0 {
								digitPlacesByCol[digit][col] = i
							} else {
								digitPlacesByCol[digit][col] = 255
							}
						}
					}
				}
				var knownSquares [81]byte
				for digit := 0; digit < 9; digit++ {
					for i := 0; i < 9; i++ {
						if digitPlacesBySquare[digit][i] != 0 &&
							digitPlacesBySquare[digit][i] != 255 {
							knownSquares[digitPlacesBySquare[digit][i]] = byte(digit + 1)
						}
						if digitPlacesByRow[digit][i] != 0 &&
							digitPlacesByRow[digit][i] != 255 {
							knownSquares[digitPlacesByRow[digit][i]] = byte(digit + 1)
						}
						if digitPlacesByCol[digit][i] != 0 &&
							digitPlacesByCol[digit][i] != 255 {
							knownSquares[digitPlacesByCol[digit][i]] = byte(digit + 1)
						}
					}
				}

				for restPos, s := range rest {
					i := s.i
					val := knownSquares[i]
					if val == 0 {
						continue
					}
					// swap it into the next slot
					t := stack[depth]
					stack[depth] = rest[restPos]
					rest[restPos] = t
					p[i] = val
					stack[depth].fixed = true
					s = stack[depth]
					row, col, square, i = s.row, s.col, s.square, s.i
					mask := Mask(1) << val
					rows[row] |= mask
					cols[col] |= mask
					squares[square] |= mask
					depth++
					if depth < maxDepth-1 {
						continue TRY // loads a *new* state, not s above
					}
				}
			}
			*/
		}
	}
}

func (p Puzzle) Blank() Puzzle {
	for i := range p {
		p[i] = 0
	}
	return p
}

// Make a puzzle up w/23 hints by random trial and error
// lots fewer hints would take smarter search or solver w/better worst-case
// perf
func MakePuzzleUp() (p Puzzle) {
	p = BlankPuzzle()
	scratch := BlankPuzzle()
RAND:
	for {
		// get a random puzzle
		p.Blank().RandomSolution(0)
		// sudoku only need 17 hints, sez the internet (but just random
		// probing doesn't find better than 23 in reasonable time)
		for i := 0; i < 81; i++ {
			rnd *= 7
			for p[rnd%81] == 0 {
				rnd *= 99
				rnd ^= 313370
			}
			overwrote := p[rnd%81]
			p[rnd%81] = 0
			copy(scratch, p)
			cnt := scratch.Solve2(2)
			if cnt > 1 { // ok, just crossed the border
				p[rnd%81] = overwrote
				if i > 57 { // we win!
					return p
				} else { // too early; poke around a bit, or retry
					rnd *= 47
					rnd ^= 313370
					i--
					if rnd%39 == 0 {
						continue RAND
					}
				}
			}
			if cnt == 0 {
				panic(
					"could not solve random solution. think about your " +
						"choices, randall",
				)
			}
		}
	}
}

func (p Puzzle) Solve(pos int) bool {
	for off, c := range p[pos:] {
		if c != 0 {
			continue
		}
		for i := byte(1); i <= 9; i++ {
			p[pos+off] = i
			if p.OK() && p.Solve(pos+off+1) {
				return true
			}
		}
		p[pos+off] = 0
		return false
	}
	return p.OK()
}

var rnd = uint(7)

func (p Puzzle) AppendSolutions(pos int, solutionsIn SolutionList) (solutionsOut SolutionList) {
	solutionsOut = solutionsIn
	rnd *= 7
	startDigit := byte(rnd)
	for off, c := range p[pos:] {
		if c != 0 {
			continue
		}
		for i := byte(1); i <= 9; i++ {
			p[pos+off] = ((startDigit + i) % 9) + 1
			if p.OK() {
				solutionsOut = p.AppendSolutions(pos+off+1, solutionsOut)
			}
		}
		p[pos+off] = 0
		return
	}
	if p.OK() {
		solutionsOut = append(solutionsOut, append(Puzzle(nil), p...))
		return
	}
	// we got nothin
	return
}

// Find an arbitrary solution (from hints or from an empty puzzle)
func (p Puzzle) RandomSolution(pos int) bool {
	rnd *= 11
	startDigit := byte(rnd % 9)
	for off, c := range p[pos:] {
		if c != 0 {
			continue
		}
		for digitDiff := byte(0); digitDiff <= 8; digitDiff++ {
			digit := ((digitDiff + startDigit) % 9) + 1
			p[pos+off] = digit
			if p.OK() && p.RandomSolution(pos+off+1) {
				return true
			}
		}
		// failed, unset digit
		p[pos+off] = 0
		return false
	}
	return p.OK()
}

func (p Puzzle) HintCount() (hc int) {
	for _, c := range p {
		if c == 0 {
			continue
		}
		hc++
	}
	return
}

func kitchenSink() {
	puzzle := WeirdPuzzle
	orig := append(Puzzle(nil), puzzle...)
	fmt.Print("The puzzle:\n", puzzle, "\n")
	fmt.Println(orig.HintCount(), "hints")
	fmt.Print("Solution:\n")
	t := time.Now()
	puzzle.Solve2(1)
	fmt.Println(time.Now().Sub(t).Nanoseconds())
	fmt.Println(puzzle)

	sol1 := append(Puzzle(nil), puzzle...)
	copy(puzzle, orig)
	puzzle.Solve(0)

	if !bytes.Equal(puzzle, sol1) {
		panic("solve and solve2 didn't agree")
	}

	copy(puzzle, orig)
	n := puzzle.Solve2(2)
	if n != 1 {
		panic("puzzle has two solutions")
	}

	t = time.Now()
	madeUp := BlankPuzzle()
	for i := 0; i < 10; i++ {
		madeUp = MakePuzzleUp()
		fmt.Print("made-up puzzle:\n", madeUp, "\n")
	}
	fmt.Println(time.Now().Sub(t).Seconds())
	n = madeUp.Solve2(2)
	if n != 1 {
		panic("crap, two solutions")
	}
	fmt.Println(madeUp.HintCount(), "hints")
	t = time.Now()
	madeUp.Solve2(1) // keeps solution around
	fmt.Println(time.Now().Sub(t).Seconds())
	fmt.Print("Its solution:\n", madeUp, "\n")
}

func main() {
	rnd = uint(time.Now().Nanosecond() | 1)
	fmt.Println(MakePuzzleUp())
}

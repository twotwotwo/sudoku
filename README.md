sudoku
======

Since of course the world needs yet another sudoku maker. Just running `sudoku` gives you a random 23-hint puzzle. In the source, you can:

* Convert puzzles into the internal format with `puz := NewPuzzle(string)`
* Make a blank puzzle with `puz := BlankPuzzle()`
* Do recursive brute-force solving with `hasSolution := puz.Solve(0)`
* Do faster and hopefully still correct brute-forcing, and learn how many possible solutions a puzzle has up to `n`, with `solutionCount := puz.Solve2(n)`
* Get a random solved puzzle (maybe starting from blank) with `puz.RandomSolution()`
* Get a 23-hint puzzle with `MakePuzzleUp()`

`Solve2`'s kinda weird. It keeps track of which digits can go in which squares (reasonable; people do it, too), but it also implements its own 'stack' of squares to look at rather than recursing properly (quite possibly useless, just the way the code came out). I tried implementing another approach that people use--figuring out, after each square is filled in, at which *positions* in a row/col/square might hold a given *digit*, rather than the other way around--but didn't immediately see gains from it, perhaps because my implementation repeated too much work after each guess, or because I wasn't using the information discovered by the new code as effectively as one could.

`MakePuzzleUp` just gets a random solved puzzle, then drops random digits from it and checks if the puzzle still has only one solution, retrying with a different digit dropped if it doesn't. It's slow to generate a puzzle with fewer than 23 hints. It could do better with a smarter search strategy, and/or a faster solver, and/or by keeping track of partly-reusable state that's currently discarded in between calls to Solve2.

Finally, I don't really do sudoku, so I don't really know if the puzzles it generates are "off" in some way that's obvious to a human that does them--too easy, too hard, or all following a similar pattern. If anyone stumbles on the repo (how?) and has anything to say about that, I'd love to hear.
package main

import (
    "fmt"
    ai "github.com/aorliche/web-nongrid-othello/ai"
)

func main() {
    board := ai.MakeTraditional(4)
    board.Premove(5, 0)
    board.Premove(6, 1)
    board.Premove(9, 1)
    board.Premove(10, 0)
    for i := 0; i < 10; i++ {
        fmt.Println(board.GetPossibleMoves())
        cand := board.GetCandidates()
        if len(cand) == 0 {
            break
        }
        board = cand[0]()
        fmt.Println(board.Eval(0), "-", board.Turn)
        board.PrintTraditional()
        board.PrintLines()
    }
}

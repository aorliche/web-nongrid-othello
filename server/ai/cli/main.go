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
    recvChan := make(chan bool)
    sendChans := make([]chan bool, 0)
    for i := 0; i < 2; i++ {
        sendChans = append(sendChans, make(chan bool))
        go ai.Loop(i, board, sendChans[i], recvChan, 20, 5000)
    }
    for {
        sendChans[board.Turn % 2] <- true
        <-recvChan
        fmt.Println(board.Eval(0), "-", board.Turn)
        board.PrintTraditional()
        if board.GameOver() {
            break
        }
    }
    fmt.Println("Game over!")
}

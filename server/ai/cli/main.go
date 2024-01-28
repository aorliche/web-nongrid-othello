package main

import (
    "fmt"
    ai "github.com/aorliche/web-nongrid-othello/ai"
)

func main() {
    board := ai.MakeTraditional(4)
    board.Points[5] = 0
    board.Points[6] = 1
    board.Points[9] = 1
    board.Points[10] = 0
    recvChan := make(chan bool)
    sendChans := make([]chan bool, 0)
    for i := 0; i < 2; i++ {
        sendChans = append(sendChans, make(chan bool))
        go ai.Loop(i, board, sendChans[i], recvChan, 20, 10000)
    }
    for {
        fmt.Println(board.Turn)
        sendChans[board.Turn % 2] <- true
        <-recvChan
        fmt.Println(board.Eval(0), board)
        if board.GameOver() {
            break
        }
    }
    fmt.Println("Game over!")
}

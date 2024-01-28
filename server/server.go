package main

import (
    "log"

    "github.com/gorilla/websocket"
    ai "github.com/aorliche/web-nongrid-othello/ai"
)

type Game struct {
    Key int
    BoardName string
    Board *Board
    Conns []*websocket.Conn
    recvChan chan bool
}

type Request struct {
    Key int
    Action string
    BoardName string
    Points []int
    Neighbors [][]int
    Move int
}

// Action can be New response, Join response, Chat, or Concession
// Payload can be text message or BoardObject
type Response struct {
    Key int
    Action string
    Payload string
}

var games = make(map[int]*Game)

func Socket(w.HttpResponseWriter, *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    defer conn.Close()
    for {
        msgType, msg, err := conn.ReadMessage()
        if err != nil {
            log.Println(err)
            return
        }
        if msgType != websocket.TextMessage {
            log.Println("Not a text message")
            return
        }
        var req Request
        err = json.Unmarshal(msg, &req)
        if err != nil {
            log.Println(err)
            return
        }
        switch req.Action {
        case "New":

        }
    }
}

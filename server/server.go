package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"

    "github.com/gorilla/websocket"
    ai "github.com/aorliche/web-nongrid-othello/ai"
)

type Game struct {
    Key int
    BoardName string
    BoardPlan string
    Board *ai.Board
    Conns []*websocket.Conn
    RecvChan chan bool
    AIGame bool
}

// Actions:
// ListBoards, LoadBoard, ListGames, NewGame, JoinGame, Move, Concede, Chat
// ListBoards: [none]
// LoadBoard: BoardName
// ListGames: [none]
// NewGame: AIGame, BoardName, Points, Neighbors (Points include initial starting points)
// JoinGame: Key
// Move: Key, Move
// Concede: Key
// Chat: Key, Text
type Request struct {
    Key int
    Action string
    BoardName string
    Points []int
    Neighbors [][]int
    Move int
    Text string
    AIGame bool
}

// Actions:
// ListBoards: BoardNames
// LoadBoard: BoardPlan
// ListGames: Keys
// NewGame: Key
// JoinGame: Key, BoardPlan, Points (neighbors implicit in BoardPlan)
// Move: Player, Move
// Concede: Player
// Chat: Player, Text
type Reply struct {
    Key int
    Player int
    Action string
    BoardPlan string
    Points []int
    BoardNames []string
    Keys []int
    Move int
}

var games = make(map[int]*Game)
var upgrader = websocket.Upgrader{} // Default options

func NextGameIdx() int {
    max := -1
    for key := range games {
        if key > max {
            max = key
        }
    }
    return max+1
}

func GetBoards() []string {
    boards := make([]string, 0)
    dir, err := os.Open("../boards")
    if err != nil {
        log.Println(err)
        return boards
    }
    files, err := dir.Readdir(0)
    if err != nil {
        log.Println(err)
        return boards
    }
    for _, v := range files {
        if v.IsDir() {
            continue
        }
        boards = append(boards, v.Name())
    }
    return boards
}

func GetBoard(name string) (string, error) {
    dat, err := os.ReadFile("../boards/" + name)
    if err != nil {
        log.Println(err)
        return "", err
    }
    return string(dat), err
}

func GameLoop(game *Game, recvChan chan bool, sendChan chan bool) {
    getLastMove := func(prev *ai.Board, cur *ai.Board) int {
        for i := 0; i < len(prev.Points); i++ {
            if prev.Points[i] == -1 && cur.Points[i] != -1 {
                return i
            }
        }
        return -1
    }
    board := game.Board
    for {
        prev := board.Clone()
        sendChan <- true
        if board.GameOver() {
            log.Println("game over")
            log.Println(board.GetScores())
            break
        }
        // Pinged by user or computer move
        // Now board should have benn updated
        keepPlaying := <- recvChan
        if !keepPlaying {
            sendChan <- false
            break
        }
        move := getLastMove(prev, board)
        player := prev.Turn % 2
        reply := Reply{Action: "Move", Player: player, Move: move}
        jsn, _ := json.Marshal(reply)
        err := game.Conns[0].WriteMessage(websocket.TextMessage, jsn)
        if err != nil {
            log.Println(err)
            continue
        }
    }
}


func Socket(w http.ResponseWriter, r *http.Request) {
    var player int
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
        // List boards
        case "ListBoards":
            boards := GetBoards()
            reply := Reply{Action: "ListBoards", BoardNames: boards}
            jsn, _ := json.Marshal(reply)
            err = conn.WriteMessage(websocket.TextMessage, jsn)
            if err != nil {
                log.Println(err)
                continue
            }
        // Load a board plan
        case "LoadBoard":
            name := req.BoardName
            board, err := GetBoard(name)
            if err != nil {
                log.Println(err)
                continue
            }
            reply := Reply{Action: "LoadBoard", BoardPlan: board}
            jsn, _ := json.Marshal(reply)
            err = conn.WriteMessage(websocket.TextMessage, jsn)
            if err != nil {
                log.Println(err)
                continue
            }
        // List available non-AI games
        case "ListGames":
            keys := make([]int, 0)
            for key,game := range games {
                // Check if game has not been joined by two players
                // and is not an AI game
                if len(game.Conns) < 2 && !game.AIGame {
                    keys = append(keys, key)
                }
            }
            reply := Reply{Action: "ListGames", Keys: keys}
            jsn, _ := json.Marshal(reply)
            err = conn.WriteMessage(websocket.TextMessage, jsn)
            if err != nil {
                log.Println(err)
                continue
            }
        // Start a new 2-player or AI game
        case "NewGame":
            player = 0
            key := NextGameIdx()
            aiGame := req.AIGame
            name := req.BoardName
            points := req.Points
            neighbors := req.Neighbors
            board := &ai.Board{
                Points: points,
                Neighbors: neighbors,
                Turn: 0,
            }
            plan, err := GetBoard(name)
            if err != nil {
                log.Println(err)
                continue
            }
            conns := make([]*websocket.Conn, 1)
            conns[0] = conn
            // Send and recv channels are from GameLoop's perspective
            recvChan := make(chan bool)
            game := &Game{Key: key, BoardName: name, BoardPlan: plan, Board: board, Conns: conns, RecvChan: recvChan, AIGame: aiGame}
            games[key] = game
            if aiGame {
                sendChan := make(chan bool)
                go ai.Loop(1, game.Board, sendChan, recvChan, 10, 2000)
                go GameLoop(game, recvChan, sendChan)
            }
            reply := Reply{Action: "NewGame", Key: key}
            jsn, _ := json.Marshal(reply)
            err = conn.WriteMessage(websocket.TextMessage, jsn)
            if err != nil {
                log.Println(err)
                continue
            }
        case "JoinGame":
            player = 1
            key := req.Key
            game := games[key]
            if game == nil {
                log.Println("Game not found")
                continue
            }
            game.Conns = append(game.Conns, conn)
            reply := Reply{Action: "JoinGame", Key: key, BoardPlan: game.BoardPlan, Points: game.Board.Points}
            jsn, _ := json.Marshal(reply)
            err := conn.WriteMessage(websocket.TextMessage, jsn)
            if err != nil {
                log.Println(err)
                continue
            }
        // Move
        case "Move":
            key := req.Key
            move := req.Move
            game := games[key]
            // Check if the move is legal and make the move
            if !game.Board.MoveIsLegal(move) {
                log.Println("Illegal move")
                continue
            }
            game.Board.MakeMove(move)
            reply := Reply{Action: "Move", Player: player, Move: move}
            jsn, _ := json.Marshal(reply)
            err := game.Conns[0].WriteMessage(websocket.TextMessage, jsn)
            if err != nil {
                log.Println(err)
                continue
            }
            if game.AIGame {
                game.RecvChan <- true
            } else if len(game.Conns) == 2 {
                err = game.Conns[1].WriteMessage(websocket.TextMessage, jsn)
                if err != nil {
                    log.Println(err)
                    continue
                }
            }
        // Concede
        case "Concede":
            key := req.Key
            game := games[key]
            reply := Reply{Action: "Concede", Player: player}
            jsn, _ := json.Marshal(reply)
            err := game.Conns[0].WriteMessage(websocket.TextMessage, jsn)
            if err != nil {
                log.Println(err)
                continue
            }
            if game.AIGame {
                game.RecvChan <- false
            } else if len(game.Conns) == 2 {
                err = game.Conns[1].WriteMessage(websocket.TextMessage, jsn)
                if err != nil {
                    log.Println(err)
                    continue
                }
            }
        }
    }
}

type HFunc func(http.ResponseWriter, *http.Request)

func Headers(fn HFunc) HFunc {
    return func (w http.ResponseWriter, req *http.Request) {
        //log.Println(req.Method)
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers",
            "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
        fn(w, req)
    }
}

func ServeStatic(w http.ResponseWriter, req *http.Request, file string) {
    w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
    http.ServeFile(w, req, file)
}

func ServeLocalFiles(dirs []string) {
    for _, dirName := range dirs {
        fsDir := "../static/" + dirName
        dir, err := os.Open(fsDir)
        if err != nil {
            log.Fatal(err)
        }
        files, err := dir.Readdir(0)
        if err != nil {
            log.Fatal(err)
        }
        for _, v := range files {
            //log.Println(v.Name(), v.IsDir())
            if v.IsDir() {
                continue
            }
            reqFile := dirName + "/" + v.Name()
            file := fsDir + "/" + v.Name()
            http.HandleFunc(reqFile, Headers(func (w http.ResponseWriter, req *http.Request) {ServeStatic(w, req, file)}))
        }
    }
}

func main() {
    log.SetFlags(0)
    ServeLocalFiles([]string{"", "/js", "/css"})
    http.HandleFunc("/ws", Socket)
    log.Fatal(http.ListenAndServe(":8003", nil))
}

package main

import (
    "encoding/json"
    //"fmt"
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
    GameOver bool
}

// Actions:
// ListBoards, LoadBoard, ListGames, NewGame, JoinGame, Move, Concede, Chat
// ListBoards: [none]
// LoadBoard: BoardName
// ListGames: [none]
// NewGame: AIGame, BoardName, Points, Neighbors
// JoinGame: Key
// Move: Key, Move
// Concede: Key
// Chat: Key, Text

type Request struct {
    Key int
    Action string
    BoardName string
    Points []ai.Point
    Neighbors [][]int
    Move int
    Text string
    AIGame bool
}

// Actions:
// ListBoards: BoardNames
// LoadBoard: BoardPlan
// ListGames: Keys
// NewGame: Key, Points, LevalMoves, GameOver
// JoinGame: Key, BoardPlan, Points, LegalMoves, GameOver
// Move: Player, Points, LegalMoves, GameOver
// Concede: Player, GameOver
// Chat: Player, Text
type Reply struct {
    Key int
    Player int
    Action string
    BoardPlan string
    Points []ai.Point
    BoardNames []string
    Keys []int
    LegalMoves []int
    GameOver bool
    Text string
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
        // Now board should have been updated
        keepPlaying := <- recvChan
        if !keepPlaying {
            sendChan <- false
            break
        }
        player := prev.Turn % 2
        moves := board.GetPossibleMoves()
        game.GameOver = len(moves) == 0
        reply := Reply{Action: "Move", Player: player, Points: board.Points, LegalMoves: moves, GameOver: game.GameOver}
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
                if len(game.Conns) < 2 && !game.AIGame && !game.GameOver {
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
            ns := req.Neighbors
            lines := ai.PointsToLinesGood(points, ns)
            lines = ai.CullShortLines(lines)
            // This and the parameter in continue lines has the potential 
            // to hang line generation
            for i := 0; i < 2; i++ {
                lines = ai.CullEqualLines(lines)
                lines = ai.CullSubsetLines(lines)
                lines = ai.CombineLines(lines)
            }
            lines = ai.CullEqualLines(lines)
            lines = ai.CullSubsetLines(lines)
            board := &ai.Board{
                Points: points,
                Lines: lines,
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
            moves := board.GetPossibleMoves()
            game.GameOver = len(moves) == 0
            reply := Reply{Action: "NewGame", Key: key, Points: board.Points, LegalMoves: moves, GameOver: game.GameOver}
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
            if game.GameOver {
                log.Println("Game is over")
                continue
            }
            game.Conns = append(game.Conns, conn)
            moves := game.Board.GetPossibleMoves()
            gameOver := len(moves) == 0
            if game.Board.Turn == 0 {
                moves = make([]int, 0)
            }
            reply := Reply{Action: "JoinGame", Key: key, BoardPlan: game.BoardPlan, Points: game.Board.Points, LegalMoves: moves, GameOver: gameOver}
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
            if game.GameOver {
                log.Println("Game is over")
                continue
            }
            // Check if the move is legal and make the move
            if !game.Board.MoveIsLegal(move) {
                log.Println("Illegal move")
                continue
            }
            game.Board.MakeMove(move)
            moves := game.Board.GetPossibleMoves()
            game.GameOver = len(moves) == 0
            reply := Reply{Action: "Move", Player: player, Points: game.Board.Points, LegalMoves: make([]int, 0), GameOver: game.GameOver}
            jsn, _ := json.Marshal(reply)
            err := game.Conns[player].WriteMessage(websocket.TextMessage, jsn)
            if err != nil {
                log.Println(err)
                continue
            }
            if game.AIGame {
                if game.GameOver {
                    game.RecvChan <- false
                } else {
                    game.RecvChan <- true
                }
            } else if len(game.Conns) == 2 {
                reply = Reply{Action: "Move", Player: player, Points: game.Board.Points, LegalMoves: moves, GameOver: game.GameOver}
                jsn, _ = json.Marshal(reply)
                err = game.Conns[1-player].WriteMessage(websocket.TextMessage, jsn)
                if err != nil {
                    log.Println(err)
                    continue
                }
            }
        // Concede
        case "Concede":
            key := req.Key
            game := games[key]
            game.GameOver = true
            reply := Reply{Action: "Concede", Player: player, GameOver: true}
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
        // Chat
        case "Chat":
            key := req.Key
            text := req.Text
            game := games[key]
            reply := Reply{Action: "Chat", Player: player, Text: text}
            jsn, _ := json.Marshal(reply)
            err := game.Conns[0].WriteMessage(websocket.TextMessage, jsn)
            if err != nil {
                log.Println(err)
                continue
            }
            if len(game.Conns) == 2 {
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

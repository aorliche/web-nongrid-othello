package ai

import (
    //"fmt"
    "math"
    //"sort"
    "time"
)

func Loop(me int, board *Board, inChan chan bool, outChan chan bool, depth int, timeMillis int) {
    for { 
        state := <- inChan 
        // false state indicates player disconnect
        // concession or two passes
        if !state || board.GameOver() {
            break
        }
        next := Search(board.Clone(), me, depth, timeMillis)
        if next == nil {
            time.Sleep(100 * time.Millisecond)
            continue
        }
        *board = *next
        outChan <- true
    }
}

// Set up iterative deepening
func Search(board *Board, me int, depth int, timeMillis int) *Board {
    startTime := time.Now()
    var res *Board
    for d := 1; d < depth; d++ {
        _, fn, fin, _ := SearchDeepAlphaBeta(board.Clone(), me, d, math.Inf(-1), math.Inf(1), true, startTime, timeMillis)
        if fn != nil && fin {
            res = fn()
        } else {
            break;
        }
    }
    return res
}

func max(a float64, b float64) float64 {
    if a > b {
        return a
    }
    return b
}

func min(a float64, b float64) float64 {
    if a < b {
        return a
    }
    return b
}

func SearchDeepAlphaBeta(board *Board, me int, depth int, alpha float64, beta float64, maxNotMin bool, startTime time.Time, timeMillis int) (*Board, func()*Board, bool, float64) {
    if depth == 0 {
        return board, nil, true, board.Eval(me)
    }
    if time.Since(startTime).Milliseconds() > int64(timeMillis) {
        return nil, nil, false, 0
    }
    fns := board.GetCandidates()
    if len(fns) == 0 {
        /*var val float64
        if maxNotMin {
            val = math.Inf(-1)
        } else {
            val = math.Inf(1)
        }
        return nil, nil, true, val*/
        return nil, nil, true, board.Eval(me)
    }
    var v float64
    var resBoard *Board
    var resFn func()*Board
    if maxNotMin {
        v = math.Inf(-1)
    } else {
        v = math.Inf(1)
    }
    for _,fn := range fns {
        next := fn()
        if next.GameOver() {
            return next, fn, true, next.Eval(me)
        }
        n, _, fin, val := SearchDeepAlphaBeta(next, me, depth-1, alpha, beta, !maxNotMin, startTime, timeMillis) 
        if !fin {
            return nil, nil, false, 0
        }
        if maxNotMin {
            if val > v {
                v = val
                resBoard = n
                resFn = fn
                alpha = max(alpha, v)
            }
            if v >= beta {
                return resBoard, resFn, true, v
            }
        } else {
            if val < v {
                v = val
                resBoard = n
                resFn = fn
                beta = min(beta, v)
            }
            if v <= alpha {
                return resBoard, resFn, true, v
            }
        }
    }
    return resBoard, resFn, true, v
}

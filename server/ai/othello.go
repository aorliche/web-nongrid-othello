package ai

import (
    //"fmt"
)

type Board struct {
    Points []int
    Neighbors [][]int
    Turn int
}

func Includes[T comparable](s []T, a T) bool {
    for _, v := range s {
        if v == a {
            return true
        }
    }
    return false
}

func Equals[T comparable](s1 []T, s2 []T) bool {
    if len(s1) != len(s2) {
        return false
    }
    for i := range s1 {
        if s1[i] != s2[i] {
            return false
        }
    }
    return true
}

func Reverse[T any](s []T) {
    for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
        s[i], s[j] = s[j], s[i]
    }
}

func MakeTraditional(n int) *Board {
    rc2p := func(r, c int) int {
        return r * n + c
    }
    points := make([]int, n * n)
    neighbors := make([][]int, n * n)
    for r := 0; r < n; r++ {
        for c := 0; c < n; c++ {
            points[rc2p(r, c)] = -1
            ns := []int{}
            if r > 0 {
                ns = append(ns, rc2p(r - 1, c))
            }
            if r < n - 1 {
                ns = append(ns, rc2p(r + 1, c))
            }
            if c > 0 {
                ns = append(ns, rc2p(r, c - 1))
            }
            if c < n - 1 {
                ns = append(ns, rc2p(r, c + 1))
            }
            if r > 0 && c > 0 {
                ns = append(ns, rc2p(r-1, c-1))
            }
            if r > 0 && c < n - 1 {
                ns = append(ns, rc2p(r-1, c+1))
            }
            if r < n - 1 && c > 0 {
                ns = append(ns, rc2p(r+1, c-1))
            }
            if r < n - 1 && c < n - 1 {
                ns = append(ns, rc2p(r+1, c+1))
            }
            neighbors[rc2p(r, c)] = ns
        }
    }
    return &Board{
        Points: points,
        Neighbors: neighbors,
        Turn: 0,
    }
}

func (board *Board) Clone() *Board {
    points := make([]int, len(board.Points))
    copy(points, board.Points)
    return &Board{
        Points: points,
        Neighbors: board.Neighbors,
        Turn: board.Turn,
    }
}

func (board *Board) GetShortestPaths(p1 int, p2 int) [][]int {
    type Node struct {
        Prev *Node
        Cur int
    }
    node2path := func(n *Node, path *[]int) {
        for n.Prev != nil {
            *path = append(*path, n.Cur)
            n = n.Prev
        }
        *path = append(*path, n.Cur)
        Reverse(*path)
    }
    visited := make([]bool, len(board.Points))
    visited[p1] = true
    start := Node{nil, p1}
    frontier := []*Node{&start}
    for len(frontier) > 0 {
        next := []*Node{}
        nextVisited := []int{}
        finished := false
        for _,node := range frontier {
            ns := board.Neighbors[node.Cur]
            for _,p := range ns {
                if visited[p] {
                    continue
                }
                nextVisited = append(nextVisited, p)
                next = append(next, &Node{node, p})
                if p == p2 {
                    finished = true
                }
            }
        }
        if finished {
            paths := [][]int{}
            for _,node := range next {
                if node.Cur == p2 {
                    path := []int{}
                    node2path(node, &path)
                    paths = append(paths, path)
                }
            }
            return paths
        }
        frontier = next
        for _,p := range nextVisited {
            visited[p] = true
        }
    }
    return [][]int{}
}

// Turn determines player
// Candidates are empty spaces next to other player's pieces
func (board *Board) GetPossibleMoves() [][2]int {
    me := board.Turn % 2
    other := 1-me 
    from := []int{}
    to := []int{}
    for p,player := range board.Points {
        if player == me {
            from = append(from, p)
        } else if player == -1 {
            for _,np := range board.Neighbors[p] {
                if board.Points[np] == other {
                    to = append(to, p)
                    break
                }
            }
        }
    }
    moves := [][2]int{}
    for _,toP := range to {
        for _,fromP := range from {
            paths := board.GetShortestPaths(fromP, toP)
            // Only those paths with all the other player's pieces are allowed
            // Other than starting and ending points
            // Also they must have length > 3
            valid := false
            nextpath:
            for _,path := range paths {
                if len(path) < 3 {
                    continue
                }
                for i := 1; i < len(path) - 1; i++ {
                    if board.Points[path[i]] != other {
                        continue nextpath
                    }
                }
                valid = true
                break
            }
            if valid {
                moves = append(moves, [2]int{fromP, toP})
            }
        }
    }
    return moves
}

func (board *Board) GameOver() bool {
    return len(board.GetPossibleMoves()) == 0
}

// Count number of pieces of a player
func (board *Board) Eval(me int) float64 {
    sum := 0
    for _,p := range board.Points {
        if p == me {
            sum += 1
        } else if p == 1-me {
            sum -= 1
        }
    }
    return float64(sum)
}

func (board *Board) GetScores() [2]int {
    scores := [2]int{}
    for _,p := range board.Points {
        if p == 0 {
            scores[0] += 1
        } else if p == 1 {
            scores[1] += 1
        }
    }
    return scores
}

func (board *Board) MoveIsLegal(to int) bool {
    moves := board.GetPossibleMoves()
    for _,move := range moves {
        if move[1] == to {
            return true
        }
    }
    return false
}

func (board *Board) MakeMove(to int) {
    moves := board.GetPossibleMoves()
    for _,move := range moves {
        if move[1] != to {
            continue
        }
        paths := board.GetShortestPaths(move[0], move[1])
        for _,path := range paths {
            for i := 1; i < len(path) - 1; i++ {
                board.Points[path[i]] = board.Points[move[0]]
            }
        }
    }
}

func (board *Board) GetCandidates(me int) []func() *Board {
    cand := make([]func() *Board, 0)
    if board.Turn % 2 != me {
        return cand
    }
    moves := board.GetPossibleMoves()
    for _,move := range moves {
        b := board.Clone()
        b.Turn += 1
        paths := b.GetShortestPaths(move[0], move[1])
        for _,path := range paths {
            for i := 1; i < len(path) - 1; i++ {
                b.Points[path[i]] = me
            }
        }
        b.Points[move[1]] = me
        cand = append(cand, func() *Board {
            return b
        })
    }
    return cand
}

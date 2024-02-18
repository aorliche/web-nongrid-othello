package ai

import (
    //"errors"
    "fmt"
    "math"
    "sort"
)

type Point struct {
    X float64
    Y float64
    Id int
    Player int
}

type Triangle struct {
    Ids [3]int
}

type Line struct {
    M float64
    Ids []int
}

const (
    NodeLine int = iota
    NodeTriangle 
)

type BoardError string

func (e BoardError) Error() string {
    return string(e)
}

type Node struct {
    Type int
    Neighbors []int
    LineId int
    TriangleId int
    TrianglePointId int
}

type Board struct {
    Points []Point
    Lines []Line
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

func ApproxEq(a float64, b float64) bool {
    // Infinities
    if a > 1000 && b > 1000 {
        return true
    }
    if a < -1000 && b < -1000 {
        return true
    }
    return a - b < 0.001 && a - b > -0.001
}

func Slope(p1 Point, p2 Point) float64 {
    return (p2.Y - p1.Y) / (p2.X - p1.X)
}
    
func (line Line) Includes(p Point) bool {
    for _,pId := range line.Ids {
        if pId == p.Id {
            return true
        }
    }
    return false
}

func Distance(p1 Point, p2 Point) float64 {
    dx := p1.X - p2.X
    dy := p1.Y - p2.Y
    return math.Sqrt(dx*dx + dy*dy)
}

func ContinueLines(points []Point, neighbors [][]int, line []int, lines *[]Line) {
    found := false
    if len(line) < 11 {
        for i := 0; i < len(points); i++ {
            if Includes(line, i) {
                continue
            }
            p := points[i]
            // Continue from start
            if Includes(neighbors[i], line[0]) {
                p0 := points[line[0]]
                p1 := points[line[1]]
                v1 := Point{p0.X - p1.X, p0.Y - p1.Y,-1,-1}
                v2 := Point{p.X - p0.X, p.Y - p0.Y,-1,-1}
                t1 := math.Atan2(v1.Y, v1.X)
                t2 := math.Atan2(v2.Y, v2.X)    
                td := t1 - t2
                if td < math.Pi/4-0.01 && td > -math.Pi/4+0.01 {
                    next := make([]int, len(line))  
                    copy(next, line)
                    Reverse(next)
                    next = append(next, i)
                    ContinueLines(points, neighbors, next, lines)
                    found = true
                }
            // Continue from end
            } else if Includes(neighbors[i], line[len(line)-1]) {
                p0 := points[line[len(line)-1]]
                p1 := points[line[len(line)-2]]
                v1 := Point{p0.X - p1.X, p0.Y - p1.Y,-1,-1}
                v2 := Point{p.X - p0.X, p.Y - p0.Y,-1,-1}
                t1 := math.Atan2(v1.Y, v1.X)
                t2 := math.Atan2(v2.Y, v2.X)
                td := t1 - t2
                if td < math.Pi/4-0.01 && td > -math.Pi/4+0.01 {
                    next := make([]int, len(line))  
                    copy(next, line)
                    next = append(next, i)
                    ContinueLines(points, neighbors, next, lines)
                    found = true
                }
            }
        }
    }
    if !found {
        // Check if lines already contains our line
        for _, ll := range *lines {
            if Equals(ll.Ids, line) {
                return
            }
            Reverse(line)
            if Equals(ll.Ids, line) {
                return
            }
        }
        *lines = append(*lines, Line{0, line})
    }
}

func PointsToLinesGood(points []Point, neighbors [][]int) []Line {
    lines := make([]Line, 0)
    for i := 0; i < len(points); i++ {
        for _,j := range neighbors[i] {
            ContinueLines(points, neighbors, []int{i, j}, &lines)
        }
    }
    return lines
}

func PointsToLines(points []Point, d float64) []Line {
    lines := make([]Line, 0)
    for i := 0; i < len(points); i++ {
        p1 := points[i]
        for j := i+1; j < len(points); j++ {
            p2 := points[j]
            m := Slope(p1, p2)
            // Find existing line
            found := false
            for k,line := range lines {
                if !ApproxEq(line.M, m) {
                    continue
                }
                if line.Includes(p1) && line.Includes(p2) {
                    found = true
                    break
                }
                if !line.Includes(p1) {
                    if ApproxEq(Slope(points[line.Ids[0]], p1), m) {
                        if d != 0 && !(ApproxEq(Distance(points[line.Ids[0]], p1), d) || ApproxEq(Distance(points[line.Ids[len(line.Ids)-1]], p1), d)) {
                            continue
                        }
                        lines[k].Ids = append(lines[k].Ids, p1.Id)
                        found = true
                    }
                } 
                if !line.Includes(p2) {
                    if ApproxEq(Slope(points[line.Ids[0]], p2), m) {
                        if d != 0 && !(ApproxEq(Distance(points[line.Ids[0]], p2), d) || ApproxEq(Distance(points[line.Ids[len(line.Ids)-1]], p2), d)) {
                            continue
                        }
                        lines[k].Ids = append(lines[k].Ids, p2.Id)
                        found = true
                    }
                }
                if found {
                    break
                }
            }
            // New line
            if !found {
                if d == 0 || ApproxEq(Distance(p1, p2), d) {
                    line := Line{m, []int{p1.Id, p2.Id}}
                    lines = append(lines, line)
                }
            }
        }
    }
    // Sort points in lines
    for _,line := range lines {
        sort.Slice(line.Ids, func(i, j int) bool {
            dx := points[line.Ids[i]].X - points[line.Ids[j]].X
            if !ApproxEq(dx, 0) {
                return dx < 0
            }
            dy := points[line.Ids[i]].Y - points[line.Ids[j]].Y
            return dy < 0
        })
    }
    return lines
}

func IsSubset(small, big []int) bool {
    yes := true
    for _,i := range small {
        if !Includes(big, i) {
            yes = false
            break
        }
    }
    return yes
}

func CullSubsetLines(lines []Line) []Line {
    keep := make([]Line, 0)
    for i := 0; i < len(lines); i++ {
        found := false
        for j := 0; j < len(lines); j++ {
            if len(lines[i].Ids) < len(lines[j].Ids) && IsSubset(lines[i].Ids, lines[j].Ids) {
                found = true
            }
        }
        if !found {
            keep = append(keep, lines[i])
        }
    }
    return keep
}

// Cull lines with only two points
func CullShortLines(lines []Line) []Line {
    keep := make([]Line, 0)
    for _,line := range lines {
        if len(line.Ids) > 2 {
            keep = append(keep, line)
        }
    }
    return keep
}

func CullLinesByNeighbors(lines []Line, neighbors [][]int) []Line {
    keep := make([]Line, 0)
    for _,line := range lines {
        k := true
        for i := 0; i < len(line.Ids)-1; i++ {
            if !Includes(neighbors[line.Ids[i]], line.Ids[i+1]) {
                k = false
                break
            }
        }
        if k {
            keep = append(keep, line)
        }
    }
    return keep
}

func MakeTraditional(n int) *Board {
    points := make([]Point, n * n)
    for r := 0; r < n; r++ {
        for c := 0; c < n; c++ {
            id := r * n + c
            p := Point{float64(r), float64(c), id, -1}
            points[id] = p
        }
    }
    lines := PointsToLines(points, 0)
    // Cull lines with d > sqrt(2)+delta
    keep := make([]Line, 0)
    for _,line := range lines {
        d := Distance(points[line.Ids[0]], points[line.Ids[1]])
        if d < math.Sqrt(2)+0.1 {
            keep = append(keep, line)
        }
    }
    // Cull lines with only two points
    keep = CullShortLines(keep)
    return &Board{
        Points: points,
        Lines: keep,
        Turn: 0,
    }
}

func (board *Board) Clone() *Board {
    points := make([]Point, len(board.Points))
    copy(points, board.Points)
    b := &Board{
        Points: points,
        Lines: board.Lines,
        Turn: board.Turn,
    }
    return b
}
    
func (board *Board) CaptureBackwards(ids []int, i int, me int, capture bool) bool {
    if i < 0 {
        return false
    }
    if capture {
        board.Points[ids[i+1]].Player = me
    }
    if board.Points[ids[i]].Player != 1-me {
        return false
    }
    for ii := i; ii >= 0; ii-- {
        if board.Points[ids[ii]].Player == -1 {
            return false
        }
        if board.Points[ids[ii]].Player == me {
            return true
        }
        if capture {
            board.Points[ids[ii]].Player = me
        }
    }
    return false
}
    
func (board *Board) CaptureForwards(ids []int, i int, me int, capture bool) bool {
    if i >= len(ids) {
        return false
    }
    if capture {
        board.Points[ids[i-1]].Player = me
    }
    if board.Points[ids[i]].Player != 1-me {
        return false
    }
    for ii := i; ii < len(ids); ii++ {
        if board.Points[ids[ii]].Player == -1 {
            return false
        }
        if board.Points[ids[ii]].Player == me {
            return true
        }
        if capture {
            board.Points[ids[ii]].Player = me
        }
    }
    return false
}

// Turn determines player
// Candidates are empty spaces next to other player's pieces
func (board *Board) GetPossibleMoves() []int {
    me := board.Turn % 2
    moves := []int{}
    for _,line := range board.Lines {
        for i,pId := range line.Ids {
            if board.Points[pId].Player != -1 {
                continue
            }
            if board.CaptureBackwards(line.Ids, i-1, me, false) || 
                board.CaptureForwards(line.Ids, i+1, me, false) {
                moves = append(moves, pId)
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
        if p.Player == me {
            sum += 1
        } else if p.Player == 1-me {
            sum -= 1
        }
    }
    return float64(sum)
}

func (board *Board) GetScores() [2]int {
    scores := [2]int{}
    for _,p := range board.Points {
        if p.Player == 0 {
            scores[0] += 1
        } else if p.Player == 1 {
            scores[1] += 1
        }
    }
    return scores
}

func (board *Board) MoveIsLegal(to int) bool {
    moves := board.GetPossibleMoves()
    for _,move := range moves {
        if move == to {
            return true
        }
    }
    return false
}

func (board *Board) Premove(to int, me int) {
    board.Points[to].Player = me
}

func (board *Board) MakeMove(to int) {
    me := board.Turn % 2
    // Now that lines have bifurcations we can longer check for legality and capture
    // in the same loop
    bwd := make([][2]int, 0)
    fwd := make([][2]int, 0)
    for j,line := range board.Lines {
        for i,pId := range line.Ids {
            if pId == to {
                if board.CaptureBackwards(line.Ids, i-1, me, false) {
                    bwd = append(bwd, [2]int{j, i})
                }
                if board.CaptureForwards(line.Ids, i+1, me, false) {
                    fwd = append(fwd, [2]int{j, i})
                }
            }
        }
    }
    for _,lp := range bwd {
        board.CaptureBackwards(board.Lines[lp[0]].Ids, lp[1]-1, me, true)
    }
    for _,lp := range fwd {
        board.CaptureForwards(board.Lines[lp[0]].Ids, lp[1]+1, me, true)
    }
    board.Turn += 1
}

func (board *Board) GetCandidates() []func() *Board {
    me := board.Turn % 2
    cand := make([]func() *Board, 0)
    if board.Turn % 2 != me {
        return cand
    }
    moves := board.GetPossibleMoves()
    for _,move := range moves {
        m := move
        cand = append(cand, func() *Board {
            b := board.Clone()
            b.MakeMove(m)
            return b
        })
    }
    return cand
}

func (board *Board) PrintTraditional() {
    n := int(math.Sqrt(float64(len(board.Points))))
    for r := 0; r < n; r++ {
        for c := 0; c < n; c++ {
            id := r * n + c
            p := board.Points[id]
            if p.Player == 0 {
                fmt.Print("X")
            } else if p.Player == 1 {
                fmt.Print("O")
            } else {
                fmt.Print(" ")
            }
        }
        fmt.Println()
    }
}

func (board *Board) PrintLines() {
    for _,line := range board.Lines {
        for _,pId := range line.Ids {
            fmt.Print(pId, ": ", board.Points[pId].Player, " ,")
        }
        fmt.Println()
    }
}

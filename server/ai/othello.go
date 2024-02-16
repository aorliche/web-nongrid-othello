package ai

import (
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

type Board struct {
    Points []Point
    Neighbors [][]int
    Triangles []Triangle
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

// Return index of keystone point in triangle or -1 if not continues
// A triangle joins two lines that have different keystone points
func (board *Board) TriangleContinuesLine(tId int, line Line) int {
    d := (1+math.Sqrt(3)/2)*Distance(board.Points[line.Ids[0]], board.Points[line.Ids[1]])
    id0 := line.Ids[0]
    ide := line.Ids[len(line.Ids)-1]
    tri := board.Triangles[tId]
    if id0 == tri.Ids[0] {
        p := board.Points[line.Ids[1]]
        p1 := board.Points[tri.Ids[1]]
        p2 := board.Points[tri.Ids[2]]
        pm := Point{(p1.X + p2.X) / 2, (p1.Y + p2.Y) / 2, -1, -1}
        if ApproxEq(Distance(pm, p), d) {
            return 0
        }
    } else if id0 == tri.Ids[1] {
        p := board.Points[line.Ids[1]]
        p1 := board.Points[tri.Ids[0]]
        p2 := board.Points[tri.Ids[2]]
        pm := Point{(p1.X + p2.X) / 2, (p1.Y + p2.Y) / 2, -1, -1}
        if ApproxEq(Distance(pm, p), d) {
            return 1
        }
    } else if id0 == tri.Ids[2] {
        p := board.Points[line.Ids[1]]
        p1 := board.Points[tri.Ids[0]]
        p2 := board.Points[tri.Ids[1]]
        pm := Point{(p1.X + p2.X) / 2, (p1.Y + p2.Y) / 2, -1, -1}
        if ApproxEq(Distance(pm, p), d) {
            return 2
        }
    } else if ide == tri.Ids[0] {
        p := board.Points[line.Ids[len(line.Ids)-2]]
        p1 := board.Points[tri.Ids[1]]
        p2 := board.Points[tri.Ids[2]]
        pm := Point{(p1.X + p2.X) / 2, (p1.Y + p2.Y) / 2, -1, -1}
        if ApproxEq(Distance(pm, p), d) {
            return 0
        }
    } else if ide == tri.Ids[1] {
        p := board.Points[line.Ids[len(line.Ids)-2]]
        p1 := board.Points[tri.Ids[0]]
        p2 := board.Points[tri.Ids[2]]
        pm := Point{(p1.X + p2.X) / 2, (p1.Y + p2.Y) / 2, -1, -1}
        if ApproxEq(Distance(pm, p), d) {
            return 1
        }
    } else if ide == tri.Ids[2] {
        p := board.Points[line.Ids[len(line.Ids)-2]]
        p1 := board.Points[tri.Ids[0]]
        p2 := board.Points[tri.Ids[1]]
        pm := Point{(p1.X + p2.X) / 2, (p1.Y + p2.Y) / 2, -1, -1}
        if ApproxEq(Distance(pm, p), d) {
            return 2
        }
    } 
    return -1
}

func (board *Board) CalculateTriangles() {
    inTriangles := func(ts []Triangle, p1 int, p2 int, p3 int) bool {
        for _,t := range ts {
            if Includes(t.Ids[:], p1) && Includes(t.Ids[:], p2) && Includes(t.Ids[:], p3) {
                return true
            }
        }
        return false
    }
    board.Triangles = make([]Triangle, 0)
    for p1,ns := range board.Neighbors {
        outer:
        for _,p2 := range ns {
            for _,p3a := range board.Neighbors[p2] {
                for _,p3b := range ns {
                    if p3a == p3b && !inTriangles(board.Triangles, p1, p2, p3a) {
                        board.Triangles = append(board.Triangles, Triangle{[3]int{p1, p2, p3a}})
                        continue outer
                    }
                }
            }
        }
    }
}

// Slope becomes incorrect
func (board *Board) ExtendLines() {
    board.CalculateTriangles()
    // tId, lineId, keypoint
    extendsTrips := make([][3]int, 0)
    extendsTris := make([]int, 0)
    for tId := range board.Triangles {
        for i,line := range board.Lines {
            j := board.TriangleContinuesLine(tId, line)
            if j != -1 {
                extendsTrips = append(extendsTrips, [3]int{tId, i, j})
                if !Includes(extendsTris, tId) {
                    extendsTris = append(extendsTris, tId)
                }
            }
        }
    }
    extendedLines := make([]Line, 0)
    linesExtended := make([]bool, len(board.Lines))
    for _,t := range extendsTris {
        lineIds := make([]int, 0)
        keypoints := make([]int, 0)
        for _,p := range extendsTrips {
            if p[0] == t {
                linesExtended[p[1]] = true
                lineIds = append(lineIds, p[1])
                keypoints = append(keypoints, p[2])
            }
        }
        for i := 0; i < len(lineIds); i++ {
            joined := false
            for j := i+1; j < len(lineIds); j++ {
               if keypoints[i] != keypoints[j] {
                   joined = true
                   // Get the proper joining for lines
                   l1 := board.Lines[lineIds[i]]
                   l2 := board.Lines[lineIds[j]]
                   id1 := board.Triangles[t].Ids[keypoints[i]]
                   id2 := board.Triangles[t].Ids[keypoints[j]]
                   nIds := make([]int, len(l1.Ids)+len(l2.Ids))
                   if l1.Ids[0] == id1 && l2.Ids[0] == id2 {
                       for k := 0; k < len(l1.Ids); k++ {
                           nIds[k] = l1.Ids[len(l1.Ids)-1-k]
                       }
                       for k := 0; k < len(l2.Ids); k++ {
                           nIds[len(l1.Ids)+k] = l2.Ids[k]
                       }
                   } else if l1.Ids[0] == id1 && l2.Ids[len(l2.Ids)-1] == id2 {
                       for k := 0; k < len(l1.Ids); k++ {
                           nIds[k] = l1.Ids[len(l1.Ids)-1-k]
                       }
                       for k := 0; k < len(l2.Ids); k++ {
                           nIds[len(l1.Ids)+k] = l2.Ids[len(l2.Ids)-1-k]
                       }
                   } else if l1.Ids[len(l1.Ids)-1] == id1 && l2.Ids[0] == id2 {
                       for k := 0; k < len(l1.Ids); k++ {
                           nIds[k] = l1.Ids[k]
                       }
                       for k := 0; k < len(l2.Ids); k++ {
                           nIds[len(l1.Ids)+k] = l2.Ids[k]
                       }
                   } else if l1.Ids[len(l1.Ids)-1] == id1 && l2.Ids[len(l2.Ids)-1] == id2 {
                       for k := 0; k < len(l1.Ids); k++ {
                           nIds[k] = l1.Ids[k]
                       }
                       for k := 0; k < len(l2.Ids); k++ {
                           nIds[len(l1.Ids)+k] = l2.Ids[len(l2.Ids)-1-k]
                       }
                   }
                   nl := Line{l1.M, nIds}
                   extendedLines = append(extendedLines, nl)
               }
            }
            // Add the other two triangle points if no joins
            // Need to duplicate the line
            if !joined {
                l1 := board.Lines[lineIds[i]]
                id1 := board.Triangles[t].Ids[keypoints[i]]
                nl1 := make([]int, 0)
                nl2 := make([]int, 0)
                var app1, app2 int
                switch keypoints[i] {
                    case 0: {
                        app1 = board.Triangles[t].Ids[1]
                        app2 = board.Triangles[t].Ids[2]
                    }
                    case 1: {
                        app1 = board.Triangles[t].Ids[0]
                        app2 = board.Triangles[t].Ids[2]
                    }
                    case 2: {
                        app1 = board.Triangles[t].Ids[0]
                        app2 = board.Triangles[t].Ids[1]
                    }
                }
                if l1.Ids[0] == id1 {
                    nl1 = append(nl1, app1)
                    nl2 = append(nl2, app2)
                    nl1 = append(nl1, l1.Ids...)
                    nl2 = append(nl2, l1.Ids...)
                } else {
                    nl1 = append(nl1, l1.Ids...)
                    nl2 = append(nl2, l1.Ids...)
                    nl1 = append(nl1, board.Triangles[t].Ids[1])
                    nl2 = append(nl2, board.Triangles[t].Ids[2])
                }
                extendedLines = append(extendedLines, Line{l1.M, nl1})
                extendedLines = append(extendedLines, Line{l1.M, nl2})
            }
        }
    }
    // Add in the non-extended lines
    for i,line := range board.Lines {
        if !linesExtended[i] {
            extendedLines = append(extendedLines, line)
        }
    }
    board.Lines = extendedLines
    board.CullShortLines()
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

func PointsToLines(points []Point) []Line {
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
                        lines[k].Ids = append(lines[k].Ids, p1.Id)
                        found = true
                    }
                } 
                if !line.Includes(p2) {
                    if ApproxEq(Slope(points[line.Ids[0]], p2), m) {
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
                line := Line{m, []int{p1.Id, p2.Id}}
                lines = append(lines, line)
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

// Cull lines with only two points
func (board *Board) CullShortLines() {
    keep := make([]Line, 0)
    for _,line := range board.Lines {
        if len(line.Ids) > 2 {
            keep = append(keep, line)
        }
    }
    board.Lines = keep
}

func (board *Board) CullLongIntervalLines(cutoff float64) {
    keep := make([]Line, 0)
    for _,line := range board.Lines {
        d := Distance(board.Points[line.Ids[0]], board.Points[line.Ids[1]])
        if d < cutoff {
            keep = append(keep, line)
        }
    }
    board.Lines = keep
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
    lines := PointsToLines(points)
    // Cull lines with d > sqrt(2)+delta
    keep := make([]Line, 0)
    for _,line := range lines {
        d := Distance(points[line.Ids[0]], points[line.Ids[1]])
        if d < math.Sqrt(2)+0.1 {
            keep = append(keep, line)
        }
    }
    b := &Board{
        Points: points,
        Lines: keep,
        Turn: 0,
    }
    // Cull length-2 lines
    b.CullShortLines()
    return b
}

// We ignore neighbors and triangles
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

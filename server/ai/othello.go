package ai

import (
    "errors"
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

type Node struct {
    Type int
    Neighbors []int
    LineId int
    TriangleId int
    TrianglePointId int
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

func GetPathsToNode(nodes []*Node, cur []int, end int, result *[][]int) {
    // No 3-triangle sections
    if len(cur) >= 3 {
        if nodes[cur[len(cur)-1]].Type == NodeTriangle && 
        nodes[cur[len(cur)-2]].Type == NodeTriangle &&
        nodes[cur[len(cur)-3]].Type == NodeTriangle {
            return
        }
    }
    // Reached the end
    if cur[len(cur)-1] == end {
        // Can't have all-triangle paths
        allTris := true
        for _, nid := range cur {
            if nodes[nid].Type != NodeTriangle {
                allTris = false
                break
            }
        }
        if allTris {
            return
        }
        *result = append(*result, cur)
        return
    }
    last := nodes[cur[len(cur)-1]]
    ns := last.Neighbors
    for _, id := range ns {
        // No loops
        if Includes(cur, id) {
            continue
        // Continue branching
        } 
        next := make([]int, len(cur)+1)
        copy(next, cur)
        next[len(cur)] = id
        GetPathsToNode(nodes, next, end, result)
    }
}

func GetGraphPaths(nodes []*Node) [][]int {
    result := [][]int{}
    // Get endpoints
    ends := make([]int, 0)
    for i, n := range nodes {
        if n.Type == NodeLine && len(n.Neighbors) == 1 {
            ends = append(ends, i)
        } else if n.Type == NodeLine && len(n.Neighbors) == 0 {
            result = append(result, []int{i})
        } else if n.Type == NodeTriangle && len(n.Neighbors) == 2 {
            ends = append(ends, i)
        }
    }
    // Make path between every pair of endpoints
    // May have multiple paths if graph contains loops
    // This should also fix a line is a loop problem?
    for i := 0; i < len(ends); i++ {
        for j := i+1; j < len(ends); j++ {
            paths := [][]int{}
            GetPathsToNode(nodes, []int{ends[i]}, ends[j], &paths)
            result = append(result, paths...)
        }
    }
    return result
}

func MakeGraph(points []Point, lines []Line, tris []Triangle) []*Node {
    nodes := []*Node{}
    for li := range lines {
        nodes = append(nodes, &Node{Type: NodeLine, LineId: li})
    }
    for ti, tri := range tris {
        nti := len(nodes)
        nodes = append(nodes, &Node{Type: NodeTriangle, TriangleId: ti, TrianglePointId: 0, Neighbors: []int{nti+1, nti+2}})
        nodes = append(nodes, &Node{Type: NodeTriangle, TriangleId: ti, TrianglePointId: 1, Neighbors: []int{nti, nti+2}})
        nodes = append(nodes, &Node{Type: NodeTriangle, TriangleId: ti, TrianglePointId: 2, Neighbors: []int{nti, nti+1}})
        for li, line := range lines {
            nli := -1
            for ni, node := range nodes {
                if node.LineId == li {
                    nli = ni
                    break
                }
            }
            tpi := TriangleContinuesLine(points, tri, line)
            if tpi != -1 {
                nodes[nti+tpi].Neighbors = append(nodes[nti+tpi].Neighbors, nli)
                nodes[nli].Neighbors = append(nodes[nli].Neighbors, nti+tpi)
            }
        }
    }
    return nodes
}

func PathsToLines(nodes []*Node, lines []Line, tris []Triangle, paths [][]int) ([]Line, error) {
    result := []Line{}
    for _,path := range paths {
        // Must be line
        if len(path) == 1 {
            if nodes[path[0]].Type != NodeLine {
                return nil, errors.New("Isolated node not a line")
            }
            result = append(result, lines[nodes[path[0]].LineId])
        // Error
        } else if len(path) == 0 {
            return nil, errors.New("Invalid path: empty")
        // Must be triangle, can ignore
        } else if len(path) == 2 {
            if nodes[path[0]].Type != NodeTriangle || nodes[path[1]].Type != NodeTriangle {
                return nil, errors.New("Invalid path: length 2 path not two triangle vertices")
            }
        // Lines and triangles
        } else {
            res := []int{}
            for _,nid := range path {
                n := nodes[nid]
                if n.Type == NodeLine {
                    if len(res) == 0 {
                        res = append(res, lines[n.LineId].Ids...)
                    } else {
                        line := lines[n.LineId].Ids
                        if res[0] == line[0] {
                            Reverse(res)
                            res = append(res, line[1:]...)
                        } else if res[len(res)-1] == line[0] {
                            res = append(res, line[1:]...)
                        } else if res[0] == line[len(line)-1] {
                            Reverse(res)
                            Reverse(line)
                            res = append(res, line[1:]...)
                        } else if res[len(res)-1] == line[len(line)-1] {
                            Reverse(line)
                            res = append(res, line[1:]...)
                        } else {
                            return nil, errors.New("Invalid path: line doesn't add properly")
                        }
                    }
                } else if n.Type == NodeTriangle {
                    tri := tris[n.TriangleId]
                    trip := tri.Ids[n.TrianglePointId]
                    if len(res) == 0 {
                        res = append(res, trip)
                    } else if !Includes(res, trip) {
                        res = append(res, trip)
                    } else if res[0] == trip {
                        Reverse(res)
                    } else if res[len(res)-1] == trip {
                        // Do nothing
                    } else {
                        return nil, errors.New("Invalid path: triangle in middle of line")
                    }
                }
            }
            result = append(result, Line{Ids: res})
        }
    }
    return result, nil
}

// Return index of keystone point in triangle or -1 if not continues
// A triangle joins two lines that have different keystone points
func TriangleContinuesLine(points []Point, tri Triangle, line Line) int {
    d := (1+math.Sqrt(3)/2)*Distance(points[line.Ids[0]], points[line.Ids[1]])
    id0 := line.Ids[0]
    ide := line.Ids[len(line.Ids)-1]
    if id0 == tri.Ids[0] {
        p := points[line.Ids[1]]
        p1 := points[tri.Ids[1]]
        p2 := points[tri.Ids[2]]
        pm := Point{(p1.X + p2.X) / 2, (p1.Y + p2.Y) / 2, -1, -1}
        if ApproxEq(Distance(pm, p), d) {
            return 0
        }
    } else if id0 == tri.Ids[1] {
        p := points[line.Ids[1]]
        p1 := points[tri.Ids[0]]
        p2 := points[tri.Ids[2]]
        pm := Point{(p1.X + p2.X) / 2, (p1.Y + p2.Y) / 2, -1, -1}
        if ApproxEq(Distance(pm, p), d) {
            return 1
        }
    } else if id0 == tri.Ids[2] {
        p := points[line.Ids[1]]
        p1 := points[tri.Ids[0]]
        p2 := points[tri.Ids[1]]
        pm := Point{(p1.X + p2.X) / 2, (p1.Y + p2.Y) / 2, -1, -1}
        if ApproxEq(Distance(pm, p), d) {
            return 2
        }
    } else if ide == tri.Ids[0] {
        p := points[line.Ids[len(line.Ids)-2]]
        p1 := points[tri.Ids[1]]
        p2 := points[tri.Ids[2]]
        pm := Point{(p1.X + p2.X) / 2, (p1.Y + p2.Y) / 2, -1, -1}
        if ApproxEq(Distance(pm, p), d) {
            return 0
        }
    } else if ide == tri.Ids[1] {
        p := points[line.Ids[len(line.Ids)-2]]
        p1 := points[tri.Ids[0]]
        p2 := points[tri.Ids[2]]
        pm := Point{(p1.X + p2.X) / 2, (p1.Y + p2.Y) / 2, -1, -1}
        if ApproxEq(Distance(pm, p), d) {
            return 1
        }
    } else if ide == tri.Ids[2] {
        p := points[line.Ids[len(line.Ids)-2]]
        p1 := points[tri.Ids[0]]
        p2 := points[tri.Ids[1]]
        pm := Point{(p1.X + p2.X) / 2, (p1.Y + p2.Y) / 2, -1, -1}
        if ApproxEq(Distance(pm, p), d) {
            return 2
        }
    } 
    return -1
}

func CalculateTriangles(neighbors [][]int) []Triangle {
    inTriangles := func(ts []Triangle, p1 int, p2 int, p3 int) bool {
        for _,t := range ts {
            if Includes(t.Ids[:], p1) && Includes(t.Ids[:], p2) && Includes(t.Ids[:], p3) {
                return true
            }
        }
        return false
    }
    tris := make([]Triangle, 0)
    for p1,ns := range neighbors {
        outer:
        for _,p2 := range ns {
            for _,p3a := range neighbors[p2] {
                for _,p3b := range ns {
                    if p3a == p3b && !inTriangles(tris, p1, p2, p3a) {
                        tris = append(tris, Triangle{[3]int{p1, p2, p3a}})
                        continue outer
                    }
                }
            }
        }
    }
    return tris
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

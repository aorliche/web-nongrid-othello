package ai

import (
    "fmt"
    "math"
    "testing"
)

func TestMakeTraditional(t *testing.T) {
    board := MakeTraditional(4)
    expect := [][]int{[]int{0,1,2,3}}
    got := board.Lines[:1]
    for i := range got {
        if !Equals(got[i].Ids, expect[i]) {
            t.Errorf("got %v, expect %v", got[i], expect[i])
        }
    }
    if len(board.Lines) != 14 {
        t.Errorf("got %v, expect %v", len(board.Lines), 14)
    }
}

func TestCaptureBackwards(t *testing.T) {
    board := MakeTraditional(4)
    board.Premove(0, 0)
    board.Premove(1, 1)
    board.Premove(4, 1)
    capt := board.CaptureBackwards(board.Lines[0].Ids, 1, 0, false)
    if capt != true {
        t.Errorf("got %v, expect %v", capt, true)
    }
}

func TestGetPossibleMoves(t *testing.T) {
    board := MakeTraditional(4)
    board.Premove(0, 0)
    board.Premove(1, 1)
    board.Premove(4, 1)
    got := board.GetPossibleMoves()
    expect := []int{2,8}
    if !Equals(got, expect) {
        t.Errorf("got %v, expect %v", got, expect)
    }
    board = MakeTraditional(4)
    board.Premove(5, 0)
    board.Premove(6, 1)
    board.Premove(9, 1)
    board.Premove(10, 0)
    got = board.GetPossibleMoves()
    expect = []int{2,7,8,13}
    if len(expect) != len(got) {
        t.Errorf("got %v, expect %v", got, expect)
    }
    for _,e := range expect {
        if !Includes(got, e) {
            t.Errorf("got %v, expect %v", got, expect)
            break
        }
    }
    board = MakeTraditional(4)
    board.Premove(5, 0)
    board.Premove(10, 0)
    board.Premove(15, 1)
    board.Turn = 1
    got = board.GetPossibleMoves()
    expect = []int{0}
    if !Equals(got, expect) {
        t.Errorf("got %v, expect %v", got, expect)
    }
}

func TestGetCandidates(t *testing.T) {
    board := MakeTraditional(4)
    board.Premove(5, 0)
    board.Premove(6, 1)
    board.Premove(9, 1)
    board.Premove(10, 0)
    cand := board.GetCandidates()
    next := cand[0]()
    if next.Turn != 1 {
        t.Errorf("got %v, expect %v", next.Turn, 1)
    }
}

/*func TestGetCandidates2(t *testing.T) {
    board := MakeTraditional(4)
    board.Premove(4, 1)
    board.Premove(5, 1)
    board.Premove(2, 0)
    board.Premove(6, 0)
    board.Premove(8, 0)
    board.Premove(9, 0)
    board.Premove(10, 0)
    board.Turn = 1
    fmt.Println(board.GetPossibleMoves())
    cand := board.GetCandidates()
    for _,c := range cand {
        next := c()
        next.PrintTraditional()
    }
}*/

func TestCalculateTriangles(t *testing.T) {
    ns := [][]int{
        []int{1,2},
        []int{0,2},
        []int{0,1},
    }
    tris := CalculateTriangles(ns)
    if len(tris) != 1 {
        t.Errorf("got %v, expect %v", len(tris), 1)
    }
    ids := tris[0].Ids
    if !Includes(ids[:], 0) || !Includes(ids[:], 1) || !Includes(ids[:], 2) {
        t.Errorf("got %v, expect %v", ids, []int{0,1,2})
    }
}

func TestCalculateTriangles2(t *testing.T) {
    ns := [][]int{
        []int{1,2,3},
        []int{0,2,4},
        []int{0,1},
        []int{0,5},
        []int{1},
        []int{6,7},
        []int{5,7},
        []int{5,6},
    }
    tris := CalculateTriangles(ns)
    if len(tris) != 2 {
        t.Errorf("got %v, expect %v", len(tris), 2)
    }
    ids0 := tris[0].Ids
    ids1 := tris[1].Ids
    if !Includes(ids0[:], 0) || !Includes(ids0[:], 1) || !Includes(ids0[:], 2) {
        t.Errorf("got %v, expect %v", ids0, []int{0,1,2})
    }
    if !Includes(ids1[:], 5) || !Includes(ids1[:], 6) || !Includes(ids1[:], 7) {
        t.Errorf("got %v, expect %v", ids1, []int{5,6,7})
    }
}

func TestCalculateTriangles3(t *testing.T) {
    ns := [][]int{
        []int{1,2,3},
        []int{0,2},
        []int{0,1,3},
        []int{0,2},
    }
    tris := CalculateTriangles(ns)
    if len(tris) != 2 {
        t.Errorf("got %v, expect %v", len(tris), 2)
    }
    ids0 := tris[0].Ids
    ids1 := tris[1].Ids
    if !Includes(ids0[:], 0) || !Includes(ids0[:], 1) || !Includes(ids0[:], 2) {
        t.Errorf("got %v, expect %v", ids0, []int{0,1,2})
    }
    if !Includes(ids1[:], 0) || !Includes(ids1[:], 2) || !Includes(ids1[:], 3) {
        t.Errorf("got %v, expect %v", ids1, []int{0,2,3})
    }
}

func TestGetGraphPaths(t *testing.T) {
    nodes := []*Node {
        &Node{Type: NodeLine, Neighbors: []int{1}},
        &Node{Type: NodeTriangle, Neighbors: []int{0,2,3}},
        &Node{Type: NodeTriangle, Neighbors: []int{1,3}},
        &Node{Type: NodeTriangle, Neighbors: []int{1,2,4}},
        &Node{Type: NodeLine, Neighbors: []int{3,5}},
        &Node{Type: NodeTriangle, Neighbors: []int{4,6,7}},
        &Node{Type: NodeTriangle, Neighbors: []int{5,9,7}},
        &Node{Type: NodeTriangle, Neighbors: []int{5,6,8}},
        &Node{Type: NodeLine, Neighbors: []int{7}},
        &Node{Type: NodeLine, Neighbors: []int{6}},
    }
    paths := GetGraphPaths(nodes)
    if len(paths) != 6 {
        t.Errorf("got %v, expect %v", len(paths), 6)
        fmt.Println(paths)
    }
}

func TestTriangleContinuesLine(t *testing.T) {
    ps := []Point{
        Point{0,0,0,-1},
        Point{math.Sqrt(3)/2,0.5,1,-1},
        Point{math.Sqrt(3)/2,-0.5,2,-1},
        Point{-1,0,3,-1},
        Point{-1-math.Sqrt(3)/2,0.5,4,-1},
        Point{-1-math.Sqrt(3)/2,-0.5,5,-1},
    }
    ns := [][]int{
        []int{1,2,3},
        []int{0,2},
        []int{0,1},
        []int{0,4,5},
        []int{3,5},
        []int{3,4},
    }
    lines := PointsToLines(ps, Distance(ps[0], ps[1]))
    lines = CullLinesByNeighbors(lines, ns)
    tris := CalculateTriangles(ns)
    if len(lines) != 7 {
        t.Errorf("got %v, expect %v", len(lines), 7)
        fmt.Println(lines)
    }
    if len(tris) != 2 {
        t.Errorf("got %v, expect %v", len(tris), 2)
    }
    n := 0
    for _,tri := range tris {
        for _,line := range lines {
            if TriangleContinuesLine(ps, tri, line) != -1 {
                n += 1
            }
        }
    }
    if n != 2 {
        t.Errorf("got %v, expect %v", n, 2)
    }
}

func TestMakeGraph(t *testing.T) {
    ps := []Point{
        Point{0,0,0,-1},
        Point{math.Sqrt(3)/2,0.5,1,-1},
        Point{math.Sqrt(3)/2,-0.5,2,-1},
        Point{-1,0,3,-1},
        Point{-1-math.Sqrt(3)/2,0.5,4,-1},
        Point{-1-math.Sqrt(3)/2,-0.5,5,-1},
    }
    ns := [][]int{
        []int{1,2,3},
        []int{0,2},
        []int{0,1},
        []int{0,4,5},
        []int{3,5},
        []int{3,4},
    }
    lines := PointsToLines(ps, Distance(ps[0], ps[1]))
    lines = CullLinesByNeighbors(lines, ns)
    tris := CalculateTriangles(ns)
    nodes := MakeGraph(ps, lines, tris)
    if len(nodes) != 13 {
        t.Errorf("got %v, expect %v", len(nodes), 13)
    }
    paths := GetGraphPaths(nodes)
    if len(paths) != 10 {
        t.Errorf("got %v, expect %v", len(paths), 10)
    }
}

func TestPathsToLines(t *testing.T) {
    ps := []Point{
        Point{0,0,0,-1},
        Point{math.Sqrt(3)/2,0.5,1,-1},
        Point{math.Sqrt(3)/2,-0.5,2,-1},
        Point{-1,0,3,-1},
        Point{-1-math.Sqrt(3)/2,0.5,4,-1},
        Point{-1-math.Sqrt(3)/2,-0.5,5,-1},
    }
    ns := [][]int{
        []int{1,2,3},
        []int{0,2},
        []int{0,1},
        []int{0,4,5},
        []int{3,5},
        []int{3,4},
    }
    lines := PointsToLines(ps, Distance(ps[0], ps[1]))
    lines = CullLinesByNeighbors(lines, ns)
    tris := CalculateTriangles(ns)
    nodes := MakeGraph(ps, lines, tris)
    if len(nodes) != 13 {
        t.Errorf("got %v, expect %v", len(nodes), 13)
    }
    paths := GetGraphPaths(nodes)
    if len(paths) != 10 {
        t.Errorf("got %v, expect %v", len(paths), 10)
    }
    plines, err := PathsToLines(nodes, lines, tris, paths)
    if err != nil {
        t.Error(err)
        for i,n := range nodes {
            fmt.Println(i, *n)
        }
        fmt.Println(paths)
    }
    if len(plines) != 10 {
        t.Errorf("got %v, expect %v", len(plines), 10)
        for i,n := range nodes {
            fmt.Println(i, *n)
        }
        fmt.Println(paths)
    }
    // Keep only length >= 3 lines
    plines = CullShortLines(plines)
    if len(plines) != 4 {
        t.Errorf("got %v, expect %v", len(plines), 4)
    }
}

func TestPathsToLines2(t *testing.T) {
    // One of the Point fields is the array index
    ps := []Point{
        Point{-1-math.Sqrt(3)/2,0.5,0,-1},
        Point{-1-math.Sqrt(3)/2,-0.5,1,-1},
        Point{math.Sqrt(3)/2,0.5,2,-1},
        Point{math.Sqrt(3)/2,-0.5,3,-1},
        Point{0,0,4,-1},
        Point{-1,0,5,-1},
    }
    ns := [][]int{
        []int{1,5},
        []int{0,5},
        []int{3,4},
        []int{2,4},
        []int{5,2,3},
        []int{0,4,1},
    }
    lines := PointsToLines(ps, Distance(ps[0], ps[1]))
    lines = CullLinesByNeighbors(lines, ns)
    tris := CalculateTriangles(ns)
    nodes := MakeGraph(ps, lines, tris)
    paths := GetGraphPaths(nodes)
    plines, err := PathsToLines(nodes, lines, tris, paths)
    if err != nil {
        t.Error(err)
    }
    plines = CullShortLines(plines)
    if len(plines) != 4 {
        t.Errorf("got %v, expect %v", len(plines), 4)
        fmt.Println(plines)
    }
}

func TestPathToLines2(t *testing.T) {
    ps := []Point{
        Point{0,0,0,-1},
        Point{math.Sqrt(3)/2,0.5,1,-1},
        Point{math.Sqrt(3)/2,-0.5,2,-1},
        Point{-1,0,3,-1},
        Point{-1-math.Sqrt(3)/2,0.5,4,-1},
        Point{-1-math.Sqrt(3)/2,-0.5,5,-1},
        Point{10,0,6,-1},
        Point{11,0,7,-1},
        Point{12,0,8,-1},
    }
    ns := [][]int{
        []int{1,2,3},
        []int{0,2},
        []int{0,1},
        []int{0,4,5},
        []int{3,5},
        []int{3,4},
        []int{7,8},
        []int{6,8},
        []int{7,6},
    }
    lines := PointsToLines(ps, Distance(ps[0], ps[1]))
    lines = CullLinesByNeighbors(lines, ns)
    tris := CalculateTriangles(ns)
    nodes := MakeGraph(ps, lines, tris)
    paths := GetGraphPaths(nodes)
    plines, err := PathsToLines(nodes, lines, tris, paths)
    if err != nil {
        t.Error(err)
    }
    plines = CullShortLines(plines)
    if len(plines) != 5 {
        t.Errorf("got %v, expect %v", len(plines), 5)
        fmt.Println(plines)
    }
}

func TestPathToLines3(t *testing.T) {
    ps := []Point{
        Point{0,0,0,-1},
        Point{math.Sqrt(3)/2,0.5,1,-1},
        Point{math.Sqrt(3)/2,-0.5,2,-1},
        Point{-1,0,3,-1},
        Point{-1-math.Sqrt(3)/2,0.5,4,-1},
        Point{-1-math.Sqrt(3)/2,-0.5,5,-1},
        Point{10,0,6,-1},
        Point{11,0,7,-1},
        Point{12,0,8,-1},
        Point{math.Sqrt(3)/2+0.5,math.Sqrt(3)/2+0.5,9,-1},
        Point{math.Sqrt(3)/2+1,math.Sqrt(3)+0.5,10,-1},
        Point{math.Sqrt(3)/2+1.5,math.Sqrt(3)*1.5+0.5,11,-1},
    }
    ns := [][]int{
        []int{1,2,3},
        []int{0,2,9,10,11},
        []int{0,1},
        []int{0,4,5},
        []int{3,5},
        []int{3,4},
        []int{7,8},
        []int{6,8},
        []int{7,6},
        []int{1,10},
        []int{9,11},
        []int{10},
    }
    lines := PointsToLines(ps, Distance(ps[0], ps[1]))
    lines = CullLinesByNeighbors(lines, ns)
    tris := CalculateTriangles(ns)
    nodes := MakeGraph(ps, lines, tris)
    paths := GetGraphPaths(nodes)
    plines, err := PathsToLines(nodes, lines, tris, paths)
    if err != nil {
        t.Error(err)
    }
    plines = CullShortLines(plines)
    if len(plines) != 6 {
        t.Errorf("got %v, expect %v", len(plines), 6)
        fmt.Println(plines)
    }
}

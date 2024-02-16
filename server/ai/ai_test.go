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

func TestGetCandidates2(t *testing.T) {
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
}

func TestCalculateTriangles(t *testing.T) {
    ps := []Point{
        Point{0,0,0,-1},
        Point{0.5,math.Sqrt(3)/2,1,-1},
        Point{0.5,-math.Sqrt(3)/2,2,-1},
    }
    ns := [][]int{
        []int{1,2},
        []int{0,2},
        []int{0,1},
    }
    board := &Board{
        Points: ps,
        Neighbors: ns,
        Turn: 0,
    }
    board.CalculateTriangles()
    if len(board.Triangles) != 1 {
        t.Errorf("got %v, expect %v", len(board.Triangles), 1)
    }
    ids := board.Triangles[0].Ids
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
    board := &Board{
        Neighbors: ns,
        Turn: 0,
    }
    board.CalculateTriangles()
    if len(board.Triangles) != 2 {
        t.Errorf("got %v, expect %v", len(board.Triangles), 2)
    }
    ids0 := board.Triangles[0].Ids
    ids1 := board.Triangles[1].Ids
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
    board := &Board{
        Neighbors: ns,
        Turn: 0,
    }
    board.CalculateTriangles()
    if len(board.Triangles) != 2 {
        t.Errorf("got %v, expect %v", len(board.Triangles), 2)
    }
    ids0 := board.Triangles[0].Ids
    ids1 := board.Triangles[1].Ids
    if !Includes(ids0[:], 0) || !Includes(ids0[:], 1) || !Includes(ids0[:], 2) {
        t.Errorf("got %v, expect %v", ids0, []int{0,1,2})
    }
    if !Includes(ids1[:], 0) || !Includes(ids1[:], 2) || !Includes(ids1[:], 3) {
        t.Errorf("got %v, expect %v", ids1, []int{0,2,3})
    }
}

func TestExtendLines(t *testing.T) {
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
    board := &Board{
        Points: ps,
        Neighbors: ns,
        Turn: 0,
    }
    board.CalculateTriangles()
    if len(board.Triangles) != 2 {
        t.Errorf("got %v, expect %v", len(board.Triangles), 2)
    }
    board.Lines = PointsToLines(board.Points)
    board.CullLongIntervalLines(1.1)
    if len(board.Lines) != 7 {
        t.Errorf("got %v, expect %v", len(board.Lines), 7)
        t.Errorf("%v", board.Lines)
    }
    board.ExtendLines()
    if len(board.Lines) != 4 {
        t.Errorf("got %v, expect %v", len(board.Lines), 10)
        t.Errorf("%v", board.Lines)
    }
    fmt.Println(board.Lines)
}

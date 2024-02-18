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

func TestPointsToLinesGood(t *testing.T) {
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
        []int{0,2,9},
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
    lines := PointsToLinesGood(ps, ns)
    lines = CullShortLines(lines)
    for i := 0; i < 3; i++ {
        lines = CombineLines(lines)
        lines = CullEqualLines(lines)
        lines = CullSubsetLines(lines)
    }
    if len(lines) != 6 {
        t.Errorf("got %v, expect %v", len(lines), 6)
        fmt.Println(lines)
    }
    fmt.Println(lines)
}

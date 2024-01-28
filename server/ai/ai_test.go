package ai

import (
    "testing"
)

func TestMakeTraditional(t *testing.T) {
    board := MakeTraditional(4)
    expect := [][]int{[]int{1,4,5}, []int{0,2,4,5,6}, []int{0,1,2,4,6,8,9,10}, []int{9,10,11,13,15}}
    got := board.Neighbors[:2]
    got = append(got, board.Neighbors[5])
    got = append(got, board.Neighbors[14])
    for i := range got {
        if len(got[i]) != len(expect[i]) {
            t.Errorf("got %v, expect %v", got[i], expect[i])
        }
        for _,ep := range expect[i] {
            if !Includes(got[i], ep) {
                t.Errorf("got %v, expect %v", got[i], expect[i])
            }
        }
    }
}

func TestShortestPaths(t *testing.T) {
    var expect [][]int
    var got [][]int
    test := func(got [][]int, expect [][]int) {
        if len(got) != len(expect) {
            t.Errorf("got %v, expect %v", got, expect)
            return
        }
        for i := range got {
            if !Equals(got[i], expect[i]) {
                t.Errorf("got %v, expect %v", got[i], expect[i])
            }
        }
    }
    board := MakeTraditional(4)
    expect = [][]int{[]int{0,1}}
    got = board.GetShortestPaths(0,1)
    test(got, expect)
    expect = [][]int{[]int{0,5,10,15}}
    got = board.GetShortestPaths(0,15)
    test(got, expect)
    // The generalized rules allow for "weird" captures on traditional boards
    // The existence of more than one shortest path on a traditional board
    // Means a capture between the two points shouldn't be allowed
    // *** This doesn't cover all cases ***
    // There is in general more than one path even for an allowed from capture to capture
    // pair of points
    expect = [][]int{[]int{0,1,6}, []int{0,5,6}}
    got = board.GetShortestPaths(0,6)
    test(got, expect)
}

func TestGetPossibleMoves(t *testing.T) {
    var expect [][2]int
    var got [][2]int
    board := MakeTraditional(4)
    board.Points[0] = 0
    board.Points[1] = 1
    expect = [][2]int{{0,2}, {0,6}}
    got = board.GetPossibleMoves()
    if !Equals(got, expect) {
        t.Errorf("got %v, expect %v", got, expect)
    }
    board.Points[4] = 0
    board.Turn = 1
    expect = [][2]int{{1,8}, {1,9}}
    got = board.GetPossibleMoves()
    if !Equals(got, expect) {
        t.Errorf("got %v, expect %v", got, expect)
    }
}

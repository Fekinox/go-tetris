package main

import "math"

type TetrisState struct {
	Grid Grid[int]

	CpIdx         int
	CpX           int
	CpY           int
	CpRot         int
	CpGrid        int
	NextPieces    []int
	HoldPiece     int
	UsedHoldPiece bool
	Airborne      bool

	LeftSnap  int
	RightSnap int
	HardDrop  int

	ClearedLines int
}

type Heuristic func(ts TetrisState) int

type NextState struct {
	Actions []Action
	State   TetrisState
}

// AI heuristics inspired by https://github.com/Tetris-Artificial-Intelligence/Tetris-Artificial-Intelligence.github.io
func (ts TetrisState) SumHeight() int {
	return 0
}

func (ts TetrisState) Bumpiness() int {
	return 0
}

func (ts TetrisState) Holes() int {
	return 0
}

func AllPossibleNextStates(ts TetrisState) []NextState {
	return nil
}

// Given a Tetris board state, find the immediate next piece placement
// that will minimize the heuristic.
func BestMove(ts TetrisState, h Heuristic) ([]Action, int) {
	// Get all possible next piece placements.
	nextStates := AllPossibleNextStates(ts)
	// If there are no possible placements (i.e. we ran out of pieces)
	// simply report the current value of the heuristic.
	if len(nextStates) == 0 {
		return nil, h(ts)
	}

	var bestActions []Action
	var bestHeuristic int = math.MaxInt
	for _, ns := range nextStates {
		_, heuristic := BestMove(ns.State, h)
		if heuristic < bestHeuristic {
			bestActions = ns.Actions
			bestHeuristic = heuristic
		}
	}

	return bestActions, bestHeuristic
}

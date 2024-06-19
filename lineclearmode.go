package main

type LineClearSettings struct {
	Lines int64
}

type LineClearObjective struct {
	Lines int64
}

func (lcs *LineClearSettings) Init(es *TetrisField) Objective {
	return &LineClearObjective{
		Lines: lcs.Lines,
	}
}

func (lco *LineClearObjective) Update(es *TetrisField) {
	if !es.gameOver {
		es.Update()
	}
}

func (lco *LineClearObjective) HandleAction(act Action, es *TetrisField) {
	es.HandleAction(act)

	if es.lines >= lco.Lines {
		es.ObjectiveComplete("Cleared all lines")
	}
}

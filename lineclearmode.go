package main

type LineClearSettings struct {
	Lines int64
}

type LineClearObjective struct {
	Lines int64

	stats []Stat
}

func (lcs *LineClearSettings) Init(es *TetrisField) Objective {
	return &LineClearObjective{
		Lines: lcs.Lines,

		stats: []Stat{
			CreateElapsedTimeStat(es),
			CreateLinesStat(es),
			CreatePiecesStat(es),
		},
	}
}

func (lco *LineClearObjective) GetStats() []Stat {
	return lco.stats
}

func (lco *LineClearObjective) Update(es *TetrisField) {
	if es.gameOver {
		return
	}
	es.Update()
}

func (lco *LineClearObjective) HandleAction(act Action, es *TetrisField) {
	es.HandleAction(act)

	if es.lines >= lco.Lines {
		es.ObjectiveComplete("Cleared all lines")
	}
}

func (lcs *LineClearSettings) CreateFormFields() []FormField {
	return []FormField{
		NewIntegerField(
			"Lines",
			lcs.Lines,
			func(value int64) {
				lcs.Lines = value
			},
			WithMin(1),
		),
	}
}

package main

type EndlessSettings struct {
}

type EndlessObjective struct {
	stats []Stat
}

func (els *EndlessSettings) Init(es *TetrisField) Objective {
	return &EndlessObjective{
		stats: []Stat {
			CreateElapsedTimeStat(es),
			CreateLinesStat(es),
			CreatePiecesStat(es),
		},
	}
}

func (eo *EndlessObjective) GetStats() []Stat {
	return eo.stats
}

func (eo *EndlessObjective) Update(es *TetrisField) {
	if es.gameOver {
		return
	}

	es.Update()
}

func (eo *EndlessObjective) HandleAction(act Action, es *TetrisField) {
	es.HandleAction(act)
}

func (els *EndlessSettings) CreateFormFields() []FormField {
	return []FormField{}
}

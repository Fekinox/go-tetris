package main

type EndlessSettings struct {
}

type EndlessObjective struct {
}

func (els *EndlessSettings) Init(es *TetrisField) Objective {
	return &EndlessObjective{}
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

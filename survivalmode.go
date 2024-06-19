package main

type SurvivalSettings struct {
	GarbageRate float64
}

type SurvivalObjective struct {
	GarbageRate float64
	GarbageTimer float64
}

func (ss *SurvivalSettings) Init(es *TetrisField) Objective {
	return &SurvivalObjective{
		GarbageRate: ss.GarbageRate,
		GarbageTimer: ss.GarbageRate,
	}
}

func (so *SurvivalObjective) Update(es *TetrisField) {
	if es.gameOver {
		return 
	}

	es.Update()

	so.GarbageTimer -= UPDATE_TICK_RATE_MS
	if so.GarbageTimer < 0 {
		so.GarbageTimer += so.GarbageRate
		es.AddGarbage(1)
	}
}

func (so *SurvivalObjective) HandleAction(act Action, es *TetrisField) {
	es.HandleAction(act)
}

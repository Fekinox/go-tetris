package main

type SurvivalSettings struct {
	GarbageRate int64
}

type SurvivalObjective struct {
	GarbageRate  float64
	GarbageTimer float64

	stats []Stat
}

func (ss *SurvivalSettings) Init(es *TetrisField) Objective {
	val := float64(ss.GarbageRate)
	return &SurvivalObjective{
		GarbageRate:  val,
		GarbageTimer: val,

		stats: []Stat{
			CreateElapsedTimeStat(es),
			CreateLinesStat(es),
			CreatePiecesStat(es),
		},
	}
}

func (so *SurvivalObjective) GetStats() []Stat {
	return so.stats
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

func (ss *SurvivalSettings) CreateFormFields() []FormField {
	return []FormField{
		NewIntegerField(
			"Garbage Rate",
			ss.GarbageRate,
			func(value int64) {
				ss.GarbageRate = value
			},
			WithMin(100),
		),
	}
}

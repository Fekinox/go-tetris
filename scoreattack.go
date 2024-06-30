package main

const INITIAL_DURATION_SECS int64 = 30

type ScoreAttackSettings struct {
	Duration int64
}

type ScoreAttackObjective struct {
	Duration int64

	stats []Stat
}

func (sas *ScoreAttackSettings) Init(es *TetrisField) Objective {
	so := &ScoreAttackObjective{
		Duration: sas.Duration,

		stats: []Stat{
			CreateCountdownStat(es, sas.Duration),
			CreateLinesStat(es),
			CreatePiecesStat(es),
		},
	}

	return so
}

func (sas *ScoreAttackSettings) CreateFormFields() []FormField {
	return []FormField{
		NewIntegerField(
			"Duration in seconds",
			sas.Duration,
			func(value int64) {
				sas.Duration = value
			},
			WithMin(1),
		),
	}
}

func (so *ScoreAttackObjective) GetStats() []Stat {
	return so.stats
}

func (so *ScoreAttackObjective) Update(es *TetrisField) {
	if es.gameOver {
		return
	}

	es.Update()

	if es.frameCount >= so.Duration*FRAMES_PER_SECOND {
		es.ObjectiveComplete("Time out")
	}
}

func (so *ScoreAttackObjective) HandleAction(act Action, es *TetrisField) {
	es.HandleAction(act)
}

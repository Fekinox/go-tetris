package main

type Gamemode int

const (
	SprintMode Gamemode = iota
	SurvivalMode
	CheeseMode
	EndlessMode
	ScoreAttackMode
)

type GlobalTetrisSettings struct {
	StartingLevel int
	MaxResets int
	LockDelay int
	BaseGravity int64
	GravityIncrease int64
}

type GamemodeSettings interface {
	gamemodeSettings()
}

type SprintModeSettings struct {
	LineLimit int
}

type SurvivalModeSettings struct {
	GarbageRate int
}

type CheeseModeSettings struct {
	GarbageLines int
}

type ScoreAttackModeSettings struct {
	LengthSecs int64
}

type EndlessModeSettings struct {
}

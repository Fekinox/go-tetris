package main

type ReplayData struct {
	Seed	int64
	TetrisSettings GlobalTetrisSettings
	Gamemode Gamemode
	GamemodeSettings GamemodeSettings
	Actions []ReplayAction
}

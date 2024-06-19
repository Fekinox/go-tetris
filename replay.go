package main

type ReplayData struct {
	Seed	int64
	TetrisSettings GlobalTetrisSettings
	Actions []ReplayAction
}

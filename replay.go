package main

type ReplayData struct {
	Seed	int64
	StartingLevel int64
	Actions []ReplayAction
}

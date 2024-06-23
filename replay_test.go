package main

import (
	"io"
	"math/rand"
	"testing"
)

func FuzzReplaysCompressed(f *testing.F) {
	for _, seed := range []int64{0, 1, 2, 3, 4, 5, 6} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, a int64) {
		seed := a
		rand := rand.New(rand.NewSource(seed))

		numFrames := rand.Int63n(36000)
		actions := make([]ReplayAction, 0)
		for i := int64(0); i < numFrames; i++ {
			if rand.Float64() < 0.1 {
				newAction := Action(rand.Intn(13))
				actions = append(actions, ReplayAction{
					Action: newAction,
					Frame:  i,
				})
			}
		}

		repData := ReplayData{
			Seed:              seed,
			TetrisSettings:    GlobalTetrisSettings{},
			ObjectiveID:       Endless,
			ObjectiveSettings: &EndlessSettings{},
			Actions:           actions,
		}

		r, w := io.Pipe()
		go func() {
			err := EncodeCompressed(&repData, w)
			if err != nil {
				t.Fatalf("Could not encode")
			}
			w.Close()
		}()

		newRepData, err := DecodeCompressed(r)
		if err != nil {
			t.Fatalf("Could not decode")
		}

		if repData.Seed != newRepData.Seed {
			t.Fatalf("Seed differs (old: %v, new: %v)", repData.Seed,
				newRepData.Seed)
		}

		if len(repData.Actions) != len(newRepData.Actions) {
			t.Fatalf("Different amounts of actions (old: %v, new: %v)",
				len(repData.Actions),
				len(newRepData.Actions),
			)
		}
		
		for i := 0; i < len(repData.Actions); i++ {
			oldAct := repData.Actions[i]
			newAct := newRepData.Actions[i]
			if oldAct.Frame != newAct.Frame || oldAct.Action != newAct.Action {
				t.Fatalf("Action discrepancy (old: %v, new: %v)", oldAct,
				newAct)
			}
		}
	})
}

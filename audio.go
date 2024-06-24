package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

type AudioEngine struct {
	Context *audio.Context
	Players map[string]*audio.Player
}

func NewAudioEngine() (*AudioEngine, error) {
	ae := &AudioEngine{
		Context: audio.NewContext(44100),
		Players: make(map[string]*audio.Player),
	}

	audioDir, err := os.Open("assets/sfx")
	if err != nil {
		return nil, err
	}

	entries, err := audioDir.ReadDir(0)
	if err != nil {
		return nil, err
	}

	for _, en := range entries {
		soundName := strings.TrimSuffix(en.Name(), filepath.Ext(en.Name()))

		file, err := os.Open(fmt.Sprintf("assets/sfx/%s", en.Name()))
		if err != nil {
			return nil, err
		}

		stream, err := vorbis.DecodeWithSampleRate(44100, file)
		if err != nil {
			return nil, err
		}

		player, err := ae.Context.NewPlayer(stream)
		if err != nil {
			return nil, err
		}

		ae.Players[soundName] = player
	}

	return ae, nil
}

func (ae *AudioEngine) PlaySound(name string) {
	player, ok := ae.Players[name]
	if !ok { return }

	player.SetPosition(0)
	player.Play()
}

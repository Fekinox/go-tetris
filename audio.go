package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

type AudioService interface {
	PlaySound(name string)
	StopSound(name string)
}

type NullAudioEngine struct {
}

type AudioEngine struct {
	Context *audio.Context
	Players map[string]*audio.Player
}

func (ne *NullAudioEngine) PlaySound(name string) {
}

func (ne *NullAudioEngine) StopSound(name string) {
}

func MustCreateAudioEngine() *AudioEngine {
	ae, err := CreateAudioEngine()
	if err != nil {
		panic(err)
	}

	return ae
}

func CreateAudioEngine() (*AudioEngine, error) {
	ae := &AudioEngine{
		Context: audio.NewContext(44100),
		Players: make(map[string]*audio.Player),
	}

	sfxDir, err := os.Open("assets/sfx")
	if err != nil {
		return nil, err
	}

	sfxEntries, err := sfxDir.ReadDir(0)
	if err != nil {
		return nil, err
	}

	for _, en := range sfxEntries {
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

	musicDir, err := os.Open("assets/music")
	if err != nil {
		return nil, err
	}

	musicEntries, err := musicDir.ReadDir(0)
	if err != nil {
		return nil, err
	}

	for _, en := range musicEntries {
		soundName := strings.TrimSuffix(en.Name(), filepath.Ext(en.Name()))

		file, err := os.Open(fmt.Sprintf("assets/music/%s", en.Name()))
		if err != nil {
			return nil, err
		}

		stream, err := vorbis.DecodeWithSampleRate(44100, file)
		if err != nil {
			return nil, err
		}

		infLoopStream := audio.NewInfiniteLoop(stream, stream.Length())

		player, err := ae.Context.NewPlayer(infLoopStream)
		if err != nil {
			return nil, err
		}

		ae.Players[soundName] = player
	}
	return ae, nil
}

func (ae *AudioEngine) PlaySound(name string) {
	player, ok := ae.Players[name]
	if !ok {
		panic("Invalid sound")
	}

	player.SetPosition(0)
	player.Play()
}

func (ae *AudioEngine) StopSound(name string) {
	player, ok := ae.Players[name]
	if !ok {
		panic("Invalid sound")
	}

	player.Pause()
}

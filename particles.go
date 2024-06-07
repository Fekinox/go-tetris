package main

import (
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
)

const NUM_PARTICLES = 256
const MIN_VISIBLE_INTENSITY = 0.05

var PARTICLE_LEVELS = []rune(
	"..--**%%##",
)

type Particle struct {
	Intensity float64
	Style     tcell.Style
	X         int
	Y         int
}

func (p Particle) GetRune() rune {
	v := max(
		0,
		min(
			len(PARTICLE_LEVELS)-1,
			int(p.Intensity*float64(len(PARTICLE_LEVELS))),
		),
	)
	return PARTICLE_LEVELS[v]
}

type ParticleSystem struct {
	Particles       []Particle
	head            int
	tail            int
	curParticles    int
	intensityJitter float64
	rand            *rand.Rand
}

func InitParticles(intensityJitter float64) ParticleSystem {
	return ParticleSystem{
		Particles:       make([]Particle, NUM_PARTICLES),
		intensityJitter: intensityJitter,
		rand:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (ps *ParticleSystem) Update() {
	for i := ps.head; i != ps.tail; i = (i + 1) % NUM_PARTICLES {
		p := &ps.Particles[i]
		if p.Intensity < 0 {
			ps.KillParticle(i)
		}
		p.Intensity -= 2.0 * float64(UPDATE_TICK_RATE_MS) / 1000.0
	}
}

func (ps *ParticleSystem) Draw(rr Area) {
	for i := ps.head; i != ps.tail; i = (i + 1) % NUM_PARTICLES {
		p := ps.Particles[i]
		if p.Intensity < MIN_VISIBLE_INTENSITY { continue }
		Screen.SetContent(
			rr.X+p.X,
			rr.Y+p.Y,
			p.GetRune(),
			nil, p.Style)
	}
}

func (ps *ParticleSystem) SpawnParticle(p Particle) {
	if p.Intensity < MIN_VISIBLE_INTENSITY { return }
	p.Intensity += (2*ps.rand.Float64() - 1.0) * ps.intensityJitter

	ps.Particles[ps.tail] = p
	if ps.curParticles == NUM_PARTICLES {
		ps.tail = (ps.tail + 1) % NUM_PARTICLES
		ps.head = (ps.head + 1) % NUM_PARTICLES
	} else {
		ps.curParticles += 1
		ps.tail = (ps.tail + 1) % NUM_PARTICLES
	}
}

func (ps *ParticleSystem) KillParticle(i int) {
	ps.Particles[i] = ps.Particles[ps.head]
	ps.head = (ps.head + 1) % NUM_PARTICLES
	ps.curParticles -= 1
}

package main

import (
	"math/rand"
)

type PieceGenerator interface {
	NextPiece() int
}

type TrueRandomPieceGenerator struct {
	rand *rand.Rand
}

func NewTrueRandomPieceGenerator(seed int64) TrueRandomPieceGenerator {
	return TrueRandomPieceGenerator{
		rand: rand.New(rand.NewSource(seed)),
	}
}

func (pg *TrueRandomPieceGenerator) NextPiece() int {
	return pg.rand.Intn(7)
}

type BagRandomizer struct {
	rand *rand.Rand
	bag  []int
	curr int
}

func NewBagRandomizer(seed int64, levels int) BagRandomizer {
	br := BagRandomizer{
		rand: rand.New(rand.NewSource(seed)),
		bag:  make([]int, 7*levels),
	}

	for i := 0; i < 7*levels; i++ {
		br.bag[i] = i % 7
	}

	br.shuffle()

	return br
}

func (br *BagRandomizer) shuffle() {
	for i := len(br.bag) - 1; i > 0; i-- {
		j := br.rand.Intn(i + 1)
		tmp := br.bag[i]
		br.bag[i] = br.bag[j]
		br.bag[j] = tmp
	}
}

func (br *BagRandomizer) NextPiece() int {
	p := br.bag[br.curr]
	br.curr += 1
	if br.curr == len(br.bag) {
		br.curr = 0
		br.shuffle()
	}

	return p
}

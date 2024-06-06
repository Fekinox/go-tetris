package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

const UPDATE_TICK_RATE_MS float64 = 1000.0 / 240.0

const BOARD_WIDTH = 10
const BOARD_HEIGHT = 22

const INITIAL_FALL_RATE = 240

const NUM_NEXT_PIECES = 5

func IsRune(ev *tcell.EventKey, r rune) bool {
	return (ev.Key() == tcell.KeyRune && ev.Rune() == r)
}

type EngineState struct {
	LastRenderDuration float64
	LastUpdateDuration float64

	grid Grid[int]

	currentPieceIdx      int
	currentPieceGrid     Grid[bool]
	currentPieceX        int
	currentPieceY        int
	currentPieceRotation int
	hardDropHeight       int

	gravityTimer int
	fallRate     int

	pieceGenerator PieceGenerator

	nextPieces        []int
	hardDropParticles ParticleSystem
}

func InitEngineState() *EngineState {
	gen := NewBagRandomizer(time.Now().UnixNano(), 2)
	es := EngineState{
		LastUpdateDuration: UPDATE_TICK_RATE_MS,

		grid:          MakeGrid(BOARD_WIDTH, BOARD_HEIGHT, 0),
		pieceGenerator: &gen,
		currentPieceX: BOARD_WIDTH / 2,
		currentPieceY: 3,
		fallRate:      INITIAL_FALL_RATE,
		nextPieces:    make([]int, NUM_NEXT_PIECES),
	}

	es.gravityTimer = es.fallRate
	es.hardDropParticles = InitParticles(0.1)

	es.FillNextPieces()
	es.GetRandomPiece()

	return &es
}

func (es *EngineState) ResetGame() {
}

func (es *EngineState) HandleInput(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyUp ||
			IsRune(ev, 'k') || IsRune(ev, 'K') ||
			IsRune(ev, 'x') || IsRune(ev, 'X') {
			es.RotateCW()
		}
		if IsRune(ev, 'z') || IsRune(ev, 'Z') {
			es.RotateCCW()
		} else if ev.Key() == tcell.KeyDown || IsRune(ev, 'j') || IsRune(ev, 'J') {
			es.SoftDrop()
		} else if ev.Key() == tcell.KeyLeft || IsRune(ev, 'h') || IsRune(ev, 'H') {
			es.MovePiece(-1)
		} else if ev.Key() == tcell.KeyRight || IsRune(ev, 'l') || IsRune(ev, 'L') {
			es.MovePiece(1)
		} else if IsRune(ev, ' ') {
			es.HardDrop()
		} else if IsRune(ev, 'r') || IsRune(ev, 'R') {
			es.HandleReset()
		}
	}
}

func (es *EngineState) Update() {
	es.hardDropParticles.Update()
	es.gravityTimer -= 1
	if es.gravityTimer <= 0 {
		es.gravityTimer = es.fallRate
		es.SoftDrop()
	}
}

func (es *EngineState) Draw(lag float64) {
	Screen.Clear()
	defer Screen.Show()
	sw, sh := Screen.Size()
	if sw < MIN_WIDTH || sh < MIN_HEIGHT {
		ShowResizeScreen(sw, sh, defStyle)
		return
	}

	rr := Area{
		X:      (sw - MIN_WIDTH) / 2,
		Y:      (sh - MIN_HEIGHT) / 2,
		Width:  MIN_WIDTH,
		Height: MIN_HEIGHT,
	}

	BorderBox(Area{
		X:      rr.X - 1,
		Y:      rr.Y - 1,
		Width:  rr.Width + 2,
		Height: rr.Height + 2,
	}, defStyle)

	es.DrawWell(rr)
	gameArea := Area{
		X:      rr.X + 1,
		Y:      rr.Y + 1,
		Width:  BOARD_WIDTH,
		Height: BOARD_HEIGHT,
	}
	es.hardDropParticles.Draw(gameArea)
	es.DrawHardDropIndicator(gameArea)

	gridOffsetX := es.currentPieceGrid.Width / 2
	gridOffsetY := es.currentPieceGrid.Height / 2

	es.DrawPiece(
		gameArea,
		es.currentPieceGrid,
		es.currentPieceIdx,
		es.currentPieceX - gridOffsetX,
		es.currentPieceY - gridOffsetY,
	)

	es.DrawGrid(gameArea)

	nextPieceArea := Area {
		X: rr.X + BOARD_WIDTH + 6,
		Y: rr.Y + 1,
	}

	es.DrawNextPieces(nextPieceArea)
}

func (es *EngineState) DrawWell(rr Area) {
	for y := 0; y < es.grid.Height+1; y++ {
		Screen.SetContent(
			rr.X,
			rr.Y+y,
			'#',
			nil, defStyle)
		Screen.SetContent(
			rr.X+es.grid.Width+1,
			rr.Y+y,
			'#',
			nil, defStyle)
	}
	for xx := 0; xx < es.grid.Width; xx++ {
		Screen.SetContent(
			rr.X+1+xx,
			rr.Y+es.grid.Height+1,
			'#',
			nil, defStyle)
	}
}

func (es *EngineState) DrawPiece(
	rr Area,
	piece Grid[bool],
	pieceIndex int,
	px, py int) {

	for yy := 0; yy < piece.Height; yy++ {
		for xx := 0; xx < piece.Width; xx++ {
			if piece.MustGet(xx, yy) {
				color := PieceColors[pieceIndex]
				style :=
					defStyle.Background(color).Foreground(tcell.ColorBlack)
				Screen.SetContent(
					rr.X + xx + px,
					rr.Y + yy + py,
					'o',
					nil, style)
			}
		}
	}
}

func (es *EngineState) DrawNextPieces(rr Area) {
	for i := 0; i < NUM_NEXT_PIECES; i++ {
		px := 0
		py := (i+1) * 4
		es.DrawPiece(
			rr,
			Pieces[es.nextPieces[i]][0],
			es.nextPieces[i],
			px, py)
	}
}

func (es *EngineState) DrawHardDropIndicator(rr Area) {
	gridOffsetX := es.currentPieceGrid.Width / 2
	gridOffsetY := es.currentPieceGrid.Height / 2

	for yy := 0; yy < es.currentPieceGrid.Height; yy++ {
		for xx := 0; xx < es.currentPieceGrid.Width; xx++ {
			if es.currentPieceGrid.MustGet(xx, yy) {
				color := PieceColors[es.currentPieceIdx]
				Screen.SetContent(
					rr.X+xx-gridOffsetX+es.currentPieceX,
					rr.Y+yy-gridOffsetY+es.hardDropHeight,
					'+',
					nil, defStyle.Foreground(color))
			}
		}
	}
}

func (es *EngineState) DrawGrid(rr Area) {
	for yy := 0; yy < es.grid.Height; yy++ {
		for xx := 0; xx < es.grid.Width; xx++ {
			if es.grid.MustGet(xx, yy) != 0 {
				color := PieceColors[es.grid.MustGet(xx, yy)-1]
				style :=
					defStyle.Background(color).Foreground(tcell.ColorBlack)
				Screen.SetContent(
					rr.X+xx,
					rr.Y+yy,
					'o',
					nil, style)
			}
		}
	}
}

func (es *EngineState) FillNextPieces() {
	for i := 0; i < NUM_NEXT_PIECES; i++ {
		es.nextPieces[i] = es.pieceGenerator.NextPiece()
	}
}

func (es *EngineState) GetRandomPiece() {
	idx := es.nextPieces[0]
	es.currentPieceIdx = idx
	es.currentPieceGrid = Pieces[idx][0]
	es.currentPieceRotation = 0
	es.currentPieceX = BOARD_WIDTH / 2
	es.currentPieceY = 1

	for i := 0; i < NUM_NEXT_PIECES-1; i++ {
		es.nextPieces[i] = es.nextPieces[i+1]
	}
	es.nextPieces[NUM_NEXT_PIECES-2] = es.pieceGenerator.NextPiece()

	es.SetHardDropHeight()
}

func (es *EngineState) SetHardDropHeight() {
	yy := es.currentPieceY
	for !es.CheckCollision(es.currentPieceGrid, es.currentPieceX, yy+1) {
		yy += 1
	}

	es.hardDropHeight = yy
}

func (es *EngineState) RotateCW() {
	rotationLength := len(Pieces[es.currentPieceIdx])
	newRotation := (es.currentPieceRotation + 1) % rotationLength
	if es.CheckCollision(
		Pieces[es.currentPieceIdx][newRotation],
		es.currentPieceX,
		es.currentPieceY,
	) {
		return
	}

	es.currentPieceRotation = newRotation
	es.currentPieceGrid = Pieces[es.currentPieceIdx][es.currentPieceRotation]
	es.SetHardDropHeight()
}

func (es *EngineState) RotateCCW() {
	rotationLength := len(Pieces[es.currentPieceIdx])
	newRotation := es.currentPieceRotation - 1
	if newRotation < 0 {
		newRotation = rotationLength - 1
	}
	if es.CheckCollision(
		Pieces[es.currentPieceIdx][newRotation],
		es.currentPieceX,
		es.currentPieceY,
	) {
		return
	}

	es.currentPieceRotation = newRotation
	es.currentPieceGrid = Pieces[es.currentPieceIdx][es.currentPieceRotation]
	es.SetHardDropHeight()
}

func (es *EngineState) HandleReset() {
	es.grid = MakeGrid(BOARD_WIDTH, BOARD_HEIGHT, 0)
	es.GetRandomPiece()
}

func (es *EngineState) CheckCollision(piece Grid[bool], px, py int) bool {
	gridOffsetX := piece.Width / 2
	gridOffsetY := piece.Height / 2

	for yy := 0; yy < piece.Height; yy++ {
		for xx := 0; xx < piece.Width; xx++ {
			if piece.MustGet(xx, yy) {
				currX := xx - gridOffsetX + px
				currY := yy - gridOffsetY + py
				cell, ok := es.grid.Get(currX, currY)
				if !ok || cell != 0 {
					return true
				}
			}
		}
	}

	return false
}

func (es *EngineState) MovePiece(dx int) {
	if es.CheckCollision(
		es.currentPieceGrid,
		es.currentPieceX+dx,
		es.currentPieceY,
	) {
		return
	}

	es.currentPieceX += dx

	es.SetHardDropHeight()
}

func (es *EngineState) SoftDrop() {
	if es.CheckCollision(
		es.currentPieceGrid,
		es.currentPieceX,
		es.currentPieceY+1,
	) {
		es.PlacePiece()
		return
	}

	es.currentPieceY += 1
}

func (es *EngineState) PlacePiece() {
	gridOffsetX := es.currentPieceGrid.Width / 2
	gridOffsetY := es.currentPieceGrid.Height / 2

	for yy := 0; yy < es.currentPieceGrid.Height; yy++ {
		for xx := 0; xx < es.currentPieceGrid.Width; xx++ {
			if es.currentPieceGrid.MustGet(xx, yy) {
				currX := xx - gridOffsetX + es.currentPieceX
				currY := yy - gridOffsetY + es.currentPieceY
				es.grid.Set(currX, currY, es.currentPieceIdx+1)
			}
		}
	}

	es.ClearLines()
	es.GetRandomPiece()
}

func (es *EngineState) HardDrop() {
	es.SpawnHardDropParticles(es.currentPieceY, es.hardDropHeight)
	es.currentPieceY = es.hardDropHeight
	es.PlacePiece()
}

func (es *EngineState) ClearLines() {
	lines := make([]int, 0)
	for y := 0; y < es.grid.Height; y++ {
		fullLine := true
	notFullLine:
		for x := 0; x < es.grid.Width; x++ {
			if es.grid.MustGet(x, y) == 0 {
				fullLine = false
				break notFullLine
			}
		}

		if fullLine {
			lines = append(lines, y)
		}
	}

	for _, lidx := range lines {
		for y := lidx; y >= 0; y-- {
			for x := 0; x < es.grid.Width; x++ {
				if y == 0 {
					es.grid.Set(x, y, 0)
				} else {
					es.grid.Set(x, y, es.grid.MustGet(x, y-1))
				}
			}
		}
	}
}

func (es *EngineState) SpawnHardDropParticles(prevHeight, hardDropHeight int) {
	gridOffsetX := es.currentPieceGrid.Width / 2
	gridOffsetY := es.currentPieceGrid.Height / 2

	blockTops := make([]int, es.currentPieceGrid.Width)
	for x := 0; x < es.currentPieceGrid.Width; x++ {
		found := false
	topFinder:
		for y := 0; y < es.currentPieceGrid.Height; y++ {
			if es.currentPieceGrid.MustGet(x, y) {
				blockTops[x] = y
				found = true
				break topFinder
			}
		}

		if !found {
			blockTops[x] = -1
		}
	}

	for z := hardDropHeight-1; z > prevHeight; z-- {
		for dx, h := range blockTops {
			if h < 0 {
				continue
			}
			i := 1 - min(1, float32(hardDropHeight - z) / 15.0)
			es.hardDropParticles.SpawnParticle(
				Particle{
					Intensity: i * i * i,
					Style:     defStyle.Foreground(PieceColors[es.currentPieceIdx]),
					X:         dx + es.currentPieceX - gridOffsetX,
					Y:         z + h - gridOffsetY,
				},
			)
		}
	}
}

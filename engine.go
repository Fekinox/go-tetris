package main

import (
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
)

const UPDATE_TICK_RATE_MS float64 = 1000.0 / 240.0

const BOARD_WIDTH = 10
const BOARD_HEIGHT = 22

func IsRune(ev *tcell.EventKey, r rune) bool {
	return (ev.Key() == tcell.KeyRune && ev.Rune() == r)
}

type EngineState struct {
	LastRenderDuration float64
	LastUpdateDuration float64

	grid Grid[int]

	currentPieceIdx int
	currentPieceGrid Grid[bool]
	currentPieceX int
	currentPieceY int
	currentPieceRotation int

	rand *rand.Rand
}

func InitEngineState() *EngineState {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	es := EngineState{
		LastUpdateDuration: UPDATE_TICK_RATE_MS,

		grid: MakeGrid(BOARD_WIDTH, BOARD_HEIGHT, 0),
		rand: rand,
		currentPieceX: BOARD_WIDTH/2,
		currentPieceY: 3,
	}

	es.GetRandomPiece()

	return &es
}

func (es *EngineState) ResetGame() {
}

func (es *EngineState) HandleInput(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyUp ||
			IsRune(ev, 'w') || IsRune(ev, 'W') ||
			IsRune(ev, 'x') || IsRune(ev, 'X') {
			es.RotateCW()
		}
		if IsRune(ev, 'z') || IsRune(ev, 'Z') {
			es.RotateCCW()
		} else if ev.Key() == tcell.KeyDown || IsRune(ev, 's') || IsRune(ev, 'S') {
			es.SoftDrop()
		} else if ev.Key() == tcell.KeyLeft || IsRune(ev, 'a') || IsRune(ev, 'A') {
			es.MovePiece(-1)
		} else if ev.Key() == tcell.KeyRight || IsRune(ev, 'd') || IsRune(ev, 'D') {
			es.MovePiece(1)
		} else if IsRune(ev, ' ') {
			es.HardDrop()
		} else if IsRune(ev, 'r') || IsRune(ev, 'R') {
			es.HandleReset()
		}
	}
}

func (es *EngineState) Update() {
	// Handle input
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
		X:		rr.X + 1,
		Y:		rr.Y + 1,
		Width:	BOARD_WIDTH,
		Height: BOARD_HEIGHT,
	}
	es.DrawCurrentPiece(gameArea)
	es.DrawGrid(gameArea)
}

func (es *EngineState) DrawWell(rr Area) {
	for y := 0; y < es.grid.Height+1; y++ {
		Screen.SetContent(
			rr.X,
			rr.Y + y,
			'#',
			nil, defStyle)
		Screen.SetContent(
			rr.X + es.grid.Width+1,
			rr.Y + y,
			'#',
			nil, defStyle)
	}
	for xx := 0; xx < es.grid.Width; xx++ {
		Screen.SetContent(
			rr.X + 1 + xx,
			rr.Y + es.grid.Height + 1,
			'#',
			nil, defStyle)
	}
}

func (es *EngineState) DrawCurrentPiece(rr Area) {
	gridOffsetX := es.currentPieceGrid.Width/2
	gridOffsetY := es.currentPieceGrid.Height/2

	for yy := 0; yy < es.currentPieceGrid.Height; yy++ {
		for xx := 0; xx < es.currentPieceGrid.Width; xx++ {
			if es.currentPieceGrid.MustGet(xx, yy) {
				color := PieceColors[es.currentPieceIdx]
				Screen.SetContent(
					rr.X + xx - gridOffsetX + es.currentPieceX,
					rr.Y + yy - gridOffsetY + es.currentPieceY,
					' ',
					nil, defStyle.Background(color))
			}
		}	
	}
}

func (es *EngineState) DrawGrid(rr Area) {
	for yy := 0; yy < es.grid.Height; yy++ {
		for xx := 0; xx < es.grid.Width; xx++ {
			if es.grid.MustGet(xx, yy) != 0 {
				color := PieceColors[es.grid.MustGet(xx, yy)-1]
				Screen.SetContent(
					rr.X + xx,
					rr.Y + yy,
					' ',
					nil, defStyle.Background(color))
			}
		}	
	}
}

func (es *EngineState) GetRandomPiece() {
	idx := (es.currentPieceIdx + 1)%7
	es.currentPieceIdx = idx
	es.currentPieceGrid = Pieces[idx][0]
	es.currentPieceRotation = 0
	es.currentPieceX = BOARD_WIDTH/2
	es.currentPieceY = 1
}

func (es *EngineState) RotateCW() {
	rotationLength := len(Pieces[es.currentPieceIdx])
	es.currentPieceRotation = (es.currentPieceRotation + 1) % rotationLength
	es.currentPieceGrid = Pieces[es.currentPieceIdx][es.currentPieceRotation]
}

func (es *EngineState) RotateCCW() {
	rotationLength := len(Pieces[es.currentPieceIdx])
	es.currentPieceRotation -= 1
	if es.currentPieceRotation < 0 { 
		es.currentPieceRotation = rotationLength - 1
	}
	es.currentPieceGrid = Pieces[es.currentPieceIdx][es.currentPieceRotation]
}

func (es *EngineState) HandleReset() {
	es.grid = MakeGrid(BOARD_WIDTH, BOARD_HEIGHT, 0)
	es.GetRandomPiece()
}

func (es *EngineState) CurrentPieceCollides(px, py int) bool {
	gridOffsetX := es.currentPieceGrid.Width/2
	gridOffsetY := es.currentPieceGrid.Height/2

	for yy := 0; yy < es.currentPieceGrid.Height; yy++ {
		for xx := 0; xx < es.currentPieceGrid.Width; xx++ {
			if es.currentPieceGrid.MustGet(xx, yy) {
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
	if es.CurrentPieceCollides(
		es.currentPieceX + dx,
		es.currentPieceY,
	) {
		return
	}

	es.currentPieceX += dx
}

func (es *EngineState) SoftDrop() {
	if es.CurrentPieceCollides(
		es.currentPieceX,
		es.currentPieceY + 1,
	) {
		es.PlacePiece()
		return
	}

	es.currentPieceY += 1
}

func (es *EngineState) PlacePiece() {
	gridOffsetX := es.currentPieceGrid.Width/2
	gridOffsetY := es.currentPieceGrid.Height/2

	for yy := 0; yy < es.currentPieceGrid.Height; yy++ {
		for xx := 0; xx < es.currentPieceGrid.Width; xx++ {
			if es.currentPieceGrid.MustGet(xx, yy) {
				currX := xx - gridOffsetX + es.currentPieceX
				currY := yy - gridOffsetY + es.currentPieceY
				es.grid.Set(currX, currY, es.currentPieceIdx + 1)
			}
		}	
	}

	es.ClearLines()
	es.GetRandomPiece()
}

func (es *EngineState) HardDrop() {
	yy := es.currentPieceY
	for !es.CurrentPieceCollides(es.currentPieceX, yy+1) {
		yy += 1
	}

	es.currentPieceY = yy
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
				break notFullLine;
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

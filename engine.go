package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
)

const UPDATE_TICK_RATE_MS float64 = 1000.0 / 240.0

const BOARD_WIDTH = 10
const BOARD_HEIGHT = 22

const INITIAL_FALL_RATE = 60

const NUM_NEXT_PIECES = 5

const SINGLE_SCORE = 100
const DOUBLE_SCORE = 300
const TRIPLE_SCORE = 500
const TETRIS_SCORE = 800

const COMBO_BASE_SCORE = 50

var COMBO_COUNTS = []int{
	0, 0,
	1, 1,
	2, 3,
	3, 3,
	4, 4, 4,
}

func IsRune(ev *tcell.EventKey, r rune) bool {
	return (ev.Key() == tcell.KeyRune && ev.Rune() == r)
}

func IsDigitRune(ev *tcell.EventKey) bool {
	if ev.Key() != tcell.KeyRune {
		return false
	}
	return ev.Rune() >= '0' && ev.Rune() <= '9'
}

type EngineState struct {
	LastRenderDuration float64
	LastUpdateDuration float64

	grid Grid[int]

	cpIdx  int
	cpGrid Grid[bool]
	cpX    int
	cpY    int
	cpRot  int

	gravityTimer int
	fallRate     int

	pieceGenerator PieceGenerator

	nextPieces []int

	dashParticles ParticleSystem

	holdPiece     int
	usedHoldPiece bool

	moveMultiplier          int
	leftSnapPosition        int
	rightSnapPosition       int
	hardDropLeftSnapHeight  int
	hardDropRightSnapHeight int
	hardDropHeight          int

	score int64
	lines int64
	combo int
}

func NewEngineState() *EngineState {
	es := EngineState{
		LastUpdateDuration: UPDATE_TICK_RATE_MS,

		fallRate:   INITIAL_FALL_RATE,
		nextPieces: make([]int, NUM_NEXT_PIECES),
		holdPiece:  8,
	}

	es.StartGame(time.Now().UnixNano())

	return &es
}

func (es *EngineState) StartGame(seed int64) {
	gen := NewBagRandomizer(seed, 1)
	es.grid = MakeGrid(BOARD_WIDTH, BOARD_HEIGHT, 0)
	es.holdPiece = 8
	es.pieceGenerator = &gen

	es.gravityTimer = es.fallRate
	es.dashParticles = InitParticles(0.1)

	es.FillNextPieces()
	es.GetRandomPiece()
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
		} else if IsRune(ev, '$') {
			es.SetMoveMultiplier(10)
		} else if IsDigitRune(ev) {
			es.SetMoveMultiplier(int(ev.Rune() - '0'))
		} else if ev.Key() == tcell.KeyDown || IsRune(ev, 'j') || IsRune(ev, 'J') {
			es.SoftDrop()
		} else if ev.Key() == tcell.KeyLeft || IsRune(ev, 'h') || IsRune(ev, 'H') {
			es.MovePiece(-1)
		} else if ev.Key() == tcell.KeyRight || IsRune(ev, 'l') || IsRune(ev, 'L') {
			es.MovePiece(1)
		} else if IsRune(ev, ' ') {
			es.HardDrop()
		} else if IsRune(ev, 'c') || IsRune(ev, 'C') {
			es.SwapHoldPiece()
		} else if IsRune(ev, 'r') || IsRune(ev, 'R') {
			es.HandleReset()
		}
	}
}

func (es *EngineState) Update() {
	es.dashParticles.Update()
	es.gravityTimer -= 1
	if es.gravityTimer <= 0 {
		es.gravityTimer = es.fallRate
		es.GravityDrop()
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

	gameArea := Area{
		X:      rr.X + 8,
		Y:      rr.Y + 1,
		Width:  BOARD_WIDTH,
		Height: BOARD_HEIGHT,
	}

	nextPieceArea := Area{
		X: rr.X + BOARD_WIDTH + 12,
		Y: rr.Y + 1,
		Width: 4,
	}

	holdPieceArea := Area{
		X: rr.X + 1,
		Y: rr.Y + 1,
	}

	scoreArea := Area{
		X: nextPieceArea.Right() + 2,
		Y: rr.Y + 1,
	}

	es.DrawWell(gameArea)

	es.dashParticles.Draw(gameArea)
	gridOffsetX := es.cpGrid.Width / 2
	gridOffsetY := es.cpGrid.Height / 2

	// Snap indicators
	if es.moveMultiplier != 0 {
		es.DrawPiece(
			es.cpGrid,
			gameArea.X+es.leftSnapPosition-gridOffsetX,
			gameArea.Y+es.cpY-gridOffsetY,
			'*',
			LightPieceStyle(es.cpIdx),
		)
		es.DrawPiece(
			es.cpGrid,
			gameArea.X+es.rightSnapPosition-gridOffsetX,
			gameArea.Y+es.cpY-gridOffsetY,
			'*',
			LightPieceStyle(es.cpIdx),
		)

		// Hard drop snap indicators
		if es.leftSnapPosition != es.cpX {
			es.DrawPiece(
				es.cpGrid,
				gameArea.X+es.leftSnapPosition-gridOffsetX,
				gameArea.Y+es.hardDropLeftSnapHeight-gridOffsetY,
				'.',
				LightPieceStyle(es.cpIdx),
			)
		}

		if es.rightSnapPosition != es.cpX {
			es.DrawPiece(
				es.cpGrid,
				gameArea.X+es.rightSnapPosition-gridOffsetX,
				gameArea.Y+es.hardDropRightSnapHeight-gridOffsetY,
				'.',
				LightPieceStyle(es.cpIdx),
			)
		}
	}

	// Hard drop indicator
	es.DrawPiece(
		es.cpGrid,
		gameArea.X+es.cpX-gridOffsetX,
		gameArea.Y+es.hardDropHeight-gridOffsetY,
		'+',
		LightPieceStyle(es.cpIdx),
	)

	es.DrawPiece(
		es.cpGrid,
		gameArea.X+es.cpX-gridOffsetX,
		gameArea.Y+es.cpY-gridOffsetY,
		'o',
		SolidPieceStyle(es.cpIdx),
	)

	es.DrawGrid(gameArea)

	es.DrawNextPieces(nextPieceArea)
	es.DrawHoldPiece(holdPieceArea)
	es.DrawScore(scoreArea)
}

func (es *EngineState) DrawWell(rr Area) {
	for y := 0; y < es.grid.Height+1; y++ {
		Screen.SetContent(
			rr.X-1,
			rr.Y+y,
			'#',
			nil, defStyle)
		Screen.SetContent(
			rr.X+es.grid.Width,
			rr.Y+y,
			'#',
			nil, defStyle)
	}
	for xx := 0; xx < es.grid.Width; xx++ {
		Screen.SetContent(
			rr.X+xx,
			rr.Y+es.grid.Height,
			'#',
			nil, defStyle)
	}
}

func (es *EngineState) DrawPiece(
	piece Grid[bool],
	px, py int,
	rune rune,
	style tcell.Style,
) {

	for yy := 0; yy < piece.Height; yy++ {
		for xx := 0; xx < piece.Width; xx++ {
			if piece.MustGet(xx, yy) {
				Screen.SetContent(
					xx+px,
					yy+py,
					rune,
					nil, style)
			}
		}
	}
}

func (es *EngineState) DrawNextPieces(rr Area) {
	for i := 0; i < NUM_NEXT_PIECES; i++ {
		px := rr.X
		py := rr.Y + i*4
		es.DrawPiece(
			Pieces[es.nextPieces[i]][0],
			px, py,
			'o',
			SolidPieceStyle(es.nextPieces[i]))
	}
}

func (es *EngineState) DrawHoldPiece(rr Area) {
	if es.holdPiece != 8 {
		es.DrawPiece(
			Pieces[es.holdPiece][0],
			rr.X, rr.Y,
			'o',
			SolidPieceStyle(es.holdPiece))
	}
}

func (es *EngineState) DrawScore(rr Area) {
	SetString(
		rr.X,
		rr.Y,
		fmt.Sprintf("SCORE: %d", es.score),
		defStyle)
	SetString(
		rr.X,
		rr.Y+2,
		fmt.Sprintf("LINES: %d", es.lines),
		defStyle)

	if es.combo > 1 {
		SetString(
			rr.X,
			rr.Y+4,
			fmt.Sprintf("%dx COMBO", es.combo),
			defStyle)
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
	es.cpIdx = idx
	es.cpGrid = Pieces[idx][0]
	es.cpRot = 0
	es.cpX = BOARD_WIDTH / 2
	es.cpY = 2

	for i := 0; i < NUM_NEXT_PIECES-1; i++ {
		es.nextPieces[i] = es.nextPieces[i+1]
	}
	es.nextPieces[NUM_NEXT_PIECES-2] = es.pieceGenerator.NextPiece()

	es.SetHardDropHeight()
	es.moveMultiplier = 0
}

func (es *EngineState) SetMoveMultiplier(val int) {
	if val == 0 || val == es.moveMultiplier {
		es.moveMultiplier = 0
	} else if val == 10 {
		es.moveMultiplier = 10
	} else {
		es.moveMultiplier = val
	}

	if es.moveMultiplier != 0 {
		es.SetSnapPositions(es.moveMultiplier)
	}
}

func (es *EngineState) RotateCW() {
	rotationLength := len(Pieces[es.cpIdx])
	newRotation := (es.cpRot + 1) % rotationLength

	if es.CheckCollision(
		Pieces[es.cpIdx][newRotation],
		es.cpX,
		es.cpY,
	) {
		return
	}

	es.cpRot = newRotation
	es.cpGrid = Pieces[es.cpIdx][es.cpRot]
	es.SetHardDropHeight()

	if es.moveMultiplier != 0 {
		es.SetSnapPositions(es.moveMultiplier)
	}
}

func (es *EngineState) RotateCCW() {

	rotationLength := len(Pieces[es.cpIdx])
	newRotation := es.cpRot - 1

	if newRotation < 0 {
		newRotation = rotationLength - 1
	}

	if es.CheckCollision(
		Pieces[es.cpIdx][newRotation],
		es.cpX,
		es.cpY,
	) {
		return
	}

	es.cpRot = newRotation
	es.cpGrid = Pieces[es.cpIdx][es.cpRot]
	es.SetHardDropHeight()

	if es.moveMultiplier != 0 {
		es.SetSnapPositions(es.moveMultiplier)
	}
}

func (es *EngineState) HandleReset() {
	es.StartGame(time.Now().UnixNano())
}

func (es *EngineState) MovePiece(dx int) {
	if es.moveMultiplier != 0 {
		if dx < 0 {
			es.LeftDashParticles(
				es.cpGrid,
				es.cpIdx,
				es.cpY,
				es.cpX, es.leftSnapPosition)
			es.cpX = es.leftSnapPosition
		} else {
			es.RightDashParticles(
				es.cpGrid,
				es.cpIdx,
				es.cpY,
				es.cpX, es.rightSnapPosition)
			es.cpX = es.rightSnapPosition
		}

		es.moveMultiplier = 0
		es.SetHardDropHeight()
		return
	}
	if es.CheckCollision(
		es.cpGrid,
		es.cpX+dx,
		es.cpY,
	) {
		return
	}

	es.cpX += dx

	es.SetHardDropHeight()
}

func (es *EngineState) SoftDrop() {
	if es.moveMultiplier == 10 {
		es.DownDashParticles(
			es.cpGrid,
			es.cpIdx,
			es.cpX,
			es.cpY, es.hardDropHeight,
		)
		es.cpY = es.hardDropHeight
		es.moveMultiplier = 0
		es.gravityTimer = es.fallRate
		return
	}
	if es.CheckCollision(
		es.cpGrid,
		es.cpX,
		es.cpY+1,
	) {
		es.LockPiece()
		return
	}

	es.cpY += 1
	es.gravityTimer = es.fallRate
}

func (es *EngineState) GravityDrop() {
	if es.CheckCollision(
		es.cpGrid,
		es.cpX,
		es.cpY+1,
	) {
		es.LockPiece()
		return
	}

	es.cpY += 1

	if es.moveMultiplier != 0 {
		es.SetSnapPositions(es.moveMultiplier)
	}
}

func (es *EngineState) HardDrop() {
	es.DownDashParticles(
		es.cpGrid,
		es.cpIdx,
		es.cpX,
		es.cpY, es.hardDropHeight,
	)
	es.cpY = es.hardDropHeight
	es.LockPiece()
}

func (es *EngineState) SwapHoldPiece() {
	if es.usedHoldPiece {
		return
	}
	tmp := es.holdPiece
	es.holdPiece = es.cpIdx
	es.cpIdx = tmp
	if es.cpIdx == 8 {
		es.GetRandomPiece()
	} else {
		es.cpGrid = Pieces[es.cpIdx][0]
		es.cpRot = 0
		es.cpX = BOARD_WIDTH / 2
		es.cpY = 2
		es.SetHardDropHeight()
		es.moveMultiplier = 0
	}

	es.usedHoldPiece = true
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

func (es *EngineState) LockPiece() {
	gridOffsetX := es.cpGrid.Width / 2
	gridOffsetY := es.cpGrid.Height / 2

	for yy := 0; yy < es.cpGrid.Height; yy++ {
		for xx := 0; xx < es.cpGrid.Width; xx++ {
			if es.cpGrid.MustGet(xx, yy) {
				currX := xx - gridOffsetX + es.cpX
				currY := yy - gridOffsetY + es.cpY
				es.grid.Set(currX, currY, es.cpIdx+1)
			}
		}
	}

	es.usedHoldPiece = false
	es.ClearLines()
	es.GetRandomPiece()
}

func (es *EngineState) SetHardDropHeight() {
	yy := es.cpY
	for !es.CheckCollision(es.cpGrid, es.cpX, yy+1) {
		yy += 1
	}

	es.hardDropHeight = yy
}

func (es *EngineState) SetSnapPositions(distance int) {
	l := es.cpX
	r := es.cpX

	for lCount := 0; (distance == 10 || lCount < distance) &&
		!es.CheckCollision(es.cpGrid, l-1, es.cpY); lCount++ {
		l -= 1
	}

	for rCount := 0; (distance == 10 || rCount < distance) &&
		!es.CheckCollision(es.cpGrid, r+1, es.cpY); rCount++ {
		r += 1
	}

	es.leftSnapPosition = l
	es.rightSnapPosition = r

	ly := es.cpY
	ry := es.cpY
	for !es.CheckCollision(es.cpGrid, es.leftSnapPosition, ly+1) {
		ly++
	}
	for !es.CheckCollision(es.cpGrid, es.rightSnapPosition, ry+1) {
		ry++
	}

	es.hardDropLeftSnapHeight = ly
	es.hardDropRightSnapHeight = ry
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

	// For each line row found, pull all the tiles above it down.
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

	// Scoring
	es.lines += int64(len(lines))
	var lineScore int
	switch len(lines) {
	case 0:
		es.combo = 0
	case 1:
		es.combo += 1
		lineScore = SINGLE_SCORE
	case 2:
		es.combo += 1
		lineScore = DOUBLE_SCORE
	case 3:
		es.combo += 1
		lineScore = TRIPLE_SCORE
	default:
		es.combo += 1
		lineScore = TETRIS_SCORE
	}

	es.score += int64(lineScore)
	var comboCount int
	if comboCount >= len(COMBO_COUNTS) {
		comboCount = 5
	} else {
		comboCount = COMBO_COUNTS[es.combo]
	}

	es.score += int64(COMBO_BASE_SCORE * comboCount)
}

func (es *EngineState) DownDashParticles(
	piece Grid[bool],
	pieceIdx int,
	x int,
	initY, finY int,
) {
	gridOffsetX := piece.Width / 2
	gridOffsetY := piece.Height / 2

	blockTops := make([]int, piece.Width)
	for x := 0; x < piece.Width; x++ {
		found := false
	topFinder:
		for y := 0; y < piece.Height; y++ {
			if piece.MustGet(x, y) {
				blockTops[x] = y
				found = true
				break topFinder
			}
		}

		if !found {
			blockTops[x] = -1
		}
	}

	for z := finY - 1; z > initY; z-- {
		for dx, h := range blockTops {
			if h < 0 {
				continue
			}
			i := 1 - min(1, float32(finY-z-1)/15.0)
			es.dashParticles.SpawnParticle(
				Particle{
					Intensity: i * i * i,
					Style:     defStyle.Foreground(PieceColors[pieceIdx]),
					X:         dx + x - gridOffsetX,
					Y:         z + h - gridOffsetY,
				},
			)
		}
	}
}

func (es *EngineState) LeftDashParticles(
	piece Grid[bool],
	pieceIdx int,
	y int,
	initX, finX int,
) {
	gridOffsetX := piece.Width / 2
	gridOffsetY := piece.Height / 2

	blockRightEdges := make([]int, piece.Height)
	for y := 0; y < piece.Height; y++ {
		found := false
	topFinder:
		for x := piece.Width - 1; x >= 0; x-- {
			if piece.MustGet(x, y) {
				blockRightEdges[y] = x
				found = true
				break topFinder
			}
		}

		if !found {
			blockRightEdges[y] = -1
		}
	}

	for z := finX + 1; z <= initX; z++ {
		for dy, w := range blockRightEdges {
			if w < 0 {
				continue
			}
			i := 1 - min(1, float32(z-finX-1)/15.0)
			es.dashParticles.SpawnParticle(
				Particle{
					Intensity: i * i * i,
					Style:     defStyle.Foreground(PieceColors[pieceIdx]),
					X:         z + w - gridOffsetX,
					Y:         dy + y - gridOffsetY,
				},
			)
		}
	}
}

func (es *EngineState) RightDashParticles(
	piece Grid[bool],
	pieceIdx int,
	y int,
	initX, finX int,
) {
	gridOffsetX := piece.Width / 2
	gridOffsetY := piece.Height / 2

	blockLeftEdges := make([]int, piece.Width)
	for y := 0; y < piece.Height; y++ {
		found := false
	topFinder:
		for x := 0; x < piece.Width; x++ {
			if piece.MustGet(x, y) {
				blockLeftEdges[y] = x
				found = true
				break topFinder
			}
		}

		if !found {
			blockLeftEdges[y] = -1
		}
	}

	for z := finX - 1; z > initX; z-- {
		for dy, w := range blockLeftEdges {
			if w < 0 {
				continue
			}
			i := 1 - min(1, float32(finX-z-1)/15.0)
			es.dashParticles.SpawnParticle(
				Particle{
					Intensity: i * i * i,
					Style:     defStyle.Foreground(PieceColors[pieceIdx]),
					X:         z + w - gridOffsetX,
					Y:         dy + y - gridOffsetY,
				},
			)
		}
	}
}

package main

import (
	"math"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
)

const UPDATE_TICK_RATE_MS float64 = 1000.0 / 240.0

const BOARD_WIDTH = 10
const BOARD_HEIGHT = 22

const MIN_SPEED = 120
const MAX_SPEED = 20

const NUM_NEXT_PIECES = 5

const SINGLE_SCORE = 100
const DOUBLE_SCORE = 300
const TRIPLE_SCORE = 500
const TETRIS_SCORE = 800

const COMBO_BASE_SCORE = 50

const MAX_MOVE_RESETS = 15
const LOCK_DELAY = 120

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

	airborne     bool
	gravityTimer int
	fallRate     int
	lockTimer    int
	moveResets   int
	floorKicked  bool

	pieceGenerator PieceGenerator

	nextPieces []int

	dashParticles ParticleSystem

	holdPiece     int
	usedHoldPiece bool

	shiftMode               bool
	leftSnapPosition        int
	rightSnapPosition       int
	hardDropLeftSnapHeight  int
	hardDropRightSnapHeight int
	hardDropHeight          int

	score         int64
	lines         int64
	combo         int
	level         int64
	startingLevel int64
}

func NewEngineState() *EngineState {
	es := EngineState{
		LastUpdateDuration: UPDATE_TICK_RATE_MS,

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
	es.usedHoldPiece = false
	es.pieceGenerator = &gen

	es.gravityTimer = es.fallRate
	es.dashParticles = InitParticles(0.1)

	es.score = 0
	es.lines = 0
	es.combo = 0
	es.startingLevel = 1
	es.level = es.startingLevel
	speedFactor := int(min(14, es.level-1))
	es.fallRate =
		MIN_SPEED + speedFactor*(MAX_SPEED-MIN_SPEED)/14

	es.FillNextPieces()
	es.GetRandomPiece()
}

func (es *EngineState) HandleInput(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyUp ||
			IsRune(ev, 'k') || IsRune(ev, 'K') ||
			IsRune(ev, 'x') || IsRune(ev, 'X') {
			es.Rotate(1)
		}
		if IsRune(ev, 'z') || IsRune(ev, 'Z') {
			es.Rotate(-1)
		} else if IsRune(ev, 'f') || IsRune(ev, 'F') {
			es.ToggleShiftMode()
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

	if es.airborne {
		es.gravityTimer -= 1
		if es.gravityTimer <= 0 {
			es.gravityTimer = es.fallRate
			es.GravityDrop()
		}
	} else {
		// Locking
		es.lockTimer -= 1
		if es.lockTimer <= 0 {
			es.LockPiece()
			return
		}
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
		X:     rr.X + BOARD_WIDTH + 12,
		Y:     rr.Y + 1,
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
	if es.shiftMode {
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
	SetString(
		rr.X,
		rr.Y+4,
		fmt.Sprintf("LEVEL: %d", es.level),
		defStyle)

	if es.combo > 1 {
		SetString(
			rr.X,
			rr.Y+6,
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

func (es *EngineState) SetPiece(idx int) {
	es.cpIdx = idx
	es.cpGrid = Pieces[idx][0]
	es.cpRot = 0
	es.cpX = BOARD_WIDTH / 2
	es.cpY = 2

	es.SetHardDropHeight()
	es.SetAirborne()
	es.floorKicked = false
	es.shiftMode = false
}

func (es *EngineState) GetRandomPiece() {
	idx := es.nextPieces[0]

	es.SetPiece(idx)

	for i := 0; i < NUM_NEXT_PIECES-1; i++ {
		es.nextPieces[i] = es.nextPieces[i+1]
	}
	es.nextPieces[NUM_NEXT_PIECES-2] = es.pieceGenerator.NextPiece()
}

func (es *EngineState) ToggleShiftMode() {
	es.shiftMode = !es.shiftMode

	if es.shiftMode {
		es.SetSnapPositions()
	}
}

func (es *EngineState) Rotate(offset int) {
	newRotation := (es.cpRot + offset) % 4
	newRotation = (newRotation + 4) % 4

	offsets := GetOffsets(es.cpIdx, es.cpRot, newRotation)

	for _, os := range offsets {
		if es.CheckCollision(
			Pieces[es.cpIdx][newRotation],
			es.cpX+os.X,
			es.cpY+os.Y,
		) {
			continue
		}

		es.cpRot = newRotation
		es.cpX += os.X
		es.cpY += os.Y

		es.cpGrid = Pieces[es.cpIdx][es.cpRot]
		es.SetHardDropHeight()

		if es.shiftMode {
			es.SetSnapPositions()
		}

		oldAirborne := es.airborne
		newAirborne := !es.CheckCollision(es.cpGrid, es.cpX, es.cpY+1)

		// If you were previously not airborne, but now you are airborne,
		// count it as a floor kick. If the piece has already floor kicked,
		// soft-drop it back down.
		if !oldAirborne && newAirborne {
			if es.floorKicked {
				es.cpY = es.hardDropHeight
				es.gravityTimer = es.fallRate
				newAirborne = false
			}

			es.floorKicked = true
		}

		if !oldAirborne && !newAirborne && es.moveResets < MAX_MOVE_RESETS {
			es.moveResets += 1
			es.lockTimer = LOCK_DELAY
		}

		es.airborne = newAirborne
		return
	}

}

func (es *EngineState) HandleReset() {
	es.StartGame(time.Now().UnixNano())
}

func (es *EngineState) MovePiece(dx int) {
	if es.shiftMode {
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

		es.shiftMode = false
		es.SetHardDropHeight()
		es.SetAirborne()
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
	oldAirborne := es.airborne
	es.SetAirborne()

	if !oldAirborne && !es.airborne && es.moveResets < MAX_MOVE_RESETS {
		es.moveResets += 1
		es.lockTimer = LOCK_DELAY
	}
}

func (es *EngineState) SoftDrop() {
	if es.shiftMode {
		es.DownDashParticles(
			es.cpGrid,
			es.cpIdx,
			es.cpX,
			es.cpY, es.hardDropHeight,
		)
		es.cpY = es.hardDropHeight
		es.shiftMode = false
		es.gravityTimer = es.fallRate
		es.SetAirborne()
		return
	}

	if !es.airborne {
		es.LockPiece()
		return
	}

	es.cpY += 1
	es.gravityTimer = es.fallRate
	es.SetAirborne()
}

func (es *EngineState) GravityDrop() {
	if es.CheckCollision(
		es.cpGrid,
		es.cpX,
		es.cpY+1,
	) {
		return
	}

	es.cpY += 1

	if es.shiftMode {
		es.SetSnapPositions()
	}

	es.SetAirborne()
}

func (es *EngineState) HardDrop() {
	// es.DownDashParticles(
	// 	es.cpGrid,
	// 	es.cpIdx,
	// 	es.cpX,
	// 	es.cpY, es.hardDropHeight,
	// )
	es.DashParticles(
		es.cpGrid,
		es.cpIdx,
		es.cpX, es.cpY,
		es.cpX, es.hardDropHeight,
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
	if tmp == 8 {
		es.GetRandomPiece()
	} else {
		es.SetPiece(tmp)
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

func (es *EngineState) SetAirborne() {
	oldAirborne := es.airborne
	newAirborne := !es.CheckCollision(es.cpGrid, es.cpX, es.cpY+1)
	es.airborne = newAirborne
	// if you were previously in the air and now you aren't in the air,
	// start the lock timer
	if oldAirborne && !newAirborne {
		es.lockTimer = LOCK_DELAY
	}

	if newAirborne {
		es.moveResets = 0
	}
}

func (es *EngineState) SetHardDropHeight() {
	yy := es.cpY
	for !es.CheckCollision(es.cpGrid, es.cpX, yy+1) {
		yy += 1
	}

	es.hardDropHeight = yy
}

func (es *EngineState) SetSnapPositions() {
	l := es.cpX
	r := es.cpX

	for lCount := 0; !es.CheckCollision(es.cpGrid, l-1, es.cpY); lCount++ {
		l -= 1
	}

	for rCount := 0; !es.CheckCollision(es.cpGrid, r+1, es.cpY); rCount++ {
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

	// Lines and levels
	if len(lines) > 0 {
		es.lines += int64(len(lines))
		es.level = (es.lines / 10) + es.startingLevel
		speedFactor := int(min(14, es.level-1))
		es.fallRate =
			MIN_SPEED + speedFactor*(MAX_SPEED-MIN_SPEED)/14
	}

	// Scoring
	var lineScore int64
	switch len(lines) {
	case 0:
		es.combo = 0
	case 1:
		es.combo += 1
		lineScore = SINGLE_SCORE * es.level
	case 2:
		es.combo += 1
		lineScore = DOUBLE_SCORE * es.level
	case 3:
		es.combo += 1
		lineScore = TRIPLE_SCORE * es.level
	default:
		es.combo += 1
		lineScore = TETRIS_SCORE * es.level
	}

	es.score += lineScore
	var comboCount int
	if comboCount >= len(COMBO_COUNTS) {
		comboCount = 5
	} else {
		comboCount = COMBO_COUNTS[es.combo]
	}

	es.score += int64(COMBO_BASE_SCORE*comboCount) * es.level
}

var dashParticleData = MakeGrid(BOARD_WIDTH, BOARD_HEIGHT, 0.0)
func (es *EngineState) DashParticles(
	piece Grid[bool],
	pieceIdx int,
	initX, initY int,
	finX, finY int,
) {
	// Reset dash particle data
	for y := 0; y < dashParticleData.Height; y++ {
		for x := 0; x < dashParticleData.Width; x++ {
			dashParticleData.Set(x, y, 0.0)	
		}	
	}

	type DashTile struct {
		X int
		Y int
		Intensity float64
	}

	gridOffsetX := piece.Width / 2
	gridOffsetY := piece.Height / 2

	distance := math.Hypot(float64(initX - finX), float64(initY - finY))

	deltaX := float64(initX - finX)/distance
	deltaY := float64(initY - finY)/distance

	curX := float64(finX)
	curY := float64(finY)

	for {
		t := float64(i)/49.0

		strength := 1 - min(1, t*distance/15.0)
		strength = math.Pow(strength, 3)

		curX := float64(finX) + t * float64(initX - finX)
		curY := float64(finY) + t * float64(initY - finY)

		touchingTiles := make([]DashTile, 0, 4)

		floorX := math.Floor(curX)
		floorY := math.Floor(curY)

		vertDiff := curY - floorY
		horizDiff := curX - floorX

		touchingTiles = append(touchingTiles, DashTile{
			X: int(curX),
			Y: int(curY),
			Intensity: (1 - vertDiff) * (1 - horizDiff),
		})

		if vertDiff > 0.1 {
			touchingTiles = append(touchingTiles, DashTile{
				X: int(curX),
				Y: int(curY) + 1,
				Intensity: vertDiff * (1 - horizDiff),
			})
		}

		if horizDiff > 0.1 {
			touchingTiles = append(touchingTiles, DashTile{
				X: int(curX) + 1,
				Y: int(curY),
				Intensity: (1 - vertDiff) * (1 - horizDiff),
			})
		}

		if vertDiff > 0.1 && horizDiff > 0.1 {
			touchingTiles = append(touchingTiles, DashTile{
				X: int(curX) + 1,
				Y: int(curY) + 1,
				Intensity: vertDiff * horizDiff,
			})
		}

		for py := 0; py < piece.Height; py++ {
			for px := 0; px < piece.Width; px++ {
				if !piece.MustGet(px, py) { continue }
				for _, tile := range touchingTiles {
					posX := tile.X + px - gridOffsetX
					posY := tile.Y + py - gridOffsetY
					val, ok := dashParticleData.Get(posX, posY)
					if !ok { continue }
					dashParticleData.Set(
						posX,
						posY,
						val + tile.Intensity * strength)
				}
			}	
		}
	}

	// Reset dash particle data
	for y := 0; y < dashParticleData.Height; y++ {
		for x := 0; x < dashParticleData.Width; x++ {
			strength := dashParticleData.MustGet(x, y)
			es.dashParticles.SpawnParticle(
				Particle{
					Intensity: strength,
					Style: defStyle.Foreground(PieceColors[pieceIdx]),
					X: x,
					Y: y,
				},
			)
		}	
	}
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
			i := 1 - min(1, float64(finY-z-1)/15.0)
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
			i := 1 - min(1, float64(z-finX-1)/15.0)
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
			i := 1 - min(1, float64(finX-z-1)/15.0)
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

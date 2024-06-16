package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
)

const UPDATE_TICK_RATE_MS float64 = 1000.0 / 60.0

const BOARD_WIDTH = 10
const BOARD_HEIGHT = 20

const MIN_SPEED = 30
const MAX_SPEED = 5

const NUM_NEXT_PIECES = 5

const SINGLE_SCORE = 100
const DOUBLE_SCORE = 300
const TRIPLE_SCORE = 500
const TETRIS_SCORE = 800

const COMBO_BASE_SCORE = 50

const MAX_MOVE_RESETS = 15
const LOCK_DELAY = 30

var GAME_OVER_PIECE_STYLE = defStyle.Background(tcell.ColorBlack).Foreground(tcell.ColorGray)

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

type LineClearHandler func(garbage, nonGarbage int)

type TetrisField struct {
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
	pieceCount    int64
	combo         int
	level         int64
	startingLevel int64
	frameCount    int64

	garbageRng   *rand.Rand
	garbageQueue []int

	lineClearHandlers []LineClearHandler
	gameOver bool
}

func NewTetrisField(startingLevel int64) *TetrisField {
	es := TetrisField{
		LastUpdateDuration: UPDATE_TICK_RATE_MS,

		nextPieces: make([]int, NUM_NEXT_PIECES),
		holdPiece:  8,
		lineClearHandlers: make([]LineClearHandler, 0),
	}

	es.StartGame(time.Now().UnixNano(), startingLevel)

	return &es
}

func (es *TetrisField) StartGame(seed int64, startingLevel int64) {
	gen := NewBagRandomizer(seed, 1)
	es.grid = MakeGrid(BOARD_WIDTH, BOARD_HEIGHT*2, 0)
	es.holdPiece = 8
	es.usedHoldPiece = false
	es.pieceGenerator = &gen

	es.score = 0
	es.lines = 0
	es.combo = 0
	es.startingLevel = startingLevel
	es.level = es.startingLevel
	speedFactor := int(min(14, es.level-1))
	es.fallRate =
		MIN_SPEED + speedFactor*(MAX_SPEED-MIN_SPEED)/14

	es.gravityTimer = es.fallRate
	es.dashParticles = InitParticles(0.1)
	es.frameCount = 0
	es.pieceCount = 0

	es.gameOver = false

	es.garbageRng = rand.New(rand.NewSource(seed))
	es.garbageQueue = make([]int, 0, 20)

	es.FillNextPieces()
	es.GetRandomPiece()
}

func (es *TetrisField) HandleInput(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if IsRune(ev, 'r') || IsRune(ev, 'R') {
			es.HandleReset()
		}
		if !es.gameOver {
			if ev.Key() == tcell.KeyUp ||
				IsRune(ev, 'w') || IsRune(ev, 'W') ||
				IsRune(ev, 'x') || IsRune(ev, 'X') {
				es.Rotate(1)
			}
			if IsRune(ev, 'z') || IsRune(ev, 'Z') {
				es.Rotate(-1)
			}
			if IsRune(ev, 'f') || IsRune(ev, 'F') {
				es.ToggleShiftMode()
			}
			if ev.Key() == tcell.KeyDown || IsRune(ev, 's') || IsRune(ev, 'S') {
				es.SoftDrop()
			}
			if ev.Key() == tcell.KeyLeft || IsRune(ev, 'a') || IsRune(ev, 'A') {
				es.MovePiece(-1)
			}
			if ev.Key() == tcell.KeyRight || IsRune(ev, 'd') || IsRune(ev, 'D') {
				es.MovePiece(1)
			}
			if IsRune(ev, ' ') {
				es.HardDrop()
			}
			if IsRune(ev, ';') || IsRune(ev, 'c') || IsRune(ev, 'C') {
				es.SwapHoldPiece()
			}
		}
	}
}

func (es *TetrisField) Update() {
	if es.gameOver {
		return
	}

	es.frameCount++

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

func (es *TetrisField) Draw(sw, sh int, rr Area, lag float64) {
	gameArea := Area{
		X:      rr.X,
		Y:      rr.Y + 2,
		Width:  BOARD_WIDTH,
		Height: BOARD_HEIGHT,
	}

	nextPieceArea := Area{
		X:     rr.X + BOARD_WIDTH + 4,
		Y:     rr.Y + 2,
		Width: 4,
	}

	holdPieceArea := Area{
		X: rr.X - 6,
		Y: rr.Y + 2,
	}

	scoreArea := Area{
		X: nextPieceArea.Right() + 2,
		Y: rr.Y + 1,
	}

	es.DrawWell(gameArea)

	if !es.gameOver {
		es.dashParticles.Draw(gameArea)
	}

	// Snap indicators
	if es.shiftMode && !es.gameOver {
		es.DrawPiece(
			es.cpGrid,
			gameArea.X+es.leftSnapPosition,
			gameArea.Y+es.cpY-BOARD_HEIGHT,
			'*',
			LightPieceStyle(es.cpIdx),
		)
		es.DrawPiece(
			es.cpGrid,
			gameArea.X+es.rightSnapPosition,
			gameArea.Y+es.cpY-BOARD_HEIGHT,
			'*',
			LightPieceStyle(es.cpIdx),
		)

		// Hard drop snap indicators
		if es.leftSnapPosition != es.cpX {
			es.DrawPiece(
				es.cpGrid,
				gameArea.X+es.leftSnapPosition,
				gameArea.Y+es.hardDropLeftSnapHeight-BOARD_HEIGHT,
				'.',
				LightPieceStyle(es.cpIdx),
			)
		}

		if es.rightSnapPosition != es.cpX {
			es.DrawPiece(
				es.cpGrid,
				gameArea.X+es.rightSnapPosition,
				gameArea.Y+es.hardDropRightSnapHeight-BOARD_HEIGHT,
				'.',
				LightPieceStyle(es.cpIdx),
			)
		}
	}

	// Hard drop indicator
	if !es.gameOver {
		es.DrawPiece(
			es.cpGrid,
			gameArea.X+es.cpX,
			gameArea.Y+es.hardDropHeight-BOARD_HEIGHT,
			'+',
			LightPieceStyle(es.cpIdx),
		)
	}

	var pieceStyle tcell.Style
	if es.gameOver {
		pieceStyle = GAME_OVER_PIECE_STYLE
	} else {
		pieceStyle =
			SolidPieceStyle(es.cpIdx)
	}
	es.DrawPiece(
		es.cpGrid,
		gameArea.X+es.cpX,
		gameArea.Y+es.cpY-BOARD_HEIGHT,
		'o',
		pieceStyle,
	)

	es.DrawGrid(gameArea)

	es.DrawNextPieces(nextPieceArea)
	es.DrawHoldPiece(holdPieceArea)
	es.DrawScore(scoreArea)

	if es.gameOver {
		es.DrawGameOver(gameArea)
	}
}

func (es *TetrisField) DrawWell(rr Area) {
	for y := 0; y < rr.Height+1; y++ {
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
			rr.Y+BOARD_HEIGHT,
			'#',
			nil, defStyle)
		Screen.SetContent(
			rr.X+xx,
			rr.Y,
			'.',
			nil, defStyle)
	}
}

func (es *TetrisField) DrawPiece(
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

func (es *TetrisField) DrawNextPieces(rr Area) {
	SetString(
		rr.X,
		rr.Y-1,
		"NEXT",
		defStyle)
	for i := 0; i < NUM_NEXT_PIECES; i++ {
		piece := Pieces[es.nextPieces[i]][0]
		gridOffsetX := piece.Width/2 + 1
		gridOffsetY := piece.Height/2 + 1

		var pieceStyle tcell.Style
		if es.gameOver {
			pieceStyle = GAME_OVER_PIECE_STYLE
		} else {
			pieceStyle =
				SolidPieceStyle(es.nextPieces[i])
		}

		px := rr.X - gridOffsetX + 2
		py := rr.Y + (i+1)*4 - gridOffsetY - 1
		es.DrawPiece(
			piece,
			px, py,
			'o',
			pieceStyle)
	}
}

func (es *TetrisField) DrawHoldPiece(rr Area) {
	SetString(
		rr.X,
		rr.Y-1,
		"HOLD",
		defStyle)
	if es.holdPiece != 8 {
		var pieceStyle tcell.Style
		if es.gameOver || es.usedHoldPiece {
			pieceStyle = GAME_OVER_PIECE_STYLE
		} else {
			pieceStyle =
				SolidPieceStyle(es.holdPiece)
		}

		piece := Pieces[es.holdPiece][0]
		gridOffsetX := piece.Width/2 + 1
		gridOffsetY := piece.Height/2 + 1

		es.DrawPiece(
			piece,
			rr.X-gridOffsetX+2, rr.Y-gridOffsetY+3,
			'o',
			pieceStyle)
	}
}

func (es *TetrisField) DrawScore(rr Area) {
	if es.combo > 1 {
		SetString(
			rr.X,
			rr.Y+6,
			fmt.Sprintf("%dx COMBO", es.combo),
			defStyle)
	}
}

func (es *TetrisField) DrawGrid(rr Area) {
	for yy := BOARD_HEIGHT; yy < es.grid.Height; yy++ {
		for xx := 0; xx < es.grid.Width; xx++ {
			if es.grid.MustGet(xx, yy) != 0 {
				color := PieceColors[es.grid.MustGet(xx, yy)-1]
				var style tcell.Style
				if es.gameOver {
					style = GAME_OVER_PIECE_STYLE
				} else {
					style =
						defStyle.Background(color).Foreground(tcell.ColorBlack)
				}
				Screen.SetContent(
					rr.X+xx,
					rr.Y+yy-BOARD_HEIGHT,
					'o',
					nil, style)
			}
		}
	}
}

func (es *TetrisField) DrawGameOver(rr Area) {
	subArea := rr.Inset(rr.Width, 4)
	for xx := rr.Left(); xx < rr.Right(); xx++ {
		Screen.SetContent(
			xx,
			subArea.Top(),
			'-',
			nil, defStyle)
		Screen.SetContent(
			xx,
			subArea.Bottom()-1,
			'-',
			nil, defStyle)
	}

	FillRegion(
		subArea.X,
		subArea.Y+1,
		subArea.Width,
		subArea.Height-2,
		' ', defStyle)

	SetCenteredString(
		subArea.X+subArea.Width/2,
		subArea.Y+1,
		"GAME",
		defStyle)

	SetCenteredString(
		subArea.X+subArea.Width/2,
		subArea.Y+2,
		"OVER",
		defStyle)
}

func (es *TetrisField) FillNextPieces() {
	for i := 0; i < NUM_NEXT_PIECES; i++ {
		es.nextPieces[i] = es.pieceGenerator.NextPiece()
	}
}

func (es *TetrisField) SetPiece(idx int) {
	es.cpIdx = idx
	es.cpGrid = Pieces[idx][0]
	es.cpRot = 0

	gridOffsetX := es.cpGrid.Width/2 + 1
	gridOffsetY := es.cpGrid.Height/2 + 1

	es.cpX = BOARD_WIDTH/2 - gridOffsetX
	es.cpY = BOARD_HEIGHT - gridOffsetY

	es.airborne = true
	es.SetHardDropHeight()
	es.SetAirborne()
	es.floorKicked = false
	es.shiftMode = false
}

func (es *TetrisField) GetRandomPiece() {
	idx := es.nextPieces[0]

	es.SetPiece(idx)

	for i := 0; i < NUM_NEXT_PIECES-1; i++ {
		es.nextPieces[i] = es.nextPieces[i+1]
	}
	es.nextPieces[NUM_NEXT_PIECES-2] = es.pieceGenerator.NextPiece()
}

func (es *TetrisField) ToggleShiftMode() {
	es.shiftMode = !es.shiftMode

	if es.shiftMode {
		es.SetSnapPositions()
	}
}

func (es *TetrisField) Rotate(offset int) {
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

		// If you are now on the ground after being airborne, start the lock
		// timer.
		if oldAirborne && !newAirborne {
			es.lockTimer = LOCK_DELAY
		}

		es.airborne = newAirborne
		return
	}

}

func (es *TetrisField) HandleReset() {
	es.StartGame(time.Now().UnixNano(), es.startingLevel)
}

func (es *TetrisField) MovePiece(dx int) {
	if es.shiftMode {
		if dx < 0 {
			es.DashParticles(
				es.cpGrid,
				es.cpIdx,
				es.cpX, es.cpY,
				es.leftSnapPosition, es.cpY)
			es.cpX = es.leftSnapPosition
		} else {
			es.DashParticles(
				es.cpGrid,
				es.cpIdx,
				es.cpX, es.cpY,
				es.rightSnapPosition, es.cpY)
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

func (es *TetrisField) SoftDrop() {
	if es.shiftMode {
		es.DashParticles(
			es.cpGrid,
			es.cpIdx,
			es.cpX, es.cpY,
			es.cpX, es.hardDropHeight)
		es.score += int64(es.hardDropHeight) - int64(es.cpY)
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
	es.score += 1
	es.gravityTimer = es.fallRate
	es.SetAirborne()
}

func (es *TetrisField) GravityDrop() {
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

func (es *TetrisField) HardDrop() {
	es.DashParticles(
		es.cpGrid,
		es.cpIdx,
		es.cpX, es.cpY,
		es.cpX, es.hardDropHeight,
	)
	es.score += 2 * (int64(es.hardDropHeight) - int64(es.cpY))
	es.cpY = es.hardDropHeight
	es.LockPiece()
}

func (es *TetrisField) SwapHoldPiece() {
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

func (es *TetrisField) CheckCollision(piece Grid[bool], px, py int) bool {
	for yy := 0; yy < piece.Height; yy++ {
		for xx := 0; xx < piece.Width; xx++ {
			if piece.MustGet(xx, yy) {
				currX := xx + px
				currY := yy + py
				cell, ok := es.grid.Get(currX, currY)
				if !ok || cell != 0 {
					return true
				}
			}
		}
	}

	return false
}

func (es *TetrisField) LockPiece() {
	for yy := 0; yy < es.cpGrid.Height; yy++ {
		for xx := 0; xx < es.cpGrid.Width; xx++ {
			if es.cpGrid.MustGet(xx, yy) {
				currX := xx + es.cpX
				currY := yy + es.cpY
				es.grid.Set(currX, currY, es.cpIdx+1)
			}
		}
	}

	es.pieceCount++

	es.usedHoldPiece = false
	es.ClearLines()

	if len(es.garbageQueue) > 0 {
		for _, gb := range es.garbageQueue {
			es.AddGarbage(gb)
		}
		es.garbageQueue = make([]int, 0, 20)

		es.SetHardDropHeight()
	}

	// Check for a game over
	pieceOverTop := false
	for x := 0; x < es.grid.Width; x++ {
		height := 0
		for y := es.grid.Height - 1; y >= 0; y-- {
			if es.grid.MustGet(x, y) != 0 {
				height = es.grid.Height - y
			}
		}
		if height > BOARD_HEIGHT {
			pieceOverTop = true
			break
		}
	}

	if pieceOverTop {
		es.gameOver = true
		return
	}

	es.GetRandomPiece()
}

func (es *TetrisField) SetAirborne() {
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

func (es *TetrisField) SetHardDropHeight() {
	yy := es.cpY
	for !es.CheckCollision(es.cpGrid, es.cpX, yy+1) {
		yy += 1
	}

	es.hardDropHeight = yy
}

func (es *TetrisField) SetSnapPositions() {
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

func (es *TetrisField) QueueGarbage(count int) {
	es.garbageQueue = append(es.garbageQueue, count)
}

func (es *TetrisField) AddGarbage(count int) {
	col := es.garbageRng.Intn(BOARD_WIDTH)
	for y := 0; y < es.grid.Height; y++ {
		for x := 0; x < es.grid.Width; x++ {
			if es.grid.Height-y-1 < count {
				var value int
				if x == col {
					value = 0
				} else {
					value = 8
				}
				es.grid.Set(x, y, value)
			} else {
				es.grid.Set(x, y, es.grid.MustGet(x, y+count))
			}
		}
	}

	es.SetHardDropHeight()

	if es.shiftMode {
		es.SetSnapPositions()
	}

	es.SetAirborne()
}

func (es *TetrisField) ClearLines() {
	lines := make([]int, 0)
	var garbage, nonGarbage int
	for y := 0; y < es.grid.Height; y++ {
		fullLine := true
		isGarbage := false
	notFullLine:
		for x := 0; x < es.grid.Width; x++ {
			if es.grid.MustGet(x, y) == 0 {
				fullLine = false
				break notFullLine
			} else if !isGarbage && es.grid.MustGet(x, y) == 8 {
				isGarbage = true
			}
		}

		if fullLine {
			lines = append(lines, y)
			if isGarbage {
				garbage++
			} else {
				nonGarbage++
			}
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

	for _, handle := range es.lineClearHandlers {
		handle(garbage, nonGarbage)
	}
}

func (es *TetrisField) DrawStats(rr Area, anchorX, anchorY int) {
	rawTime := float64(es.frameCount) * UPDATE_TICK_RATE_MS
	timeMinutes := math.Trunc(rawTime / (60 * 1000))
	timeSeconds := math.Trunc((rawTime - timeMinutes*60*1000) / 1000)
	timeMillis := math.Trunc((rawTime - timeMinutes*60*1000 -
		timeSeconds*1000))

	timeString := fmt.Sprintf(
		"%0d:%02d.%03d",
		int64(timeMinutes),
		int64(timeSeconds),
		int64(timeMillis),
	)

	SetStringArray(
		anchorX,
		anchorY-1,
		defStyle,
		true,
		"TIME",
		timeString)

	pieceCountString := fmt.Sprintf(
		"%d",
		es.pieceCount,
	)

	piecesPerSecondString := fmt.Sprintf(
		"%.2f p/s",
		float64(es.pieceCount)/(rawTime/(1000)),
	)

	SetStringArray(
		anchorX,
		anchorY-5,
		defStyle,
		true,
		"PIECES",
		pieceCountString,
		piecesPerSecondString)

	// Draw lines & lines per second
	linesString := fmt.Sprintf(
		"%d",
		es.lines)

	linesPerMinuteString := fmt.Sprintf(
		"%.2f l/m",
		float64(es.lines)/(rawTime/(1000*60)))

	SetStringArray(
		anchorX,
		anchorY-9,
		defStyle,
		true,
		"LINES",
		linesString,
		linesPerMinuteString)

	// Draw score and level

	centerBottomAnchorX := rr.Left() + rr.Width/2
	centerBottomAnchorY := rr.Bottom() - 1
	SetCenteredString(
		centerBottomAnchorX,
		centerBottomAnchorY,
		fmt.Sprintf("%d", es.score),
		defStyle)
	SetCenteredString(
		centerBottomAnchorX,
		centerBottomAnchorY+1,
		fmt.Sprintf("%d", es.level),
		defStyle)
}

var dashParticleData = MakeGrid(BOARD_WIDTH, BOARD_HEIGHT, 0.0)

func (es *TetrisField) DashParticles(
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

	distance := math.Hypot(float64(initX-finX), float64(initY-finY))

	deltaX := float64(finX-initX) / distance
	deltaY := float64(finY-initY) / distance

	prevFloorX := -1
	prevFloorY := -1

	t := 0.0
	done := false

	for !done {
		if t-distance > 0.01 {
			t = distance
			done = true
		}

		f := t / distance
		strength := 1 - min(1, (1-f)*distance/20.0)
		strength = math.Pow(strength, 3)

		curX := float64(initX) + t*deltaX
		curY := float64(initY) + t*deltaY

		floorX := int(math.Floor(curX))
		floorY := int(math.Floor(curY))

		// First case
		dirty := true
		var stamp Grid[bool]
		if prevFloorX == -1 || prevFloorY == -1 {
			stamp = piece
		} else if floorX != prevFloorX || floorY != prevFloorY {
			stamp = ShiftedDifference(piece, floorX-prevFloorX, floorY-prevFloorY)
		} else {
			dirty = false
		}

		if dirty {
			for py := 0; py < stamp.Height; py++ {
				for px := 0; px < stamp.Width; px++ {
					if !piece.MustGet(px, py) {
						continue
					}
					posX := floorX + px
					posY := floorY + py
					dashParticleData.Set(
						posX,
						posY-BOARD_HEIGHT,
						strength)
				}
			}
		}

		prevFloorX = floorX
		prevFloorY = floorY
		t += 1
	}

	// Reset dash particle data
	for y := 0; y < dashParticleData.Height; y++ {
		for x := 0; x < dashParticleData.Width; x++ {
			strength := dashParticleData.MustGet(x, y)
			es.dashParticles.SpawnParticle(
				Particle{
					Intensity: strength,
					Style:     defStyle.Foreground(PieceColors[pieceIdx]),
					X:         x,
					Y:         y,
				},
			)
		}
	}
}

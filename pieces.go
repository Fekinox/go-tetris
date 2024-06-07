package main

import "github.com/gdamore/tcell/v2"

var (
	IPieces = []Grid[bool]{
		GridFromSlice(5, 5,
			false, false, false, false, false,
			false, false, false, false, false,
			false, true, true, true, true,
			false, false, false, false, false,
			false, false, false, false, false,
		),
		GridFromSlice(5, 5,
			false, false, false, false, false,
			false, false, true, false, false,
			false, false, true, false, false,
			false, false, true, false, false,
			false, false, true, false, false,
		),
		GridFromSlice(5, 5,
			false, false, false, false, false,
			false, false, false, false, false,
			true, true, true, true, false,
			false, false, false, false, false,
			false, false, false, false, false,
		),
		GridFromSlice(5, 5,
			false, false, true, false, false,
			false, false, true, false, false,
			false, false, true, false, false,
			false, false, true, false, false,
			false, false, false, false, false,
		),
	}
	JPieces = []Grid[bool]{
		GridFromSlice(3, 3,
			true, false, false,
			true, true, true,
			false, false, false,
		),
		GridFromSlice(3, 3,
			false, true, true,
			false, true, false,
			false, true, false,
		),
		GridFromSlice(3, 3,
			false, false, false,
			true, true, true,
			false, false, true,
		),
		GridFromSlice(3, 3,
			false, true, false,
			false, true, false,
			true, true, false,
		),
	}
	LPieces = []Grid[bool]{
		GridFromSlice(3, 3,
			false, false, true,
			true, true, true,
			false, false, false,
		),
		GridFromSlice(3, 3,
			false, true, false,
			false, true, false,
			false, true, true,
		),
		GridFromSlice(3, 3,
			false, false, false,
			true, true, true,
			true, false, false,
		),
		GridFromSlice(3, 3,
			true, true, false,
			false, true, false,
			false, true, false,
		),
	}
	OPieces = []Grid[bool]{
		GridFromSlice(3, 3,
			false, true, true,
			false, true, true,
			false, false, false,
		),
		GridFromSlice(3, 3,
			false, false, false,
			false, true, true,
			false, true, true,
		),
		GridFromSlice(3, 3,
			false, false, false,
			true, true, false,
			true, true, false,
		),
		GridFromSlice(3, 3,
			true, true, false,
			true, true, false,
			false, false, false,
		),
	}
	SPieces = []Grid[bool]{
		GridFromSlice(3, 3,
			false, true, true,
			true, true, false,
			false, false, false,
		),
		GridFromSlice(3, 3,
			false, true, false,
			false, true, true,
			false, false, true,
		),
		GridFromSlice(3, 3,
			false, false, false,
			false, true, true,
			true, true, false,
		),
		GridFromSlice(3, 3,
			true, false, false,
			true, true, false,
			false, true, false,
		),
	}
	TPieces = []Grid[bool]{
		GridFromSlice(3, 3,
			false, false, false,
			true, true, true,
			false, true, false,
		),
		GridFromSlice(3, 3,
			false, true, false,
			true, true, false,
			false, true, false,
		),
		GridFromSlice(3, 3,
			false, true, false,
			true, true, true,
			false, false, false,
		),
		GridFromSlice(3, 3,
			false, true, false,
			false, true, true,
			false, true, false,
		),
	}
	ZPieces = []Grid[bool]{
		GridFromSlice(3, 3,
			true, true, false,
			false, true, true,
			false, false, false,
		),
		GridFromSlice(3, 3,
			false, false, true,
			false, true, true,
			false, true, false,
		),
		GridFromSlice(3, 3,
			false, false, false,
			true, true, false,
			false, true, true,
		),
		GridFromSlice(3, 3,
			false, true, false,
			true, true, false,
			true, false, false,
		),
	}
)

var JLSTZOffsets = [][]Position{
	[]Position{
		{},
		{},
		{},
		{},
	},
	[]Position{
		{},
		{X: 1},
		{},
		{X: -1},
	},
	[]Position{
		{},
		{X: 1, Y: 1},
		{},
		{X: -1, Y: 1},
	},
	[]Position{
		{},
		{Y: -2},
		{},
		{Y: -2},
	},
	[]Position{
		{},
		{X: 1, Y: -2},
		{},
		{X: -1, Y: -2},
	},
}

var IOffsets = [][]Position{
	[]Position{
		{},
		{X: -1},
		{X: -1, Y: -1},
		{Y: -1},
	},
	[]Position{
		{X: -1},
		{},
		{X: 1, Y: -1},
		{Y: -1},
	},
	[]Position{
		{X: 2},
		{},
		{X: -2, Y: -1},
		{Y: -1},
	},
	[]Position{
		{X: -1},
		{Y: -1},
		{X: 1},
		{Y: 1},
	},
	[]Position{
		{X: 2},
		{Y: 2},
		{X: -2},
		{Y: -2},
	},
}

var OOffsets = [][]Position{
	[]Position{
		{},
		{Y: 1},
		{X: -1, Y: 1},
		{X: -1},
	},
}

func GetOffsets(pieceIdx int, startRot int, endRot int) []Position {
	positions := make([]Position, 0)
	var offsetData [][]Position
	if pieceIdx == 0 {
		offsetData = IOffsets
	} else if pieceIdx == 3 {
		offsetData = OOffsets
	} else {
		offsetData = JLSTZOffsets
	}

	for _, offsetCol := range offsetData {
		positions = append(positions, Position{
			X: offsetCol[startRot].X - offsetCol[endRot].X,
			Y: offsetCol[startRot].Y - offsetCol[endRot].Y,
		})
	}

	return positions
}

var Pieces = [][]Grid[bool]{
	IPieces,
	JPieces,
	LPieces,
	OPieces,
	SPieces,
	TPieces,
	ZPieces,
}

var PieceColors = []tcell.Color{
	// I
	tcell.ColorAqua,
	// J
	tcell.ColorBlue,
	// L
	tcell.ColorWhite,
	// O
	tcell.ColorYellow,
	// S
	tcell.ColorLime,
	// T
	tcell.ColorFuchsia,
	// Z
	tcell.ColorRed,
}

func SolidPieceStyle(pieceIdx int) tcell.Style {
	return defStyle.Foreground(tcell.ColorBlack).Background(PieceColors[pieceIdx])
}

func LightPieceStyle(pieceIdx int) tcell.Style {
	return defStyle.Foreground(PieceColors[pieceIdx])
}

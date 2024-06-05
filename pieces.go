package main

import "github.com/gdamore/tcell/v2"

var (
	IPieces = []Grid[bool]{
		GridFromSlice(4, 4,
			false, false, false, false,
			true, true, true, true,
			false, false, false, false,
			false, false, false, false,
		),
		GridFromSlice(4, 4,
			false, false, true, false,
			false, false, true, false,
			false, false, true, false,
			false, false, true, false,
		),
	}
	JPieces = []Grid[bool]{
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
		GridFromSlice(3, 3,
			false, false, false,
			true, false, false,
			true, true, true,
		),
		GridFromSlice(3, 3,
			false, true, true,
			false, true, false,
			false, true, false,
		),
	}
	LPieces = []Grid[bool]{
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
		GridFromSlice(3, 3,
			false, false, false,
			false, false, true,
			true, true, true,
		),
		GridFromSlice(3, 3,
			false, true, false,
			false, true, false,
			false, true, true,
		),
	}
	OPieces = []Grid[bool]{
		GridFromSlice(4, 4,
			false, false, false, false,
			false, true, true, false,
			false, true, true, false,
			false, false, false, false,
		),
	}
	SPieces = []Grid[bool]{
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
			false, false, false,
			true, true, false,
			false, true, true,
		),
		GridFromSlice(3, 3,
			false, false, true,
			false, true, true,
			false, true, false,
		),
	}
)

var Pieces = [][]Grid[bool]{
	IPieces,
	JPieces,
	ZPieces,
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
	tcell.ColorOlive,
	// O
	tcell.ColorYellow,
	// S
	tcell.ColorLime,
	// T
	tcell.ColorPurple,
	// Z
	tcell.ColorRed,
}

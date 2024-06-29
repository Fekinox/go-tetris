package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type FormField struct {
	Name  string
	Field EditableField
}

type EditableField interface {
	HandleInput(tcell.Event)
	HandleAction(Action)
	Draw(x, y int, editing bool)
}

type BooleanField struct {
	Value    bool
	OnChange func(value bool)
}

func (bf *BooleanField) SetValue(value bool) {
	bf.Value = value
	bf.OnChange(bf.Value)
}

func (bf *BooleanField) HandleInput(evt tcell.Event) {
}

func (bf *BooleanField) HandleAction(act Action) {
}

func (bf *BooleanField) Draw(x, y int, editing bool) {
	if bf.Value {
		SetString(
			x, y,
			"X",
			defStyle,
		)
	} else {
		SetString(
			x, y,
			"_",
			defStyle,
		)
	}
}

type IntegerField struct {
	Value  int64
	HasMin bool
	Min    int64
	HasMax bool
	Max    int64

	CurrentValue string
	Cursor       int
	Valid        bool
	OnChange     func(value int64)
}

type IntegerFieldOption func(nf *IntegerField) *IntegerField

func WithMin(value int64) IntegerFieldOption {
	return func(nf *IntegerField) *IntegerField {
		nf.HasMin = true
		nf.Min = value
		return nf
	}
}

func WithMax(value int64) IntegerFieldOption {
	return func(nf *IntegerField) *IntegerField {
		nf.HasMax = true
		nf.Max = value
		return nf
	}
}

func (nf *IntegerField) UpdateValue() {
	// Remove leading/trailing whitespace
	nf.CurrentValue = strings.TrimSpace(nf.CurrentValue)
	// Remove leading zeroes
	nf.CurrentValue = strings.TrimLeft(nf.CurrentValue, "0")

	nf.Valid = true

	for _, r := range nf.CurrentValue {
		if r < '0' || r > '9' {
			nf.Valid = false
			break
		}
	}

	nf.Valid = nf.Valid && len(nf.CurrentValue) > 0

	i, err := strconv.ParseInt(nf.CurrentValue, 10, 64)
	if err != nil {
		nf.Valid = false
		return
	}

	nf.Value = i

	if nf.HasMin && nf.Value < nf.Min {
		nf.Value = nf.Min
	}

	if nf.HasMax && nf.Value > nf.Max {
		nf.Value = nf.Max
	}
	nf.CurrentValue = fmt.Sprint(nf.Value)

	nf.OnChange(nf.Value)
}

func (nf *IntegerField) HandleInput(evt tcell.Event) {
	switch evt := evt.(type) {
	case *tcell.EventKey:
		switch evt.Key() {
		case tcell.KeyRune:
			if evt.Rune() < '0' || evt.Rune() > '9' {
				return
			}
			if evt.Rune() == '0' && nf.Cursor == 0 && len(nf.CurrentValue) > 0 {
				return
			}
			runeArr := []rune(nf.CurrentValue)
			nf.CurrentValue = strings.Join(
				[]string{
					string(runeArr[:nf.Cursor]),
					string(evt.Rune()),
					string(runeArr[nf.Cursor:]),
				}, "",
			)
			nf.UpdateValue()

			nf.Cursor = min(
				len([]rune(nf.CurrentValue)),
				nf.Cursor+1,
			)
		case tcell.KeyLeft:
			nf.Cursor = max(0, nf.Cursor-1)
		case tcell.KeyRight:
			nf.Cursor = min(
				len([]rune(nf.CurrentValue)),
				nf.Cursor+1,
			)
		case tcell.KeyBackspace:
			fallthrough
		case tcell.KeyBackspace2:
			if nf.Cursor == 0 || len(nf.CurrentValue) == 0 {
				return
			}
			runeArr := []rune(nf.CurrentValue)
			nf.CurrentValue = strings.Join(
				[]string{
					string(runeArr[:nf.Cursor-1]),
					string(runeArr[nf.Cursor:]),
				}, "",
			)
			nf.UpdateValue()

			nf.Cursor = max(0, nf.Cursor-1)
		case tcell.KeyDelete:
			if len(nf.CurrentValue) == 0 || nf.Cursor == len(nf.CurrentValue) {
				return
			}
			runeArr := []rune(nf.CurrentValue)

			nf.CurrentValue = strings.Join(
				[]string{
					string(runeArr[:nf.Cursor-1]),
					string(runeArr[nf.Cursor:]),
				}, "",
			)
			nf.UpdateValue()

			nf.Cursor = min(
				len([]rune(nf.CurrentValue)),
				nf.Cursor,
			)
		}
	}
}

func (nf *IntegerField) HandleAction(act Action) {
}

func (nf *IntegerField) Draw(x, y int, editing bool) {
	SetString(
		x, y,
		nf.CurrentValue,
		defStyle,
	)

	if !nf.Valid {
		style := defStyle.Foreground(tcell.ColorRed)
		for i := 0; i < max(1, runewidth.StringWidth(nf.CurrentValue)); i++ {
			Screen.SetContent(
				x+i,
				y+1,
				'-',
				nil, style,
			)
		}
	}

	if editing {
		cell, cm, style, _ := Screen.GetContent(x+nf.Cursor, y)
		Screen.SetContent(
			x+nf.Cursor,
			y,
			cell, cm, style.Reverse(true),
		)
	}
}

func NewBooleanField(
	name string,
	value bool,
	onChange func(value bool),
) FormField {
	return FormField{
		Name: name,
		Field: &BooleanField{
			Value:    value,
			OnChange: onChange,
		},
	}
}

func NewIntegerField(
	name string, value int64, onChange func(value int64),
	options ...IntegerFieldOption,
) FormField {
	field := &IntegerField{
		Value:        value,
		OnChange:     onChange,
		CurrentValue: fmt.Sprintf("%v", value),
		Valid:        true,
	}

	field.Cursor = len(field.CurrentValue)

	for _, opt := range options {
		field = opt(field)
	}

	return FormField{
		Name:  name,
		Field: field,
	}
}

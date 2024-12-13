package arcade

import (
	"image/color"
)

const (
	SIZE = 64
)

type Screen struct {
	Cells [SIZE][SIZE]color.Color
}

type State struct {
	Won    bool
	Lose   bool
	Screen Screen
}

type Arcade interface {
	Feed(input byte) error
	State() State
}

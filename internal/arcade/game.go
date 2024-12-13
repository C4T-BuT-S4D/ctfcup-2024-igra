package arcade

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenSize = 64
)

type Result int

const (
	ResultUnknown Result = iota
	ResultWon
	ResultLost
)

type State struct {
	Won    bool
	Result Result
	Screen [ScreenSize][ScreenSize]color.Color
}

type Game interface {
	Start() error
	Stop() error
	Feed([]ebiten.Key) error
	State() *State
}

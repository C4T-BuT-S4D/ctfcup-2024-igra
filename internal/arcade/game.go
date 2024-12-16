package arcade

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	gameserverpb "github.com/c4t-but-s4d/ctfcup-2024-igra/proto/go/gameserver"
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
	Result Result
	Screen [ScreenSize][ScreenSize]color.Color
}

type Game interface {
	Start() error
	Stop() error
	Feed([]ebiten.Key) error
	State() *State
}

func (s *State) ToProto() *gameserverpb.ArcadeState {
	screen := make([]*gameserverpb.Color, ScreenSize*ScreenSize)
	for i := 0; i < len(s.Screen); i++ {
		for j := 0; j < len(s.Screen[i]); j++ {
			r, g, b, a := s.Screen[i][j].RGBA()
			screen[i*ScreenSize+j] = &gameserverpb.Color{
				R: r,
				G: g,
				B: b,
				A: a,
			}
		}
	}
	return &gameserverpb.ArcadeState{
		Result: gameserverpb.ArcadeState_Result(s.Result),
		Screen: screen,
	}
}

func StateFromProto(s *gameserverpb.ArcadeState) *State {
	state := State{
		Result: Result(s.Result),
		Screen: [ScreenSize][ScreenSize]color.Color{},
	}
	for i, val := range s.Screen {
		state.Screen[i/ScreenSize][i%ScreenSize] = color.RGBA{
			R: uint8(val.R),
			G: uint8(val.G),
			B: uint8(val.B),
			A: uint8(val.A),
		}
	}

	return &state
}

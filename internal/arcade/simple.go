package arcade

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
)

type cell int

const (
	empty  cell = iota
	player      = iota
	enemy       = iota
	finish      = iota
)

type move int

const (
	nop move = iota
	up
	down
	left
	right
)

func newSimpleGame() *Simple {
	var s Simple
	return &s
}

type Simple struct {
	state [ScreenSize][ScreenSize]cell
	lost  bool
	won   bool
}

func (s *Simple) Start() error {
	s.reset()
	return nil
}

func (s *Simple) Stop() error {
	return nil
}

func (s *Simple) Feed(keys []ebiten.Key) error {
	var m move
	if len(keys) == 0 {
		m = nop
	} else {
		m = s.toMove(keys[0])
	}
	s.step(m)
	return nil
}

func (s *Simple) reset() {
	// empty the state.
	s.state = [ScreenSize][ScreenSize]cell{}
	for i := 0; i < ScreenSize; i += 2 {
		s.state[0][i] = enemy
	}
	for i := 1; i < ScreenSize; i += 2 {
		s.state[1][i] = enemy
	}

	s.state[ScreenSize/2][ScreenSize-1] = finish
	s.state[ScreenSize/4][0] = player
}

func (s *Simple) step(m move) {
	for row := ScreenSize - 1; row >= 0; row-- {
		for col := 0; col < ScreenSize; col++ {
			switch s.state[row][col] {
			case enemy:
				// move enemy
				s.state[row][col] = empty
				if row != ScreenSize-1 {
					if s.state[row+1][col] == player {
						s.lost = true
					}
					s.state[row+1][col] = enemy
				}
			case player:
				// move player
				x, y := col, row
				switch m {
				case up:
					y--
				case down:
					y++
				case left:
					x--
				case right:
					x++
				default:
					// nop
				}
				x = max(0, min(ScreenSize-1, x))
				y = max(0, min(ScreenSize-1, y))
				if s.state[y][x] == enemy {
					s.lost = true
				}
				if s.state[y][x] == finish {
					s.won = true
				}
				s.state[row][col], s.state[y][x] = empty, player
			default:
				// nop
			}
		}
	}
}

func (s *Simple) toMove(key ebiten.Key) move {
	switch key {
	case ebiten.KeyA:
		return left
	case ebiten.KeyD:
		return right
	case ebiten.KeyW:
		return up
	case ebiten.KeyS:
		return down
	default:
		return nop
	}
}

func (s *Simple) State() *State {
	var state State
	state.Won = s.won
	if s.won {
		state.Result = ResultWon
	}
	if s.lost {
		state.Result = ResultLost
	}
	for i := 0; i < ScreenSize; i++ {
		for j := 0; j < ScreenSize; j++ {
			switch s.state[i][j] {
			case player:
				state.Screen[i][j] = color.RGBA{0, 0, 255, 255}
			case enemy:
				state.Screen[i][j] = color.RGBA{255, 0, 0, 255}
			case finish:
				state.Screen[i][j] = color.RGBA{0, 255, 0, 255}
			default:
				state.Screen[i][j] = color.RGBA{0, 0, 0, 0}
			}
		}
	}

	return &state
}

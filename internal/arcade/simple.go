package arcade

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
)

type cell int

const (
	empty  cell = iota
	enemy       = iota
	player      = iota
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

type coord struct {
	x, y int
}

func newSimpleGame() *Simple {
	var s Simple
	return &s
}

type Simple struct {
	state     [ScreenSize][ScreenSize]cell
	nextState [ScreenSize][ScreenSize]cell
	lost      bool
	won       bool
	finish    coord
	player    coord
}

func (s *Simple) Start() error {
	s.reset()
	return nil
}

func (s *Simple) Stop() error {
	return nil
}

func (s *Simple) Feed(keys []ebiten.Key) error {
	for row := 0; row < ScreenSize; row++ {
		for col := 0; col < ScreenSize; col++ {
			if s.state[row][col] == enemy {
				// move enemy
				s.nextState[row][col] = empty
				if row != ScreenSize-1 {
					s.nextState[row+1][col] = enemy
				}
			}
		}
	}

	x, y := s.player.x, s.player.y
	for _, key := range keys {
		m := s.toMove(key)
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
	}
	x = max(0, min(ScreenSize-1, x))
	y = max(0, min(ScreenSize-1, y))

	if s.nextState[y][x] == enemy {
		s.lost = true
	}
	if s.finish == s.player {
		s.won = true
	}
	s.nextState[s.player.y][s.player.x] = empty
	s.player = coord{x, y}
	s.nextState[s.player.y][s.player.x] = player
	s.nextState[s.finish.y][s.finish.x] = finish
	s.state = s.nextState
	return nil
}

func (s *Simple) reset() {
	s.won = false
	s.lost = false
	// empty the state.
	s.state = [ScreenSize][ScreenSize]cell{}
	for i := 0; i < ScreenSize; i += 2 {
		s.state[0][i] = enemy
	}
	for i := 1; i < ScreenSize; i += 2 {
		s.state[2][i] = enemy
	}

	s.player = coord{0, ScreenSize / 2}
	s.finish = coord{ScreenSize - 1, ScreenSize / 4}
	s.nextState = s.state
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

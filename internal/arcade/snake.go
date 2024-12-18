package arcade

import (
	"image/color"
	"math/rand"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	SnakeScreenH = 8
	SnakeScreenW = 8
)

type snakeCell int

const (
	snakeEmpty snakeCell = iota
	snakeSnake           = iota
	snakeApple           = iota
)

type snakeCoord struct {
	x, y int
}

type snakeDir int

const (
	snakeUp    snakeDir = iota
	snakeDown           = iota
	snakeLeft           = iota
	snakeRight          = iota
	snakeNop            = iota
)

func newSnakeGame() *Snake {
	return &Snake{}
}

type Snake struct {
	screen   [SnakeScreenH][SnakeScreenW]snakeCell
	snake    []snakeCoord
	apple    *snakeCoord
	won      bool
	snakeDir snakeDir
	lost     bool
}

func (s *Snake) Start() error {
	s.reset()
	return nil
}

func (s *Snake) Stop() error {
	return nil
}

func (s *Snake) spawnApple() {
	emptyCells := make([]snakeCoord, 0)
	for y := 0; y < SnakeScreenH; y++ {
		for x := 0; x < SnakeScreenW; x++ {
			if s.screen[y][x] == snakeEmpty {
				emptyCells = append(emptyCells, snakeCoord{x, y})
			}
		}
	}

	s.apple = &emptyCells[rand.Intn(len(emptyCells))]
	s.screen[s.apple.y][s.apple.x] = snakeApple
}

func (s *Snake) Feed(keys []ebiten.Key) error {
	if s.lost || s.won {
		return nil
	}
	for _, key := range keys {
		d := s.toDir(key)
		if d != snakeNop {
			s.snakeDir = d
		}
	}

	var delta snakeCoord
	switch s.snakeDir {
	case snakeUp:
		delta = snakeCoord{
			x: 0,
			y: -1,
		}
	case snakeDown:
		delta = snakeCoord{
			x: 0,
			y: +1,
		}
	case snakeLeft:
		delta = snakeCoord{
			x: -1,
			y: 0,
		}
	case snakeRight:
		delta = snakeCoord{
			x: +1,
			y: 0,
		}
	}
	newCoord := snakeCoord{
		x: s.snake[len(s.snake)-1].x + delta.x,
		y: s.snake[len(s.snake)-1].y + delta.y,
	}

	if newCoord.x < 0 || newCoord.x >= SnakeScreenW || newCoord.y < 0 || newCoord.y >= SnakeScreenH {
		s.lost = true
		return nil
	}

	if s.screen[newCoord.y][newCoord.x] != snakeApple {
		s.screen[s.snake[0].y][s.snake[0].x] = snakeEmpty
		if s.screen[newCoord.y][newCoord.x] == snakeSnake {
			s.lost = true
			return nil
		}
		s.snake = s.snake[1:]
	} else {
		s.apple = nil
	}
	s.snake = append(s.snake, newCoord)
	s.screen[newCoord.y][newCoord.x] = snakeSnake
	if len(s.snake) == SnakeScreenH*SnakeScreenW {
		s.won = true
		return nil
	}

	if s.apple == nil {
		s.spawnApple()
	}

	return nil
}

func (s *Snake) reset() {
	s.lost = false
	s.won = false
	s.snake = []snakeCoord{
		{0, 0},
		{1, 0},
		{2, 0},
		{3, 0},
	}

	for y := 0; y < SnakeScreenH; y++ {
		for x := 0; x < SnakeScreenW; x++ {
			if slices.Contains(s.snake, snakeCoord{x, y}) {
				s.screen[y][x] = snakeSnake
			} else {
				s.screen[y][x] = snakeEmpty
			}
		}
	}
	s.snakeDir = snakeRight
	s.spawnApple()
}

func (s *Snake) toDir(key ebiten.Key) snakeDir {
	switch key {
	case ebiten.KeyA:
		return snakeLeft
	case ebiten.KeyD:
		return snakeRight
	case ebiten.KeyW:
		return snakeUp
	case ebiten.KeyS:
		return snakeDown
	default:
		return snakeNop
	}
}

func (s *Snake) State() *State {
	var state State
	if s.won {
		state.Result = ResultWon
	}
	if s.lost {
		state.Result = ResultLost
	}
	for y := 0; y < ScreenSize; y++ {
		for x := 0; x < ScreenSize; x++ {
			switch s.screen[y*SnakeScreenH/ScreenSize][x*SnakeScreenW/ScreenSize] {
			case snakeSnake:
				state.Screen[y][x] = color.RGBA{0, 240, 100, 255}
			case snakeApple:
				state.Screen[y][x] = color.RGBA{255, 0, 0, 255}
			case snakeEmpty:
				state.Screen[y][x] = color.RGBA{255, 255, 255, 255}
			default:
				state.Screen[y][x] = color.RGBA{255, 255, 255, 255}
			}
		}
	}

	return &state
}

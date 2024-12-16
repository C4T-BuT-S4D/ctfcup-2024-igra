package arcade

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
	"image/color"
)

func NewClientGame() *ClientGame {
	return &ClientGame{
		s: &State{},
	}
}

type ClientGame struct {
	s *State
}

func (c *ClientGame) Start() error {
	logrus.Info("Starting client game")
	c.s = &State{
		Screen: [ScreenSize][ScreenSize]color.Color{},
		Result: ResultUnknown,
	}
	// Fill the screen with back color
	for i := 0; i < ScreenSize; i++ {
		for j := 0; j < ScreenSize; j++ {
			c.s.Screen[i][j] = color.Black
		}
	}
	return nil
}

func (c *ClientGame) Stop() error {
	return nil
}

func (c *ClientGame) Feed(keys []ebiten.Key) error {
	return nil
}

func (c *ClientGame) State() *State {
	return c.s
}

func (c *ClientGame) SetState(s *State) {
	c.s = s
}

package arcade

import (
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/item"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
	"github.com/hajimehoshi/ebiten/v2"
)

type Machine struct {
	*object.Base
	Game         Game          `msgpack:"-"`
	Image        *ebiten.Image `msgpack:"-"`
	LinkedItem   *item.Item    `msgpack:"-"`
	ProvidesItem string
}

func (m *Machine) Type() object.Type {
	return object.Arcade
}

func New(origin *geometry.Point, img *ebiten.Image, width, height float64, game Game, item string) *Machine {
	return &Machine{
		Base: &object.Base{
			Origin: origin,
			Width:  width,
			Height: height,
		},
		Image:        img,
		Game:         game,
		ProvidesItem: item,
	}
}

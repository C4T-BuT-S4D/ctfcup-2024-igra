package arcade

import (
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/item"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
	"github.com/hajimehoshi/ebiten/v2"
)

type Machine struct {
	*object.Rendered
	Game         Game       `msgpack:"-"`
	LinkedItem   *item.Item `msgpack:"-"`
	ProvidesItem string
}

func New(origin *geometry.Point, img *ebiten.Image, width, height float64, game Game, item string) *Machine {
	return &Machine{
		Rendered:     object.NewRendered(origin, img, width, height),
		ProvidesItem: item,
		Game:         game,
	}
}

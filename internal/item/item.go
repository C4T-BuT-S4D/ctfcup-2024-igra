package item

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
)

type Item struct {
	*object.Rendered `json:"-"`

	Name      string `json:"name"`
	Important bool   `json:"important"`
	Collected bool   `json:"collected"`
}

func New(origin geometry.Point, width, height float64, img *ebiten.Image, name string, important bool) *Item {
	return &Item{
		Rendered:  object.NewRendered(origin, img, width, height),
		Name:      name,
		Important: important,
	}
}

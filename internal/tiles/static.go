package tiles

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
)

type StaticTile struct {
	*object.Object

	Image *ebiten.Image
}

func NewStaticTile(origin *geometry.Point, width, height int, image *ebiten.Image) *StaticTile {
	return &StaticTile{
		Object: &object.Object{
			Origin: origin,
			Width:  float64(width),
			Height: float64(height),
		},
		Image: image,
	}
}

func (s *StaticTile) Type() object.Type {
	return object.StaticTileType
}

package tiles

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
)

type StaticTile struct {
	*object.Rendered
}

func NewStaticTile(origin *geometry.Point, width, height int, image *ebiten.Image) *StaticTile {
	return &StaticTile{
		Rendered: object.NewRendered(origin, image, float64(width), float64(height)),
	}
}

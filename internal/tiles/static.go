package tiles

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
)

type Flips struct {
	Horizontal bool
	Vertical   bool
	Diagonal   bool
}

type StaticTile struct {
	*object.Rendered
	Flips Flips
}

func NewStaticTile(origin geometry.Point, width, height int, image *ebiten.Image, flips Flips) *StaticTile {
	return &StaticTile{
		Rendered: object.NewRendered(origin, image, float64(width), float64(height)),
		Flips:    flips,
	}
}

func (t *StaticTile) ApplyFlips(op *ebiten.DrawImageOptions) {
	// Yes, if's, not else-if's. Do not question this.
	if t.Flips.Horizontal {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(t.Width, 0)
	}
	if t.Flips.Vertical {
		op.GeoM.Scale(1, -1)
		op.GeoM.Translate(0, t.Height)
	}
	if t.Flips.Diagonal {
		op.GeoM.Rotate(-math.Pi / 2)
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(t.Width, t.Height)
	}
}

package damage

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
)

type Spike struct {
	*object.Rendered
	Damageable
}

func NewSpike(origin *geometry.Point, img *ebiten.Image, width, height float64) *Spike {
	return &Spike{
		Rendered:   object.NewRendered(origin, img, width, height),
		Damageable: NewDamageable(100),
	}
}

package damage

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
)

type Spike struct {
	*object.Base
	Damageable

	Image *ebiten.Image `msgpack:"-"`
}

func (s *Spike) Type() object.Type {
	return object.Spike
}

func NewSpike(origin *geometry.Point, img *ebiten.Image, width, height float64) *Spike {
	return &Spike{
		Base: &object.Base{
			Origin: origin,
			Width:  width,
			Height: height,
		},
		Image:      img,
		Damageable: NewDamageable(100),
	}
}

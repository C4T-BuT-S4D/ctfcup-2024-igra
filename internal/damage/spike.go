package damage

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/physics"
)

type Spike struct {
	*physics.MovingObject
	Damageable
}

func NewSpike(origin geometry.Point, img *ebiten.Image, width, height float64) *Spike {
	return &Spike{
		MovingObject: physics.NewMovingObject(origin, width, height, img, physics.PathHorizontal, 0, 0),
		Damageable:   NewDamageable(100),
	}
}

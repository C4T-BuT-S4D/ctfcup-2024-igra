package platform

import (
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/physics"
	"github.com/hajimehoshi/ebiten/v2"
)

type Platform struct {
	*physics.MovingObject

	disableCollisions bool
}

func New(origin geometry.Point, width, height float64, image *ebiten.Image, path physics.MovementPath, distance, speed int) *Platform {
	return &Platform{
		MovingObject: physics.NewMovingObject(origin, width, height, image, path, distance, speed),
	}
}

func (p *Platform) CollisionsDisabled() bool {
	return p.disableCollisions
}

func (p *Platform) DisableCollisions(b bool) {
	p.disableCollisions = b
}

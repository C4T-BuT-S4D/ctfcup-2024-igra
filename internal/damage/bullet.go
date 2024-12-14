package damage

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
)

const (
	BulletWidth  = 1
	BulletHeight = 1
)

type Bullet struct {
	*object.Rendered
	Damageable `msgpack:"-"`
	Direction  *geometry.Vector
	Triggered  bool
}

func NewBullet(origin *geometry.Point, img *ebiten.Image, damage int, direction *geometry.Vector) *Bullet {
	return &Bullet{
		Rendered:   object.NewRendered(origin, img, BulletWidth, BulletHeight),
		Damageable: NewDamageable(damage),
		Direction:  direction,
	}
}

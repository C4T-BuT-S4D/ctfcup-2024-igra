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
	*object.Base
	Image      *ebiten.Image `msgpack:"-"`
	Damageable `msgpack:"-"`
	Direction  *geometry.Vector
	Triggered  bool
}

func (e *Bullet) Type() object.Type {
	return object.EnemyBullet
}

func NewBullet(origin *geometry.Point, img *ebiten.Image, damage int, direction *geometry.Vector) *Bullet {
	return &Bullet{
		Base: &object.Base{
			Origin: origin,
			Width:  BulletWidth,
			Height: BulletHeight,
		},
		Image:      img,
		Damageable: NewDamageable(damage),
		Direction:  direction,
	}
}

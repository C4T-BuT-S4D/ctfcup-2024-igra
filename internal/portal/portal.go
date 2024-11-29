package portal

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
)

type Portal struct {
	*object.Object
	Image      *ebiten.Image `msgpack:"-"`
	PortalTo   string
	TeleportTo *geometry.Point
	Boss       string
}

func (p *Portal) Type() object.Type {
	return object.Portal
}

func New(origin *geometry.Point, img *ebiten.Image, width, height float64, portalTo string, teleportTo *geometry.Point, boss string) *Portal {
	return &Portal{
		Object: &object.Object{
			Origin: origin,
			Width:  width,
			Height: height,
		},
		PortalTo:   portalTo,
		TeleportTo: teleportTo,
		Image:      img,
		Boss:       boss,
	}
}

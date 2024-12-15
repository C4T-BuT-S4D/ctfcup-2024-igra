package portal

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
)

type Portal struct {
	*object.Rendered
	PortalTo   string
	TeleportTo geometry.Point
	Boss       string
}

func New(origin geometry.Point, img *ebiten.Image, width, height float64, portalTo string, boss string) *Portal {
	return &Portal{
		Rendered: object.NewRendered(origin, img, width, height),
		PortalTo: portalTo,
		Boss:     boss,
	}
}

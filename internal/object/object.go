package object

import (
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/hajimehoshi/ebiten/v2"
)

type Base struct {
	Origin *geometry.Point
	Width  float64
	Height float64
}

func (b *Base) GetOrigin() *geometry.Point {
	if b == nil {
		return nil
	}
	return b.Origin
}

func (b *Base) Rectangle() *geometry.Rectangle {
	return &geometry.Rectangle{
		LeftX:   b.Origin.X,
		TopY:    b.Origin.Y,
		RightX:  b.Origin.X + b.Width,
		BottomY: b.Origin.Y + b.Height,
	}
}

func (b *Base) Move(d *geometry.Vector) *Base {
	b.Origin = b.Origin.Add(d)
	return b
}

func (b *Base) MoveTo(p *geometry.Point) *Base {
	b.Origin = p
	return b
}

type Rendered struct {
	*Base
	StaticImage *ebiten.Image `msgpack:"-"`
}

func NewRendered(origin *geometry.Point, img *ebiten.Image, width, height float64) *Rendered {
	return &Rendered{
		Base: &Base{
			Origin: origin,
			Width:  width,
			Height: height,
		},
		StaticImage: img,
	}
}

func (r *Rendered) Image() *ebiten.Image {
	return r.StaticImage
}

type Collidable interface {
	Rectangle() *geometry.Rectangle
}

type Drawable interface {
	Image() *ebiten.Image
}

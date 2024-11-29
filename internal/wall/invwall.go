package wall

import (
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
)

type InvWall struct {
	*object.Object
}

func NewInvWall(origin *geometry.Point, width, height float64) *InvWall {
	return &InvWall{
		Object: &object.Object{
			Origin: origin,
			Width:  width,
			Height: height,
		},
	}
}

func (s *InvWall) Type() object.Type {
	return object.InvWall
}

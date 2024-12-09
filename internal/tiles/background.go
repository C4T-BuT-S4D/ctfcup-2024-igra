package tiles

import (
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
)

type BackgroundImage struct {
	StaticTile
}

func (s *BackgroundImage) Type() object.Type {
	return object.BackgroundImage
}

package resources

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type TileBundle struct {
	*imageBundle
}

func newTileBundle() *TileBundle {
	return &TileBundle{imageBundle: newImageBundle()}
}

func (tb *TileBundle) GetTile(path string) *ebiten.Image {
	return tb.getImage(fmt.Sprintf("tiles/%s", path))
}

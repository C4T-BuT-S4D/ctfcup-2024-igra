package resources

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Bundle struct {
	*SpriteBundle
	*TileBundle
	*FontBundle
	*MusicBundle
}

func NewBundle(withMusic bool) *Bundle {
	b := &Bundle{
		SpriteBundle: newSpriteBundle(),
		TileBundle:   newTileBundle(),
		FontBundle:   newFontBundle(),
	}

	if withMusic {
		b.MusicBundle = newMusicBundle()
	}

	return b
}

type imageBundle struct {
	cache map[string]*ebiten.Image
	m     sync.Mutex
}

func newImageBundle() *imageBundle {
	return &imageBundle{cache: make(map[string]*ebiten.Image)}
}

func (ib *imageBundle) getImage(path string) *ebiten.Image {
	ib.m.Lock()
	defer ib.m.Unlock()

	if sprite, ok := ib.cache[path]; ok {
		return sprite
	}

	eimg, _, err := ebitenutil.NewImageFromFileSystem(EmbeddedFS, path)
	if err != nil {
		panic(err)
	}

	ib.cache[path] = eimg
	return eimg
}

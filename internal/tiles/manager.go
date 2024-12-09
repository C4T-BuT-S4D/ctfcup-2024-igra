package tiles

import (
	"fmt"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/resources"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"sync"
)

type Manager struct {
	cache map[string]*ebiten.Image
	m     sync.Mutex
}

func (m *Manager) getImage(path string) *ebiten.Image {
	m.m.Lock()
	defer m.m.Unlock()

	if sprite, ok := m.cache[path]; ok {
		return sprite
	}

	eimg, _, err := ebitenutil.NewImageFromFileSystem(resources.EmbeddedFS, path)
	if err != nil {
		panic(err)
	}

	m.cache[path] = eimg
	return eimg
}

func (m *Manager) Get(path string) *ebiten.Image {
	return m.getImage(fmt.Sprintf("tiles/%s", path))
}

func NewManager() *Manager {
	return &Manager{
		cache: make(map[string]*ebiten.Image),
	}
}

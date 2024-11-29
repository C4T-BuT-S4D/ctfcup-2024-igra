package sprites

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/resources"

	// Register PNG decoder.
	_ "image/png"
	"sync"
)

type Manager struct {
	cache map[string]*ebiten.Image
	m     sync.Mutex
}

func (m *Manager) getSprite(path string) *ebiten.Image {
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

func (m *Manager) GetSprite(spriteType Type) *ebiten.Image {
	return m.getSprite(fmt.Sprintf("sprites/%s.png", spriteType))
}

func (m *Manager) GetAnimationSprite(spriteType Type, animation string) *ebiten.Image {
	return m.getSprite(fmt.Sprintf("sprites/%s_%s.png", spriteType, animation))
}

func NewManager() *Manager {
	return &Manager{
		cache: make(map[string]*ebiten.Image),
	}
}

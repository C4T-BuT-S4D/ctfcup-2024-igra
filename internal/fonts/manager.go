package fonts

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/resources"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Manager struct {
	cache map[Type]text.Face
	m     sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		cache: make(map[Type]text.Face),
	}
}

func (m *Manager) Get(t Type) text.Face {
	m.m.Lock()
	defer m.m.Unlock()

	if face, ok := m.cache[t]; ok {
		return face
	}

	f, err := resources.EmbeddedFS.ReadFile(fmt.Sprintf("fonts/%s", t))
	if err != nil {
		panic(err)
	}

	source, err := text.NewGoTextFaceSource(bytes.NewReader(f))
	if err != nil {
		panic(err)
	}

	face := &text.GoTextFace{
		Source:    source,
		Direction: text.DirectionLeftToRight,
		Size:      72,
	}

	if t == Dialog {
		face.Size = 24
	}

	m.cache[t] = face
	return face
}

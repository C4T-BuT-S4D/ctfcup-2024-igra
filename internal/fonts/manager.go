package fonts

import (
	"fmt"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/resources"
)

type Manager struct {
	cache map[Type]font.Face
	m     sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		cache: make(map[Type]font.Face),
	}
}

func (m *Manager) Get(t Type) font.Face {
	m.m.Lock()
	defer m.m.Unlock()

	if face, ok := m.cache[t]; ok {
		return face
	}

	f, err := resources.EmbeddedFS.ReadFile(fmt.Sprintf("fonts/%s", t))
	if err != nil {
		panic(err)
	}

	ff, err := opentype.Parse(f)
	if err != nil {
		panic(err)
	}

	opts := &opentype.FaceOptions{
		Size:    72,
		DPI:     72,
		Hinting: font.HintingFull,
	}
	if t == Dialog {
		opts.Size = 24
	}
	face, err := opentype.NewFace(ff, opts)
	if err != nil {
		panic(err)
	}

	m.cache[t] = face
	return face
}

package resources

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type FontType string

const (
	FontSouls  FontType = "DSOULS.ttf"
	FontDialog FontType = "Dialog.ttf"
)

type FontBundle struct {
	cache map[FontType]text.Face
	m     sync.Mutex
}

func newFontBundle() *FontBundle {
	return &FontBundle{cache: make(map[FontType]text.Face)}
}

func (m *FontBundle) GetFontFace(t FontType) text.Face {
	m.m.Lock()
	defer m.m.Unlock()

	if face, ok := m.cache[t]; ok {
		return face
	}

	f, err := EmbeddedFS.ReadFile(fmt.Sprintf("fonts/%s", t))
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

	if t == FontDialog {
		face.Size = 24
	}

	m.cache[t] = face
	return face
}

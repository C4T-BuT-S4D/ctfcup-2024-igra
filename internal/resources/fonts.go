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
	cache map[string]text.Face
	m     sync.Mutex
}

func newFontBundle() *FontBundle {
	return &FontBundle{cache: make(map[string]text.Face)}
}

func (m *FontBundle) GetFontFace(t FontType, size float64) text.Face {
	m.m.Lock()
	defer m.m.Unlock()
	cacheKey := fmt.Sprintf("%s-%.2f", t, size)

	if face, ok := m.cache[cacheKey]; ok {
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
		//Size:      72,
		Size: size,
	}

	//if t == FontDialog {
	//	face.Size = 24
	//}

	m.cache[cacheKey] = face
	return face
}

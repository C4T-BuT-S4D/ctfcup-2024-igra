package music

import (
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/resources"
)

type Manager struct {
	audioContext *audio.Context
	cache        map[string]*audio.Player
	m            sync.Mutex
}

func (m *Manager) getPlayer(path string) *audio.Player {
	m.m.Lock()
	defer m.m.Unlock()

	if player, ok := m.cache[path]; ok {
		return player
	}

	f, err := resources.EmbeddedFS.Open(path)
	if err != nil {
		panic(err)
	}

	stream, err := mp3.DecodeWithoutResampling(f)
	if err != nil {
		panic(err)
	}

	player, err := m.audioContext.NewPlayer(stream)
	if err != nil {
		panic(err)
	}

	m.cache[path] = player
	return player
}

func (m *Manager) GetPlayer(musicType Type) *audio.Player {
	return m.getPlayer(fmt.Sprintf("music/%s.mp3", musicType))
}

func NewManager() *Manager {
	return &Manager{
		audioContext: audio.NewContext(44100),
		cache:        make(map[string]*audio.Player),
	}
}

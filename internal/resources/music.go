package resources

import (
	"fmt"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

type MusicType string

const (
	MusicBackground MusicType = "background"
)

type MusicBundle struct {
	m sync.Mutex

	audioContext *audio.Context
	cache        map[string]*audio.Player
}

func newMusicBundle() *MusicBundle {
	return &MusicBundle{
		audioContext: audio.NewContext(44100),
		cache:        make(map[string]*audio.Player),
	}
}

func (mb *MusicBundle) GetMusicPlayer(t MusicType) *audio.Player {
	return mb.getPlayer(fmt.Sprintf("music/%s.mp3", t))
}

func (mb *MusicBundle) getPlayer(path string) *audio.Player {
	mb.m.Lock()
	defer mb.m.Unlock()

	if player, ok := mb.cache[path]; ok {
		return player
	}

	f, err := EmbeddedFS.Open(path)
	if err != nil {
		panic(err)
	}

	stream, err := mp3.DecodeWithoutResampling(f)
	if err != nil {
		panic(err)
	}

	player, err := mb.audioContext.NewPlayer(stream)
	if err != nil {
		panic(err)
	}

	mb.cache[path] = player
	return player
}

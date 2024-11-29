package server

import (
	"errors"
	"fmt"
	"image/color"
	"os"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/font"

	gameserverpb "github.com/c4t-but-s4d/ctfcup-2023-igra/proto/go/gameserver"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/camera"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/engine"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/fonts"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/input"
)

var ErrGameShutdown = errors.New("game is shut down")

type Game struct {
	IsWin       bool
	WasCheating bool

	fontManager *fonts.Manager
	engine      *engine.Engine

	lock sync.Mutex

	snapshotsDir string
	shutdown     chan struct{}
}

func NewGame(snapshotsDir string, fontManager *fonts.Manager) *Game {
	return &Game{
		snapshotsDir: snapshotsDir,
		shutdown:     make(chan struct{}),
		fontManager:  fontManager,
	}
}

func (g *Game) processEvent(event *gameserverpb.ClientEvent) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.engine == nil {
		return nil
	}

	logrus.Debugf("new update from client: %v", event)
	inp := input.NewFromProto(event.KeysPressed)

	if inp.IsKeyNewlyPressed(ebiten.KeySlash) {
		s, err := g.engine.MakeSnapshot()
		if err != nil {
			return fmt.Errorf("making snapshot: %w", err)
		}

		if err := g.engine.SaveSnapshot(s); err != nil {
			return fmt.Errorf("saving snapshot: %w", err)
		}
	}

	if err := g.engine.Update(inp); err != nil {
		return fmt.Errorf("updating engine state: %w", err)
	}

	return nil
}

func (g *Game) setEngine(eng *engine.Engine) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.engine = eng
}

func (g *Game) getEngine() *engine.Engine {
	g.lock.Lock()
	defer g.lock.Unlock()
	return g.engine
}

func (g *Game) resetEngine() {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.engine = nil
}

// Update doesn't do anything, because the game state is updated by the server.
func (g *Game) Update() error {
	select {
	case <-g.shutdown:
		return ErrGameShutdown
	default:
		return nil
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.engine != nil {
		g.engine.Draw(screen)
	} else {
		face := g.fontManager.Get(fonts.DSouls)

		c := color.RGBA{0xff, 0xff, 0xff, 0xff}
		team := strings.Split(os.Getenv("AUTH_TOKEN"), ":")[0]
		action := "disconnected"
		if g.IsWin {
			c = color.RGBA{0x00, 0xff, 0x00, 0xff}
			action = "won"
		}
		if g.WasCheating {
			c = color.RGBA{0xff, 0x00, 0x00, 0xff}
			action = "cheated"
		}

		txt := fmt.Sprintf("Team %s: %s", team, action)
		width := font.MeasureString(face, txt)

		text.Draw(screen, txt, face, camera.WIDTH/2-width.Floor()/2, camera.HEIGHT/2, c)
	}
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return camera.WIDTH, camera.HEIGHT
}

func (g *Game) Shutdown() {
	close(g.shutdown)
}

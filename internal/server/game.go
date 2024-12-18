package server

import (
	"errors"
	"fmt"
	"image/color"
	"os"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/sirupsen/logrus"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/camera"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/engine"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/input"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/resources"
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2024-igra/proto/go/gameserver"
)

var ErrGameShutdown = errors.New("game is shut down")

type Game struct {
	IsWin       bool
	WasCheating bool

	fontBundle *resources.FontBundle
	engine     *engine.Engine

	lock sync.Mutex

	snapshotsDir string
	shutdown     chan struct{}
}

func NewGame(snapshotsDir string, fontBundle *resources.FontBundle) *Game {
	return &Game{
		snapshotsDir: snapshotsDir,
		shutdown:     make(chan struct{}),
		fontBundle:   fontBundle,
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
		s := g.engine.MakeSnapshot()

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
		face := g.fontBundle.GetFontFace(resources.FontSouls)

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
		width, _ := text.Measure(txt, face, 0)

		textOp := &text.DrawOptions{}
		textOp.GeoM.Translate(camera.WIDTH/2-width/2, camera.HEIGHT/2)
		textOp.ColorScale.ScaleWithColor(c)
		text.Draw(screen, txt, face, textOp)
	}
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return camera.WIDTH, camera.HEIGHT
}

func (g *Game) Shutdown() {
	close(g.shutdown)
}

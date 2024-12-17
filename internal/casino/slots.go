package casino

import (
	"math/rand/v2"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
	"github.com/hajimehoshi/ebiten/v2"
)

type SlotMachine struct {
	*object.Rendered
	Coins         int
	Cooldown      int
	Probability   float64
	LastTriggered int

	rnd *rand.Rand
}

func NewSlotMachine(origin geometry.Point, img *ebiten.Image, width, height float64, coins, cooldown, seed int, probability float64) *SlotMachine {
	sm := &SlotMachine{
		Rendered:    object.NewRendered(origin, img, width, height),
		Coins:       coins,
		Cooldown:    cooldown,
		Probability: probability,
		rnd:         rand.New(rand.NewPCG(uint64(seed), uint64(seed))),
	}

	sm.Reset()

	return sm
}

func (sm *SlotMachine) Reset() {
	sm.LastTriggered = -60_000_000
}

func (sm *SlotMachine) Won() bool {
	return sm.rnd.Float64() < sm.Probability
}

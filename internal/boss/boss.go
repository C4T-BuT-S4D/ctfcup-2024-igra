package boss

import (
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/damage"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
)

type TickState struct {
	CurrentTick int
}

type TickResult struct {
	Dead    bool
	Bullets []*damage.Bullet
}

type HealthState struct {
	Health    int
	MaxHealth int
}

type BOSS interface {
	Object() *object.Rendered
	Tick(s *TickState) *TickResult
	Health() *HealthState
	Reset()
}

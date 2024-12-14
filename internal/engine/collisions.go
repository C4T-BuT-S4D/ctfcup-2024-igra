package engine

import (
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/player"
)

func (e *Engine) Collisions(r *geometry.Rectangle) []object.Generic {
	var result []object.Generic

	// Collision order is important for rendering:
	// - Background is rendered first
	// - Player is rendered on top of everything except bullets
	result = collideGeneric(result, r, e.BackgroundImages)
	result = collideGeneric(result, r, e.Tiles)
	result = collideGeneric(result, r, e.Items)
	result = collideGeneric(result, r, e.Portals)
	result = collideGeneric(result, r, e.Spikes)
	result = collideGeneric(result, r, e.NPCs)
	result = collideGeneric(result, r, e.Arcades)
	result = collideGeneric(result, r, []*player.Player{e.Player})
	result = collideGeneric(result, r, e.EnemyBullets)

	return result
}

func Collide[O object.Generic](r *geometry.Rectangle, objects []O) []O {
	var result []O
	for _, o := range objects {
		if o.Rectangle().Intersects(r) {
			result = append(result, o)
		}
	}
	return result
}

func collideGeneric[O object.Generic](result []object.Generic, r *geometry.Rectangle, objects []O) []object.Generic {
	for _, o := range objects {
		if o.Rectangle().Intersects(r) {
			result = append(result, o)
		}
	}
	return result
}

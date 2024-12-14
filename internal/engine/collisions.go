package engine

import (
	"slices"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/player"
)

func collide[T object.Generic](result []object.Generic, r *geometry.Rectangle, objects []T, filter []object.Type) []object.Generic {
	var t T
	if len(filter) > 0 && !slices.Contains(filter, t.Type()) {
		return result
	}

	for _, o := range objects {
		if o.Rectangle().Intersects(r) {
			result = append(result, o)
		}
	}

	return result
}

func (e *Engine) Collisions(r *geometry.Rectangle, filter ...object.Type) []object.Generic {
	var result []object.Generic

	// Collision order is important for rendering:
	// - Background is rendered first
	// - Player is rendered on top of everything except bullets
	result = collide(result, r, e.BackgroundImages, filter)
	result = collide(result, r, e.Tiles, filter)
	result = collide(result, r, e.Items, filter)
	result = collide(result, r, e.Portals, filter)
	result = collide(result, r, e.Spikes, filter)
	result = collide(result, r, e.NPCs, filter)
	result = collide(result, r, e.Arcades, filter)
	result = collide(result, r, []*player.Player{e.Player}, filter)
	result = collide(result, r, e.EnemyBullets, filter)

	return result
}

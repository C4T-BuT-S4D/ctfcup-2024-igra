package engine

import (
	"slices"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
)

func (e *Engine) Collisions(r *geometry.Rectangle, filter ...object.Type) []object.Generic {
	var result []object.Generic

	// Background image should be rendered first.
	for _, bg := range e.BackgroundImages {
		if bg.Rectangle().Intersects(r) {
			result = append(result, bg)
		}
	}

	for _, t := range e.Tiles {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	for _, t := range e.Items {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	for _, t := range e.Portals {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	for _, t := range e.Spikes {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	for _, t := range e.NPCs {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	for _, t := range e.Arcades {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	// Render player on top of everything except bullets.
	if e.Player.Rectangle().Intersects(r) {
		result = append(result, e.Player)
	}

	for _, t := range e.EnemyBullets {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	if len(filter) == 0 {
		return result
	}

	var filtered []object.Generic
	for _, o := range result {
		if slices.Contains(filter, o.Type()) {
			filtered = append(filtered, o)
		}
	}

	return filtered
}

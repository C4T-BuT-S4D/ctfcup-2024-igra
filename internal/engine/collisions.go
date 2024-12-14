package engine

import (
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
)

func (e *Engine) Collisions(r *geometry.Rectangle) []object.GenericObject {
	var result []object.GenericObject

	// Background image should be rendered first.
	for _, bg := range e.BackgroundImages {
		if bg.Rectangle().Intersects(r) {
			result = append(result, bg)
		}
	}

	if e.Player.Rectangle().Intersects(r) {
		result = append(result, e.Player)
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

	for _, t := range e.InvWalls {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	for _, t := range e.NPCs {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	for _, t := range e.ArcadeMachines {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	for _, t := range e.EnemyBullets {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	return result
}

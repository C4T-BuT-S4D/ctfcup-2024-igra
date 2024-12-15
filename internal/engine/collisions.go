package engine

import (
	"iter"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/player"
)

func (e *Engine) Collisions(r *geometry.Rectangle) []object.Collidable {
	var result []object.Collidable

	// Collision order is important for rendering:
	// - Background is rendered first
	// - Player is rendered on top of everything except bullets
	result = collideGeneric(result, r, e.BackgroundImages)
	result = collideGeneric(result, r, e.Tiles)
	result = collideGeneric(result, r, e.Items)
	result = collideGeneric(result, r, e.Portals)
	result = collideGeneric(result, r, e.Spikes)
	result = collideGeneric(result, r, e.Platforms)
	result = collideGeneric(result, r, e.NPCs)
	result = collideGeneric(result, r, e.Arcades)
	result = collideGeneric(result, r, []*player.Player{e.Player})
	result = collideGeneric(result, r, e.EnemyBullets)

	return result
}

func Collide[O object.Collidable](r *geometry.Rectangle, objects []O) iter.Seq[O] {
	return func(yield func(O) bool) {
		for _, o := range objects {
			if o.Rectangle().Intersects(r) {
				if !yield(o) {
					return
				}
			}
		}
	}
}

func Collide2[O1, O2 object.Collidable](r *geometry.Rectangle, o1s []O1, o2s []O2) iter.Seq[object.Collidable] {
	return func(yield func(object.Collidable) bool) {
		for _, o1 := range o1s {
			if o1.Rectangle().Intersects(r) {
				if !yield(o1) {
					return
				}
			}
		}

		for _, o2 := range o2s {
			if o2.Rectangle().Intersects(r) {
				if !yield(o2) {
					return
				}
			}
		}
	}
}

func collideGeneric[O object.Collidable](result []object.Collidable, r *geometry.Rectangle, objects []O) []object.Collidable {
	for _, o := range objects {
		if o.Rectangle().Intersects(r) {
			result = append(result, o)
		}
	}
	return result
}

package resources

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type SpriteType string

const (
	SpriteSpike  SpriteType = "spike"
	SpritePlayer SpriteType = "player"
	SpritePortal SpriteType = "portal"
	SpriteBullet SpriteType = "bullet"
	SpriteHP     SpriteType = "hp"
	SpriteBG     SpriteType = "bg"
	SpriteArcade SpriteType = "arcade"
)

type SpriteBundle struct {
	*imageBundle
}

func newSpriteBundle() *SpriteBundle {
	return &SpriteBundle{imageBundle: newImageBundle()}
}

func (sb *SpriteBundle) GetSprite(t SpriteType) *ebiten.Image {
	return sb.getImage(fmt.Sprintf("sprites/%s.png", t))
}

func (sb *SpriteBundle) GetAnimationSprite(t SpriteType, animation string) *ebiten.Image {
	return sb.getImage(fmt.Sprintf("sprites/%s_%s.png", t, animation))
}

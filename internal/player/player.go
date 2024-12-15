package player

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/item"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/physics"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/resources"
)

const (
	Width             = 32
	Height            = 32
	DefaultHealth     = 100
	StandingAnimation = "standing"
	RunningAnimation  = "running"
	JumpingAnimation  = "jumping"
	FallingAnimation  = "falling"
)

type Player struct {
	*object.Base
	*physics.Physical

	animations               map[string][]*ebiten.Image
	currentAnimationName     string
	currentAnimationIndex    int
	currentAnimationDuration int

	Inventory *Inventory

	LooksRight bool
	Health     int

	prevGround     object.Collidable
	onGround       object.Collidable
	onGroundCoyote bool
	coyoteTick     int
}

func New(origin geometry.Point, spriteBundle *resources.SpriteBundle) (*Player, error) {
	animations := make(map[string][]*ebiten.Image)

	for anim, numAnims := range map[string]int{
		StandingAnimation: 1,
		RunningAnimation:  2,
		JumpingAnimation:  1,
		FallingAnimation:  1,
	} {
		for i := 0; i < numAnims; i++ {
			img := spriteBundle.GetAnimationSprite(resources.SpritePlayer, fmt.Sprintf("%s_%d", anim, i))
			animations[anim] = append(animations[anim], img)
		}
	}

	return &Player{
		Base: &object.Base{
			Origin: origin,
			Width:  Width,
			Height: Height,
		},
		Physical:   &physics.Physical{},
		Inventory:  &Inventory{},
		Health:     DefaultHealth,
		animations: animations,
	}, nil
}

func (p *Player) SetOnGround(ground object.Collidable, tick int) {
	p.prevGround = p.onGround
	if ground != nil {
		p.coyoteTick = 0
		p.onGround = ground
		p.onGroundCoyote = true
		return
	}
	p.onGround = ground
	if !p.onGroundCoyote {
		return
	}
	if p.coyoteTick == 0 {
		p.coyoteTick = tick
	}
	if tick-p.coyoteTick > 7 {
		p.onGroundCoyote = false
		p.coyoteTick = 0
	}
}

func (p *Player) OnGround() object.Collidable {
	return p.onGround
}

func (p *Player) PrevGround() object.Collidable {
	return p.prevGround
}

func (p *Player) OnGroundCoyote() bool {
	return p.onGroundCoyote
}

func (p *Player) ResetCoyote() {
	p.onGroundCoyote = false
	p.coyoteTick = 0
}

func (p *Player) IsDead() bool {
	return p.Health <= 0
}

func (p *Player) Collect(it *item.Item) {
	it.Collected = true
	p.Inventory.Items = append(p.Inventory.Items, it)
}

func (p *Player) Image() *ebiten.Image {
	prevAnimationName := p.currentAnimationName

	if p.OnGroundCoyote() {
		if p.Speed.X == 0 {
			p.currentAnimationName = StandingAnimation
		} else {
			p.currentAnimationName = RunningAnimation
		}
	} else {
		if p.Speed.Y <= 0 {
			p.currentAnimationName = JumpingAnimation
		} else {
			p.currentAnimationName = FallingAnimation
		}
	}

	switch {
	case p.currentAnimationName != prevAnimationName:
		p.currentAnimationIndex = 0
		p.currentAnimationDuration = 0
	case p.currentAnimationDuration >= 10:
		p.currentAnimationIndex = (p.currentAnimationIndex + 1) % len(p.animations[p.currentAnimationName])
		p.currentAnimationDuration = 0
	default:
		p.currentAnimationDuration++
	}

	return p.animations[p.currentAnimationName][p.currentAnimationIndex]
}

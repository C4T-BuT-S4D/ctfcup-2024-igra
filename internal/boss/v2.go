package boss

import (
	"math"
	"math/rand/v2"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/damage"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/platform"
	"github.com/hajimehoshi/ebiten/v2"
)

var _ BOSS = &V2{}

type V2 struct {
	*object.Rendered `msgpack:"object"`

	BulletImg *ebiten.Image

	platforms         []*platform.Platform
	disabledPlatforms []*platform.Platform

	prevMove geometry.Vector

	health          int            `msgpack:"health"`
	startTick       int            `msgpack:"start_tick"`
	initialLocation geometry.Point `msgpack:"initial_location"`
	rnd             *rand.Rand
}

func NewV2(obj *object.Rendered, bulletImg *ebiten.Image) *V2 {
	v2 := &V2{
		Rendered:        obj,
		BulletImg:       bulletImg,
		initialLocation: obj.Origin,
	}
	v2.Reset()
	return v2
}

func (b *V2) Object() *object.Rendered {
	return b.Rendered
}

func (b *V2) SetPlatforms(platforms []*platform.Platform) {
	b.platforms = platforms
}

func (b *V2) Tick(s *TickState) *TickResult {
	if b.rnd == nil {
		b.startTick = s.CurrentTick
		b.rnd = rand.New(rand.NewPCG(0, uint64(s.CurrentTick)))
	}

	if b.health <= 0 {
		return &TickResult{Dead: true}
	}

	const (
		bulletCount           = 2
		bulletSpawnSquareSize = 50
		bulletWiggle          = 1
		bulletSpeed           = 5
		bulletDamage          = 5
		maxPlatformsDisabled  = 3
	)

	res := &TickResult{}

	// Shoot.
	if b.rnd.IntN(10) == 0 {
		for i := 0; i < bulletCount; i++ {
			boxdx := float64(randInt(b.rnd, -bulletSpawnSquareSize, bulletSpawnSquareSize))
			boxdy := float64(randInt(b.rnd, -bulletSpawnSquareSize, bulletSpawnSquareSize))

			dirdy := float64(randInt(b.rnd, -bulletWiggle, bulletWiggle))
			dir := geometry.Vector{X: -bulletSpeed, Y: dirdy}

			res.Bullets = append(res.Bullets, damage.NewBullet(
				b.Origin.Add(geometry.Vector{X: boxdx, Y: boxdy}),
				b.BulletImg,
				bulletDamage,
				dir,
				0,
			))
		}
	}

	// Switch platform.
	if b.rnd.IntN(20) == 0 {
		platformIndex := b.rnd.IntN(len(b.platforms))
		p := b.platforms[platformIndex]
		p.DisableCollisions(true)
		b.disabledPlatforms = append(b.disabledPlatforms, p)
		for len(b.disabledPlatforms) > maxPlatformsDisabled {
			p := b.disabledPlatforms[0]
			p.DisableCollisions(false)
			b.disabledPlatforms = b.disabledPlatforms[1:]
		}
	}

	if (s.CurrentTick-b.startTick)%30 == 0 {
		b.health -= 4
	}

	res.Dead = b.health <= 0

	const maxDY = 350
	if !res.Dead {
		move := b.prevMove
		dy := b.Origin.SubPoint(b.initialLocation).Y
		if dy > maxDY {
			move.Y = -math.Abs(dy)
		}
		if dy < -maxDY {
			move.Y = math.Abs(dy)
		}
		b.Rendered.Move(move)
		b.prevMove = move
	}

	return res
}

func (b *V2) Health() *HealthState {
	return &HealthState{
		Health:    b.health,
		MaxHealth: 300,
	}
}

func (b *V2) Reset() {
	b.health = 300
	b.Rendered.MoveTo(b.initialLocation)
	b.prevMove = geometry.Vector{X: 0, Y: -5}
	b.rnd = nil
}

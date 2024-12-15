package platform

import (
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/physics"
	"github.com/hajimehoshi/ebiten/v2"
)

type PlatformPath int

const (
	PathVertical PlatformPath = iota
	PathHorizontal
)

type Platform struct {
	*object.Rendered
	*physics.Physical
	// Used to delay acceleration by 1 tick since the speed changes on the next tick
	// after the platform reaches the end of the path,
	// so its acceleration should be observable only on the next tick, too.
	nextAcceleration *geometry.Vector
	start            float64
	end              float64
	Path             PlatformPath
}

func New(origin *geometry.Point, width, height int, image *ebiten.Image, path PlatformPath, distance, speed int) *Platform {
	start := origin.X
	speedVector := geometry.Vector{X: float64(speed), Y: 0}
	if path == PathVertical {
		start = origin.Y - float64(distance) // swap start and end if vertical
		speedVector = geometry.Vector{X: 0, Y: float64(speed)}
	}

	end := start + float64(distance)

	return &Platform{
		Rendered:         object.NewRendered(origin, image, float64(width), float64(height)),
		Physical:         &physics.Physical{Speed: &speedVector, Acceleration: &geometry.Vector{}},
		nextAcceleration: &geometry.Vector{},
		start:            start,
		end:              end,
		Path:             path,
	}
}

func (p *Platform) MoveX() {
	p.Acceleration.X = p.nextAcceleration.X
	p.ApplyAccelerationX()
	p.nextAcceleration.X = 0
	if p.Path == PathHorizontal {
		p.move()
	}
}

func (p *Platform) MoveY() {
	p.Acceleration.Y = p.nextAcceleration.Y
	p.ApplyAccelerationY()
	p.nextAcceleration.Y = 0
	if p.Path == PathVertical {
		p.move()
	}
}

func (p *Platform) move() {
	cur := p.Origin.Y
	if p.Path == PathHorizontal {
		cur = p.Origin.X
	}

	speed := p.Speed.X
	if p.Path == PathVertical {
		speed = p.Speed.Y
	}

	next := cur + float64(speed)
	switch {
	case speed > 0 && next > p.end:
		p.nextAcceleration = p.Speed.Neg().Multiply(2)
		next = p.end
	case speed < 0 && next < p.start:
		p.nextAcceleration = p.Speed.Neg().Multiply(2)
		next = p.start
	}

	if p.Path == PathVertical {
		p.Origin.Y = next
	} else {
		p.Origin.X = next
	}
}

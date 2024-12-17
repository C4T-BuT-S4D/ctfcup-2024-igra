package physics

import (
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
	"github.com/hajimehoshi/ebiten/v2"
)

type MovementPath int

const (
	PathVertical MovementPath = iota
	PathHorizontal
)

func ParsePath(path string) MovementPath {
	if path == "vertical" {
		return PathVertical
	}
	return PathHorizontal
}

type Physical struct {
	Speed        geometry.Vector
	Acceleration geometry.Vector
}

const GravityAcceleration = 1.0 * 2.0 / 6.0

func (o *Physical) ApplyAccelerationX() *Physical {
	o.Speed.X += o.Acceleration.X
	return o
}

func (o *Physical) ApplyAccelerationY() *Physical {
	o.Speed.Y += o.Acceleration.Y
	return o
}

func (o *Physical) SpeedVec() geometry.Vector {
	return o.Speed
}

type Moving interface {
	SpeedVec() geometry.Vector
}

type MovingObject struct {
	*object.Rendered
	*Physical
	// Used to delay acceleration by 1 tick since the speed changes on the next tick
	// after the object reaches the end of the path,
	// so its acceleration should be observable only on the next tick, too.
	nextAcceleration geometry.Vector
	start            float64
	end              float64
	static           bool
	Path             MovementPath
}

func NewMovingObject(origin geometry.Point, width, height float64, image *ebiten.Image, path MovementPath, distance, speed int) *MovingObject {
	start := origin.X
	speedVector := geometry.Vector{X: float64(speed), Y: 0}
	if path == PathVertical {
		start = origin.Y - float64(distance) // swap start and end if vertical
		speedVector = geometry.Vector{X: 0, Y: float64(speed)}
	}

	end := start + float64(distance)

	return &MovingObject{
		Rendered:         object.NewRendered(origin, image, width, height),
		Physical:         &Physical{Speed: speedVector, Acceleration: geometry.Vector{}},
		nextAcceleration: geometry.Vector{},
		start:            start,
		end:              end,
		static:           speed == 0,
		Path:             path,
	}
}

func (p *MovingObject) MoveX() {
	if p.static {
		return
	}

	p.Acceleration.X = p.nextAcceleration.X
	p.ApplyAccelerationX()
	p.nextAcceleration.X = 0
	if p.Path == PathHorizontal {
		p.move()
	}
}

func (p *MovingObject) MoveY() {
	if p.static {
		return
	}

	p.Acceleration.Y = p.nextAcceleration.Y
	p.ApplyAccelerationY()
	p.nextAcceleration.Y = 0
	if p.Path == PathVertical {
		p.move()
	}
}

func (p *MovingObject) move() {
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
	case speed > 0 && next >= p.end:
		p.nextAcceleration = p.Speed.Neg().Multiply(2)
		next = p.end
	case speed < 0 && next <= p.start:
		p.nextAcceleration = p.Speed.Neg().Multiply(2)
		next = p.start
	}

	if p.Path == PathVertical {
		p.Origin.Y = next
	} else {
		p.Origin.X = next
	}
}

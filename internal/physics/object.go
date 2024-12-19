package physics

import (
	"fmt"
	"math"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
	"github.com/hajimehoshi/ebiten/v2"
)

type MovementPath int

const (
	PathVertical MovementPath = iota
	PathHorizontal
	PathCircular
)

func (p MovementPath) Vertical() bool {
	return p == PathVertical || p == PathCircular
}

func (p MovementPath) Horizontal() bool {
	return p == PathHorizontal || p == PathCircular
}

func ParsePath(path string) MovementPath {
	switch path {
	case "circular":
		return PathCircular
	case "vertical":
		return PathVertical
	case "horizontal":
		return PathHorizontal
	default:
		panic(fmt.Errorf("unknown path: %s", path))
	}
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
	start            geometry.Point
	end              geometry.Point
	mid              geometry.Point
	static           bool
	Path             MovementPath
}

func NewMovingObject(origin geometry.Point, width, height float64, image *ebiten.Image, path MovementPath, distance, speed int) *MovingObject {
	obj := &MovingObject{
		Rendered:         object.NewRendered(origin, image, width, height),
		Physical:         &Physical{Speed: geometry.Vector{X: 0, Y: 0}, Acceleration: geometry.Vector{}},
		nextAcceleration: geometry.Vector{},
		static:           speed == 0,
		Path:             path,
	}

	switch path {
	case PathVertical:
		obj.start = geometry.Point{X: origin.X, Y: origin.Y - float64(distance)}
		obj.end = geometry.Point{X: origin.X, Y: origin.Y}
		obj.Physical.Speed = geometry.Vector{X: 0, Y: float64(speed)}
	case PathHorizontal:
		obj.start = geometry.Point{X: origin.X, Y: origin.Y}
		obj.end = geometry.Point{X: origin.X + float64(distance), Y: origin.Y}
		obj.Physical.Speed = geometry.Vector{X: float64(speed), Y: 0}
	case PathCircular:
		obj.Origin.X -= float64(distance) / 2
		obj.start = geometry.Point{X: origin.X - float64(distance)/2, Y: origin.Y - float64(distance)/2}
		obj.end = geometry.Point{X: origin.X + float64(distance)/2, Y: origin.Y + float64(distance)/2}
		obj.mid = geometry.Point{X: origin.X, Y: origin.Y}
		obj.Physical.Speed = geometry.Vector{X: float64(speed), Y: float64(speed)}
	}

	return obj
}

func (p *MovingObject) MoveX() {
	if p.static {
		return
	}

	p.Acceleration.X = p.nextAcceleration.X
	p.ApplyAccelerationX()
	p.nextAcceleration.X = 0
	if p.Path.Horizontal() {
		p.moveX()
	}
}

func (p *MovingObject) MoveY() {
	if p.static {
		return
	}

	p.Acceleration.Y = p.nextAcceleration.Y
	p.ApplyAccelerationY()
	p.nextAcceleration.Y = 0
	if p.Path.Vertical() {
		p.moveY()
	}
}

func (p *MovingObject) moveX() {
	cur := p.Origin.X
	speed := p.Speed.X

	if p.Path == PathCircular {
		if p.Origin.Y >= p.mid.Y {
			speed = math.Abs(speed)
		} else {
			speed = -math.Abs(speed)
		}
	}

	next := cur + speed
	switch {
	case speed > 0 && next >= p.end.X:
		p.nextAcceleration.X = -speed * 2
		next = p.end.X
	case speed < 0 && next <= p.start.X:
		p.nextAcceleration.X = -speed * 2
		next = p.start.X
	}

	p.Origin.X = next
}

func (p *MovingObject) moveY() {
	cur := p.Origin.Y
	speed := p.Speed.Y

	if p.Path == PathCircular {
		if p.Origin.X >= p.mid.X {
			speed = -math.Abs(speed)
		} else {
			speed = math.Abs(speed)
		}
	}

	next := cur + speed
	switch {
	case speed > 0 && next >= p.end.Y:
		p.nextAcceleration.Y = -speed * 2
		next = p.end.Y
	case speed < 0 && next <= p.start.Y:
		p.nextAcceleration.Y = -speed * 2
		next = p.start.Y
	}

	p.Origin.Y = next
}

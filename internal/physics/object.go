package physics

import "github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"

type Physical struct {
	Speed        *geometry.Vector
	Acceleration *geometry.Vector
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

func (o *Physical) SpeedVec() *geometry.Vector {
	return o.Speed
}

func NewPhysical() *Physical {
	return &Physical{
		Speed:        &geometry.Vector{},
		Acceleration: &geometry.Vector{},
	}
}

type Moving interface {
	SpeedVec() *geometry.Vector
}

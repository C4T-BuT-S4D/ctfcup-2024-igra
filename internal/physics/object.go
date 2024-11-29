package physics

import "github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"

type Physical struct {
	Speed        *geometry.Vector
	Acceleration *geometry.Vector
}

const GravityAcceleration = 1.0 * 2.0 / 6.0

func (o *Physical) ApplyAcceleration() *Physical {
	o.Speed = o.Speed.Add(o.Acceleration)
	return o
}

func NewPhysical() *Physical {
	return &Physical{
		Speed:        &geometry.Vector{},
		Acceleration: &geometry.Vector{},
	}
}

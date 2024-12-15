package geometry

type Vector struct {
	X, Y float64
}

func (v Vector) Add(other Vector) Vector {
	return Vector{X: v.X + other.X, Y: v.Y + other.Y}
}

func (v Vector) Neg() Vector {
	return Vector{X: -v.X, Y: -v.Y}
}

func (v Vector) LengthSquared() float64 {
	return v.X*v.X + v.Y*v.Y
}

func (v Vector) Multiply(m float64) Vector {
	return Vector{X: v.X * m, Y: v.Y * m}
}

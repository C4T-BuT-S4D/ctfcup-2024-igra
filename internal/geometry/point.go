package geometry

var Origin = &Point{X: 0, Y: 0}

type Point struct {
	X float64
	Y float64
}

func (p Point) Add(v Vector) Point {
	return Point{X: p.X + v.X, Y: p.Y + v.Y}
}

func (p Point) SubPoint(p2 Point) Vector {
	return Vector{X: p.X - p2.X, Y: p.Y - p2.Y}
}

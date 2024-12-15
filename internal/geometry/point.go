package geometry

var Origin = &Point{X: 0, Y: 0}

type Point struct {
	X float64
	Y float64
}

func (p Point) Add(v Vector) Point {
	return Point{X: p.X + v.X, Y: p.Y + v.Y}
}

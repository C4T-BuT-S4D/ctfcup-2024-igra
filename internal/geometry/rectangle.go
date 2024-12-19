package geometry

type Rectangle struct {
	LeftX   float64
	TopY    float64
	RightX  float64
	BottomY float64
}

func (r *Rectangle) Extended(delta float64) *Rectangle {
	return &Rectangle{
		LeftX:   r.LeftX - delta,
		TopY:    r.TopY - delta,
		RightX:  r.RightX + delta,
		BottomY: r.BottomY + delta,
	}
}

func (r *Rectangle) Center() Point {
	return Point{
		X: (r.LeftX + r.RightX) / 2,
		Y: (r.TopY + r.BottomY) / 2,
	}
}

func (r *Rectangle) AddVector(other *Vector) *Rectangle {
	return &Rectangle{
		LeftX:   r.LeftX + other.X,
		TopY:    r.TopY + other.Y,
		RightX:  r.RightX + other.X,
		BottomY: r.BottomY + other.Y,
	}
}

func (r *Rectangle) Sub(other *Rectangle) Vector {
	return Vector{
		X: r.LeftX - other.LeftX,
		Y: r.TopY - other.TopY,
	}
}

func (r *Rectangle) Intersects(b *Rectangle) bool {
	return r.RightX > b.LeftX && b.RightX > r.LeftX && r.BottomY > b.TopY && b.BottomY > r.TopY
}

func (r *Rectangle) PushVectorX(b *Rectangle) Vector {
	return r.pushVector(b, []Vector{
		{X: r.RightX - b.LeftX, Y: 0},
		{X: r.LeftX - b.RightX, Y: 0},
	}, Vector{X: r.RightX - b.RightX, Y: r.LeftX - b.LeftX})
}

func (r *Rectangle) PushVectorY(b *Rectangle) Vector {
	return r.pushVector(b, []Vector{
		{X: 0, Y: r.BottomY - b.TopY},
		{X: 0, Y: r.TopY - b.BottomY},
	}, Vector{X: r.BottomY - b.BottomY, Y: r.TopY - b.TopY})
}

func (r *Rectangle) pushVector(b *Rectangle, vecs []Vector, check Vector) Vector {
	if !r.Intersects(b) || check.Length() < 1e-6 {
		return Vector{}
	}

	v := vecs[0]
	if v1 := vecs[1]; v1.Length() < v.Length() {
		v = v1
	}

	return v
}

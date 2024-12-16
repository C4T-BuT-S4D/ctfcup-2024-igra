package geometry

type Rectangle struct {
	LeftX   float64
	TopY    float64
	RightX  float64
	BottomY float64
}

func (a *Rectangle) Extended(delta float64) *Rectangle {
	return &Rectangle{
		LeftX:   a.LeftX - delta,
		TopY:    a.TopY - delta,
		RightX:  a.RightX + delta,
		BottomY: a.BottomY + delta,
	}
}

func (a *Rectangle) AddVector(other *Vector) *Rectangle {
	return &Rectangle{
		LeftX:   a.LeftX + other.X,
		TopY:    a.TopY + other.Y,
		RightX:  a.RightX + other.X,
		BottomY: a.BottomY + other.Y,
	}
}

func (a *Rectangle) Sub(other *Rectangle) Vector {
	return Vector{
		X: a.LeftX - other.LeftX,
		Y: a.TopY - other.TopY,
	}
}

func (a *Rectangle) Intersects(b *Rectangle) bool {
	return a.RightX > b.LeftX && b.RightX > a.LeftX && a.BottomY > b.TopY && b.BottomY > a.TopY
}

func (a *Rectangle) PushVectorX(b *Rectangle) Vector {
	return a.pushVector(b, []Vector{
		{X: a.RightX - b.LeftX, Y: 0},
		{X: a.LeftX - b.RightX, Y: 0},
	}, Vector{X: a.RightX - b.RightX, Y: a.LeftX - b.LeftX})
}

func (a *Rectangle) PushVectorY(b *Rectangle) Vector {
	return a.pushVector(b, []Vector{
		{X: 0, Y: a.BottomY - b.TopY},
		{X: 0, Y: a.TopY - b.BottomY},
	}, Vector{X: a.BottomY - b.BottomY, Y: a.TopY - b.TopY})
}

func (a *Rectangle) pushVector(b *Rectangle, vecs []Vector, check Vector) Vector {
	if !a.Intersects(b) || check.Length() < 1e-6 {
		return Vector{}
	}

	v := vecs[0]
	if v1 := vecs[1]; v1.Length() < v.Length() {
		v = v1
	}

	return v
}

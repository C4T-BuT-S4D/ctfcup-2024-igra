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

func (a *Rectangle) Sub(other *Rectangle) *Vector {
	return &Vector{
		X: a.LeftX - other.LeftX,
		Y: a.TopY - other.TopY,
	}
}

func (a *Rectangle) Intersects(b *Rectangle) bool {
	return !(a.RightX <= b.LeftX || b.RightX <= a.LeftX || a.BottomY <= b.TopY || b.BottomY <= a.TopY)
}

func (a *Rectangle) PushVectorX(b *Rectangle) *Vector {
	if !a.Intersects(b) {
		return &Vector{}
	}

	var vecs []*Vector

	if b.RightX > a.LeftX {
		vecs = append(vecs, &Vector{X: a.LeftX - b.RightX, Y: 0})
	}

	if b.LeftX < a.RightX {
		vecs = append(vecs, &Vector{X: a.RightX - b.LeftX, Y: 0})
	}

	v := vecs[0]
	for _, nv := range vecs {
		if nv.LengthSquared() < v.LengthSquared() {
			v = nv
		}
	}

	return v
}

func (a *Rectangle) PushVectorY(b *Rectangle) *Vector {
	if !a.Intersects(b) {
		return &Vector{}
	}

	var vecs []*Vector

	if b.BottomY > a.TopY {
		vecs = append(vecs, &Vector{X: 0, Y: a.TopY - b.BottomY})
	}

	if b.TopY < a.BottomY {
		vecs = append(vecs, &Vector{X: 0, Y: a.BottomY - b.TopY})
	}

	v := vecs[0]
	for _, nv := range vecs {
		if nv.LengthSquared() < v.LengthSquared() {
			v = nv
		}
	}

	return v
}

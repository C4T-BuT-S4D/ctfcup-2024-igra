package object

import "github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"

type Type int

const (
	StaticTileType Type = iota
	PlayerType
	Item
	Portal
	Spike
	InvWall
	NPC
	ArcadeMachine
	BossV1
	BossV2
	EnemyBullet
	BackgroundImage
)

func (t Type) String() string {
	switch t {
	case StaticTileType:
		return "StaticTileType"
	case PlayerType:
		return "PlayerType"
	case Item:
		return "Item"
	case Portal:
		return "Portal"
	case Spike:
		return "Spike"
	default:
		panic("unknown type")
	}
}

type Object struct {
	Origin *geometry.Point
	Width  float64
	Height float64
}

func (o *Object) GetOrigin() *geometry.Point {
	if o == nil {
		return nil
	}
	return o.Origin
}

func (o *Object) Rectangle() *geometry.Rectangle {
	return &geometry.Rectangle{
		LeftX:   o.Origin.X,
		TopY:    o.Origin.Y,
		RightX:  o.Origin.X + o.Width,
		BottomY: o.Origin.Y + o.Height,
	}
}

func (o *Object) Move(d *geometry.Vector) *Object {
	o.Origin = o.Origin.Add(d)
	return o
}

func (o *Object) MoveTo(p *geometry.Point) *Object {
	o.Origin = p
	return o
}

type GenericObject interface {
	Rectangle() *geometry.Rectangle
	Type() Type
}

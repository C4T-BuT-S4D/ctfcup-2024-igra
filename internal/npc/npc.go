package npc

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/dialog"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/item"
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/object"
)

type NPC struct {
	*object.Rendered

	Dialog      dialog.Dialog `msgpack:"-"`
	DialogImage *ebiten.Image `msgpack:"-"`
	LinkedItem  *item.Item    `msgpack:"-"`
	ReturnsItem string
}

func New(origin *geometry.Point, img *ebiten.Image, dialogImage *ebiten.Image, width, height float64, dialog dialog.Dialog, item string) *NPC {
	return &NPC{
		Rendered:    object.NewRendered(origin, img, width, height),
		DialogImage: dialogImage,
		Dialog:      dialog,
		ReturnsItem: item,
	}
}

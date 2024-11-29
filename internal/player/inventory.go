package player

import (
	"github.com/c4t-but-s4d/ctfcup-2024-igra/internal/item"
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2024-igra/proto/go/gameserver"
)

type Inventory struct {
	Items []*item.Item `json:"items"`
}

func (inv *Inventory) ToProto() *gameserverpb.Inventory {
	items := make([]*gameserverpb.Inventory_Item, 0, len(inv.Items))

	for _, it := range inv.Items {
		items = append(items, &gameserverpb.Inventory_Item{
			Name:      it.Name,
			Important: it.Important,
		})
	}

	return &gameserverpb.Inventory{
		Items: items,
	}
}

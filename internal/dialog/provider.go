package dialog

import (
	"fmt"
)

type Provider interface {
	Get(id string) (Dialog, error)
	DisplayInput() bool
}

type StandardProvider struct {
	ShowInput bool
}

func NewStandardProvider(showInput bool) *StandardProvider {
	return &StandardProvider{
		ShowInput: showInput,
	}
}

func (sp *StandardProvider) DisplayInput() bool {
	return sp.ShowInput
}

func (sp *StandardProvider) Get(id string) (Dialog, error) {
	switch id {
	case "test-npc":
		return NewDummy("Hello, I'm a test NPC!\n 2 + 2 = ?", "4"), nil
	default:
		return nil, fmt.Errorf("unknown dialog id: %s", id)
	}
}

type ClientProvider struct {
}

func (cp *ClientProvider) Get(_ string) (Dialog, error) {
	return NewClientDialog(), nil
}

func (cp *ClientProvider) DisplayInput() bool {
	return true
}

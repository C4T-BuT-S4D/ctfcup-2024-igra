package arcade

import "fmt"

type Provider interface {
	Get(id string) (Arcade, error)
}

type StandardProvider struct {
}

func (sp *StandardProvider) Get(id string) (Arcade, error) {
	return nil, fmt.Errorf("unknown arcade id: %s", id)
}

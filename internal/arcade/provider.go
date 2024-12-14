package arcade

import "fmt"

type Provider interface {
	Get(id string) (Game, error)
}

type LocalProvider struct{}

func (sp *LocalProvider) Get(id string) (Game, error) {
	if id == "brodilka" {
		return newBinaryGame("brodilka")
	}

	return nil, fmt.Errorf("unknown arcade id: %s", id)
}
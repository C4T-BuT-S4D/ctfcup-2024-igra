package arcade

import "fmt"

type Provider interface {
	Get(id string) (Game, error)
}

type LocalProvider struct{}

func (sp *LocalProvider) Get(id string) (Game, error) {
	switch id {
	case "brodilka":
		return newBinaryGame("./internal/resources/arcades/brodilka.py"), nil
	case "simple":
		return newSimpleGame(), nil
	case "maze":
		return newBinaryGame("./internal/resources/arcades/internal/maze/maze"), nil
	default:
		return nil, fmt.Errorf("unknown arcade id: %s", id)
	}
}

type ClientProvider struct{}

func (cp *ClientProvider) Get(_ string) (Game, error) {
	return NewClientGame(), nil
}

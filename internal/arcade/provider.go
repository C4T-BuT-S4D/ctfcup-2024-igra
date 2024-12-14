package arcade

import "fmt"

type Provider interface {
	Get(id string) (Game, error)
}

type LocalProvider struct{}

func (sp *LocalProvider) Get(id string) (Game, error) {
	switch id {
	case "brodilka":
		return newBinaryGame("./internal/resources/arcades/brodilka"), nil
	case "simple":
		return newSimpleGame(), nil
	default:
		return nil, fmt.Errorf("unknown arcade id: %s", id)
	}
}

package arcade

const (
	SIZE = 64
)

type Screen struct {
	Cells [SIZE][SIZE]uint8
}

type State struct {
	Won    bool
	Lose   bool
	Screen Screen
}

type Arcade interface {
	Feed(input byte) error
	State() State
}

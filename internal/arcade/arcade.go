package arcade

type Screen struct {
}

type Header struct {
	Height int
	Width  int
}

type State struct {
	Won  bool
	Lose bool
}
type Arcade interface {
	Header() Header
	Feed(input string) error
	State() State
}

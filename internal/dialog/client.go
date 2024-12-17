package dialog

func NewClientDialog() Dialog {
	return &ClientDialog{
		s: State{},
	}
}

type ClientDialog struct {
	s State
}

func (c *ClientDialog) Greeting() {
	// No need to greet.
}

func (c *ClientDialog) Feed(_ string, _ int) {
	// No need to feed.
}

func (c *ClientDialog) State() *State {
	return &c.s
}

func (c *ClientDialog) SetState(s *State) {
	c.s = *s
}

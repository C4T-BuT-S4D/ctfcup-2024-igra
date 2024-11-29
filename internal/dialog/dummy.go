package dialog

import (
	"fmt"
	"strings"
)

func NewDummy(s string, answer string) Dialog {
	return &dummyDialog{greet: s, answer: answer}
}

type dummyDialog struct {
	s      State
	greet  string
	answer string
}

func (d *dummyDialog) Greeting() {
	d.s.Text = d.greet
}

func (d *dummyDialog) Feed(text string) {
	if strings.EqualFold(text, d.answer) {
		d.s.Text += fmt.Sprintf("\n Answer '%s' is correct!", text)
		d.s.GaveItem = true
		d.s.Finished = true
	} else {
		d.s.Text += fmt.Sprintf("\n Answer '%s' is incorrect!", text)
	}
}

func (d *dummyDialog) State() *State {
	return &d.s
}

func (d *dummyDialog) SetState(s *State) {
	d.s = *s
}

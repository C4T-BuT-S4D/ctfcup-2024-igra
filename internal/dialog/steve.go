package dialog

import (
	"strings"
)

func NewSteve() Dialog {
	return &steveDialog{}
}

type steveDialog struct {
	s State
}

func (d *steveDialog) Greeting() {
	d.s.Text = "Wow.\nYou must have been practicing parkour since you were a child.\nI'm impressed.\nPlease take this branch of honour."
}

func (d *steveDialog) Feed(text string) {
	if strings.EqualFold(text, "Thank you") {
		d.s.GaveItem = true
	}

	d.s.Finished = true
}

func (d *steveDialog) State() *State {
	return &d.s
}

func (d *steveDialog) SetState(s *State) {
	d.s = *s
}

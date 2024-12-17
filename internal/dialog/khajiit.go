package dialog

import (
	"fmt"
	"strings"
)

func NewKhajiit(item string, price int) Dialog {
	return &khajiitDialog{
		item:  item,
		price: price,
	}
}

type khajiitDialog struct {
	s     State
	item  string
	price int
}

func (d *khajiitDialog) Greeting() {
	d.s.Text = fmt.Sprintf("Khajiit has wares, if you have %d coin.", d.price)
}

func (d *khajiitDialog) Feed(text string, coins int) {
	if !strings.EqualFold(text, "What have you got for sale?") {
		d.s.Text = "May the sun keep you warm, even in this land of bitter cold."
		return
	}

	if coins >= d.price {
		d.s.GaveItem = true
		d.s.Finished = true
		d.s.Text = "Thank you for the purchase, traveller, enjoy your new " + d.item + "."
	} else {
		d.s.Text = "Your coins are not enough. Until our next meeting, if such is fated."
	}
}

func (d *khajiitDialog) State() *State {
	return &d.s
}

func (d *khajiitDialog) SetState(s *State) {
	d.s = *s
}

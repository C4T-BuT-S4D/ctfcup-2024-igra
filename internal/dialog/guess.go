package dialog

import (
	cryptoRand "crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"math/rand/v2"
	"strconv"
)

var ErrRandomInit = errors.New("failed to parse int in random inicialization")

func NewGuess(greet string) (Dialog, error) {
	hi := big.NewInt(31337)
	if _, ok := hi.SetString("0x10000000000000000", 0); !ok {
		return nil, ErrRandomInit
	}
	seed1, err := cryptoRand.Int(cryptoRand.Reader, hi)
	if err != nil {
		return nil, err
	}
	return &GuessDialog{greet: greet, rand: rand.New(rand.NewPCG(seed1.Uint64(), 0))}, nil
}

type GuessDialog struct {
	s     State
	rand  *rand.Rand
	greet string
}

func (d *GuessDialog) Greeting() {
	d.s.Text = d.greet
}

func (d *GuessDialog) Feed(text string, _ int) {
	answer := d.rand.Uint64N(1000_000_000_000)
	if guess, err := strconv.ParseUint(text, 10, 64); err != nil {
		d.s.Text += fmt.Sprintf("\n encountered error converting int '%s'!", err.Error())
	} else if answer != guess {
		d.s.Text += fmt.Sprintf("\n incorrect: correct answer is '%d'!", answer)
	} else {
		d.s.Text += fmt.Sprintf("\n correct!")
		d.s.GaveItem = true
		d.s.Finished = true
	}
}

func (d *GuessDialog) State() *State {
	return &d.s
}

func (d *GuessDialog) SetState(s *State) {
	d.s = *s
}

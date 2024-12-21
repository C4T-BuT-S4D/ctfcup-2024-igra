package dialog

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
)

func NewBinary(binaryPath, greet, target string, hex bool) Dialog {
	return &BinaryDialog{binaryPath: binaryPath, greet: greet, target: target, hex: hex}
}

type BinaryDialog struct {
	s          State
	hex        bool
	binaryPath string
	greet      string
	target     string
}

func (d *BinaryDialog) Greeting() {
	d.s.Text = d.greet
}

func (d *BinaryDialog) getBinaryOutput(input string) (string, error) {
	cmd := exec.Command(d.binaryPath)

	if d.hex {
		inputBytes, err := hex.DecodeString(input)
		if err != nil {
			return "", err
		}
		cmd.Stdin = bytes.NewReader(inputBytes)
	} else {
		cmd.Stdin = strings.NewReader(input)
	}
	var out strings.Builder
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func (d *BinaryDialog) Feed(text string, _ int) {
	binaryOutput, err := d.getBinaryOutput(text)
	binaryOutput = strings.TrimSpace(binaryOutput)
	if err != nil {
		d.s.Text += fmt.Sprintf("\n Encountered error '%s'!", err.Error())
	} else if binaryOutput != d.target {
		d.s.Text += fmt.Sprintf("\n incorrect!")
	} else {
		d.s.Text += fmt.Sprintf("\n correct!")
		d.s.GaveItem = true
		d.s.Finished = true
	}
}

func (d *BinaryDialog) State() *State {
	return &d.s
}

func (d *BinaryDialog) SetState(s *State) {
	d.s = *s
}

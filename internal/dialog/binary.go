package dialog

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os/exec"
	"strings"
)

func NewInteractiveBinary(binaryPath string, greet string, target string, finishMarkers ...string) Dialog {
	return &InteractiveBinary{
		s:             State{},
		greet:         greet,
		binaryPath:    binaryPath,
		target:        target,
		finishMarkers: finishMarkers,
	}
}

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

type InteractiveBinary struct {
	s             State
	cmd           *exec.Cmd
	greet         string
	binaryPath    string
	target        string
	finishMarkers []string
	writer        io.WriteCloser
	reader        io.ReadCloser
}

func (d *InteractiveBinary) State() *State {
	return &d.s
}

func (d *InteractiveBinary) SetState(s *State) {
	d.s = *s
}

func (d *InteractiveBinary) Greeting() {
	d.s.Text = d.greet
	d.s.Finished = false
	d.restart()
	d.s.Text += "\n" + d.getOutput()
}

func (d *InteractiveBinary) Feed(text string, _ int) {
	if d.cmd == nil {
		d.s.Text += "\n Error: No binary running."
		return
	}

	if d.writer == nil {
		d.s.Text += "\n Error: No writer set."
	}

	_, err := d.writer.Write([]byte(text + "\n"))
	if err != nil {
		d.s.Text += "\n Write error."
	}

	out := d.getOutput()
	for _, finishMarker := range d.finishMarkers {
		if strings.Contains(out, finishMarker) {
			d.s.Finished = true
			d.stopBinary()
		}
	}

	if strings.Contains(out, d.target) {
		d.s.GaveItem = true
	}

	d.s.Text += "\n" + out
}

func (d *InteractiveBinary) getOutput() string {
	if d.reader == nil {
		return "Error: No reader set."
	}

	buf := make([]byte, 1024)
	n, err := d.reader.Read(buf)
	if err != nil {
		return "read error"
	}
	return string(buf[:n])
}

func (d *InteractiveBinary) stopBinary() {
	if d.writer != nil {
		d.writer.Close()
		d.writer = nil
	}
	if d.reader != nil {
		d.reader.Close()
		d.reader = nil
	}
	if d.cmd != nil && d.cmd.Process != nil {
		d.cmd.Process.Kill()
		d.cmd.Wait()
		d.cmd = nil
	}
}

func (d *InteractiveBinary) restart() {
	d.stopBinary()

	var err error
	d.cmd = exec.Command(d.binaryPath)
	d.writer, err = d.cmd.StdinPipe()
	if err != nil {
		logrus.Error(err)
		d.s.Text += "\n Encountered error!"
		return
	}
	d.reader, err = d.cmd.StdoutPipe()
	if err != nil {
		logrus.Error(err)
		d.s.Text += "\n Encountered error!"
		return
	}

	err = d.cmd.Start()
	if err != nil {
		logrus.Error(err)
		d.s.Text += "\n Encountered error!"
		return
	}
}

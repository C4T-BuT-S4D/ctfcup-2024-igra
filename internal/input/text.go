package input

import (
	"bufio"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func WrapLine(t string, face text.Face, width int) (lines []string) {
	var buf strings.Builder

	scan := bufio.NewScanner(strings.NewReader(t))
	scan.Split(bufio.ScanWords)
	for scan.Scan() {
		word := scan.Text()
		wnew, _ := text.Measure(buf.String()+" "+word, face, 0)
		if wnew > float64(width) {
			lines = append(lines, buf.String())
			buf.Reset()
			buf.WriteString(word)
		} else {
			buf.WriteString(" ")
			buf.WriteString(word)
		}
	}
	if buf.Len() > 0 {
		lines = append(lines, buf.String())
	}
	return
}

func AutoWrap(t string, face text.Face, width int) []string {
	var lines []string

	lineScan := bufio.NewScanner(strings.NewReader(t))
	lineScan.Split(bufio.ScanLines)
	for lineScan.Scan() {
		lines = append(lines, WrapLine(lineScan.Text(), face, width)...)
	}

	return lines
}

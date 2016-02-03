package gocropy

import (
	"bytes"
	"fmt"
)

type Char struct {
	Box  Bbox
	Rune rune
}

func (char Char) String() string {
	return fmt.Sprintf("%c %v", char.Rune, char.Box)
}

func TokenizeChars(chars []Char) (string, Bbox) {
	if len(chars) <= 0 {
		return "", Bbox{}
	}
	var bbox Bbox
	var buffer bytes.Buffer
	for i := range chars {
		if i == 0 {
			bbox = chars[i].Box
		} else {
			bbox.Add(chars[i].Box)
		}
		buffer.WriteRune(chars[i].Rune)
	}
	return buffer.String(), bbox

}

func CharsFromLlocs(llocs []Lloc, bbox Bbox) []Char {
	chars := make([]Char, 0, len(llocs))
	for i := range llocs {
		chars = append(chars, makeChar(llocs[i], bbox))
		if i > 0 {
			chars[i-1].Box.Right = chars[i].Box.Left - 1
		}
	}
	if len(chars) > 0 {
		chars[len(chars)-1].Box.Right = bbox.Right
	}
	return chars
}

func makeChar(lloc Lloc, bbox Bbox) Char {
	return Char{
		Bbox{
			bbox.Left + int(lloc.Adjustment),
			bbox.Top,
			0,
			bbox.Bottom,
		},
		lloc.Codepoint,
	}
}

package gocropy

import (
	"fmt"
	"strconv"
	"unicode"
)

type Bbox struct {
	Left, Top, Right, Bottom int64
}

func (bbox Bbox) String() string {
	return fmt.Sprintf(
		"bbox %v %v %v %v",
		bbox.Left,
		bbox.Top,
		bbox.Right,
		bbox.Bottom,
	)
}

func (bbox *Bbox) Sanitize(page Bbox) {
	h := page.Bottom
	bbox.Top = h - bbox.Bottom - 1
	bbox.Bottom = h - bbox.Top - 1
}

func (bbox *Bbox) Scan(state fmt.ScanState, verb rune) error {
	state.Token(true, func(r rune) bool {
		return r == 'b' || r == 'o' || r == 'x'
	})
	bbox.Left = scanInt(state)
	bbox.Top = scanInt(state)
	bbox.Right = scanInt(state)
	bbox.Bottom = scanInt(state)
	return nil
}

func scanInt(state fmt.ScanState) int64 {
	token, err := state.Token(true, func(r rune) bool {
		return unicode.IsDigit(r)
	})
	if err != nil {
		return -1
	}
	res, err := strconv.ParseInt(string(token), 0, 64)
	if err != nil {
		return -1
	}
	return res
}

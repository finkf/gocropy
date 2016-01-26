package gocropy

import (
	"fmt"
	"strconv"
	"unicode"
)

type Bbox struct {
	Left, Top, Right, Bottom int
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func (bbox *Bbox) Add(other Bbox) {
	bbox.Left = min(bbox.Left, other.Left)
	bbox.Top = min(bbox.Top, other.Top)
	bbox.Right = max(bbox.Right, other.Right)
	bbox.Bottom = max(bbox.Bottom, other.Bottom)
}

func (bbox *Bbox) AddSlice(bboxes []Bbox) {
	for _, b := range bboxes {
		bbox.Add(b)
	}
}

func (bbox *Bbox) Sanitize(page Bbox) {
	h := page.Bottom
	bbox.Top = h - bbox.Bottom - 1
	bbox.Bottom = h - bbox.Top - 1
}

func (bbox Bbox) String() string {
	return fmt.Sprintf(
		"%v %v %v %v",
		bbox.Left,
		bbox.Top,
		bbox.Right,
		bbox.Bottom,
	)
}

func (bbox *Bbox) Scan(state fmt.ScanState, verb rune) error {
	var err error
	bbox.Left, err = scanInt(state)
	if err != nil {
		return err
	}
	bbox.Top, err = scanInt(state)
	if err != nil {
		return err
	}
	bbox.Right, err = scanInt(state)
	if err != nil {
		return err
	}
	bbox.Bottom, err = scanInt(state)
	if err != nil {
		return err
	}
	return nil
}

func scanInt(state fmt.ScanState) (int, error) {
	token, err := state.Token(true, func(r rune) bool {
		return unicode.IsDigit(r)
	})
	if err != nil {
		res, err := strconv.ParseInt(string(token), 0, 64)
		return int(res), err
	}
	return 0, err
}

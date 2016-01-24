package gocropy

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"unicode"
)

type Lloc struct {
	Codepoint  rune
	Adjustment float64
}

func (lloc Lloc) String() string {
	return fmt.Sprintf("%c %e", lloc.Codepoint, lloc.Adjustment)
}

func (lloc *Lloc) Scan(state fmt.ScanState, verb rune) error {
	r, _, err := state.ReadRune()
	if err != nil {
		return err
	}
	token, err := state.Token(true, func(r rune) bool {
		return unicode.IsDigit(r) || r == '.' || r == '+' || r == '-'
	})
	if err != nil {
		return err
	}
	lloc.Codepoint = r
	lloc.Adjustment, err = strconv.ParseFloat(string(token), 64)
	return err
}

func ReadLlocs(file string) ([]Lloc, error) {
	in, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer in.Close()
	var llocs []Lloc
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		var lloc Lloc
		_, err := fmt.Sscanf(scanner.Text(), "%v", &lloc)
		if err != nil {
			return nil, err
		}
		llocs = append(llocs, lloc)
	}
	return llocs, nil
}

func MustReadLlocs(file string) []Lloc {
	llocs, err := ReadLlocs(file)
	if err != nil {
		panic(err)
	}
	return llocs
}

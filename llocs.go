package gocropy

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Lloc struct {
	Codepoint  rune
	Adjustment float64
}

func (lloc Lloc) String() string {
	return fmt.Sprintf("%c\t%f", lloc.Codepoint, lloc.Adjustment)
}

// This implementation is weird. There seems to be some bug with skipping
// whitespace and the fmt.ScanState
func (lloc *Lloc) Scan(state fmt.ScanState, verb rune) error {
	token, err := state.Token(false, func(r rune) bool {
		return r != '\n'
	})
	if err != nil {
		return err
	}
	if strings.ContainsRune(string(token), '\t') {
		_, err = fmt.Sscanf(
			string(token),
			"%c\t%f",
			&lloc.Codepoint,
			&lloc.Adjustment,
		)
	} else {
		lloc.Codepoint = ' '
		_, err = fmt.Sscanf(
			string(token),
			"%f",
			&lloc.Adjustment,
		)
	}
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

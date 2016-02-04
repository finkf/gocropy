package gocropy

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	//	"strings"
	//	"unicode/utf8"
)

type Lloc struct {
	Codepoint  rune
	Adjustment float64
}

func (lloc Lloc) String() string {
	return fmt.Sprintf("%c\t%f", lloc.Codepoint, lloc.Adjustment)
}

var splitRe = regexp.MustCompile("\\s+")

func isOkLlocLine(splits []string) bool {
	return len(splits) == 2 && len(splits[1]) > 0
}

func llocFromString(line string) (Lloc, error) {
	var lloc Lloc
	splits := splitRe.Split(line, 2)

	if !isOkLlocLine(splits) {
		return lloc, fmt.Errorf("invalid llocs line `%s`", line)
	}
	if len(splits[0]) == 0 {
		lloc.Codepoint = ' '
	} else {
		_, err := fmt.Sscanf(splits[0], "%c", &lloc.Codepoint)
		if err != nil {
			return lloc, err
		}
	}
	_, err := fmt.Sscanf(splits[1], "%f", &lloc.Adjustment)
	return lloc, err
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
		lloc, err := llocFromString(scanner.Text())
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

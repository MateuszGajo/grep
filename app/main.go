package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Ensures gofmt doesn't remove the "bytes" import above (feel free to remove this!)
var _ = bytes.ContainsAny

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}
	fmt.Println("1.0")

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}
	fmt.Println("1")

	ok, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
	fmt.Println("1")

	if !ok {
		os.Exit(1)
	}
	fmt.Println("3")
	// default exit code is 0 which means success
}

func hasDigit(line []byte) bool {
	for _, b := range line {
		if b >= '0' && b <= '9' {
			return true
		}
	}

	return false
}

func hasAlphaNumeric(line []byte) bool {
	for _, b := range line {
		if b >= '0' && b <= '9' {
			return true
		}
		if b >= 'a' && b <= 'z' {
			return true
		}

		if b >= 'A' && b <= 'Z' {
			return true
		}

		if b == '_' {
			return true
		}
	}

	return false
}

type RangeStruct struct {
	start byte
	end   byte
}

func groupSearch(line []byte, pattern string) (bool, error) {
	ranges := []RangeStruct{}
	singleChars := []byte{}

	pattern = pattern[1 : len(pattern)-1]

	for i := 0; i < len(pattern); i++ {
		if i+1 < len(pattern) && pattern[i+1] == '-' {
			if i+2 >= len(pattern) {
				return false, fmt.Errorf("invalid group structure expected START-TO format")
			}
			singleRange := RangeStruct{
				start: pattern[i],
				end:   pattern[i+2],
			}
			ranges = append(ranges, singleRange)
			i += 2
		} else {
			singleChars = append(singleChars, pattern[i])
		}
	}

	for _, b := range line {
		for _, c := range ranges {
			if b >= c.start && b <= c.end {
				return true, nil
			}
		}
		for _, c := range singleChars {
			if c == b {
				return true, nil
			}
		}
	}

	return false, nil
}

func matchLine(line []byte, pattern string) (bool, error) {

	if pattern[0] == '[' && pattern[len(pattern)-1] != ']' {
		return false, fmt.Errorf("postivie characters group should end with ]")
	}

	var ok bool
	var err error
	switch {
	case pattern == "\\d":
		ok = hasDigit(line)
	case pattern == "\\w":
		ok = hasAlphaNumeric(line)
	case pattern[0] == '[':
		ok, err = groupSearch(line, pattern)

	default:
		ok = bytes.ContainsAny(line, pattern)
	}

	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	return ok, err
}

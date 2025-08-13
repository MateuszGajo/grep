package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2)
	}

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	ok, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if !ok {
		os.Exit(1)
	}

}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func reverseBytes(b []byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}

func matchLine(line []byte, pattern string) (bool, error) {
	parser := Parser{}
	if pattern[0] == '^' {
		parser.isStartAnchor = true
		pattern = pattern[1:]
	}
	if pattern[len(pattern)-1] == '$' {
		parser.isEndAnchor = true
		pattern = pattern[:len(pattern)-1]
	}

	if parser.isStartAnchor {
		return match2(line, pattern, parser, true, parser.isEndAnchor)
	} else if !parser.isEndAnchor {
		return match2(line, pattern, parser, false, false)
	}
	pattern = reverseString(pattern)
	reverseBytes(line)

	return match2(line, pattern, parser, true, false)

}

func match2(line []byte, pattern string, parser Parser, checkFirstIteration bool, checkTillEnd bool) (bool, error) {
	err := parser.parse(pattern)
	if err != nil {
		panic(err)
	}
mainLoop:
	for i := 0; i < len(line); i++ {
		startIndex := i
		pi := 0

		for pi < len(parser.patterns) {
			item := parser.patterns[pi]
			_, endIndex := item.Match(line, startIndex)
			if endIndex == -1 {
				if checkFirstIteration {
					return false, nil
				}

				// else {
				// 	continue mainLoop
				// }

				if startIndex+1 <= len(line) {
					startIndex++
					continue
				}

				continue mainLoop

			}
			startIndex = endIndex + 1
			pi++
		}
		// for _, item := range parser.patterns {
		// 	// we can add extra sliding window if idnex from to
		// 	// get index from parse.patterns and iterate till the end, if error decrement index and try again, till startIndex, if index == startIndex, exit and run rest of code as normal, so if we only has one matching index do not extra sliding window
		// 	// }

		// }
		if checkTillEnd {
			if startIndex == len(line) {
				return true, nil
			} else {
				return false, nil
			}
		} else {
			return true, nil
		}

	}

	return false, nil
}

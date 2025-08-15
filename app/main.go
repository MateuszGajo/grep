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
		return matchLineInternal(line, pattern, parser, true, parser.isEndAnchor)
	} else if !parser.isEndAnchor {
		return matchLineInternal(line, pattern, parser, false, false)
	}
	pattern = reverseString(pattern)
	reverseBytes(line)

	return matchLineInternal(line, pattern, parser, true, false)
}

func matchLineInternal(line []byte, pattern string, parser Parser, checkFirstIteration bool, checkTillEnd bool) (bool, error) {
	err := parser.parse(pattern)
	if err != nil {
		panic(err)
	}
mainLoop:
	for i := 0; i < len(line); i++ {
		startIndex := i
		endIndex := matchPatterns(parser.patterns, 0, line, startIndex)
		if endIndex == -1 {
			if checkFirstIteration {
				return false, nil
			} else {
				continue mainLoop
			}

		}
		startIndex = endIndex
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

func matchPatterns(items []Pattern, patternId int, line []byte, startIndex int) int {
	var previousPattern Pattern
	if patternId > 0 {
		previousPattern = items[patternId-1]
	}
	start, endIndex := items[patternId].Match(line, startIndex, previousPattern)
	if endIndex == -1 {
		return -1
	}
	if patternId == len(items)-1 {
		return endIndex + 1
	}

	patternId++
	for i := endIndex + 1; i >= start+1; i-- {
		returnVal := matchPatterns(items, patternId, line, i)

		if returnVal != -1 {
			return returnVal
		}

	}

	return -1

}

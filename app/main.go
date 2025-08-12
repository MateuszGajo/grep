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

// func hasAlphaNumeric(line []byte) bool {
// 	for _, b := range line {
// 		if b >= '0' && b <= '9' {
// 			return true
// 		}
// 		if b >= 'a' && b <= 'z' {
// 			return true
// 		}

// 		if b >= 'A' && b <= 'Z' {
// 			return true
// 		}

// 		if b == '_' {
// 			return true
// 		}
// 	}

// 	return false
// }

type RangeStruct struct {
	start byte
	end   byte
}

// func getGroups(line []byte, pattern string) ([]RangeStruct, []byte, error) {
// 	ranges := []RangeStruct{}
// 	singleChars := []byte{}
// 	for i := 0; i < len(pattern); i++ {
// 		if i+1 < len(pattern) && pattern[i+1] == '-' {
// 			if i+2 >= len(pattern) {
// 				return nil, nil, fmt.Errorf("invalid group structure expected START-TO format")
// 			}
// 			singleRange := RangeStruct{
// 				start: pattern[i],
// 				end:   pattern[i+2],
// 			}
// 			ranges = append(ranges, singleRange)
// 			i += 2
// 		} else {
// 			singleChars = append(singleChars, pattern[i])
// 		}
// 	}

// 	return ranges, singleChars, nil
// }

// func groupSearch(line []byte, pattern string, isNegativeLookup bool) (bool, error) {

// 	ranges, singleChars, err := getGroups(line, pattern)
// 	if err != nil {
// 		return false, err
// 	}

// 	if !isNegativeLookup {
// 		for _, b := range line {
// 			for _, c := range ranges {
// 				if b >= c.start && b <= c.end {
// 					return true, nil
// 				}
// 			}
// 			for _, c := range singleChars {
// 				if c == b {
// 					return true, nil
// 				}
// 			}
// 		}

// 		return false, nil
// 	} else {

// 	mainLoop:
// 		for _, b := range line {
// 			for _, c := range ranges {
// 				if b >= c.start && b <= c.end {
// 					continue mainLoop
// 				}
// 			}
// 			for _, c := range singleChars {
// 				if c == b {
// 					continue mainLoop
// 				}
// 			}

// 			return true, nil
// 		}

// 		return false, nil
// 	}

// }

// func handleGroup(line []byte, pattern string) (bool, error) {
// 	if pattern[len(pattern)-1] != ']' {
// 		return false, fmt.Errorf("gorup pattern should end with ]")
// 	}
// 	isNegative := false
// 	pattern = pattern[1 : len(pattern)-1]

// 	if pattern[0] == '^' {
// 		isNegative = true
// 		pattern = pattern[1:]
// 	}

// 	return groupSearch(line, pattern, isNegative)
// }

// groups [fsd] group negative lookup[^]
// single characters lookup: abc
// regex lookup \d \w
// start anchor: ^abcd end anchor: $fdsf

//1. We start with pattern look for first index [0]
// if [ it means its group lookup, validate if there is ]
// if ^ is start anchor
// if \d regex
// rest match chars

// anchor end???

// lets start with parser
// parser collects information how many diffrent phases it need to pass
// then we go phase by phase
// e.g [A-Z] group matches, then returns index next phase is able to pick it up and verify
// there is somewhere main lookup, that takes an input and patterns, matched for first pattern is match it goes to next phase, if only one phase return result

func matchLine(line []byte, pattern string) (bool, error) {

	parser := Parser{}

	parser.parse(pattern)
mainLoop:
	for i := 0; i < len(line); i++ {
		startIndex := i
		for _, item := range parser.patterns {
			endIndex := item.Match(line, startIndex) // index its the last char that matched
			if endIndex == -1 {
				continue mainLoop
			}
			startIndex = endIndex + 1
		}
		return true, nil
	}

	return false, nil
}

// func matchLine(line []byte, pattern string) (bool, error) {

// 	if pattern[0] == '[' && pattern[len(pattern)-1] != ']' {
// 		return false, fmt.Errorf("postivie characters group should end with ]")
// 	}

// 	// treat it as a special case for now, deal with it later
// 	if pattern[0] == '[' {
// 		return handleGroup(line, pattern)
// 	}

// 	startAnchor := false

// 	if (pattern[0]) == '^' {
// 		startAnchor = true
// 		pattern = pattern[1:]
// 	}

// mainLoop:
// 	for i := 0; i < len(line); i++ {
// 		var ok bool
// 		// var err error
// 		tmpI := i
// 		for j := 0; j < len(pattern); j++ {

// 			if pattern[j] == '\\' {
// 				j = j + 1
// 				if pattern[j] == 'd' {
// 					ok = hasDigit([]byte{line[tmpI]})
// 				} else if pattern[j] == 'w' {
// 					ok = hasAlphaNumeric([]byte{line[tmpI]})
// 				}

// 			} else {
// 				ok = pattern[j] == line[tmpI]
// 			}

// 			if !ok {
// 				if startAnchor {
// 					return false, nil
// 				}
// 				continue mainLoop
// 			} else {
// 				tmpI = tmpI + 1
// 				if tmpI == len(line) && j == (len(pattern)-1) {
// 					return true, nil
// 				} else if tmpI >= len(line) {
// 					if startAnchor {
// 						return false, nil
// 					}
// 					continue mainLoop
// 				}
// 			}
// 		}

// 		return true, nil
// 	}

// 	// You can use print statements as follows for debugging, they'll be visible when running tests.
// 	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

// 	// return ok, err
// 	return false, nil
// }

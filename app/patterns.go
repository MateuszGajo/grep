package main

import (
	"fmt"
	"slices"
)

type Pattern interface {
	Match(input []byte, startIndex int, previousPattern Pattern) (int, int)
}

type Range struct {
	from byte
	to   byte
}

type GroupPattern struct {
	ranges     []Range
	extraChars []byte
	isNegative bool
}

func removeDuplicates(input []byte) []byte {
	seen := make(map[byte]bool)
	result := make([]byte, 0, len(input))

	for _, b := range input {
		if !seen[b] {
			seen[b] = true
			result = append(result, b)
		}
	}

	return result
}

func (g *GroupPattern) AddPatterns(pattern string) error {
	chars := []byte{}
	for i := 0; i < len(pattern); i++ {
		if i+2 < len(pattern) && pattern[i+1] == '-' {
			start := pattern[i]
			end := pattern[i+2]

			if start > end {
				return fmt.Errorf("invalid range end")
			}

			for j := start; j <= end; j++ {
				chars = append(chars, j)
			}

			i = i + 2
			continue
		}

		chars = append(chars, pattern[i])
	}
	slices.Sort(chars)
	chars = removeDuplicates(chars)
	rangeStart := 0
	for i := 0; i < len(chars); i++ {
		if i+1 < len(chars) && chars[i]+1 == chars[i+1] {
			continue
		} else {
			if rangeStart < i {
				start := chars[rangeStart]
				end := chars[i]
				newRange := Range{
					from: start,
					to:   end,
				}
				g.ranges = append(g.ranges, newRange)
			} else {
				g.extraChars = append(g.extraChars, chars[i])
			}
			rangeStart = i + 1
		}
	}
	return nil
}

func (g GroupPattern) Match(input []byte, startindex int, previousPattern Pattern) (int, int) {
	if !g.isNegative {
		return g.MatchPositive(input, startindex)
	} else {
		return g.MatchNegative(input, startindex)
	}
}

func (g GroupPattern) MatchPositive(input []byte, startIndex int) (int, int) {

	for i := startIndex; i < len(input); i++ {
		for _, item := range g.ranges {
			if input[i] >= item.from && input[i] <= item.to {
				return i, i
			}
		}

		for _, b := range g.extraChars {
			if input[i] == b {
				return i, i
			}
		}
	}

	return -1, -1
}

func (g GroupPattern) MatchNegative(input []byte, startIndex int) (int, int) {
mainLoop:
	for i := startIndex; i < len(input); i++ {
		for _, item := range g.ranges {
			if input[i] >= item.from && input[i] <= item.to {
				continue mainLoop
			}
		}

		for _, b := range g.extraChars {
			if input[i] == b {
				continue mainLoop
			}
		}

		return i, i
	}

	return -1, -1
}

type CharsPattern struct {
	chars []byte
}

func (g CharsPattern) Match(input []byte, startindex int, previousPattern Pattern) (int, int) {

	for i := startindex; i < len(input); i++ {
		for j := 0; j < len(g.chars); j++ {
			if i+j == len(input) {
				return -1, -1
			}
			if g.chars[j] != input[i+j] {
				return -1, -1
			}

			if j == len(g.chars)-1 {
				return i + j, i + j
			}
		}

	}
	return -1, -1
}

func (g *CharsPattern) AddPaterns(pattern string) error {
	for i := 0; i < len(pattern); i++ {
		g.chars = append(g.chars, pattern[i])
	}

	return nil
}

type DigitPatterns struct {
	length int
}

func (d DigitPatterns) Match(input []byte, startindex int, previousPattern Pattern) (int, int) {
	for i := startindex; i < startindex+d.length; i++ {
		if input[i] < '0' || input[i] > '9' {
			return -1, -1
		}
	}
	return startindex + d.length - 1, startindex + d.length - 1
}

func (d *DigitPatterns) AddPatterns(pattern string) {

	d.length = len(pattern) / 2
}

type AlphaNumericPatterns struct {
	length int
}

func (d AlphaNumericPatterns) Match(input []byte, startindex int, previousPattern Pattern) (int, int) {
	for i := startindex; i < startindex+d.length; i++ {
		if !hasAlphaNumeric(input[i]) {
			return -1, -1
		}
	}
	return startindex + d.length - 1, startindex + d.length - 1
}

func (d *AlphaNumericPatterns) AddPatterns(pattern string) {

	d.length = len(pattern) / 2
}

func hasAlphaNumeric(b byte) bool {
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

	return false
}

type OneOrMorePatterns struct {
}

func (d OneOrMorePatterns) Match(input []byte, startindexPassed int, previousPattern Pattern) (int, int) {
	previousEndIndex := 0
	previousStartIndex := 0
	initStartIndex := startindexPassed - 1
	index := initStartIndex
	for {
		si, ei := previousPattern.Match(input, index, previousPattern)
		if ei == -1 {
			if previousStartIndex != previousEndIndex {
				panic("invalid state")
			}
			break
		}
		previousStartIndex = si
		previousEndIndex = ei
		index++
	}

	return initStartIndex, previousEndIndex

}

func (d *OneOrMorePatterns) AddPatterns(pattern string) {

}

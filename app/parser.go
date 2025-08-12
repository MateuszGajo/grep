package main

import (
	"fmt"
	"strings"
)

type Parser struct {
	patterns []Pattern
}

func (p *Parser) parse(pattern string) error {
	for len(pattern) > 0 {
		switch pattern[0] {
		case '[':
			group := GroupPattern{}
			end := strings.Index(pattern, "]")

			if end == -1 {
				return fmt.Errorf("invalid group syntax")
			}
			// pattern = pattern[1:end]
			if pattern[1] == '^' {
				group.isNegative = true
				pattern = pattern[2:]
				end -= 2
			} else {
				pattern = pattern[1:]
				end -= 1
			}
			err := group.AddPatterns(pattern[:end])
			if err != nil {
				return err
			}
			pattern = pattern[end+1:]
			p.patterns = append(p.patterns, group)
		case '\\':
			regexPatterns := pattern[1]
			i := 0
			for i = 0; i < len(pattern); i += 2 {
				if pattern[i] != '\\' || pattern[i+1] != regexPatterns {
					break
				}
			}
			foundPattern := pattern[:i]
			pattern = pattern[i:]
			switch regexPatterns {
			case 'd':
				digitPattern := DigitPatterns{}
				digitPattern.AddPatterns(foundPattern)
				p.patterns = append(p.patterns, digitPattern)
			case 'w':
				alphaNumericPattern := AlphaNumericPatterns{}
				alphaNumericPattern.AddPatterns(foundPattern)
				p.patterns = append(p.patterns, alphaNumericPattern)
			default:
				panic("not handled \\ for letter, %v" + string(regexPatterns))
			}
		default:
			charPattern := CharsPattern{}
			i := 0
			for i = 0; i < len(pattern); i++ {
				if pattern[i] == '\\' || pattern[i] == '[' {

					break
				}
			}
			charPattern.AddPaterns(pattern[:i])
			pattern = pattern[i:]
			p.patterns = append(p.patterns, charPattern)

		}

	}

	// now what happens if there is more than one pattern????????\
	// either create a new method matchAll and invoke based on number of patterns\
	// What matchAll returns??

	return nil
}

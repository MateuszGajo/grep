package main

import (
	"fmt"
	"strings"
)

type Parser struct {
	patterns      []Pattern
	isStartAnchor bool
	isEndAnchor   bool
}

func (p *Parser) parse(pattern string) error {
	p.patterns = []Pattern{}
	for len(pattern) > 0 {
		switch pattern[0] {
		case '[':
			group := GroupPattern{}
			end := strings.Index(pattern, "]")

			if end == -1 {
				return fmt.Errorf("invalid group syntax")
			}

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
		case '^':
			return fmt.Errorf("anchor ^ only should be at begging")
		case '$':
			return fmt.Errorf("anchor & only should be at the end")
		case '+':
			pattern = pattern[1:]
			oneOrMorePattern := OneOrMorePatterns{previousPattern: p.patterns[len(p.patterns)-1]}
			p.patterns = p.patterns[:len(p.patterns)-1]
			p.patterns = append(p.patterns, oneOrMorePattern)
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
				fmt.Println("youre escpaing not supported option: " + string(regexPatterns) + ";")
			}
		default:

			i := 0
			for i = 0; i < len(pattern); i++ {
				if pattern[i] == '\\' || pattern[i] == '[' || pattern[i] == '$' {
					charPattern := CharsPattern{}
					charPattern.AddPaterns(pattern[:i])
					pattern = pattern[i:]
					p.patterns = append(p.patterns, charPattern)

					break
				}
				if i == len(pattern)-1 {
					charPattern := CharsPattern{}
					charPattern.AddPaterns(pattern[:i+1])
					pattern = pattern[i+1:]
					p.patterns = append(p.patterns, charPattern)
				}
				if i+1 < len(pattern) && pattern[i+1] == '+' {
					charPattern := CharsPattern{}
					charBeforeSpecial := CharsPattern{}
					charPattern.AddPaterns(pattern[:i])
					charBeforeSpecial.AddPaterns(pattern[i : i+1])
					pattern = pattern[i+1:]
					p.patterns = append(p.patterns, charPattern)
					p.patterns = append(p.patterns, charBeforeSpecial)
					break
				}
			}
		}

	}

	return nil
}

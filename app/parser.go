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

func findCorrespondingBracketIndex(pattern string) int {
	openingBracketCount := 0
	closingBracketCount := 0
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '(' {
			openingBracketCount++
		}
		if pattern[i] == ')' {
			closingBracketCount++
		}

		if openingBracketCount == closingBracketCount {
			return i
		}
	}

	return -1
}

func (p *Parser) parse(pattern string) error {
	if p.patterns == nil {

		p.patterns = []Pattern{}
	}
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
		case '?':
			pattern = pattern[1:]
			oneOrZeroPattern := OneOrZeroPatterns{previousPattern: p.patterns[len(p.patterns)-1]}
			p.patterns = p.patterns[:len(p.patterns)-1]
			p.patterns = append(p.patterns, oneOrZeroPattern)
		case '.':
			pattern = pattern[1:]
			wildcardPattern := WildcardPatterns{}
			p.patterns = append(p.patterns, wildcardPattern)
		case '(':
			// ^I see (\\d (cat|dog|cow)s?(, | and )?)+$
			end := findCorrespondingBracketIndex(pattern)
			start := strings.Index(pattern, ")")
			bracketPatternInput := pattern[1:end]
			//
			//
			// () // do we need bracket type or something???? there is a chance we dont
			// we take (, count how many are of ( and then take appropriate ), and whole content run with parse
			// how to handle alternation now??
			// one way is to check then content inside this function agains | if it its alternation running as below
			// second idea is somehow case it agains | so it would be chars cat then dog then cow, but the problem is there is no alternation at the end, so i wouldnt have a way to remove it fro mpatterns and add as alternation (same as we did for +)
			// okay lets check alternation here
			//    (\\d (cat|dog|cow)s?(, | and )?)?
			//
			//

			alternationIndex := strings.Index(bracketPatternInput, "|")
			if alternationIndex != -1 && start == end {
				tmpParser := Parser{}
				patterns := strings.Split(bracketPatternInput, "|")
				for i := 0; i < len(patterns); i++ {
					tmpParser.parse(patterns[i])
				}

				alternationPattern := AlternationPatterns{
					patterns: tmpParser.patterns,
				}
				p.patterns = append(p.patterns, alternationPattern)
			} else {
				tmpParser := Parser{}
				tmpParser.parse(bracketPatternInput)
				bracketPattern := BracketPatterns{
					patterns:  tmpParser.patterns,
					patternId: 0,
				}
				p.patterns = append(p.patterns, &bracketPattern)
			}
			pattern = pattern[end+1:]

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
				if pattern[i] == '\\' || pattern[i] == '[' || pattern[i] == '$' || pattern[i] == '.' || pattern[i] == '(' {
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
				if i+1 < len(pattern) && (pattern[i+1] == '+' || pattern[i+1] == '?') {
					charPattern := CharsPattern{}
					// charBeforeSpecial := CharsPattern{}

					// charBeforeSpecial.AddPaterns(pattern[i : i+1])
					pattern = pattern[i+1:]
					if len(pattern[:i]) != 0 {
						charPattern.AddPaterns(pattern[:i])
						p.patterns = append(p.patterns, charPattern)
					}

					// p.patterns = append(p.patterns, charBeforeSpecial)
					break
				}
			}
		}

	}

	return nil
}

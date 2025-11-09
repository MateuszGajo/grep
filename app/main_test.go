package main

import (
	"fmt"
	"testing"
)

type Data struct {
	pattern string
	input   string
	matches []string
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestLiteralMatching(t *testing.T) {
	data := []Data{
		{
			pattern: "a",
			input:   "apple",
			matches: []string{"a"},
		},
		{
			pattern: "a",
			input:   "dog",
			matches: []string{},
		},
		{
			pattern: "d",
			input:   "dog",
			matches: []string{"d"},
		},
	}

	for _, item := range data {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			regexEngine, _ := NewRegexEngine(item.pattern)
			matches := regexEngine.matchLine([]byte(item.input))
			matchesString := []string{}
			for _, item := range matches {
				matchesString = append(matchesString, string(item))
			}

			if !stringSliceEqual(matchesString, item.matches) {
				t.Errorf("Expected to find these matches: %v, got: %v", item.matches, matchesString)
			}
		})
	}
}

func TestDigitMatching(t *testing.T) {
	data := []Data{
		{
			pattern: "\\d",
			input:   "1",
			matches: []string{"1"},
		},
		{
			pattern: "\\d",
			input:   "123",
			matches: []string{"1", "2", "3"},
		},
		{
			pattern: "\\d",
			input:   "a3",
			matches: []string{"3"},
		},
		{
			pattern: "\\d",
			input:   "a",
			matches: []string{},
		},
	}

	for _, item := range data {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			regexEngine, _ := NewRegexEngine(item.pattern)
			matches := regexEngine.matchLine([]byte(item.input))
			matchesString := []string{}
			for _, item := range matches {
				matchesString = append(matchesString, string(item))
			}

			if !stringSliceEqual(matchesString, item.matches) {
				t.Errorf("Expected to find these matches: %v, got: %v", item.matches, matchesString)
			}
		})
	}

}

func TestWordMatching(t *testing.T) {
	data := []Data{
		{
			pattern: "\\w",
			input:   "12a",
			matches: []string{"1", "2", "a"},
		},
		{
			pattern: "\\w\\w",
			input:   "12ab",
			matches: []string{"12", "ab"},
		},
		{
			pattern: "\\w",
			input:   "1$2",
			matches: []string{"1", "2"},
		},
		{
			pattern: "\\w",
			input:   "%$#",
			matches: []string{},
		},
	}

	for _, item := range data {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			regexEngine, _ := NewRegexEngine(item.pattern)
			matches := regexEngine.matchLine([]byte(item.input))
			matchesString := []string{}
			for _, item := range matches {
				matchesString = append(matchesString, string(item))
			}

			if !stringSliceEqual(matchesString, item.matches) {
				t.Errorf("Expected to find these matches: %v, got: %v", item.matches, matchesString)
			}
		})
	}
}

func TestCharGroupMatching(t *testing.T) {
	data := []Data{
		{
			pattern: "[\\w]",
			input:   "12a",
			matches: []string{"1", "2", "a"},
		},
		{
			pattern: "[\\w\\w]",
			input:   "12ab",
			matches: []string{"1", "2", "a", "b"},
		},
		{
			pattern: "[\\w]",
			input:   "1$2",
			matches: []string{"1", "2"},
		},
		{
			pattern: "[\\2]",
			input:   "%$#",
			matches: []string{},
		},
		{
			pattern: "[^abcd]",
			input:   "abcde",
			matches: []string{"e"},
		},
	}

	for _, item := range data {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			regexEngine, _ := NewRegexEngine(item.pattern)
			matches := regexEngine.matchLine([]byte(item.input))
			matchesString := []string{}
			for _, item := range matches {
				matchesString = append(matchesString, string(item))
			}

			if !stringSliceEqual(matchesString, item.matches) {
				t.Errorf("Expected to find these matches: %v, got: %v", item.matches, matchesString)
			}
		})
	}
}

func TestCombingCharClass(t *testing.T) {
	data := []Data{
		{
			pattern: "\\d\\d\\d apples",
			input:   "sally has 124 apples",
			matches: []string{"124 apples"},
		},
		{
			pattern: "\\d \\w\\w\\ws",
			input:   "sally has 3 dogs",
			matches: []string{"3 dogs"},
		},
		{
			pattern: "\\d\\\\d\\\\d apples",
			input:   "sally has 12 apples",
			matches: []string{},
		},
	}

	for _, item := range data {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			regexEngine, _ := NewRegexEngine(item.pattern)
			matches := regexEngine.matchLine([]byte(item.input))
			matchesString := []string{}
			for _, item := range matches {
				matchesString = append(matchesString, string(item))
			}

			if !stringSliceEqual(matchesString, item.matches) {
				t.Errorf("Expected to find these matches: %v, got: %v", item.matches, matchesString)
			}
		})
	}
}

func TestAnchor(t *testing.T) {
	data := []Data{
		{
			pattern: "^12",
			input:   "123",
			matches: []string{"12"},
		},
		{
			pattern: "^123",
			input:   "234",
			matches: []string{},
		},
		{
			pattern: "^12$",
			input:   "123",
			matches: []string{},
		},
		{
			pattern: "^12$",
			input:   "12",
			matches: []string{"12"},
		},
		{
			pattern: "^log",
			input:   "slog",
			matches: []string{},
		},
	}

	for _, item := range data {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			regexEngine, _ := NewRegexEngine(item.pattern)
			matches := regexEngine.matchLine([]byte(item.input))
			matchesString := []string{}
			for _, item := range matches {
				matchesString = append(matchesString, string(item))
			}

			if !stringSliceEqual(matchesString, item.matches) {
				t.Errorf("Expected to find these matches: %v, got: %v", item.matches, matchesString)
			}
		})
	}
}

func TestPlus(t *testing.T) {
	data := []Data{
		{
			pattern: "a+",
			input:   "aaaa",
			matches: []string{"aaaa"},
		},
		{
			pattern: "ca+t",
			input:   "caat",
			matches: []string{"caat"},
		},
		{
			pattern: "ca+t",
			input:   "caart",
			matches: []string{},
		},
	}

	for _, item := range data {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			regexEngine, _ := NewRegexEngine(item.pattern)
			matches := regexEngine.matchLine([]byte(item.input))
			matchesString := []string{}
			for _, item := range matches {
				matchesString = append(matchesString, string(item))
			}

			if !stringSliceEqual(matchesString, item.matches) {
				t.Errorf("Expected to find these matches: %v, got: %v", item.matches, matchesString)
			}
		})
	}
}

func TestQuestionmark(t *testing.T) {
	data := []Data{
		{
			pattern: "a?",
			input:   "aaaa",
			matches: []string{"a", "a", "a", "a"},
		},
		{
			pattern: "a?b",
			input:   "b",
			matches: []string{"b"},
		},
		{
			pattern: "a?c",
			input:   "b",
			matches: []string{},
		},
	}

	for _, item := range data {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			regexEngine, _ := NewRegexEngine(item.pattern)
			matches := regexEngine.matchLine([]byte(item.input))
			matchesString := []string{}
			for _, item := range matches {
				matchesString = append(matchesString, string(item))
			}

			if !stringSliceEqual(matchesString, item.matches) {
				t.Errorf("Expected to find these matches: %v, got: %v", item.matches, matchesString)
			}
		})
	}
}

func TestAsterik(t *testing.T) {
	data := []Data{
		{
			pattern: "ca*t",
			input:   "ct",
			matches: []string{"ct"},
		},
		{
			pattern: "ca*t",
			input:   "caaat",
			matches: []string{"caaat"},
		},
		{
			pattern: "ca*t",
			input:   "dog",
			matches: []string{},
		},
		{
			pattern: "k\\d*t",
			input:   "kt",
			matches: []string{"kt"},
		},
		{
			pattern: "k\\d*t",
			input:   "k1t",
			matches: []string{"k1t"},
		},
		{
			pattern: "k[abc]*t",
			input:   "kt",
			matches: []string{"kt"},
		},
		{
			pattern: "k[abc]*t",
			input:   "kat",
			matches: []string{"kat"},
		},
		{
			pattern: "k[abc]*t",
			input:   "kabct",
			matches: []string{"kabct"},
		},
		{
			pattern: "k[abc]*t",
			input:   "kxt",
			matches: []string{},
		},
	}

	for _, item := range data {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			regexEngine, _ := NewRegexEngine(item.pattern)
			matches := regexEngine.matchLine([]byte(item.input))
			matchesString := []string{}
			for _, item := range matches {
				matchesString = append(matchesString, string(item))
			}

			if !stringSliceEqual(matchesString, item.matches) {
				t.Errorf("Expected to find these matches: %v, got: %v", item.matches, matchesString)
			}
		})
	}
}

func TestDot(t *testing.T) {
	data := []Data{
		{
			pattern: "a.b",
			input:   "aab",
			matches: []string{"aab"},
		},
		{
			pattern: "a.",
			input:   "aa",
			matches: []string{"aa"},
		},
		{
			pattern: "a.",
			input:   "b",
			matches: []string{},
		},
		{
			pattern: "a.+",
			input:   "accc",
			matches: []string{"accc"},
		},
	}

	for _, item := range data {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			regexEngine, _ := NewRegexEngine(item.pattern)
			matches := regexEngine.matchLine([]byte(item.input))
			matchesString := []string{}
			for _, item := range matches {
				matchesString = append(matchesString, string(item))
			}

			if !stringSliceEqual(matchesString, item.matches) {
				t.Errorf("Expected to find these matches: %v, got: %v", item.matches, matchesString)
			}
		})
	}
}

func TestGroup(t *testing.T) {
	data := []Data{
		{
			pattern: "(a+)",
			input:   "aaa",
			matches: []string{"aaa"},
		},
		{
			pattern: "(a)",
			input:   "a",
			matches: []string{"a"},
		},
		{
			pattern: "(b)",
			input:   "a",
			matches: []string{},
		},
		{
			pattern: "^I see (\\d (cat|dog|cow)s?(, | and )?)+$",
			input:   "I see 1 cat, 2 dogs and 3 cows",
			matches: []string{"I see 1 cat, 2 dogs and 3 cows"},
		},
		{
			pattern: "^I see (\\d (cat|dog|cow)(, | and )?)+$",
			input:   "I see 1 cat, 2 dogs and 3 cows",
			matches: []string{},
		},
	}

	for _, item := range data {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			regexEngine, _ := NewRegexEngine(item.pattern)
			matches := regexEngine.matchLine([]byte(item.input))
			matchesString := []string{}
			for _, item := range matches {
				matchesString = append(matchesString, string(item))
			}

			if !stringSliceEqual(matchesString, item.matches) {
				t.Errorf("Expected to find these matches: %v, got: %v", item.matches, matchesString)
			}
		})
	}
}

func TestAlternation(t *testing.T) {
	data := []Data{
		{
			pattern: "(a|b)",
			input:   "ab",
			matches: []string{"a", "b"},
		},
		{
			pattern: "(abc|def)",
			input:   "abc",
			matches: []string{"abc"},
		},
		{
			pattern: "(abc|r)",
			input:   "aa",
			matches: []string{},
		},
		{
			pattern: "a|b",
			input:   "a",
			matches: []string{"a"},
		},
	}

	for _, item := range data {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			regexEngine, _ := NewRegexEngine(item.pattern)
			matches := regexEngine.matchLine([]byte(item.input))
			matchesString := []string{}
			for _, item := range matches {
				matchesString = append(matchesString, string(item))
			}

			if !stringSliceEqual(matchesString, item.matches) {
				t.Errorf("Expected to find these matches: %v, got: %v", item.matches, matchesString)
			}
		})
	}
}

func TestBackreference(t *testing.T) {
	data := []Data{
		{
			pattern: "(cat) and \\1",
			input:   "cat and cat",
			matches: []string{"cat and cat"},
		},
		{
			pattern: "((c.t|d.g) and (f..h|b..d)), \\2 with \\3, \\1",
			input:   "bat and fish, bat with fish, bat and fish",
			matches: []string{},
		},

		{
			pattern: "^((\\w+) (\\w+)) is made of \\2 and \\3. love \\1$",
			input:   "apple pie is made of apple and pie. love apple pie",
			matches: []string{"apple pie is made of apple and pie. love apple pie"},
		},
		{
			pattern: "((\\w\\w\\w\\w) (\\d\\d\\d)) is doing \\2 \\3 times, and again \\1 times",
			input:   "grep 101 is doing grep 101 times, and again grep 101 times",
			matches: []string{"grep 101 is doing grep 101 times, and again grep 101 times"},
		},
	}

	for _, item := range data {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			regexEngine, _ := NewRegexEngine(item.pattern)
			matches := regexEngine.matchLine([]byte(item.input))
			matchesString := []string{}
			for _, item := range matches {
				matchesString = append(matchesString, string(item))
			}

			if !stringSliceEqual(matchesString, item.matches) {
				t.Errorf("Expected to find these matches: %v, got: %v", item.matches, matchesString)
			}
		})
	}
}

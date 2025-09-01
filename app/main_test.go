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
	}

	for _, item := range data {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			matches, _, _ := matchLine([]byte(item.input), item.pattern)
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
			matches, _, _ := matchLine([]byte(item.input), item.pattern)
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
			matches, _, _ := matchLine([]byte(item.input), item.pattern)
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
			matches, _, _ := matchLine([]byte(item.input), item.pattern)
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
			matches, _, _ := matchLine([]byte(item.input), item.pattern)
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
			matches, _, _ := matchLine([]byte(item.input), item.pattern)
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
			matches, _, _ := matchLine([]byte(item.input), item.pattern)
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
			matches, _, _ := matchLine([]byte(item.input), item.pattern)
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
			matches, _, _ := matchLine([]byte(item.input), item.pattern)
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
		// {
		// 	pattern: "(abc)(def)\\1",
		// 	input:   "abcdef1",
		// 	matches: []string{},
		// },
	}

	for _, item := range data {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			matches, _, _ := matchLine([]byte(item.input), item.pattern)
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
			matches, _, _ := matchLine([]byte(item.input), item.pattern)
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

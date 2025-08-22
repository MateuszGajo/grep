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
			matches, _ := matchLine([]byte(item.input), item.pattern)
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
			matches, _ := matchLine([]byte(item.input), item.pattern)
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

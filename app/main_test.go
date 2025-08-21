package main

import (
	"fmt"
	"testing"
)

type Data struct {
	pattern string
	input   string
}

func TestLiteralMatching(t *testing.T) {
	matchingCases := []Data{
		{
			pattern: "a",
			input:   "apple",
		},
	}

	for _, item := range matchingCases {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			ok, _ := matchLine([]byte(item.input), item.pattern)

			if !ok {
				t.Errorf("Expected %v to match %v pattern", item.input, item.pattern)
			}
		})
	}

	nonMatchingCases := []Data{
		{
			pattern: "a",
			input:   "dog",
		},
	}

	for _, item := range nonMatchingCases {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			ok, _ := matchLine([]byte(item.input), item.pattern)

			if ok {
				t.Errorf("Expected %v to not match %v pattern", item.input, item.pattern)
			}
		})
	}
}

func TestDigitMatching(t *testing.T) {
	matchingCases := []Data{
		{
			pattern: "\\d",
			input:   "1",
		},
		{
			pattern: "\\d",
			input:   "123",
		},
		{
			pattern: "\\d",
			input:   "a3",
		},
	}

	for _, item := range matchingCases {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			ok, _ := matchLine([]byte(item.input), item.pattern)

			if !ok {
				t.Errorf("Expected %v to match %v pattern", item.input, item.pattern)
			}
		})
	}

	nonMatchingCases := []Data{
		{
			pattern: "a",
			input:   "\\d",
		},
	}

	for _, item := range nonMatchingCases {
		t.Run(fmt.Sprintf("Checking input %v, for pattern %v", item.input, item.pattern), func(t *testing.T) {
			ok, _ := matchLine([]byte(item.input), item.pattern)

			if ok {
				t.Errorf("Expected %v to not match %v pattern", item.input, item.pattern)
			}
		})
	}
}

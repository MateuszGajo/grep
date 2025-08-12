package main

import (
	"testing"
)

func TestRanges(t *testing.T) {
	group := GroupPattern{}
	err := group.AddPatterns("A-CDab")

	if err != nil {
		t.Errorf("Error occured %v", err)
	}

	if group.ranges[0].from != 'A' || group.ranges[0].to != 'D' {
		t.Errorf("Expected first range to be from A-D, got: %v-%v", string(group.ranges[0].from), string(group.ranges[0].to))
	}

	if group.ranges[1].from != 'a' || group.ranges[1].to != 'b' {
		t.Errorf("Expected first range to be from a-b, got: %v-%v", string(group.ranges[1].from), string(group.ranges[1].to))

	}

	if len(group.extraChars) != 0 {
		t.Errorf("single chars array should have length 0")
	}

}

func TestRangesAndSingleChars(t *testing.T) {
	group := GroupPattern{}
	err := group.AddPatterns("a-cdgi")

	if err != nil {
		t.Errorf("Error occured %v", err)
	}

	if group.ranges[0].from != 'a' || group.ranges[0].to != 'd' {
		t.Errorf("Expected first range to be from a-b, got: %v-%v", string(group.ranges[0].from), string(group.ranges[0].to))
	}

	if len(group.extraChars) != 2 {
		t.Errorf("single chars array should have length 2")
	}

	if group.extraChars[0] != 'g' {
		t.Errorf("expected first single char to be g")
	}

	if group.extraChars[1] != 'i' {
		t.Errorf("expected first single char to be i")
	}

}

func TestRangesAndSingleChars1(t *testing.T) {
	group := GroupPattern{}
	err := group.AddPatterns("abcA-Z")

	if err != nil {
		t.Errorf("Error occured %v", err)
	}

	if group.ranges[0].from != 'A' || group.ranges[0].to != 'Z' {
		t.Errorf("Expected first range to be from a-c, got: %v-%v", string(group.ranges[0].from), string(group.ranges[0].to))
	}

	if group.ranges[1].from != 'a' || group.ranges[1].to != 'c' {
		t.Errorf("Expected first range to be from A-Z, got: %v-%v", string(group.ranges[1].from), string(group.ranges[1].to))

	}

	if len(group.extraChars) != 0 {
		t.Errorf("single chars array should have length 0")
	}

}

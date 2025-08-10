package main

import "testing"

func TestSingleChar(t *testing.T) {
	input := "apple"
	pattern := "a"
	ok, err := matchLine([]byte(input), pattern)

	if !ok {
		t.Errorf("Expected to found digit in %q, using pattern: %q, got err: %v", input, pattern, err)
	}
}

func TestDigit(t *testing.T) {
	input := "apple123"
	pattern := "\\d"
	ok, err := matchLine([]byte(input), pattern)

	if !ok {
		t.Errorf("Expected to found digit in %q, using pattern: %q, got err: %v", input, pattern, err)
	}
}

func TestAlphanumericShouldPass(t *testing.T) {
	inputs := []string{"dsa", "32??", "alpha-nume3ic", "FDFADSA", "122121", "_"}
	pattern := "\\w"

	for _, input := range inputs {
		t.Run("Should pass for input"+input, func(t *testing.T) {
			ok, err := matchLine([]byte(input), pattern)

			if !ok {
				t.Errorf("Expected to found alphanumeric in %q, using pattern: %q, got err: %v", input, pattern, err)
			}
		})
	}

}

func TestAlphanumericShouldFail(t *testing.T) {
	inputs := []string{"!@#$%^&*()", "😂🔥💥", "-----"}
	pattern := "\\w"

	for _, input := range inputs {
		t.Run("Should pass for input"+input, func(t *testing.T) {
			ok, err := matchLine([]byte(input), pattern)

			if ok {
				t.Errorf("Expected to found not found alphanumeric in %q, using pattern: %q, got err: %v", input, pattern, err)
			}
		})
	}

}

func TestPositiveCharacterGroupsShouldPass(t *testing.T) {
	inputs := []string{"abc", "a", "FDSFSD", "aaaG"}
	pattern := "[abcA-Z]"

	for _, input := range inputs {
		t.Run("Should pass for input"+input, func(t *testing.T) {
			ok, err := matchLine([]byte(input), pattern)

			if !ok {
				t.Errorf("Expected to found postiive character group in %q, using pattern: %q, got err: %v", input, pattern, err)
			}
		})
	}

}

func TestPositiveCharacterGroupsShouldNotPassCharspattern(t *testing.T) {
	inputs := []string{"", "de", "123"}
	pattern := "[abcA-Z]"

	for _, input := range inputs {
		t.Run("Should pass for input"+input, func(t *testing.T) {
			ok, err := matchLine([]byte(input), pattern)

			if ok {
				t.Errorf("Expected to not found postiive character groupin %q, using pattern: %q, got err: %v", input, pattern, err)
			}
		})
	}

	inputs = []string{"[]"}
	pattern = "[blueberry]"

	for _, input := range inputs {
		t.Run("Should pass for input"+input, func(t *testing.T) {
			ok, err := matchLine([]byte(input), pattern)

			if ok {
				t.Errorf("Expected to not found postiive character groupin %q, using pattern: %q, got err: %v", input, pattern, err)
			}
		})
	}

}

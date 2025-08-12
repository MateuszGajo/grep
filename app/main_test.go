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

func TestDigitShouldFail(t *testing.T) {
	input := "apple"
	pattern := "\\d"
	ok, err := matchLine([]byte(input), pattern)

	if ok {
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
	inputs := []string{"a", "FDSFSD", "aaaG"}
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

func TestPositiveCharacterGroupsWithSinglecharsShouldPass(t *testing.T) {
	inputs := []string{"Bab", "aab", "Zab", "FDSFSDab"}
	pattern := "[abcA-Z]ab"

	for _, input := range inputs {
		t.Run("Should pass for input"+input, func(t *testing.T) {
			ok, err := matchLine([]byte(input), pattern)

			if !ok {
				t.Errorf("Expected to found postiive character group in %q, using pattern: %q, got err: %v", input, pattern, err)
			}
		})
	}

}

func TestPositiveCharacterGroupsWithSinglecharsShouldNotPass(t *testing.T) {
	inputs := []string{"ABa"}
	pattern := "[abcA-Z]ab"

	for _, input := range inputs {
		t.Run("Should pass for input"+input, func(t *testing.T) {
			ok, err := matchLine([]byte(input), pattern)

			if ok {
				t.Errorf("Expected to not be found postiive character group in %q, using pattern: %q, got err: %v", input, pattern, err)
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

func TestNegativeCharacterGroupsShouldNotPass(t *testing.T) {
	inputs := []string{"dog", "apple"}
	pattern := "[^abc]"

	for _, input := range inputs {
		t.Run("Should pass for input"+input, func(t *testing.T) {
			ok, err := matchLine([]byte(input), pattern)

			if !ok {
				t.Errorf("Expected to found postiive character group in %q, using pattern: %q, got err: %v", input, pattern, err)
			}
		})
	}
}

func TestNegativeCharacterGroupsShouldNotPassCharspattern(t *testing.T) {
	inputs := []string{"bac"}
	pattern := "[^abc]"

	for _, input := range inputs {
		t.Run("Should pass for input"+input, func(t *testing.T) {
			ok, err := matchLine([]byte(input), pattern)

			if ok {
				t.Errorf("Expected to found postiive character group in %q, using pattern: %q, got err: %v", input, pattern, err)
			}
		})
	}

}

type InputPattern struct {
	input   string
	pattern string
}

func TestCombinationCharacterClassesShouldPass(t *testing.T) {
	data := []InputPattern{
		{
			input:   "1 apple",
			pattern: "\\d apple",
		},
		{
			input:   "100 apple",
			pattern: "\\d\\d\\d apple",
		},
		{
			input:   "3 dogs",
			pattern: "\\d \\w\\w\\ws",
		},
		{
			input:   "sally has 124 apples",
			pattern: "\\d\\d\\d apples",
		},
	}

	for _, d := range data {
		t.Run("Should pass for input"+d.input, func(t *testing.T) {
			ok, err := matchLine([]byte(d.input), d.pattern)

			if !ok {
				t.Errorf("Expected to found combnations characters in %q, using pattern: %q, got err: %v", d.input, d.pattern, err)
			}
		})
	}
}

func TestCombinationCharacterClassesShouldNotPass(t *testing.T) {
	data := []InputPattern{
		{
			input:   "1 orange",
			pattern: "\\d apple",
		},
		{
			input:   "1 apple",
			pattern: "\\d\\d\\d apple",
		},
		{
			input:   "3 dog",
			pattern: "\\d \\w\\w\\ws",
		},
	}

	for _, d := range data {
		t.Run("Should pass for input"+d.input, func(t *testing.T) {
			ok, err := matchLine([]byte(d.input), d.pattern)

			if ok {
				t.Errorf("Expected to not found combnations characters in %q, using pattern: %q, got err: %v", d.input, d.pattern, err)
			}
		})
	}
}

func TestStartAnchorShouldNotPass(t *testing.T) {
	data := []InputPattern{
		{
			input:   "slog",
			pattern: "^log",
		},
	}

	for _, d := range data {
		t.Run("Should pass for input"+d.input, func(t *testing.T) {
			ok, err := matchLine([]byte(d.input), d.pattern)

			if ok {
				t.Errorf("Expected to not found combnations characters in %q, using pattern: %q, got err: %v", d.input, d.pattern, err)
			}
		})
	}
}

func TestStartAnchorShouldPass(t *testing.T) {
	data := []InputPattern{
		{
			input:   "log",
			pattern: "^log",
		},
	}

	for _, d := range data {
		t.Run("Should pass for input"+d.input, func(t *testing.T) {
			ok, err := matchLine([]byte(d.input), d.pattern)

			if !ok {
				t.Errorf("Expected to found in %q, using pattern: %q, got err: %v", d.input, d.pattern, err)
			}
		})
	}
}

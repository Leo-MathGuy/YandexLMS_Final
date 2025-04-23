package validation

import (
	"fmt"
	"regexp"
)

const (
	number = iota
	operator
	parentheses
	space
)

// Helper Functions

func e(err error) bool {
	return err != nil
}

func validChar(char rune) (bool, error) {
	return regexp.Match("[0-9+\\-*/()^ .]", []byte{byte(char)})
}

func runeType(char rune) (int, error) {
	type RuneType struct {
		pattern string
		number  int
	}

	matches := []RuneType{
		{"[0-9.]", number},
		{"()", parentheses},
		{"+-*/^", operator},
		{" ", space},
	}

	for _, v := range matches {
		match, err := regexp.Match(v.pattern, []byte{byte(char)})

		if e(err) {
			return -1, err
		}

		if match {
			return v.number, nil
		}
	}

	return -1, fmt.Errorf("no match")
}

// Main Steps

func Validate(expression string) error {
	expr := []rune(expression)

	if len(expr) == 0 {
		return fmt.Errorf("empty expression")
	}

	for _, v := range expr {
		charValid, err := validChar(v)

		if e(err) {
			return fmt.Errorf("error regexing %v", v)
		}

		if !charValid {
			return fmt.Errorf("invalid character: %v", v)
		}
	}

	return nil
}

func Separate(expression []rune) [][]rune {
	result := make([][]rune, 0)
	current := make([]rune, 0)

	//currentType := -1

	for _, v := range expression {
		if len(current) == 0 {
			current = append(current, v)
			continue
		}

	}
	return result
}

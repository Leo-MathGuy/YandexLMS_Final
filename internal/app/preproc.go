package app

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	number = iota
	operator
	parentheses
	space
)

// // Helper Functions

// Error
func e(err error) bool {
	return err != nil
}

func validChar(char rune) (bool, error) {
	return regexp.Match("[0-9\\+\\-*/()^ .]", []byte{byte(char)})
}

func charType(char rune) (int, error) {
	type RuneType struct {
		pattern string
		number  int
	}

	matches := []RuneType{
		{"[0-9.]", number},
		{"[()]", parentheses},
		{"[+\\-*\\/^]", operator},
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

// // Main Steps

// Big function
func Validate(expression string) error {
	expr := []rune(expression)

	if len(expr) == 0 {
		return fmt.Errorf("empty expression")
	}

	for _, v := range expr {
		charValid, err := validChar(v)

		if e(err) {
			return fmt.Errorf("error regexing %v - %s", v, err.Error())
		}

		if !charValid {
			return fmt.Errorf("invalid character: %v", v)
		}
	}

	separated, err := Separate(expr)
	if e(err) {
		return fmt.Errorf("cannot separate: %s", err.Error())
	}

	types := make([]int, 0)

	for _, v := range separated {
		t, err := charType(v[0])
		if e(err) {
			return fmt.Errorf("unexpected error during number check")
		}
		if t == number && !validateNumber(string(v)) {
			return fmt.Errorf("invalid number: %s", string(v))
		}

		types = append(types, t)
	}

	if len(types) == 1 {
		if types[0] == number {
			return nil
		} else {
			return fmt.Errorf("parsing failure")
		}
	}

	if !parse(types, separated) {
		return fmt.Errorf("parsing failure")
	}

	return nil
}

// Separate, not validate
func Separate(expression []rune) ([][]rune, error) {
	result := make([][]rune, 0)
	current := make([]rune, 0)

	currentType := -1

	for _, v := range expression {
		t, err := charType(v)

		if e(err) {
			return result, fmt.Errorf("cannot understand %v", v)
		}

		if len(current) == 0 {
			current = append(current, v)
			currentType = t
			continue
		}

		if currentType == t && t == number {
			current = append(current, v)
			continue
		} else {
			result = append(result, current)
			current = []rune{v}
			currentType = t
		}
	}

	result = append(result, current)
	return result, nil
}

// Assuming that previous steps were done
func validateNumber(str string) bool {
	return str[len(str)-1] != '.' && strings.Count(str, ".") <= 1
}

// Simulate calculator tokenization with basic what-fails-after-what logic and parentheses checking
func parse(types []int, separated [][]rune) bool {
	var prev int = -1
	op := false
	unary := false
	parenStack := 0
	spaceOffset := 0

	for i, cur := range types {
		if cur == parentheses {
			if separated[i][0] == '(' {
				parenStack++
			} else {
				parenStack--
			}
		}

		if prev == -1 {
			if cur != space {
				prev = cur
			}
			continue
		}

		switch prev {
		case number:
			switch cur {
			case number:
				return false // Num after num
			case parentheses:
				if separated[i-1-spaceOffset][0] == '(' {
					return false // ( after num
				}
			}
		case parentheses:
			switch cur {
			case number:
				if separated[i-1-spaceOffset][0] == ')' {
					return false // Num after )
				}
			case parentheses:
				if separated[i-1-spaceOffset][0] != separated[i][0] {
					return false // () or )(
				}
			}
		case operator:
			switch cur {
			case operator:
				if separated[i][0] != '-' {
					return false // Op after op
				} else if !unary {
					unary = true
				} else {
					return false // Too many -
				}
			case parentheses:
				if separated[i][0] == ')' {
					return false // ) after op
				}
			}
		}

		if cur == parentheses {
			spaceOffset = 0
		}

		if cur != operator {
			unary = false
		}

		if cur != space {
			prev = cur
			op = cur == operator
		} else {
			spaceOffset++
		}
	}

	return !op && parenStack == 0
}

// // Tokenization
type ExprToken struct {
	tokenType int
	valueF    *float64
	valueI    *int
}

// Tokenize an expression for generation
func Tokenize(separated [][]rune) ([]ExprToken, error) {
	result := make([]ExprToken, 0)

	for _, v := range separated {
		ct, err := charType(v[0])
		if e(err) {
			return result, fmt.Errorf("charType error in tokenization")
		}

		switch ct {
		case number:
			value, err := strconv.ParseFloat(string(v), 64)
			if e(err) {
				return result, fmt.Errorf("cannot parse number: %s", string(v))
			}
			result = append(result, ExprToken{ct, &value, nil})
		case parentheses, operator:
			value := int(rune(v[0]))
			result = append(result, ExprToken{ct, nil, &value})
		}
	}

	return result, nil
}

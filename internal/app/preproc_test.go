package preproc

import (
	"strconv"
	"strings"
	"sync"
	"testing"
)

type CharValidationTest struct {
	char     rune
	expected bool
}

var charFail []CharValidationTest = []CharValidationTest{
	{'$', false},
	{'&', false},
	{'\\', false},
	{';', false},
	{'⠀', false}, // U+2800
	{'🗿', false},
}

var charPass []CharValidationTest = []CharValidationTest{
	{'+', true},
	{'-', true},
	{' ', true},
	{'/', true},
	{'^', true},
	{'*', true},
}

func TestValidChar(t *testing.T) {
	t.Parallel()

	var tests []CharValidationTest = make([]CharValidationTest, 0)

	tests = append(
		tests,
		charPass...,
	)

	for i := range 10 {
		tests = append(tests, CharValidationTest{rune(strconv.FormatInt(int64(i), 10)[0]), true})
	}

	tests = append(tests, charFail...)

	wg := sync.WaitGroup{}

	for _, test := range tests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if real, err := validChar(test.char); real != test.expected {
				t.Errorf("'%s' - got %t want %t", string(test.char), real, test.expected)
			} else if err != nil {
				t.Errorf("'%s' - %s", string(test.char), err.Error())
			}
		}()
	}
	wg.Wait()
}

type CharTypeTest struct {
	char     rune
	expected int
}

var typeFail []CharTypeTest = []CharTypeTest{
	{'$', -1},
	{'&', -1},
	{'\\', -1},
	{';', -1},
	{'⠀', -1}, // U+2800
	{'🗿', -1},
}

var typePass []CharTypeTest = []CharTypeTest{
	{'-', operator},
	{'+', operator},
	{'/', operator},
	{'*', operator},
	{'^', operator},
	{'(', parentheses},
	{')', parentheses},
	{' ', space},
}

func TestCharType(t *testing.T) {
	t.Parallel()

	var tests []CharTypeTest = make([]CharTypeTest, 0)

	tests = append(
		tests,
		typePass...,
	)

	for i := range 10 {
		tests = append(tests, CharTypeTest{rune(strconv.FormatInt(int64(i), 10)[0]), number})
	}

	tests = append(tests, typeFail...)

	wg := sync.WaitGroup{}

	for _, test := range tests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if real, err := charType(test.char); real != test.expected {
				t.Errorf("'%s' - got %d want %d", string(test.char), real, test.expected)
			} else if err != nil && test.expected != -1 {
				t.Errorf("'%s' - %s", string(test.char), err.Error())
			}
		}()
	}
	wg.Wait()
}

type SepTest struct {
	name     string
	expr     string
	expected string
}

var sepPass []SepTest = []SepTest{
	{"simple", "2+2", "2,+,2"},
	{"long", "5158812-125", "5158812,-,125"},
	{"complex", "(-2^3 + (5 * (-4))^2) / -3", "(,-,2,^,3, ,+, ,(,5, ,*, ,(,-,4,),),^,2,), ,/, ,-,3"},
	{"big", "-(42^52 * ((-533 * 155) + (-451^2531))) ^ 251", "-,(,42,^,52, ,*, ,(,(,-,533, ,*, ,155,), ,+, ,(,-,451,^,2531,),),), ,^, ,251"},
}
var sepFail []SepTest = []SepTest{
	{"symbol", "2+&2", ""},
	{"symbollong", "5158🗿8112-125", ""},
}

func TestSeparate(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}

	for _, test := range sepPass {
		wg.Add(1)
		go func() {
			defer wg.Done()
			separated, err := Separate([]rune(test.expr))

			if e(err) {
				t.Errorf("Test %s: %s", test.name, err.Error())
			}

			resulting := ""
			for _, word := range separated {
				resulting += string(word) + ","
			}
			resulting = strings.TrimRight(resulting, ",")

			if resulting != test.expected {
				t.Errorf("Test %s:\n - got  %s\n - want %s", test.name, resulting, test.expected)
			}
		}()
	}

	for _, test := range sepFail {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if _, err := Separate([]rune(test.expr)); !e(err) {
				t.Errorf("Test %s: Expected error", test.name)
			}
		}()
	}
	wg.Wait()
}

type NumTest struct {
	expr     string
	expected bool
}

var numTests []NumTest = []NumTest{
	{"2", true},
	{"5916", true},
	{"2.51", true},
	{"21524.035216", true},
	{".15335", true},
	{"13751.", false},
	{"135.1536.13", false},
}

func TestValidateNumber(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}

	for _, test := range numTests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := validateNumber(test.expr)

			if result != test.expected {
				t.Errorf("Test %s:\n - got  %t\n - want %t", test.expr, result, test.expected)
			}
		}()
	}

	wg.Wait()
}

type ParseTest struct {
	expr     string
	expected bool
}

var parseTests []ParseTest = []ParseTest{
	{"(, , ,-, ,15.2, , ,^, , ,2, , ,), , ,*, , ,(, , ,-, ,27.4, , ,+, , ,(, , ,8.6, , ,*, , ,-, ,4.1, , ,), , ,)", true},
	{"-, ,(, ,9.7, , ,^, , ,0.4, , ,) , ,+, , ,21.3, , ,^, ,(, ,-, ,0.5, , ,)", true},
	{"(, ,-, ,31.5, , ,*, , ,(, ,6.9, , ,^, , ,2, , ,) , ,) , ,/, , ,(, ,3.8, , ,+, , ,-, ,7.2, , ,)", true},
	{"(, ,(, ,4.7, , ,+, , ,-, ,12.8, , ,) , ,^, , ,3, , ,) , ,-, ,-, ,(, ,25.6, , ,*, , ,2.3, , ,)", true},
	{"-, ,(, ,(, ,17.4, , ,/, , ,-, ,5.5, , ,) , ,^, , ,2, , ,) , ,+, , ,(, ,7.1, , ,^, , ,1.7, , ,)", true},
}

func TestParse(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}

	for _, test := range parseTests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			expr1 := strings.Split(test.expr, ",")
			expr2 := make([][]rune, 0)
			for _, v := range expr1 {
				expr2 = append(expr2, []rune(v))
			}

			types := make([]int, 0)
			for _, v := range expr2 {
				ct, err := charType(v[0])
				if e(err) {
					t.Errorf("charType error - %s", err.Error())
				}
				types = append(types, ct)
			}

			result := parse(types, expr2)

			if result != test.expected {
				t.Errorf("Test %s:\n - got  %t\n - want %t", test.expr, result, test.expected)
			}
		}()
	}

	wg.Wait()
}

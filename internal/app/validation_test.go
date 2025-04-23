package validation

import (
	"strconv"
	"sync"
	"testing"
)

type ValidationTest struct {
	char     rune
	expected bool
}

var testsFail []ValidationTest = []ValidationTest{
	{'$', false},
	{'&', false},
	{'\\', false},
	{';', false},
	{'⠀', false}, // U+2800
	{'🗿', false},
}

func TestValidChar(t *testing.T) {
	t.Parallel()

	var tests []ValidationTest = make([]ValidationTest, 0)

	tests = append(
		tests,
		ValidationTest{'+', true},
		ValidationTest{'-', true},
		ValidationTest{' ', true},
		ValidationTest{'/', true},
		ValidationTest{'*', true},
		ValidationTest{'^', true},
	)

	for i := range 10 {
		tests = append(tests, ValidationTest{rune(strconv.FormatInt(int64(i), 10)[0]), true})
	}

	tests = append(tests, testsFail...)

	wg := sync.WaitGroup{}

	for _, test := range tests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if real, err := validChar(test.char); real != test.expected {
				t.Errorf("'%s' - %t =/= %t", string(test.char), real, test.expected)
			} else if err != nil {
				t.Errorf("'%s' - %s", string(test.char), err.Error())
			}
		}()
	}
	wg.Wait()
}

func TestCharType(t *testing.T) {

}

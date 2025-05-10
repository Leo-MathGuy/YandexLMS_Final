package processing

import (
	"strconv"
	"sync"
	"testing"
)

func genTokens(expr string, t *testing.T) []ExprToken {
	sep, err := Separate([]rune(expr))
	if e(err) {
		t.Errorf("Error separating %s: %s", expr, err.Error())
		return nil
	}

	tokens, err := Tokenize(sep)
	if e(err) {
		t.Errorf("Error tokenizing %s: %s", expr, err.Error())
		return nil
	}

	return tokens
}

type EvalTest struct {
	expr     string
	expected string
}

var evalTests []EvalTest = []EvalTest{
	{
		"1*(2.5*2-4/(1+3))",
		"[1*[[2.5*2]-[4/[1+3]]]]",
	},
	{
		"-3 * (4 + -2)",
		"[-3*[4+-2]]",
	},
	{
		"((5 - 1) / 2) * 3",
		"[[[5-1]/2]*3]",
	},
	{
		"-2.5 + 3 * 4 - 1",
		"[[-2.5+[3*4]]-1]",
	},
	{
		"-(2 * (3.5 / -7 + 1))",
		"[-1*[2*[[3.5/-7]+1]]]",
	}, {
		"-(5--3*5+-(4/-2))",
		"[-1*[[5-[-3*5]]+[-1*[4/-2]]]]",
	},
}

func TestEval(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}

	for _, test := range evalTests {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if tokens := genTokens(test.expr, t); tokens != nil {
				res, err := Eval(tokens, NodeGen)

				if res.toString() != test.expected {
					t.Errorf("Got:      " + res.toString())
					t.Errorf("Expected: " + test.expected)
				}

				if e(err) {
					t.Error(err.Error())
				}
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if _, err := Eval(make([]ExprToken, 0), NodeGen); err == nil {
			t.Errorf("Expected error on empty list")
		}
	}()

	wg.Wait()
}

type NodeGenTest struct {
	expr   string
	expect []string
	mode   int
}

var nodeGenTests []NodeGenTest = []NodeGenTest{
	{
		"2+2",
		[]string{
			"2", "2",
		},
		expr,
	},
	{
		"2*2",
		[]string{
			"2*2",
		},
		expr,
	},
	{
		"(2+2)",
		[]string{
			"(2+2)",
		},
		expr,
	},
	{
		"2.5 + 3 * 4 - 1",
		[]string{
			"2.5", "3*4", "1",
		},
		expr,
	},
	{
		"2.5 * (3) / (4 - 1)",
		[]string{
			"2.5", "(3)", "(4-1)",
		},
		term,
	},
}

func fakeNodeGen(result *[][]ExprToken) nodeproc {
	return func(tokens []ExprToken, u uint, n nodeproc, i int) *Node {
		*result = append(*result, tokens)
		return NodeGen(tokens, u, NodeGen, i)
	}
}

func TestNodeGen(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}

	for _, test := range nodeGenTests {
		wg.Add(1)
		go func() {
			defer wg.Done()

			result := make([][]ExprToken, 0)
			if tokens := genTokens(test.expr, t); tokens != nil {
				NodeGen(tokens, uint(test.mode), fakeNodeGen(&result), 0)

				resultTokens := make([]string, 0)

				for _, v := range result {
					r := ""
					for _, w := range v {
						if w.tokenType == number {
							r += strconv.FormatFloat(*w.valueF, 'f', -1, 64)
						} else {
							r += string(rune(*w.valueI))
						}
					}
					resultTokens = append(resultTokens, r)
				}

				if len(result) != len(test.expect) {
					t.Errorf("Lengths dont match: exp: %d got: %d\n Expected:\n", len(tokens), len(result))

					for _, v := range test.expect {
						t.Errorf(" - %s\n", v)
					}

					t.Errorf(" Got:\n")

					for _, v := range resultTokens {
						t.Errorf("\n - %s", v)
					}

					return
				}

				for i, v := range test.expect {
					if v != resultTokens[i] {
						t.Errorf("got: %s need: %s", resultTokens[i], v)
						return
					}
				}
			}
		}()
	}

	wg.Wait()
}

package app

import (
	"sync"
	"testing"
)

type EvalTest struct {
	name     string
	expr     []ExprToken
	expected string
}

var evalTests []EvalTest = []EvalTest{
	{
		"1*(2.5*2-4/(1+3))",
		[]ExprToken{
			{number, FloatPtr(1.0), nil},
			{operator, nil, IntPtr(int('*'))},
			{parentheses, nil, IntPtr(int('('))},
			{number, FloatPtr(2.5), nil},
			{operator, nil, IntPtr(int('*'))},
			{number, FloatPtr(2.0), nil},
			{operator, nil, IntPtr(int('-'))},
			{number, FloatPtr(4.0), nil},
			{operator, nil, IntPtr(int('/'))},
			{parentheses, nil, IntPtr(int('('))},
			{number, FloatPtr(1.0), nil},
			{operator, nil, IntPtr(int('+'))},
			{number, FloatPtr(3.0), nil},
			{parentheses, nil, IntPtr(int(')'))},
			{parentheses, nil, IntPtr(int(')'))},
		},
		"[1*[[2.5*2]-[4/[1+3]]]]",
	},
	{
		"-3 * (4 + -2)",
		[]ExprToken{
			{operator, nil, IntPtr(int('-'))},
			{number, FloatPtr(3.0), nil},
			{operator, nil, IntPtr(int('*'))},
			{parentheses, nil, IntPtr(int('('))},
			{number, FloatPtr(4.0), nil},
			{operator, nil, IntPtr(int('+'))},
			{operator, nil, IntPtr(int('-'))},
			{number, FloatPtr(2.0), nil},
			{parentheses, nil, IntPtr(int(')'))},
		},
		"[-3*[4+-2]]",
	},
	{
		"((5 - 1) / 2) * 3",
		[]ExprToken{
			{parentheses, nil, IntPtr(int('('))},
			{parentheses, nil, IntPtr(int('('))},
			{number, FloatPtr(5.0), nil},
			{operator, nil, IntPtr(int('-'))},
			{number, FloatPtr(1.0), nil},
			{parentheses, nil, IntPtr(int(')'))},
			{operator, nil, IntPtr(int('/'))},
			{number, FloatPtr(2.0), nil},
			{parentheses, nil, IntPtr(int(')'))},
			{operator, nil, IntPtr(int('*'))},
			{number, FloatPtr(3.0), nil},
		},
		"[[[5-1]/2]*3]",
	},
	{
		"-2.5 + 3 * 4 - 1",
		[]ExprToken{
			{operator, nil, IntPtr(int('-'))},
			{number, FloatPtr(2.5), nil},
			{operator, nil, IntPtr(int('+'))},
			{number, FloatPtr(3.0), nil},
			{operator, nil, IntPtr(int('*'))},
			{number, FloatPtr(4.0), nil},
			{operator, nil, IntPtr(int('-'))},
			{number, FloatPtr(1.0), nil},
		},
		"[-2.5+[[3*4]-1]]",
	},
	{
		"-(2 * (3.5 / -7 + 1))",
		[]ExprToken{
			{operator, nil, IntPtr(int('-'))},
			{parentheses, nil, IntPtr(int('('))},
			{number, FloatPtr(2.0), nil},
			{operator, nil, IntPtr(int('*'))},
			{parentheses, nil, IntPtr(int('('))},
			{number, FloatPtr(3.5), nil},
			{operator, nil, IntPtr(int('/'))},
			{operator, nil, IntPtr(int('-'))},
			{number, FloatPtr(7.0), nil},
			{operator, nil, IntPtr(int('+'))},
			{number, FloatPtr(1.0), nil},
			{parentheses, nil, IntPtr(int(')'))},
			{parentheses, nil, IntPtr(int(')'))},
		},
		"[-1*[2*[[3.5/-7]+1]]]",
	}, {
		"-(5--3*5+-(4/-2))",
		[]ExprToken{
			{operator, nil, IntPtr(int('-'))},
			{parentheses, nil, IntPtr(int('('))},
			{number, FloatPtr(5.0), nil},
			{operator, nil, IntPtr(int('-'))},
			{operator, nil, IntPtr(int('-'))},
			{number, FloatPtr(3.0), nil},
			{operator, nil, IntPtr(int('*'))},
			{number, FloatPtr(5.0), nil},
			{operator, nil, IntPtr(int('+'))},
			{operator, nil, IntPtr(int('-'))},
			{parentheses, nil, IntPtr(int('('))},
			{number, FloatPtr(4.0), nil},
			{operator, nil, IntPtr(int('/'))},
			{operator, nil, IntPtr(int('-'))},
			{number, FloatPtr(2.0), nil},
			{parentheses, nil, IntPtr(int(')'))},
			{parentheses, nil, IntPtr(int(')'))},
		},
		"[-1*[5-[[-3*5]+[-1*[4/-2]]]]]",
	},
}

func TestEval(t *testing.T) {
	wg := sync.WaitGroup{}

	for _, test := range evalTests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := Eval(test.expr, NodeGen)

			if res.toString() != test.expected {
				t.Errorf("Got:      " + res.toString())
				t.Errorf("Expected: " + test.expected)
			}

			if e(err) {
				t.Error(err.Error())
			}
		}()
	}

	wg.Wait()
}

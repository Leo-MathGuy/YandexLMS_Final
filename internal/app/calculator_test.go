package app

import (
	"fmt"
	"testing"
)

func TestNodeGen(t *testing.T) {
	// 1*(2.5*2.0-4/(1+3))
	test := []ExprToken{
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
	}

	res, _ := Eval(test, NodeGen)
	fmt.Println("\n-----------------------------")
	fmt.Println("Got:      " + res.toString())
	fmt.Println("Expected: 1*(2.5*2-4/(1+3))")
}

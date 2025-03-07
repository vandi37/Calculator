// This tests cover
//
// 1. `../pkg/parsing/` (lexer, parser, ast)
//
// 2. `../pkg/calc/` (running ast)
//
// 3. `../internal/agent/do` (doing the short tasks)
//
// In this case i think it isn't necessary to create separate tests for this modules
package main

import (
	"testing"

	"github.com/vandi37/Calculator/internal/agent/do"
	"github.com/vandi37/Calculator/internal/agent/module"
	"github.com/vandi37/Calculator/pkg/calc"
	"github.com/vandi37/Calculator/pkg/parsing/tree"
)

func send(arg1, arg2 float64, operation tree.ExprSep) (<-chan float64, error) {
	getter := make(chan float64)
	go func() {
		getter <- do.Do(module.Request{Id: 0, Arg1: arg1, Arg2: arg2, Operation: operation, OperationTimeMs: 1})
	}()
	return getter, nil
}

func TestCalc(t *testing.T) {
	testCases := []struct {
		expression string
		expected   float64
	}{
		{"1+1", 1 + 1},
		{"3+3*6", 3 + 3*6},
		{"1+8/2*4", 1 + 8/2*4},
		{"(1+1) *2", (1 + 1) * 2},
		{"((1+4) * (1+2) +10) *4", ((1+4)*(1+2) + 10) * 4},
		{"(4+3+2)/(1+2) * 10 / 3", (4 + 3 + 2) / (1 + 2) * 10 / 3},
		{"(70/7) * 10 /((3+2) * (3+7)) -2", (70/7)*10/((3+2)*(3+7)) - 2},
		{"((7+1) / (2+2) * 4) / 8 * (32 - ((4+12)*2)) -1", ((7+1)/(2+2)*4)/8*(32-((4+12)*2)) - 1},
		{"-1", -1},
		{"+5", 5},
		{"5+5+5+5+5", 5 + 5 + 5 + 5 + 5},
		{"(1)", 1},
		{"(1+2*(10) + 10)", (1 + 2*(10) + 10)},
		{"((1+2)*(5*(7+3) - 70 / (3+4) * (1+2)) - (8-1)) + (10 * (5-1 * (2+3)))", ((1+2)*(5*(7+3)-70/(3+4)*(1+2)) - (8 - 1)) + (10 * (5 - 1*(2+3)))},
		{"-1+2", -1 + 2},
		{"5+ -1", 5 + -1},
		{"5+ -5 + 7 - -6", 5 + -5 + 7 - -6},
		{"-(5+5)", -(5 + 5)},
		{"-90+90", -90 + 90},
		{"9*-1", 9 * -1},
		{"10*(10/10*-10)", 10 * 10 / 10 * -10},
		{"10*-10", 10 * -10},
	}

	for _, testCase := range testCases {
		t.Run(testCase.expression, func(t *testing.T) {
			pre, err := calc.Pre(testCase.expression)
			if err != nil {
				t.Errorf("Pre(%s) error: %v", testCase.expression, err)
			}
			result, err := calc.Calc(pre, send)
			if err != nil {
				t.Errorf("Calc(%s) error: %v", testCase.expression, err)
			} else if result != testCase.expected {
				t.Errorf("Calc(%s) = %v, want %v", testCase.expression, result, testCase.expected)
			}
		})
	}
}

func TestCalcErrors(t *testing.T) {
	testCases := []string{
		"2*(10+9",
		"not numbs",
		"2r+10b",
		"10*(10+2*(10+2*(3+4) + 3 * (1+3) + 8 )",
		"10**2",
		"67^21",
		"((((((((((1)))))))))",
		"",
		"()",
		"*10",
		"-+",
		"-",
		"'10",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			_, err := calc.Pre(testCase)
			if err == nil {
				t.Errorf("Pre(%s) error is not nil", testCase)
			}
		})
	}
}

func TestZero(t *testing.T) {
	pre, err := calc.Pre("10/0")
	if err != nil {
		t.Errorf("Pre(%s) error: %v", "10/0", err)
	}
	_, err = calc.Calc(pre, send)
	if err == nil {
		t.Errorf("Calc(%s) error is not nil", "10/0")
	}
}

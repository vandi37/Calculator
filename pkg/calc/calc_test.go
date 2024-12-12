package calc_test

import (
	"testing"

	"github.com/VandiKond/Calculator/pkg/calc"
)

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
	}

	for _, testCase := range testCases {
		t.Run(testCase.expression, func(t *testing.T) {
			result, err := calc.Calc(testCase.expression)
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
		"10/0",
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
			_, err := calc.Calc(testCase)
			if err == nil {
				t.Errorf("Calc(%s) error is not nil", testCase)
			}
		})
	}
}

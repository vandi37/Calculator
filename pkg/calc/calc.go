package calc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vandi37/vanerrors"
)

// Error names
const (
	ErrorDivideByZero             = "divide by zero not allowed"
	ErrorUnknownOperator          = "unknown operator"
	ErrorParsingNumber            = "number parsing error"
	ErrorDoingOperation           = "error doing operation"
	ErrorCompletingOrderOperation = "error completing order operation"
	ErrorExpressionCompleting     = "error completing the expression"
	ErrorBracketShouldBeClosed    = "bracket should be closed"
	ErrorBracketOf                = "error getting rid of brackets"
	ErrorBracketShouldBeOpened    = "bracket should be opened"
)

// Creating a new error writer
var errorW = vanerrors.NewW(vanerrors.Options{ShowMessage: true, ShowCause: true, ShowAsJson: true}, vanerrors.EmptyLoggerOptions, vanerrors.EmptyHandler)

var operators = []string{"*", "+", "/", "-"}

// Operation struct
type operation struct {
	num1 float64
	// Operation type (one of operations)
	symbol string
	num2   float64
}

func (o operation) run() (float64, error) {
	var num float64

	switch o.symbol {
	// In case of *, +, - Returning the operation result
	case "*":
		num = o.num1 * o.num2
	case "+":
		num = o.num1 + o.num2
	case "-":
		num = o.num1 - o.num2
	// If it is dividing checking that the number is not zero 0. And making the operation
	case "/":
		if o.num2 == 0 {
			return 0, vanerrors.NewSimple(ErrorDivideByZero, fmt.Sprintf("you can't divide by zero in operation '%s'", o.String()))
		}
		num = o.num1 / o.num2

	default:
		return 0, vanerrors.NewSimple(ErrorUnknownOperator, fmt.Sprintf("the operator '%s' is not allowed in operation '%s'", o.symbol, o.String()))
	}

	return num, nil
}

func (o operation) String() string {
	return strconv.FormatFloat(o.num1, 'f', -1, 64) + o.symbol + strconv.FormatFloat(o.num2, 'f', -1, 64)
}

// Getting operator index
//
// Using find type to find is the result index has more ore priority than the current (for example checking which is bigger/lover)
//
// using stdRes as default result for the func
//
// using getIndex to find index of operator (if index == -1 it means that it does not exist)
func findIndex(str string, findType func(index, result_index int) bool, stdRes int, getIndex func(str, operator string) int) (int, string) {
	ResultIndex := stdRes
	var operatorResult string

	for _, operator := range operators {

		index := getIndex(str, operator)

		// If index == -1 it means that it does not exist
		if index == -1 {
			continue
		}

		// If there are multiple minuses (-1+2; 2*-3; 4/-3) doing after operations
		if operator == "-" && index == 0 {

			index, operator = findIndex(str[1:], findType, stdRes, getIndex)
			index++
		}

		if findType(index, ResultIndex) {
			ResultIndex = index
			operatorResult = operator
		}

	}

	return ResultIndex, operatorResult
}

// Order operation completing
//
// It is expressions like (10+30-2+2; 1-54+43-32+100)
func calcOrdered(expression string) (string, error) {
	for {

		// If the expression can be converted to float it means, that the operation is completed
		_, err := strconv.ParseFloat(expression, 64)
		if err == nil {
			return expression, nil
		}

		index, operator := findIndex(expression, func(i1, i2 int) bool { return i1 < i2 }, len(expression), strings.Index)

		num1, err := strconv.ParseFloat(expression[:index], 64)
		if err != nil {
			return expression, vanerrors.NewSimple(ErrorParsingNumber, fmt.Sprintf("'%s' is not a number in expression '%s'", expression[:index], expression))
		}

		expressionTilEnd := expression[index+1:]

		indexOfEnd, _ := findIndex(expressionTilEnd, func(i1, i2 int) bool { return i1 < i2 }, len(expressionTilEnd), strings.Index)

		num2, err := strconv.ParseFloat(expressionTilEnd[:indexOfEnd], 64)
		if err != nil {
			return expression, vanerrors.NewSimple(ErrorParsingNumber, fmt.Sprintf("'%s' is not a number in expression '%s'", expression[:indexOfEnd], expression))
		}

		opr := operation{num1: num1, symbol: operator, num2: num2}

		result, err := opr.run()
		if err != nil {
			return expression, errorW.New(vanerrors.ErrorData{Name: ErrorDoingOperation, Message: fmt.Sprintf("could not do operation '%s' in expression '%s'", opr.String(), expression), Cause: err})
		}

		expression = strings.Replace(expression, opr.String(), strconv.FormatFloat(result, 'f', -1, 64), 1)
	}
}

// Manages the operation order without brackets
//
// It expect expressions like (10+32*12; 455/4-49*7; 40+32/32)
func calcNotOrdered(expression string) (string, error) {
	for {
		// If the expression can be converted to float it means, that the operation is completed
		_, err := strconv.ParseFloat(expression, 64)
		if err == nil {
			return expression, nil
		}

		indexMul := strings.Index(expression, "*")
		indexDiv := strings.Index(expression, "/")

		// If the indexes do not exist (equal -1)
		if indexDiv == -1 {
			indexDiv = len(expression)
		}
		if indexMul == -1 {
			indexMul = len(expression)
		}

		// If they are the same (they don't exist)
		if indexDiv == indexMul {
			expression, err := calcOrdered(expression)
			if err != nil {
				return expression, errorW.New(vanerrors.ErrorData{Name: ErrorCompletingOrderOperation, Message: fmt.Sprintf("could not do order operation of expression '%s'", expression), Cause: err})
			}
			return expression, nil
		}

		index := -1
		oprType := ""

		// Checking that index of multiplying has more priority than index of division (checking that is les and it exist)
		if indexMul < indexDiv && indexMul != len(expression) {
			index = indexMul
			oprType = "*"
		} else {
			index = indexDiv
			oprType = "/"
		}

		expressionBe4 := expression[:index]
		expressionAfter := expression[index+1:]

		indexBe4, _ := findIndex(expressionBe4, func(i1, i2 int) bool { return i1 > i2 }, -1, strings.LastIndex)
		indexAfter, _ := findIndex(expressionAfter, func(i1, i2 int) bool { return i1 < i2 }, len(expressionAfter), strings.Index)

		num1, err := strconv.ParseFloat(expressionBe4[indexBe4+1:], 64)
		if err != nil {
			return expression, vanerrors.NewSimple(ErrorParsingNumber, fmt.Sprintf("'%s' is not a number in expression '%s'", expression[:indexBe4+1], expression))
		}

		num2, err := strconv.ParseFloat(expressionAfter[:indexAfter], 64)
		if err != nil {
			return expression, vanerrors.NewSimple(ErrorParsingNumber, fmt.Sprintf("'%s' is not a number in expression '%s'", expression[:indexAfter], expression))
		}

		opr := operation{num1: num1, symbol: oprType, num2: num2}

		result, err := opr.run()
		if err != nil {
			return expression, errorW.New(vanerrors.ErrorData{Name: ErrorDoingOperation, Message: fmt.Sprintf("could not do operation '%s' in expression '%s'", opr.String(), expression), Cause: err})
		}

		expression = strings.Replace(expression, opr.String(), strconv.FormatFloat(result, 'f', -1, 64), 1)

	}
}

// Calculates with brackets
//
// Can get any operation. Will return an expression that can be calculated with calcNotOrder.
//
// Can get any expressions with expressions for calcNotOrder in side of  multiple brackets
func calcWithBrackets(expression string) (string, error) {
	for {
		indexClose := strings.Index(expression, ")")

		if indexClose == -1 {
			countOpen := strings.Count(expression, "(")
			if countOpen > 0 {
				return expression, vanerrors.NewSimple(ErrorBracketShouldBeClosed, fmt.Sprintf("not closed bracket in expression '%s'", expression))
			}

			return expression, nil
		}

		subExpression := expression[:indexClose]

		indexOpen := strings.LastIndex(subExpression, "(")

		if indexOpen == -1 {
			return expression, vanerrors.NewSimple(ErrorBracketShouldBeOpened, fmt.Sprintf("not opened bracket in expression '%s'", subExpression))
		}

		subExpression = subExpression[indexOpen+1:]

		subExpression, err := calcNotOrdered(subExpression)
		if err != nil {
			return expression, errorW.New(vanerrors.ErrorData{Name: ErrorExpressionCompleting, Message: fmt.Sprintf("could not complete expression '%s'", subExpression), Cause: err})
		}

		expression = strings.Replace(expression, expression[indexOpen:indexClose+1], subExpression, 1)
	}
}

// Calculator
//
// It does the same thing as calcWithBrackets, but will do the calculation till error or float64
func Calc(expression string) (float64, error) {

	// Deleting spaces
	expression = strings.Replace(expression, " ", "", -1)

	expression, err := calcWithBrackets(expression)
	if err != nil {
		return float64(0), errorW.New(vanerrors.ErrorData{Name: ErrorBracketOf, Message: fmt.Sprintf("error getting rid of brackets in expression '%s'", expression), Cause: err})
	}

	expression, err = calcNotOrdered(expression)
	if err != nil {
		return float64(0), errorW.New(vanerrors.ErrorData{Name: ErrorExpressionCompleting, Message: fmt.Sprintf("could not completing of expression '%s'", expression), Cause: err})
	}

	return strconv.ParseFloat(expression, 64)
}

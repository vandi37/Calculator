package calc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/VandiKond/vanerrors"
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

var errorW = vanerrors.NewW(vanerrors.Options{ShowMessage: true, ShowCause: true}, vanerrors.EmptyLoggerOptions, vanerrors.EmptyHandler)

// All allowed operations
var operators = []string{"*", "+", "/", "-"}

// Operation struct
type Operation struct {
	// Thirst number
	Num1 float64
	// Operation type (one of operations)
	Symbol string
	// Second number
	Num2 float64
}

// Method to do operation
func (Op Operation) ParseOpr() (float64, error) {
	// Making the result num
	var num float64
	// Switching operation types
	switch Op.Symbol {
	// In case of *, +, - Returning the operation result
	case "*":
		num = Op.Num1 * Op.Num2
	case "+":
		num = Op.Num1 + Op.Num2
	case "-":
		num = Op.Num1 - Op.Num2
	// If it is dividing checking that the number is not zero 0. And making the operation
	case "/":
		if Op.Num2 == 0 {
			return 0, vanerrors.NewSimple(ErrorDivideByZero, fmt.Sprintf("you can't divide by zero in operation '%s'", Op.FormatToString()))
		}
		num = Op.Num1 / Op.Num2
	// Default returning an error
	default:
		return 0, vanerrors.NewSimple(ErrorUnknownOperator, fmt.Sprintf("the operator '%s' is not allowed in operation '%s'", Op.Symbol, Op.FormatToString()))
	}
	// returning the result
	return num, nil
}

// Method to create a string of the operation
func (Op Operation) FormatToString() string {
	return strconv.FormatFloat(Op.Num1, 'f', -1, 64) + Op.Symbol + strconv.FormatFloat(Op.Num2, 'f', -1, 64)
}

// Getting operator index by findType and getIndex
func findIndex(str string, findType func(int, int) bool, stand_result int, getIndex func(str, operator string) int) (int, string) {
	// Setting the standard result
	ResultIndex := stand_result
	var operatorResult string
	// Checking all operators
	for _, operator := range operators {
		// Getting the index
		index := getIndex(str, operator)

		// Checking the index existence
		if index == -1 {
			continue
		}

		if operator == "-" && index == 0 {
			// Getting the next operator after -
			index, operator = findIndex(str[1:], findType, stand_result, getIndex)
			index++
		}

		if findType(index, ResultIndex) {
			ResultIndex = index
			operatorResult = operator
		}

	}

	// return the result
	return ResultIndex, operatorResult
}

// Order operation completing
func OrderOperations(expression string) (string, error) {
	for {
		// Checking is the result a number
		_, err := strconv.ParseFloat(expression, 64)
		if err == nil {
			return expression, nil
		}

		// Getting the index
		index, operator := findIndex(expression, func(i1, i2 int) bool { return i1 < i2 }, len(expression), strings.Index)

		// Tying to convert the result to a number
		num1, err := strconv.ParseFloat(expression[:index], 64)
		if err != nil {
			return expression, vanerrors.NewSimple(ErrorParsingNumber, fmt.Sprintf("'%s' is not a number in expression '%s'", expression[:index], expression))
		}

		// Getting the index of the second operator
		expressionTilEnd := expression[index+1:]
		indexOfEnd, _ := findIndex(expressionTilEnd, func(i1, i2 int) bool { return i1 < i2 }, len(expressionTilEnd), strings.Index)
		num2, err := strconv.ParseFloat(expressionTilEnd[:indexOfEnd], 64)
		if err != nil {
			return expression, vanerrors.NewSimple(ErrorParsingNumber, fmt.Sprintf("'%s' is not a number in expression '%s'", expression[:indexOfEnd], expression))
		}

		// Creating the operation data
		opr := Operation{Num1: num1, Symbol: operator, Num2: num2}

		// Doing the operation
		result, err := opr.ParseOpr()
		if err != nil {
			return expression, errorW.New(vanerrors.ErrorData{Name: ErrorDoingOperation, Message: fmt.Sprintf("could not do operation '%s' in expression '%s'", opr.FormatToString(), expression), Cause: err})
		}
		// Replacing the operation with the result
		expression = strings.Replace(expression, opr.FormatToString(), strconv.FormatFloat(result, 'f', -1, 64), 1)
	}
}

// Manages the operation order without brackets
func ManageOrder(expression string) (string, error) {
	for {
		// Checking is the result a number
		_, err := strconv.ParseFloat(expression, 64)
		if err == nil {
			return expression, nil
		}

		// Getting the index of the dividing and multiplying
		indexMul := strings.Index(expression, "*")
		indexDiv := strings.Index(expression, "/")

		// Is they don't exist replacing with the string length
		if indexDiv == -1 {
			indexDiv = len(expression)
		}
		if indexMul == -1 {
			indexMul = len(expression)
		}

		// If they are the same (they don't exist)
		if indexDiv == indexMul {
			expression, err := OrderOperations(expression)
			if err != nil {
				return expression, errorW.New(vanerrors.ErrorData{Name: ErrorCompletingOrderOperation, Message: fmt.Sprintf("could not do order operation of expression '%s'", expression), Cause: err})
			}
			return expression, nil
		}

		// Creating empty index and type
		index := -1
		oprType := ""

		// Setting the type and index
		if indexMul < indexDiv && indexMul != len(expression) {
			index = indexMul
			oprType = "*"
		} else {
			index = indexDiv
			oprType = "/"
		}

		// getting the expression before and after
		expressionBe4 := expression[:index]
		expressionAfter := expression[index+1:]

		// Getting the nearest operation index
		indexBe4, _ := findIndex(expressionBe4, func(i1, i2 int) bool { return i1 > i2 }, -1, strings.LastIndex)
		indexAfter, _ := findIndex(expressionAfter, func(i1, i2 int) bool { return i1 < i2 }, len(expressionAfter), strings.Index)

		// Getting the numbers for the operation
		num1, err1 := strconv.ParseFloat(expressionBe4[indexBe4+1:], 64)
		num2, err2 := strconv.ParseFloat(expressionAfter[:indexAfter], 64)
		if err1 != nil {
			return expression, vanerrors.NewSimple(ErrorParsingNumber, fmt.Sprintf("'%s' is not a number in expression '%s'", expression[:indexBe4+1], expression))
		}
		if err2 != nil {
			return expression, vanerrors.NewSimple(ErrorParsingNumber, fmt.Sprintf("'%s' is not a number in expression '%s'", expression[:indexAfter], expression))
		}

		// Creating the operation
		opr := Operation{Num1: num1, Symbol: oprType, Num2: num2}

		// Doing the operation
		result, err := opr.ParseOpr()
		if err != nil {
			return expression, errorW.New(vanerrors.ErrorData{Name: ErrorDoingOperation, Message: fmt.Sprintf("could not do operation '%s' in expression '%s'", opr.FormatToString(), expression), Cause: err})
		}

		// Replacing the operation result
		expression = strings.Replace(expression, opr.FormatToString(), strconv.FormatFloat(result, 'f', -1, 64), 1)

	}
}

// Gets rid of brackets
func BracketOf(expression string) (string, error) {
	for {
		// Getting the index of the closing bracket
		indexClose := strings.Index(expression, ")")

		// Checking non bracket variant
		if indexClose == -1 {
			countOpen := strings.Count(expression, "(")
			if countOpen > 0 {
				return expression, vanerrors.NewSimple(ErrorBracketShouldBeClosed, fmt.Sprintf("not closed bracket in expression '%s'", expression))
			}

			return expression, nil
		}

		// Creating a sub expression before the closing bracket
		subExpression := expression[:indexClose]

		// Getting the last index of opening bracket in the sub expression
		indexOpen := strings.LastIndex(subExpression, "(")

		if indexOpen == -1 {
			return expression, vanerrors.NewSimple(ErrorBracketShouldBeOpened, fmt.Sprintf("not opened bracket in expression '%s'", subExpression))
		}

		// Creating the expression between two brackets
		subExpression = subExpression[indexOpen+1:]

		// Doing the operation
		subExpression, err := ManageOrder(subExpression)
		if err != nil {
			return expression, errorW.New(vanerrors.ErrorData{Name: ErrorExpressionCompleting, Message: fmt.Sprintf("could not complete expression '%s'", subExpression), Cause: err})
		}

		// Replacing the result
		expression = strings.Replace(expression, expression[indexOpen:indexClose+1], subExpression, 1)
	}
}

// Calculator
func Calc(expression string) (float64, error) {
	// Creating empty error
	var err error = nil

	// Deleting spaces
	expression = strings.Replace(expression, " ", "", -1)

	// Finding the first bracket
	index := strings.Index(expression, "(")
	if index != -1 {

		// Getting rid of brackets
		expression, err = BracketOf(expression)
		if err != nil {
			return float64(0), errorW.New(vanerrors.ErrorData{Name: ErrorBracketOf, Message: fmt.Sprintf("error getting rid of brackets in expression '%s'", expression), Cause: err})
		}
	}

	// Managing the operation without the brackets
	expression, err = ManageOrder(expression)
	if err != nil {
		return float64(0), errorW.New(vanerrors.ErrorData{Name: ErrorExpressionCompleting, Message: fmt.Sprintf("could not completing of expression '%s'", expression), Cause: err})
	}

	return strconv.ParseFloat(expression, 64)
}

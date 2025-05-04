package lexer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vandi37/Calculator/pkg/parsing/lexer"
	"github.com/vandi37/Calculator/pkg/parsing/tokens"
)

func TestLexer_Next(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokens.Token
		wantErr  bool
		errMsg   string
	}{
		{
			name:  "Single addition",
			input: "+",
			expected: []tokens.Token{
				{Kind: tokens.Addition},
			},
		},
		{
			name:  "Single subtraction",
			input: "-",
			expected: []tokens.Token{
				{Kind: tokens.Subtraction},
			},
		},
		{
			name:  "Single multiplication",
			input: "*",
			expected: []tokens.Token{
				{Kind: tokens.Multiplication},
			},
		},
		{
			name:  "Single division",
			input: "/",
			expected: []tokens.Token{
				{Kind: tokens.Division},
			},
		},
		{
			name:  "Single open bracket",
			input: "(",
			expected: []tokens.Token{
				{Kind: tokens.BracketOpen},
			},
		},
		{
			name:  "Single close bracket",
			input: ")",
			expected: []tokens.Token{
				{Kind: tokens.BracketClose},
			},
		},

		{
			name:  "Single digit",
			input: "5",
			expected: []tokens.Token{
				{Kind: tokens.Number, Value: 5.0},
			},
		},
		{
			name:  "Multiple digits",
			input: "123",
			expected: []tokens.Token{
				{Kind: tokens.Number, Value: 123.0},
			},
		},
		{
			name:  "Decimal number with dot",
			input: "3.14",
			expected: []tokens.Token{
				{Kind: tokens.Number, Value: 3.14},
			},
		},
		{
			name:  "Decimal number with comma",
			input: "3,14",
			expected: []tokens.Token{
				{Kind: tokens.Number, Value: 3.14},
			},
		},
		{
			name:  "Number starting with zero",
			input: "0.5",
			expected: []tokens.Token{
				{Kind: tokens.Number, Value: 0.5},
			},
		},

		{
			name:  "Simple expression",
			input: "1+2",
			expected: []tokens.Token{
				{Kind: tokens.Number, Value: 1.0},
				{Kind: tokens.Addition},
				{Kind: tokens.Number, Value: 2.0},
			},
		},
		{
			name:  "Expression with brackets",
			input: "(1+2)*3",
			expected: []tokens.Token{
				{Kind: tokens.BracketOpen},
				{Kind: tokens.Number, Value: 1.0},
				{Kind: tokens.Addition},
				{Kind: tokens.Number, Value: 2.0},
				{Kind: tokens.BracketClose},
				{Kind: tokens.Multiplication},
				{Kind: tokens.Number, Value: 3.0},
			},
		},
		{
			name:  "Complex expression",
			input: "3.14*(2+5.67)/1.2",
			expected: []tokens.Token{
				{Kind: tokens.Number, Value: 3.14},
				{Kind: tokens.Multiplication},
				{Kind: tokens.BracketOpen},
				{Kind: tokens.Number, Value: 2.0},
				{Kind: tokens.Addition},
				{Kind: tokens.Number, Value: 5.67},
				{Kind: tokens.BracketClose},
				{Kind: tokens.Division},
				{Kind: tokens.Number, Value: 1.2},
			},
		},

		{
			name:    "Invalid character",
			input:   "1@2",
			wantErr: true,
			errMsg:  lexer.UnexpectedChar,
		},
		{
			name:    "Multiple decimal points",
			input:   "1.2.3",
			wantErr: true,
			errMsg:  lexer.UnexpectedChar,
		},
		{
			name:    "Decimal point without digits",
			input:   ".",
			wantErr: true,
			errMsg:  lexer.UnexpectedChar,
		},
		{
			name:    "Comma without digits",
			input:   ",",
			wantErr: true,
			errMsg:  lexer.UnexpectedChar,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New([]rune(tt.input))
			ts, err := l.GetTokens()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.Equal(t, len(tt.expected), len(ts), "unexpected number of tokens")

			for i, expectedToken := range tt.expected {
				assert.Equal(t, expectedToken.Kind, ts[i].Kind, "token kind mismatch at position %d", i)
				if expectedToken.Kind == tokens.Number {
					assert.Equal(t, expectedToken.Value, ts[i].Value, "token value mismatch at position %d", i)
				}
			}
		})
	}
}

func TestLexer_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Empty input",
			input:    "",
			expected: true,
		},
		{
			name:     "Non-empty input",
			input:    "1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New([]rune(tt.input))
			assert.Equal(t, tt.expected, l.IsEmpty())
		})
	}
}

func TestLexer_NumberParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "Integer zero",
			input:    "0",
			expected: 0,
		},
		{
			name:     "Positive integer",
			input:    "42",
			expected: 42.0,
		},
		{
			name:     "Decimal with dot",
			input:    "3.14159",
			expected: 3.14159,
		},
		{
			name:     "Decimal with comma",
			input:    "3,14159",
			expected: 3.14159,
		},
		{
			name:     "Decimal starting with zero",
			input:    "0.123",
			expected: 0.123,
		},
		{
			name:     "Large number",
			input:    "1234567890",
			expected: 1234567890,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New([]rune(tt.input))
			token, err := l.Next()
			require.NoError(t, err)
			assert.Equal(t, tokens.Number, token.Kind)
			assert.Equal(t, tt.expected, token.Value)
		})
	}
}

func TestLexer_Combinations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []tokens.Token
	}{
		{
			name:  "Multiple operations",
			input: "1+2-3*4/5",
			expected: []tokens.Token{
				{Kind: tokens.Number, Value: 1.0},
				{Kind: tokens.Addition},
				{Kind: tokens.Number, Value: 2.0},
				{Kind: tokens.Subtraction},
				{Kind: tokens.Number, Value: 3.0},
				{Kind: tokens.Multiplication},
				{Kind: tokens.Number, Value: 4.0},
				{Kind: tokens.Division},
				{Kind: tokens.Number, Value: 5.0},
			},
		},
		{
			name:  "Nested brackets",
			input: "((1+2)*(3-4))",
			expected: []tokens.Token{
				{Kind: tokens.BracketOpen},
				{Kind: tokens.BracketOpen},
				{Kind: tokens.Number, Value: 1.0},
				{Kind: tokens.Addition},
				{Kind: tokens.Number, Value: 2.0},
				{Kind: tokens.BracketClose},
				{Kind: tokens.Multiplication},
				{Kind: tokens.BracketOpen},
				{Kind: tokens.Number, Value: 3.0},
				{Kind: tokens.Subtraction},
				{Kind: tokens.Number, Value: 4.0},
				{Kind: tokens.BracketClose},
				{Kind: tokens.BracketClose},
			},
		},
		{
			name:  "Decimal numbers in expression",
			input: "1.5+2.5*3.0",
			expected: []tokens.Token{
				{Kind: tokens.Number, Value: 1.5},
				{Kind: tokens.Addition},
				{Kind: tokens.Number, Value: 2.5},
				{Kind: tokens.Multiplication},
				{Kind: tokens.Number, Value: 3.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New([]rune(tt.input))
			ts, err := l.GetTokens()
			require.NoError(t, err)
			require.Equal(t, len(tt.expected), len(ts))

			for i, expectedToken := range tt.expected {
				assert.Equal(t, expectedToken.Kind, ts[i].Kind, "token kind mismatch at position %d", i)
				if expectedToken.Kind == tokens.Number {
					assert.Equal(t, expectedToken.Value, ts[i].Value, "token value mismatch at position %d", i)
				}
			}
		})
	}
}

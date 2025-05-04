package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/vandi37/Calculator-Models"
	"github.com/vandi37/Calculator/pkg/parsing/binding"
	"github.com/vandi37/Calculator/pkg/parsing/lexer"
	"github.com/vandi37/Calculator/pkg/parsing/parser"
	"github.com/vandi37/Calculator/pkg/parsing/tokens"
	"github.com/vandi37/Calculator/pkg/parsing/tree"
)

func TestParser_PrimExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected tree.ExpressionType
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Single number",
			input:    "42",
			expected: tree.Num(42),
		},
		{
			name:     "Zero",
			input:    "0",
			expected: tree.Num(0),
		},
		{
			name:     "Decimal number",
			input:    "3.14",
			expected: tree.Num(3.14),
		},

		{
			name:     "Unary plus",
			input:    "+5",
			expected: tree.Num(5),
		},
		{
			name:  "Unary minus",
			input: "-5",
			expected: tree.Expression{
				Left:  tree.Num(0),
				Operation:   tree.Operation(pb.Operation_SUBTRACT),
				Right: tree.Num(5),
			},
		},
		{
			name:  "Multiple unary operators",
			input: "--5",
			expected: tree.Expression{
				Left: tree.Num(0),
				Operation:  tree.Operation(pb.Operation_SUBTRACT),
				Right: tree.Expression{
					Left:  tree.Num(0),
					Operation:   tree.Operation(pb.Operation_SUBTRACT),
					Right: tree.Num(5),
				},
			},
		},
		{
			name:  "Expression in brackets",
			input: "(5+3)",
			expected: tree.Expression{
				Left:  tree.Num(5),
				Operation:   tree.Operation(pb.Operation_ADD),
				Right: tree.Num(3),
			},
		},
		{
			name:     "Nested brackets",
			input:    "((5))",
			expected: tree.Num(5),
		},
		{
			name:    "Empty input",
			input:   "",
			wantErr: true,
			errMsg:  parser.UnexpectedEOF,
		},
		{
			name:    "Unclosed bracket",
			input:   "(5",
			wantErr: true,
			errMsg:  parser.ExpectedKind,
		},
		// { // Wouldn't work in primary expression
		// 	name:    "Extra closing bracket",
		// 	input:   "5)",
		// 	wantErr: true,
		// 	errMsg:  parser.UnexpectedToken,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New([]rune(tt.input))
			tokens, err := l.GetTokens()
			require.NoError(t, err)

			p := parser.New(tokens)
			result, err := p.PrimExpression()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_Expression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected tree.ExpressionType
		wantErr  bool
	}{
		{
			name:  "Simple addition",
			input: "1+2",
			expected: tree.Expression{
				Left:  tree.Num(1),
				Operation:   tree.Operation(pb.Operation_ADD),
				Right: tree.Num(2),
			},
		},
		{
			name:  "Simple subtraction",
			input: "5-3",
			expected: tree.Expression{
				Left:  tree.Num(5),
				Operation:   tree.Operation(pb.Operation_SUBTRACT),
				Right: tree.Num(3),
			},
		},
		{
			name:  "Simple multiplication",
			input: "2*3",
			expected: tree.Expression{
				Left:  tree.Num(2),
				Operation:   tree.Operation(pb.Operation_MULTIPLY),
				Right: tree.Num(3),
			},
		},
		{
			name:  "Simple division",
			input: "6/2",
			expected: tree.Expression{
				Left:  tree.Num(6),
				Operation:   tree.Operation(pb.Operation_DIVIDE),
				Right: tree.Num(2),
			},
		},
		{
			name:  "Multiplication before addition",
			input: "2+3*4",
			expected: tree.Expression{
				Left: tree.Num(2),
				Operation:  tree.Operation(pb.Operation_ADD),
				Right: tree.Expression{
					Left:  tree.Num(3),
					Operation:   tree.Operation(pb.Operation_MULTIPLY),
					Right: tree.Num(4),
				},
			},
		},
		{
			name:  "Parentheses change priority",
			input: "(2+3)*4",
			expected: tree.Expression{
				Left: tree.Expression{
					Left:  tree.Num(2),
					Operation:   tree.Operation(pb.Operation_ADD),
					Right: tree.Num(3),
				},
				Operation:   tree.Operation(pb.Operation_MULTIPLY),
				Right: tree.Num(4),
			},
		},
		{
			name:  "Complex expression 1",
			input: "1+2*3-4/2",
			expected: tree.Expression{
				Left: tree.Expression{
					Left: tree.Num(1),
					Operation:  tree.Operation(pb.Operation_ADD),
					Right: tree.Expression{
						Left:  tree.Num(2),
						Operation:   tree.Operation(pb.Operation_MULTIPLY),
						Right: tree.Num(3),
					},
				},
				Operation: tree.Operation(pb.Operation_SUBTRACT),
				Right: tree.Expression{
					Left:  tree.Num(4),
					Operation:   tree.Operation(pb.Operation_DIVIDE),
					Right: tree.Num(2),
				},
			},
		},
		{
			name:  "Complex expression with unary",
			input: "-1+2*-3",
			expected: tree.Expression{
				Left: tree.Expression{
					Left:  tree.Num(0),
					Operation:   tree.Operation(pb.Operation_SUBTRACT),
					Right: tree.Num(1),
				},
				Operation: tree.Operation(pb.Operation_ADD),
				Right: tree.Expression{
					Left: tree.Num(2),
					Operation:  tree.Operation(pb.Operation_MULTIPLY),
					Right: tree.Expression{
						Left:  tree.Num(0),
						Operation:   tree.Operation(pb.Operation_SUBTRACT),
						Right: tree.Num(3),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New([]rune(tt.input))
			tokens, err := l.GetTokens()
			require.NoError(t, err)

			p := parser.New(tokens)
			result, err := p.Expression(binding.Lowest)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_Build(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected tree.Ast
		wantErr  bool
	}{
		{
			name:  "Simple expression",
			input: "1+1",
			expected: tree.Ast{
				Expression: tree.Expression{
					Left:  tree.Num(1),
					Operation:   tree.Operation(pb.Operation_ADD),
					Right: tree.Num(1),
				},
			},
		},
		{
			name:  "Ignores whitespace",
			input: " 1 + 1 ",
			expected: tree.Ast{
				Expression: tree.Expression{
					Left:  tree.Num(1),
					Operation:   tree.Operation(pb.Operation_ADD),
					Right: tree.Num(1),
				},
			},
		},
		{
			name:    "Invalid expression",
			input:   "1+",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Build(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_Methods(t *testing.T) {
	t.Run("Next and Peek", func(t *testing.T) {
		ts := []tokens.Token{
			{Kind: tokens.Number, Value: 1},
			{Kind: tokens.Addition},
			{Kind: tokens.Number, Value: 2},
		}

		p := parser.New(ts)
		token, ok := p.Peek()
		require.True(t, ok)
		assert.Equal(t, tokens.Number, token.Kind)
		assert.Equal(t, 1.0, token.Value)

		token, ok = p.Next()
		require.True(t, ok)
		assert.Equal(t, tokens.Number, token.Kind)
		assert.Equal(t, 1.0, token.Value)

		token, ok = p.Peek()
		require.True(t, ok)
		assert.Equal(t, tokens.Addition, token.Kind)

		p.Move()
		token, ok = p.Peek()
		require.True(t, ok)
		assert.Equal(t, tokens.Number, token.Kind)
		assert.Equal(t, 2.0, token.Value)

		p.Move()
		token, ok = p.Next()
		require.False(t, ok)
		assert.Equal(t, tokens.EOFToken, token)
	})

	t.Run("expectKindError", func(t *testing.T) {
		ts := []tokens.Token{
			{Kind: tokens.BracketOpen},
			{Kind: tokens.Number, Value: 1},
			{Kind: tokens.BracketClose},
		}

		p := parser.New(ts)
		p.Move()

		err := p.ExpectKindError(tokens.Number)
		require.NoError(t, err)

		err = p.ExpectKindError(tokens.BracketClose)
		require.NoError(t, err)

		err = p.ExpectKindError(tokens.Addition)
		require.Error(t, err)
		assert.Contains(t, err.Error(), parser.ExpectedKind)
	})
}

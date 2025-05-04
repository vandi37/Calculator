package do_test

import (
	"agent/pkg/do"
	"math"
	"testing"
	"time"

	pb "github.com/vandi37/Calculator-Models"
)

const epsilon = 1e-10

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestDo_Addition(t *testing.T) {
	tests := []struct {
		name     string
		arg1     float64
		arg2     float64
		expected float64
	}{
		{"positive numbers", 5.5, 3.2, 8.7},
		{"negative numbers", -2.5, -3.1, -5.6},
		{"mixed numbers", -4.2, 6.3, 2.1},
		{"zero addition", 0, 7.8, 7.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pb.Task{
				Operation:     pb.Operation_ADD,
				Arg1:          tt.arg1,
				Arg2:          tt.arg2,
				OperationTime: 50,
			}

			result, err := do.Do(req)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !floatEquals(result, tt.expected) {
				t.Errorf("Expected %f, got %f (diff: %e)", tt.expected, result, math.Abs(result-tt.expected))
			}
		})
	}
}

func TestDo_Subtraction(t *testing.T) {
	tests := []struct {
		name     string
		arg1     float64
		arg2     float64
		expected float64
	}{
		{"positive numbers", 8.5, 3.2, 5.3},
		{"negative numbers", -2.5, -3.1, 0.6},
		{"mixed numbers", -4.2, 6.3, -10.5},
		{"zero subtraction", 7.8, 0, 7.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pb.Task{
				Operation:     pb.Operation_SUBTRACT,
				Arg1:          tt.arg1,
				Arg2:          tt.arg2,
				OperationTime: 50,
			}

			result, err := do.Do(req)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !floatEquals(result, tt.expected) {
				t.Errorf("Expected %f, got %f (diff: %e)", tt.expected, result, math.Abs(result-tt.expected))
			}
		})
	}
}

func TestDo_Multiplication(t *testing.T) {
	tests := []struct {
		name     string
		arg1     float64
		arg2     float64
		expected float64
	}{
		{"positive numbers", 2.5, 4.0, 10.0},
		{"negative numbers", -3.0, -2.0, 6.0},
		{"mixed numbers", -4.0, 2.5, -10.0},
		{"zero multiplication", 7.8, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pb.Task{
				Operation:     pb.Operation_MULTIPLY,
				Arg1:          tt.arg1,
				Arg2:          tt.arg2,
				OperationTime: 30,
			}

			result, err := do.Do(req)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !floatEquals(result, tt.expected) {
				t.Errorf("Expected %f, got %f (diff: %e)", tt.expected, result, math.Abs(result-tt.expected))
			}
		})
	}
}

func TestDo_Division(t *testing.T) {
	tests := []struct {
		name        string
		arg1        float64
		arg2        float64
		expected    float64
		expectError bool
	}{
		{"normal division", 10.0, 2.0, 5.0, false},
		{"fractional result", 5.0, 2.0, 2.5, false},
		{"negative numbers", -10.0, -2.0, 5.0, false},
		{"mixed signs", -10.0, 2.0, -5.0, false},
		{"divide by zero", 10.0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pb.Task{
				Operation:     pb.Operation_DIVIDE,
				Arg1:          tt.arg1,
				Arg2:          tt.arg2,
				OperationTime: 25,
			}

			result, err := do.Do(req)
			if tt.expectError {
				if err != do.DivisionByZero {
					t.Errorf("Expected DivisionByZero error, got %v", err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !floatEquals(result, tt.expected) {
				t.Errorf("Expected %f, got %f (diff: %e)", tt.expected, result, math.Abs(result-tt.expected))
			}
		})
	}
}

func TestDo_UnknownOperation(t *testing.T) {
	req := &pb.Task{
		Operation:     pb.Operation(-1),
		Arg1:          10,
		Arg2:          2,
		OperationTime: 10,
	}

	result, err := do.Do(req)
	if err != do.UnknownOperation {
		t.Errorf("Expected UnknownOperation error, got %v", err)
	}

	if result != 0 {
		t.Errorf("Expected 0 result for unknown operation, got %f", result)
	}
}

func TestDo_TimingAccuracy(t *testing.T) {
	testCases := []struct {
		name          string
		operationTime int32
	}{
		{"short delay", 50},
		{"medium delay", 100},
		{"longer delay", 500},
	}
	// testCases = slices.Repeat(testCases, 200) // repeat to test more times
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := &pb.Task{
				Operation:     pb.Operation_ADD,
				Arg1:          1,
				Arg2:          1,
				OperationTime: tc.operationTime,
			}

			start := time.Now()
			_, err := do.Do(req)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			elapsed := time.Since(start)
			expected := time.Duration(tc.operationTime) * time.Millisecond
			tolerance := time.Duration(float64(expected) * 0.1) // 10% tolerance, however it still can since testing time is not accurate

			min := expected - tolerance
			max := expected + tolerance

			if elapsed < min || elapsed > max {
				t.Errorf("Expected execution time between %v and %v, got %v", min, max, elapsed)
			}
		})
	}
}

package tests

import (
	"loan-service/internal/utils"
	"math/big"
	"testing"
)

func TestEqual(t *testing.T) {
	tests := []struct {
		name     string
		input    [2]big.Float
		expected bool
	}{
		{
			name:     "case #1: return TRUE if 1 == 1",
			input:    [2]big.Float{*big.NewFloat(1), *big.NewFloat(1)},
			expected: true,
		},
		{
			name:     "case #2: return FALSE if 1 != 0",
			input:    [2]big.Float{*big.NewFloat(1), *big.NewFloat(0)},
			expected: false,
		},
		{
			name:     "case #3: return FALSE if 0 == 0",
			input:    [2]big.Float{*big.NewFloat(0), *big.NewFloat(0)},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.Equal(&tt.input[0], &tt.input[1])
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGreaterThan(t *testing.T) {
	tests := []struct {
		name     string
		input    [2]big.Float
		expected bool
	}{
		{
			name:     "case #1: return TRUE if 10 > 1",
			input:    [2]big.Float{*big.NewFloat(10), *big.NewFloat(1)},
			expected: true,
		},
		{
			name:     "case #2: return FALSE if 1 == 1",
			input:    [2]big.Float{*big.NewFloat(1), *big.NewFloat(1)},
			expected: false,
		},
		{
			name:     "case #3: return FALSE if -1 < 1",
			input:    [2]big.Float{*big.NewFloat(-1), *big.NewFloat(1)},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.GreaterThan(&tt.input[0], &tt.input[1])
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLessThan(t *testing.T) {
	tests := []struct {
		name     string
		input    [2]big.Float
		expected bool
	}{
		{
			name:     "case #1: return TRUE if 1 < 10",
			input:    [2]big.Float{*big.NewFloat(1), *big.NewFloat(10)},
			expected: true,
		},
		{
			name:     "case #2: return FALSE if 1 == 1",
			input:    [2]big.Float{*big.NewFloat(1), *big.NewFloat(1)},
			expected: false,
		},
		{
			name:     "case #3: return FALSE if 1 > -1",
			input:    [2]big.Float{*big.NewFloat(1), *big.NewFloat(-1)},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.LessThan(&tt.input[0], &tt.input[1])
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

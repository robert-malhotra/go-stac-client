package cql2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextParser(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Expression
		expectError bool
	}{
		{
			name:  "basic comparison",
			input: "temperature > 30.5",
			expected: &Comparison{
				Operator: ">",
				Left:     Property{Name: "temperature"},
				Right:    Literal{Value: 30.5},
			},
		},
		{
			name:  "logical AND",
			input: "temp > 30 AND humidity < 80",
			expected: &LogicalOperator{
				Operator: "AND",
				Left: &Comparison{
					Operator: ">",
					Left:     Property{Name: "temp"},
					Right:    Literal{Value: 30.0},
				},
				Right: &Comparison{
					Operator: "<",
					Left:     Property{Name: "humidity"},
					Right:    Literal{Value: 80.0},
				},
			},
		},
		{
			name:  "complex expression",
			input: `(a > 5 OR b < 10) AND NOT status = "active"`, // Changed to double quotes
			expected: &LogicalOperator{
				Operator: "AND",
				Left: &LogicalOperator{
					Operator: "OR",
					Left: &Comparison{
						Operator: ">",
						Left:     Property{Name: "a"},
						Right:    Literal{Value: 5.0},
					},
					Right: &Comparison{
						Operator: "<",
						Left:     Property{Name: "b"},
						Right:    Literal{Value: 10.0},
					},
				},
				Right: &Not{
					Expression: &Comparison{
						Operator: "=",
						Left:     Property{Name: "status"},
						Right:    Literal{Value: "active"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseText(tt.input)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, expr)
		})
	}
}

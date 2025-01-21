package cql2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextSerializer(t *testing.T) {
	tests := []struct {
		name     string
		expr     Expression
		expected string
	}{
		{
			name: "simple comparison",
			expr: Comparison{
				Operator: ">",
				Left:     Property{Name: "temperature"},
				Right:    Literal{Value: 30.5},
			},
			expected: `temperature > 30.5`,
		},
		{
			name: "logical operators with precedence",
			expr: LogicalOperator{
				Operator: "AND",
				Left: Comparison{
					Operator: ">",
					Left:     Property{Name: "temp"},
					Right:    Literal{Value: 30},
				},
				Right: LogicalOperator{
					Operator: "OR",
					Left: Comparison{
						Operator: "<",
						Left:     Property{Name: "humidity"},
						Right:    Literal{Value: 50},
					},
					Right: Not{
						Expression: Comparison{
							Operator: "=",
							Left:     Property{Name: "status"},
							Right:    Literal{Value: "active"},
						},
					},
				},
			},
			expected: `temp > 30 AND (humidity < 50 OR NOT status = "active")`,
		},
		{
			name: "complex nested expressions",
			expr: LogicalOperator{
				Operator: "OR",
				Left: LogicalOperator{
					Operator: "AND",
					Left: Comparison{
						Operator: ">",
						Left:     Property{Name: "a"},
						Right:    Literal{Value: 5},
					},
					Right: Comparison{
						Operator: "<",
						Left:     Property{Name: "b"},
						Right:    Literal{Value: 10},
					},
				},
				Right: Not{
					Expression: LogicalOperator{
						Operator: "OR",
						Left: Comparison{
							Operator: "=",
							Left:     Property{Name: "x"},
							Right:    Literal{Value: 1},
						},
						Right: Comparison{
							Operator: "=",
							Left:     Property{Name: "y"},
							Right:    Literal{Value: 2},
						},
					},
				},
			},
			expected: `(a > 5 AND b < 10) OR NOT (x = 1 OR y = 2)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, err := SerializeText(tt.expr)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, text)
		})
	}
}

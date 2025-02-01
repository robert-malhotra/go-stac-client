package cql2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseText(t *testing.T) {
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
			input: `(a > 5 OR b < 10) AND NOT status = "active"`,
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

func TestParseText_Literals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		propName string
	}{
		{"string literal", `name = "John Doe"`, "name"},
		{"number literal", `age = 30`, "age"},
		{"boolean literal", `active = true`, "active"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseText(tt.input)
			require.NoError(t, err)
			comp, ok := expr.(*Comparison)
			require.True(t, ok, "expected a Comparison")
			prop, ok := comp.Left.(Property)
			require.True(t, ok, "expected left operand to be a Property")
			assert.Equal(t, tt.propName, prop.Name)
		})
	}
}

func TestParseText_Invalid(t *testing.T) {
	invalid := []string{
		`this is not a valid expression`,
		`(unclosed parenthesis`,
		`name == "John"`,
	}
	for _, input := range invalid {
		t.Run(input, func(t *testing.T) {
			_, err := ParseText(input)
			assert.Error(t, err)
		})
	}
}

func TestParseText_Grouped(t *testing.T) {
	expr, err := ParseText(`(a > 5 OR b < 10) AND NOT (status = "active")`)
	require.NoError(t, err)
	_, ok := expr.(*LogicalOperator)
	assert.True(t, ok, "expected a LogicalOperator")
}

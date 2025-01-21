package cql2

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONParser(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Expression
		expectError bool
	}{
		{
			name: "simple comparison",
			input: `{
				"op": ">",
				"args": [
					{"property": "temperature"},
					30.5
				]
			}`,
			expected: &Comparison{
				Operator: ">",
				Left:     Property{Name: "temperature"}, // Struct instead of pointer
				Right:    Literal{Value: 30.5},          // Struct instead of pointer
			},
		},
		{
			name: "nested logical",
			input: `{
                "op": "AND",
                "args": [
                    {"op": ">", "args": [{"property": "temp"}, 30]},
                    {
                        "op": "OR",
                        "args": [
                            {"op": "<", "args": [{"property": "humidity"}, 50]},
                            {"op": "NOT", "args": [{"op": "=", "args": [{"property": "status"}, "active"]}]}
                        ]
                    }
                ]
            }`,
			expected: &LogicalOperator{
				Operator: "AND",
				Left: &Comparison{
					Operator: ">",
					Left:     Property{Name: "temp"},
					Right:    Literal{Value: 30.0},
				},
				Right: &LogicalOperator{
					Operator: "OR",
					Left: &Comparison{
						Operator: "<",
						Left:     Property{Name: "humidity"},
						Right:    Literal{Value: 50.0},
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var compactJSON bytes.Buffer
			err := json.Compact(&compactJSON, []byte(tt.input))
			require.NoError(t, err)

			expr, err := ParseJSON(compactJSON.Bytes())
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, expr)
		})
	}
}

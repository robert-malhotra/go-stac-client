package cql2

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONSerializer(t *testing.T) {
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
			expected: `{
                "op": ">",
                "args": [
                    {"property": "temperature"},
                    30.5
                ]
            }`,
		},
		{
			name: "nested logical",
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
			expected: `{
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test serialization
			jsonData, err := SerializeJSON(tt.expr)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(jsonData))

			// Test deserialization
			parsed, err := DeserializeJSON(jsonData)
			require.NoError(t, err)

			// Normalize for comparison
			originalJSON, _ := json.Marshal(tt.expr)
			parsedJSON, _ := json.Marshal(parsed)
			assert.JSONEq(t, string(originalJSON), string(parsedJSON))
		})
	}
}

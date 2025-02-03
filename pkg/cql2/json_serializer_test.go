package cql2

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJSONRoundTrip verifies that various expressions are correctly
// serialized and then deserialized back to an equivalent expression.
// Some test cases also verify the exact JSON output.
func TestJSONRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		expr     Expression
		expected string // optional expected JSON (whitespace-insensitive)
	}{
		{
			name: "simple comparison",
			expr: Comparison{
				Operator: ">",
				Left:     "temperature",
				Right:    30.5,
			},
			expected: `{"op":">","args":[{"property":"temperature"},30.5]}`,
		},
		{
			name: "nested logical",
			expr: LogicalOperator{
				Operator: "AND",
				Left: Comparison{
					Operator: ">",
					Left:     "temp",
					Right:    30,
				},
				Right: LogicalOperator{
					Operator: "OR",
					Left: Comparison{
						Operator: "<",
						Left:     "humidity",
						Right:    50,
					},
					Right: Not{
						Expression: Comparison{
							Operator: "=",
							Left:     "status",
							Right:    "active",
						},
					},
				},
			},
			expected: `{"op":"AND","args":[{"op":">","args":[{"property":"temp"},30]},{"op":"OR","args":[{"op":"<","args":[{"property":"humidity"},50]},{"op":"NOT","args":[{"op":"=","args":[{"property":"status"},"active"]}]}]}]}`,
		},
		{
			name: "string literal",
			expr: Comparison{
				Operator: "=",
				Left:     "name",
				Right:    "Alice",
			},
		},
		{
			name: "number literal",
			expr: Comparison{
				Operator: ">",
				Left:     "age",
				Right:    30.0,
			},
		},
		{
			name: "boolean literal",
			expr: Comparison{
				Operator: "=",
				Left:     "active",
				Right:    true,
			},
		},
		{
			name: "logical operator",
			expr: LogicalOperator{
				Operator: "AND",
				Left: Comparison{
					Operator: "=",
					Left:     "status",
					Right:    "open",
				},
				Right: Comparison{
					Operator: "<",
					Left:     "priority",
					Right:    5.0,
				},
			},
		},
		{
			name: "NOT operator",
			expr: Not{
				Expression: Comparison{
					Operator: "=",
					Left:     "closed",
					Right:    false,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			// Serialize the expression to JSON.
			data, err := SerializeJSON(tt.expr)
			require.NoError(t, err)
			if tt.expected != "" {
				assert.JSONEq(t, tt.expected, string(data))
			}
			// Deserialize the JSON back to an expression.
			parsed, err := DeserializeJSON(data)
			require.NoError(t, err)

			// Compare the JSON representations of the original and the round-tripped expressions.
			origJSON, err := json.Marshal(tt.expr)
			require.NoError(t, err)
			roundTripJSON, err := json.Marshal(parsed)
			require.NoError(t, err)
			assert.JSONEq(t, string(origJSON), string(roundTripJSON))
		})
	}
}

// TestJSONErrors verifies that error conditions are reported.
func TestJSONErrors(t *testing.T) {
	// Serializing a nil expression should return an error.
	t.Run("nil expression", func(t *testing.T) {
		_, err := SerializeJSON(nil)
		require.Error(t, err)
	})

	errorCases := []struct {
		name string
		data []byte
	}{
		{
			name: "invalid AND args",
			data: []byte(`{"op": "AND", "args": [{"property": "a"}]}`),
		},
		{
			name: "invalid comparison args",
			data: []byte(`{"op": "=", "args": [{"property": "a"}]}`),
		},
	}

	for _, tc := range errorCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			_, err := DeserializeJSON(tc.data)
			require.Error(t, err)
		})
	}
}

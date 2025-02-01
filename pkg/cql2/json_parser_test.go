package cql2

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Original and extended test cases for the JSON parser.

func TestJSONParser_OriginalCases(t *testing.T) {
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
				Left:     Property{Name: "temperature"},
				Right:    Literal{Value: 30.5},
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
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			var compactJSON bytes.Buffer
			err := json.Compact(&compactJSON, []byte(tt.input))
			require.NoError(t, err, "JSON compaction failed")
			expr, err := ParseJSON(compactJSON.Bytes())
			if tt.expectError {
				require.Error(t, err, "Expected error for test case %s", tt.name)
				return
			}
			require.NoError(t, err, "Unexpected error for test case %s", tt.name)
			assert.Equal(t, tt.expected, expr)
		})
	}
}

// When no "op" field is provided, the parser will treat the JSON as a comparison with an empty operator.
func TestJSONParser_MissingOp(t *testing.T) {
	input := `{"args": [ {"property": "temp"}, 25 ]}`
	expr, err := ParseJSON([]byte(input))
	require.NoError(t, err, "Did not expect an error when 'op' field is missing")
	expected := &Comparison{
		Operator: "",
		Left:     Property{Name: "temp"},
		Right:    Literal{Value: 25.0},
	}
	assert.Equal(t, expected, expr)
}

// When an unknown operator is provided, the parser returns a comparison with that operator.
func TestJSONParser_UnknownOperator(t *testing.T) {
	input := `{"op": "UNKNOWN", "args": [ {"property": "temp"}, 25 ]}`
	expr, err := ParseJSON([]byte(input))
	require.NoError(t, err, "Did not expect an error for unknown operator")
	expected := &Comparison{
		Operator: "UNKNOWN",
		Left:     Property{Name: "temp"},
		Right:    Literal{Value: 25.0},
	}
	assert.Equal(t, expected, expr)
}

// Test the case where the right operand is a timestamp object.
// The parser returns the timestamp object as a literal (a map) rather than unwrapping it.
func TestJSONParser_TimestampTransformation(t *testing.T) {
	input := `{"op": ">=", "args": [{"property": "datetime"}, {"timestamp": "2021-04-08T04:39:23Z"}]}`
	expr, err := ParseJSON([]byte(input))
	require.NoError(t, err, "Expected no error for valid timestamp transformation")

	comp, ok := expr.(*Comparison)
	require.True(t, ok, "Expected expression to be a Comparison")
	lit, ok := comp.Right.(Literal)
	require.True(t, ok, "Expected right operand to be a Literal")
	// Instead of unwrapping to a string, the literal is a map with a "timestamp" key.
	tsMap, ok := lit.Value.(map[string]interface{})
	require.True(t, ok, "Expected literal value to be a map")
	assert.Equal(t, "2021-04-08T04:39:23Z", tsMap["timestamp"])
}

// Test a deeply nested expression.
func TestJSONParser_DeeplyNested(t *testing.T) {
	input := `{
		"op": "AND",
		"args": [
			{"op": "=", "args": [{"property": "a"}, 1]},
			{
				"op": "OR",
				"args": [
					{"op": ">", "args": [{"property": "b"}, 2]},
					{
						"op": "AND",
						"args": [
							{"op": "<", "args": [{"property": "c"}, 3]},
							{"op": "=", "args": [{"property": "d"}, 4]}
						]
					}
				]
			}
		]
	}`
	_, err := ParseJSON([]byte(input))
	require.NoError(t, err, "Expected no error for a valid deeply nested expression")
}

// Test for the spatial intersect operator ("s_intersects").
func TestJSONParser_SpatialIntersect(t *testing.T) {
	input := `{
		"op": "s_intersects",
		"args": [
			{"property": "geometry"},
			{
				"type": "Polygon",
				"coordinates": [[[0,0], [1,0], [1,1], [0,1], [0,0]]]
			}
		]
	}`
	expr, err := ParseJSON([]byte(input))
	require.NoError(t, err, "Expected no error for a valid spatial intersect expression")
	expected := &Comparison{
		Operator: "s_intersects",
		Left:     Property{Name: "geometry"},
		Right: Literal{Value: map[string]interface{}{
			"type":        "Polygon",
			"coordinates": []interface{}{[]interface{}{[]interface{}{float64(0), float64(0)}, []interface{}{float64(1), float64(0)}, []interface{}{float64(1), float64(1)}, []interface{}{float64(0), float64(1)}, []interface{}{float64(0), float64(0)}}},
		}},
	}
	assert.Equal(t, expected, expr)
}

// TODO add parsing as time structs
// Test for a datetime comparison operation (using a literal datetime string).
func TestJSONParser_DatetimeComparison(t *testing.T) {
	input := `{
		"op": "=",
		"args": [
			{"property": "datetime"},
			"2021-04-08T04:39:23Z"
		]
	}`
	expr, err := ParseJSON([]byte(input))
	require.NoError(t, err, "Expected no error for a valid datetime comparison expression")
	expected := &Comparison{
		Operator: "=",
		Left:     Property{Name: "datetime"},
		Right:    Literal{Value: "2021-04-08T04:39:23Z"},
	}
	assert.Equal(t, expected, expr)
}

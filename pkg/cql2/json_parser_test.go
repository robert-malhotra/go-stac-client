// File: json_parser_test.go
package cql2

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func compactJSON(t *testing.T, in string) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, json.Compact(&buf, []byte(in)))
	return buf.Bytes()
}

func TestJSONParser_Success(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		verify func(t *testing.T, expr Expression)
	}{
		{
			name:  "simple comparison",
			input: `{"op": ">", "args": [{"property": "temperature"}, 30.5]}`,
			verify: func(t *testing.T, expr Expression) {
				expected := &Comparison{
					Operator: ">",
					Left:     Property{Name: "temperature"},
					Right:    Literal{Value: 30.5},
				}
				assert.Equal(t, expected, expr)
			},
		},
		{
			name: "nested logical",
			input: `{
				"op": "AND",
				"args": [
					{"op": ">", "args": [{"property": "temp"}, 30]},
					{"op": "OR", "args": [
						{"op": "<", "args": [{"property": "humidity"}, 50]},
						{"op": "NOT", "args": [{"op": "=", "args": [{"property": "status"}, "active"]}]}
					]}
				]
			}`,
			verify: func(t *testing.T, expr Expression) {
				expected := &LogicalOperator{
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
				}
				assert.Equal(t, expected, expr)
			},
		},
		{
			name:  "missing op",
			input: `{"args": [{"property": "temp"}, 25]}`,
			verify: func(t *testing.T, expr Expression) {
				expected := &Comparison{
					Operator: "",
					Left:     Property{Name: "temp"},
					Right:    Literal{Value: 25.0},
				}
				assert.Equal(t, expected, expr)
			},
		},
		{
			name:  "unknown operator",
			input: `{"op": "UNKNOWN", "args": [{"property": "temp"}, 25]}`,
			verify: func(t *testing.T, expr Expression) {
				expected := &Comparison{
					Operator: "UNKNOWN",
					Left:     Property{Name: "temp"},
					Right:    Literal{Value: 25.0},
				}
				assert.Equal(t, expected, expr)
			},
		},
		{
			name:  "timestamp transformation",
			input: `{"op": ">=", "args": [{"property": "datetime"}, {"timestamp": "2021-04-08T04:39:23Z"}]}`,
			verify: func(t *testing.T, expr Expression) {
				comp, ok := expr.(*Comparison)
				require.True(t, ok)
				lit, ok := comp.Right.(Literal)
				require.True(t, ok)
				tsMap, ok := lit.Value.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "2021-04-08T04:39:23Z", tsMap["timestamp"])
			},
		},
		{
			name: "spatial intersect",
			input: `{
				"op": "s_intersects",
				"args": [
					{"property": "geometry"},
					{"type": "Polygon", "coordinates": [[[0,0],[1,0],[1,1],[0,1],[0,0]]]}
				]
			}`,
			verify: func(t *testing.T, expr Expression) {
				expected := &Comparison{
					Operator: "s_intersects",
					Left:     Property{Name: "geometry"},
					Right: Literal{Value: map[string]interface{}{
						"type": "Polygon",
						"coordinates": []interface{}{
							[]interface{}{
								[]interface{}{float64(0), float64(0)},
								[]interface{}{float64(1), float64(0)},
								[]interface{}{float64(1), float64(1)},
								[]interface{}{float64(0), float64(1)},
								[]interface{}{float64(0), float64(0)},
							},
						},
					}},
				}
				assert.Equal(t, expected, expr)
			},
		},
		{
			name: "datetime comparison",
			input: `{
				"op": "=",
				"args": [
					{"property": "datetime"},
					"2021-04-08T04:39:23Z"
				]
			}`,
			verify: func(t *testing.T, expr Expression) {
				expected := &Comparison{
					Operator: "=",
					Left:     Property{Name: "datetime"},
					Right:    Literal{Value: "2021-04-08T04:39:23Z"},
				}
				assert.Equal(t, expected, expr)
			},
		},
		{
			name: "deeply nested",
			input: `{
				"op": "AND",
				"args": [
					{"op": "=", "args": [{"property": "a"}, 1]},
					{"op": "OR", "args": [
						{"op": ">", "args": [{"property": "b"}, 2]},
						{"op": "AND", "args": [
							{"op": "<", "args": [{"property": "c"}, 3]},
							{"op": "=", "args": [{"property": "d"}, 4]}
						]}
					]}
				]
			}`,
			verify: func(t *testing.T, expr Expression) {
				// Only verifying no error occurred.
			},
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			expr, err := ParseJSON(compactJSON(t, tc.input))
			require.NoError(t, err)
			tc.verify(t, expr)
		})
	}
}

func TestJSONParser_Errors(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		errSubstr string
	}{
		{"Invalid JSON format", `this is not json`, "invalid character"},
		{"Comparison missing arguments", `{"op": "=","args": [{"property": "temp"}]}`, "comparison requires exactly 2 arguments"},
		{"Logical operator with too many arguments", `{"op": "AND", "args": [{"property": "temp"}, {"property": "humidity"}, {"property": "pressure"}]}`, "AND requires 2 arguments"},
		{"Not operator with too many arguments", `{"op": "NOT", "args": [{"property": "x"}, {"property": "y"}]}`, "NOT requires 1 argument"},
		{"Not operator with zero arguments", `{"op": "NOT", "args": []}`, "NOT requires 1 argument"},
		{"Bad JSON argument", `{"op": "=","args": [{"property": "x"}, bad]}`, "invalid character"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseJSON([]byte(tc.input))
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errSubstr)
		})
	}
}

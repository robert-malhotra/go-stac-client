// pkg/filter/filter_test.go

package filter

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/twpayne/go-geom"
)

func TestBasicOperators(t *testing.T) {
	tests := []struct {
		name string
		json string
		want Expression
	}{
		{
			name: "Equal",
			json: `{"op": "=", "args": [{"property": "type"}, "satellite"]}`,
			want: Comparison{Op: OpEqual, Property: "type", Value: "satellite"},
		},
		{
			name: "NotEqual",
			json: `{"op": "<>", "args": [{"property": "active"}, false]}`,
			want: Comparison{Op: OpNotEqual, Property: "active", Value: false},
		},
		{
			name: "LessThan",
			json: `{"op": "<", "args": [{"property": "cloud_cover"}, 20]}`,
			want: Comparison{Op: OpLessThan, Property: "cloud_cover", Value: float64(20)},
		},
		{
			name: "IsNull",
			json: `{"op": "isNull", "args": [{"property": "end_date"}]}`,
			want: IsNull{Property: "end_date"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExpression([]byte(tt.json))
			if err != nil {
				t.Fatalf("ParseExpression() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseExpression() = %v, want %v", got, tt.want)
			}

			// Test serialization
			serialized, err := SerializeExpression(got)
			if err != nil {
				t.Fatalf("SerializeExpression() error = %v", err)
			}

			var gotMap, wantMap map[string]interface{}
			if err := json.Unmarshal(serialized, &gotMap); err != nil {
				t.Fatalf("Failed to unmarshal serialized JSON: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.json), &wantMap); err != nil {
				t.Fatalf("Failed to unmarshal test JSON: %v", err)
			}

			if !reflect.DeepEqual(gotMap, wantMap) {
				t.Errorf("SerializeExpression() = %v, want %v", gotMap, wantMap)
			}
		})
	}
}

func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		name string
		json string
		want Expression
	}{
		{
			name: "And",
			json: `{
                "op": "and",
                "args": [
                    {"op": "=", "args": [{"property": "type"}, "satellite"]},
                    {"op": "<", "args": [{"property": "cloud_cover"}, 20]}
                ]
            }`,
			want: Logical{
				Op: OpAnd,
				Children: []Expression{
					Comparison{Op: OpEqual, Property: "type", Value: "satellite"},
					Comparison{Op: OpLessThan, Property: "cloud_cover", Value: float64(20)},
				},
			},
		},
		{
			name: "Or",
			json: `{
                "op": "or",
                "args": [
                    {"op": "=", "args": [{"property": "status"}, "active"]},
                    {"op": "=", "args": [{"property": "status"}, "pending"]}
                ]
            }`,
			want: Logical{
				Op: OpOr,
				Children: []Expression{
					Comparison{Op: OpEqual, Property: "status", Value: "active"},
					Comparison{Op: OpEqual, Property: "status", Value: "pending"},
				},
			},
		},
		{
			name: "Not",
			json: `{
                "op": "not",
                "args": [
                    {"op": "=", "args": [{"property": "active"}, false]}
                ]
            }`,
			want: Logical{
				Op: OpNot,
				Children: []Expression{
					Comparison{Op: OpEqual, Property: "active", Value: false},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExpression([]byte(tt.json))
			if err != nil {
				t.Fatalf("ParseExpression() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTIntersects(t *testing.T) {
	startTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)

	json := `{
        "op": "t_intersects",
        "args": [
            {"property": "datetime"},
            {"interval": ["2023-01-01T00:00:00Z", "2023-12-31T23:59:59Z"]}
        ]
    }`

	want := TIntersects{
		Property: "datetime",
		Interval: TimeInterval{
			Start: startTime,
			End:   endTime,
		},
	}

	got, err := ParseExpression([]byte(json))
	if err != nil {
		t.Fatalf("ParseExpression() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseExpression() = %v, want %v", got, want)
	}
}

func TestSIntersects(t *testing.T) {
	polygon := geom.NewPolygon(geom.XY).MustSetCoords([][]geom.Coord{
		{
			{-77.0824, 38.7886},
			{-77.0189, 38.7886},
			{-77.0189, 38.8351},
			{-77.0824, 38.8351},
			{-77.0824, 38.7886},
		},
	})

	json := `{
        "op": "s_intersects",
        "args": [
            {"property": "geometry"},
            {
                "type": "Polygon",
                "coordinates": [[
                    [-77.0824, 38.7886],
                    [-77.0189, 38.7886],
                    [-77.0189, 38.8351],
                    [-77.0824, 38.8351],
                    [-77.0824, 38.7886]
                ]]
            }
        ]
    }`

	expr, err := ParseExpression([]byte(json))
	if err != nil {
		t.Fatalf("ParseExpression() error = %v", err)
	}

	sIntersects, ok := expr.(SIntersects)
	if !ok {
		t.Fatalf("Expected SIntersects expression")
	}

	if sIntersects.Property != "geometry" {
		t.Errorf("Expected property 'geometry', got '%s'", sIntersects.Property)
	}

	gotPolygon, ok := sIntersects.Geometry.(*geom.Polygon)
	if !ok {
		t.Fatalf("Expected Polygon geometry")
	}

	if !reflect.DeepEqual(gotPolygon.Coords(), polygon.Coords()) {
		t.Errorf("Polygon coordinates mismatch")
	}
}

func TestFunction(t *testing.T) {
	json := `{
        "op": "casei",
        "args": [
            {"property": "provider"},
            "coolsat"
        ]
    }`

	want := Function{
		Name: "casei",
		Args: []interface{}{
			map[string]interface{}{"property": "provider"},
			"coolsat",
		},
	}

	got, err := ParseExpression([]byte(json))
	if err != nil {
		t.Fatalf("ParseExpression() error = %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseExpression() = %v, want %v", got, want)
	}
}

func TestErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr string
	}{
		{
			name:    "Missing operator",
			json:    `{"args": [{"property": "type"}, "satellite"]}`,
			wantErr: "unsupported operator: ",
		},
		{
			name:    "Invalid operator",
			json:    `{"op": "invalid", "args": [{"property": "type"}, "satellite"]}`,
			wantErr: "unsupported operator: invalid",
		},
		{
			name:    "Missing property",
			json:    `{"op": "=", "args": [{"wrongkey": "type"}, "satellite"]}`,
			wantErr: "failed to unmarshal property: missing 'property' field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseExpression([]byte(tt.json))
			if err == nil {
				t.Fatal("Expected error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("Expected error '%s', got '%s'", tt.wantErr, err.Error())
			}
		})
	}
}

func TestLikeOperator(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Expression
		wantErr bool
	}{
		{
			name: "simple pattern",
			json: `{"op": "like", "args": [{"property": "name"}, "test%"]}`,
			want: Like{Property: "name", Pattern: "test%"},
		},
		{
			name:    "missing pattern",
			json:    `{"op": "like", "args": [{"property": "name"}]}`,
			wantErr: true,
		},
		{
			name: "complex pattern",
			json: `{"op": "like", "args": [{"property": "name"}, "%test_123%"]}`,
			want: Like{Property: "name", Pattern: "%test_123%"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExpression([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBetweenOperator(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Expression
		wantErr bool
	}{
		{
			name: "numeric between",
			json: `{"op": "between", "args": [{"property": "age"}, 18, 65]}`,
			want: Between{Property: "age", Lower: float64(18), Upper: float64(65)},
		},
		{
			name: "string between",
			json: `{"op": "between", "args": [{"property": "name"}, "A", "Z"]}`,
			want: Between{Property: "name", Lower: "A", Upper: "Z"},
		},
		{
			name:    "missing upper bound",
			json:    `{"op": "between", "args": [{"property": "age"}, 18]}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExpression([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInOperator(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Expression
		wantErr bool
	}{
		{
			name: "string values",
			json: `{"op": "in", "args": [{"property": "status"}, ["active", "pending", "closed"]]}`,
			want: In{Property: "status", Values: []interface{}{"active", "pending", "closed"}},
		},
		{
			name: "numeric values",
			json: `{"op": "in", "args": [{"property": "code"}, [100, 200, 300]]}`,
			want: In{Property: "code", Values: []interface{}{float64(100), float64(200), float64(300)}},
		},
		{
			name: "empty values",
			json: `{"op": "in", "args": [{"property": "status"}, []]}`,
			want: In{Property: "status", Values: []interface{}{}},
		},
		{
			name:    "missing values array",
			json:    `{"op": "in", "args": [{"property": "status"}]}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExpression([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComplexQueries(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{
			name: "nested and/or",
			json: `{
                "op": "and",
                "args": [
                    {
                        "op": "or",
                        "args": [
                            {"op": "=", "args": [{"property": "type"}, "satellite"]},
                            {"op": "=", "args": [{"property": "type"}, "aerial"]}
                        ]
                    },
                    {
                        "op": "<",
                        "args": [{"property": "cloud_cover"}, 20]
                    },
                    {
                        "op": "in",
                        "args": [
                            {"property": "status"},
                            ["active", "pending"]
                        ]
                    }
                ]
            }`,
		},
		{
			name: "temporal and spatial",
			json: `{
                "op": "and",
                "args": [
                    {
                        "op": "t_intersects",
                        "args": [
                            {"property": "datetime"},
                            {"interval": ["2023-01-01T00:00:00Z", "2023-12-31T23:59:59Z"]}
                        ]
                    },
                    {
                        "op": "s_intersects",
                        "args": [
                            {"property": "geometry"},
                            {
                                "type": "Polygon",
                                "coordinates": [[
                                    [-77.0824, 38.7886],
                                    [-77.0189, 38.7886],
                                    [-77.0189, 38.8351],
                                    [-77.0824, 38.8351],
                                    [-77.0824, 38.7886]
                                ]]
                            }
                        ]
                    }
                ]
            }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseExpression([]byte(tt.json))
			if err != nil {
				t.Fatalf("ParseExpression() error = %v", err)
			}

			// Test serialization round-trip
			serialized, err := SerializeExpression(expr)
			if err != nil {
				t.Fatalf("SerializeExpression() error = %v", err)
			}

			parsed, err := ParseExpression(serialized)
			if err != nil {
				t.Fatalf("ParseExpression() error after serialization = %v", err)
			}

			if !reflect.DeepEqual(expr, parsed) {
				t.Errorf("Expression changed after serialization/parsing cycle")
			}
		})
	}
}

func TestValidateTimeInterval(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "valid interval",
			json: `{
                "op": "t_intersects",
                "args": [
                    {"property": "datetime"},
                    {"interval": ["2023-01-01T00:00:00Z", "2023-12-31T23:59:59Z"]}
                ]
            }`,
			wantErr: false,
		},
		{
			name: "invalid start time",
			json: `{
                "op": "t_intersects",
                "args": [
                    {"property": "datetime"},
                    {"interval": ["invalid", "2023-12-31T23:59:59Z"]}
                ]
            }`,
			wantErr: true,
		},
		{
			name: "invalid end time",
			json: `{
                "op": "t_intersects",
                "args": [
                    {"property": "datetime"},
                    {"interval": ["2023-01-01T00:00:00Z", "invalid"]}
                ]
            }`,
			wantErr: true,
		},
		{
			name: "missing interval",
			json: `{
                "op": "t_intersects",
                "args": [
                    {"property": "datetime"},
                    {}
                ]
            }`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseExpression([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseExpression() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBetweenComprehensive(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    Expression
		wantErr bool
	}{
		{
			name: "numeric between",
			json: `{
                "op": "between",
                "args": [
                    {"property": "size"},
                    10,
                    100
                ]
            }`,
			want: Between{
				Property: "size",
				Lower:    float64(10),
				Upper:    float64(100),
			},
		},
		{
			name: "string between",
			json: `{
                "op": "between",
                "args": [
                    {"property": "name"},
                    "A",
                    "Z"
                ]
            }`,
			want: Between{
				Property: "name",
				Lower:    "A",
				Upper:    "Z",
			},
		},
		{
			name: "date between",
			json: `{
                "op": "between",
                "args": [
                    {"property": "date"},
                    "2023-01-01",
                    "2023-12-31"
                ]
            }`,
			want: Between{
				Property: "date",
				Lower:    "2023-01-01",
				Upper:    "2023-12-31",
			},
		},
		{
			name: "mixed types",
			json: `{
                "op": "between",
                "args": [
                    {"property": "value"},
                    0,
                    "100"
                ]
            }`,
			want: Between{
				Property: "value",
				Lower:    float64(0),
				Upper:    "100",
			},
		},
		{
			name:    "missing upper bound",
			json:    `{"op": "between", "args": [{"property": "value"}, 0]}`,
			wantErr: true,
		},
		{
			name:    "missing lower bound",
			json:    `{"op": "between", "args": [{"property": "value"}]}`,
			wantErr: true,
		},
		{
			name:    "too many arguments",
			json:    `{"op": "between", "args": [{"property": "value"}, 0, 100, 200]}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExpression([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseExpression() = %v, want %v", got, tt.want)
			}

			if !tt.wantErr {
				// Test serialization round-trip
				serialized, err := SerializeExpression(got)
				if err != nil {
					t.Fatalf("SerializeExpression() error = %v", err)
				}

				parsed, err := ParseExpression(serialized)
				if err != nil {
					t.Fatalf("ParseExpression() error after serialization = %v", err)
				}

				if !reflect.DeepEqual(got, parsed) {
					t.Errorf("Expression changed after serialization/parsing cycle")
				}
			}
		})
	}
}

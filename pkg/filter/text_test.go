package filter

import (
	"reflect"
	"testing"
	"time"

	"github.com/twpayne/go-geom"
)

func TestTextParser(t *testing.T) {
	parser, err := NewTextParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		want    Expression
		wantErr bool
	}{
		{
			name:  "simple equals",
			input: `collection = "landsat"`,
			want: Comparison{
				Op:       OpEqual,
				Property: "collection",
				Value:    "landsat",
			},
		},
		{
			name:  "numeric less than",
			input: `cloudCover < 10.5`,
			want: Comparison{
				Op:       OpLessThan,
				Property: "cloudCover",
				Value:    10.5,
			},
		},
		{
			name:  "between with numbers",
			input: `resolution BETWEEN 10 AND 30`,
			want: Between{
				Property: "resolution",
				Lower:    float64(10),
				Upper:    float64(30),
			},
		},
		{
			name:  "like pattern",
			input: `name LIKE "%landsat%"`,
			want: Like{
				Property: "name",
				Pattern:  "%landsat%",
			},
		},
		{
			name:  "in values",
			input: `status IN ("active", "pending")`,
			want: In{
				Property: "status",
				Values:   []interface{}{"active", "pending"},
			},
		},
		{
			name:  "is null",
			input: `deletedAt IS NULL`,
			want: IsNull{
				Property: "deletedAt",
			},
		},
		{
			name:  "simple and",
			input: `AND(cloudCover < 10, quality = "good")`,
			want: Logical{
				Op: OpAnd,
				Children: []Expression{
					Comparison{Op: OpLessThan, Property: "cloudCover", Value: float64(10)},
					Comparison{Op: OpEqual, Property: "quality", Value: "good"},
				},
			},
		},
		{
			name:  "simple or",
			input: `OR(quality = "good", quality = "excellent")`,
			want: Logical{
				Op: OpOr,
				Children: []Expression{
					Comparison{Op: OpEqual, Property: "quality", Value: "good"},
					Comparison{Op: OpEqual, Property: "quality", Value: "excellent"},
				},
			},
		},
		{
			name:  "not equal",
			input: `status <> "cancelled"`,
			want: Comparison{
				Op:       OpNotEqual,
				Property: "status",
				Value:    "cancelled",
			},
		},
		{
			name:  "spatial intersects",
			input: `footprint S_INTERSECTS POINT(10.5 20.5)`,
			want: SIntersects{
				Property: "footprint",
				Geometry: geom.NewPointFlat(geom.XY, []float64{10.5, 20.5}),
			},
		},
		{
			name:  "temporal intersects",
			input: `datetime T_INTERSECTS ["2024-01-01T00:00:00Z"/"2024-12-31T23:59:59Z"]`,
			want: TIntersects{
				Property: "datetime",
				Interval: TimeInterval{
					Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				},
			},
		},
		{
			name:  "complex nested expression",
			input: `AND(collection = "landsat", cloudCover < 20, OR(quality = "good", quality = "excellent"))`,
			want: Logical{
				Op: OpAnd,
				Children: []Expression{
					Comparison{Op: OpEqual, Property: "collection", Value: "landsat"},
					Comparison{Op: OpLessThan, Property: "cloudCover", Value: float64(20)},
					Logical{
						Op: OpOr,
						Children: []Expression{
							Comparison{Op: OpEqual, Property: "quality", Value: "good"},
							Comparison{Op: OpEqual, Property: "quality", Value: "excellent"},
						},
					},
				},
			},
		},
		// Error cases
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid operator",
			input:   `collection INVALID "landsat"`,
			wantErr: true,
		},
		{
			name:    "incomplete comparison",
			input:   `cloudCover <`,
			wantErr: true,
		},
		{
			name:    "invalid between syntax",
			input:   `resolution BETWEEN 10`,
			wantErr: true,
		},
		{
			name:    "invalid point syntax",
			input:   `footprint S_INTERSECTS POINT(10.5)`,
			wantErr: true,
		},
		{
			name:    "invalid temporal format",
			input:   `datetime T_INTERSECTS ["invalid"/"date"]`,
			wantErr: true,
		},
		{
			name:    "missing parentheses",
			input:   `AND cloudCover < 10, quality = "good"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test helper functions for comparing specific expression types
func TestTextParserValueTypes(t *testing.T) {
	parser, err := NewTextParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tests := []struct {
		name  string
		input string
		check func(t *testing.T, expr Expression)
	}{
		{
			name:  "string value",
			input: `name = "test"`,
			check: func(t *testing.T, expr Expression) {
				comp, ok := expr.(Comparison)
				if !ok {
					t.Errorf("Expected Comparison, got %T", expr)
					return
				}
				str, ok := comp.Value.(string)
				if !ok {
					t.Errorf("Expected string value, got %T", comp.Value)
					return
				}
				if str != "test" {
					t.Errorf("Expected value 'test', got '%s'", str)
				}
			},
		},
		{
			name:  "numeric value",
			input: `score = 42.5`,
			check: func(t *testing.T, expr Expression) {
				comp, ok := expr.(Comparison)
				if !ok {
					t.Errorf("Expected Comparison, got %T", expr)
					return
				}
				num, ok := comp.Value.(float64)
				if !ok {
					t.Errorf("Expected float64 value, got %T", comp.Value)
					return
				}
				if num != 42.5 {
					t.Errorf("Expected value 42.5, got %v", num)
				}
			},
		},
		{
			name:  "boolean value",
			input: `active = true`,
			check: func(t *testing.T, expr Expression) {
				comp, ok := expr.(Comparison)
				if !ok {
					t.Errorf("Expected Comparison, got %T", expr)
					return
				}
				b, ok := comp.Value.(bool)
				if !ok {
					t.Errorf("Expected bool value, got %T", comp.Value)
					return
				}
				if !b {
					t.Errorf("Expected value true, got false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			tt.check(t, got)
		})
	}
}

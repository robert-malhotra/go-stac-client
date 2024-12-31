package filter

import (
	"reflect"
	"testing"
	"time"
)

func TestBuilder(t *testing.T) {
	tests := []struct {
		name     string
		build    func(*Builder) Expression
		expected Expression
	}{
		{
			name: "single equal",
			build: func(b *Builder) Expression {
				return b.Equal("type", "satellite").Build()
			},
			expected: Comparison{Op: OpEqual, Property: "type", Value: "satellite"},
		},
		{
			name: "and two equals",
			build: func(b *Builder) Expression {
				return b.Equal("type", "satellite").
					Equal("provider", "test").
					Build()
			},
			expected: Logical{
				Op: OpAnd,
				Children: []Expression{
					Comparison{Op: OpEqual, Property: "type", Value: "satellite"},
					Comparison{Op: OpEqual, Property: "provider", Value: "test"},
				},
			},
		},
		{
			name: "or two equals",
			build: func(b *Builder) Expression {
				return b.Or(
					Comparison{Op: OpEqual, Property: "type", Value: "satellite"},
					Comparison{Op: OpEqual, Property: "type", Value: "aerial"},
				).Build()
			},
			expected: Logical{
				Op: OpOr,
				Children: []Expression{
					Comparison{Op: OpEqual, Property: "type", Value: "satellite"},
					Comparison{Op: OpEqual, Property: "type", Value: "aerial"},
				},
			},
		},
		{
			name: "not equal",
			build: func(b *Builder) Expression {
				return b.Not(
					Comparison{Op: OpEqual, Property: "type", Value: "satellite"},
				).Build()
			},
			expected: Logical{
				Op: OpNot,
				Children: []Expression{
					Comparison{Op: OpEqual, Property: "type", Value: "satellite"},
				},
			},
		},
		{
			name: "comparison operators",
			build: func(b *Builder) Expression {
				return b.LessThan("cloud_cover", 20.0).
					GreaterThan("quality", 80.0).
					Build()
			},
			expected: Logical{
				Op: OpAnd,
				Children: []Expression{
					Comparison{Op: OpLessThan, Property: "cloud_cover", Value: 20.0},
					Comparison{Op: OpGreaterThan, Property: "quality", Value: 80.0},
				},
			},
		},
		{
			name: "between",
			build: func(b *Builder) Expression {
				return b.Between("value", 0.0, 100.0).Build()
			},
			expected: Between{Property: "value", Lower: 0.0, Upper: 100.0},
		},
		{
			name: "like",
			build: func(b *Builder) Expression {
				return b.Like("name", "test%").Build()
			},
			expected: Like{Property: "name", Pattern: "test%"},
		},
		{
			name: "in",
			build: func(b *Builder) Expression {
				return b.In("status", []interface{}{"active", "pending"}).Build()
			},
			expected: In{Property: "status", Values: []interface{}{"active", "pending"}},
		},
		{
			name: "complex nested",
			build: func(b *Builder) Expression {
				return b.And(
					Logical{
						Op: OpOr,
						Children: []Expression{
							Comparison{Op: OpEqual, Property: "type", Value: "satellite"},
							Comparison{Op: OpEqual, Property: "type", Value: "aerial"},
						},
					},
					Comparison{Op: OpLessThan, Property: "cloud_cover", Value: 20.0},
					In{Property: "status", Values: []interface{}{"active", "pending"}},
				).Build()
			},
			expected: Logical{
				Op: OpAnd,
				Children: []Expression{
					Logical{
						Op: OpOr,
						Children: []Expression{
							Comparison{Op: OpEqual, Property: "type", Value: "satellite"},
							Comparison{Op: OpEqual, Property: "type", Value: "aerial"},
						},
					},
					Comparison{Op: OpLessThan, Property: "cloud_cover", Value: 20.0},
					In{Property: "status", Values: []interface{}{"active", "pending"}},
				},
			},
		},
		{
			name: "temporal",
			build: func(b *Builder) Expression {
				return b.TIntersects("datetime", TimeInterval{
					Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
				}).Build()
			},
			expected: TIntersects{
				Property: "datetime",
				Interval: TimeInterval{
					Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					End:   time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilder()
			result := tt.build(builder)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Builder output mismatch.\nGot: %v\nWant: %v", result, tt.expected)
			}

			// Test serialization roundtrip
			serialized, err := SerializeExpression(result)
			if err != nil {
				t.Fatalf("SerializeExpression() error = %v", err)
			}

			parsed, err := ParseExpression(serialized)
			if err != nil {
				t.Fatalf("ParseExpression() error = %v", err)
			}

			if !reflect.DeepEqual(result, parsed) {
				t.Errorf("Expression changed after serialization/parsing cycle")
			}
		})
	}
}

func TestBuilderEmptyOperations(t *testing.T) {
	tests := []struct {
		name     string
		build    func(*Builder) Expression
		expected Expression
	}{
		{
			name: "empty and",
			build: func(b *Builder) Expression {
				return b.And().Build()
			},
			expected: nil,
		},
		{
			name: "empty or",
			build: func(b *Builder) Expression {
				return b.Or().Build()
			},
			expected: nil,
		},
		// {
		// 	name: "and with nil",
		// 	build: func(b *Builder) Expression {
		// 		return b.And(nil).Build()
		// 	},
		// 	expected: nil,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilder()
			result := tt.build(builder)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Builder output mismatch.\nGot: %v\nWant: %v", result, tt.expected)
			}
		})
	}
}

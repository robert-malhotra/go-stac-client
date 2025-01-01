// pkg/filter/terminal_test.go

package filter

import (
	"reflect"
	"testing"
	"time"

	"github.com/twpayne/go-geom"
)

func TestTerminalOperationMethods(t *testing.T) {
	now := time.Now()
	interval := TimeInterval{Start: now, End: now.Add(time.Hour)}
	poly := geom.NewPolygonFlat(geom.XY, []float64{0, 0, 1, 0, 1, 1, 0, 1, 0, 0}, []int{10})

	tests := []struct {
		name      string
		op        TerminalOperation
		wantProp  string
		wantValue interface{}
		wantOp    Operator
	}{
		{
			name: "comparison equals",
			op: Comparison{
				Op:       OpEqual,
				Property: "name",
				Value:    "test",
			},
			wantProp:  "name",
			wantValue: "test",
			wantOp:    OpEqual,
		},
		{
			name: "between",
			op: Between{
				Property: "age",
				Lower:    18,
				Upper:    65,
			},
			wantProp: "age",
			wantValue: map[string]interface{}{
				"lower": 18,
				"upper": 65,
			},
			wantOp: OpBetween,
		},
		{
			name: "like",
			op: Like{
				Property: "email",
				Pattern:  "%@example.com",
			},
			wantProp:  "email",
			wantValue: "%@example.com",
			wantOp:    OpLike,
		},
		{
			name: "in",
			op: In{
				Property: "status",
				Values:   []interface{}{"active", "pending"},
			},
			wantProp:  "status",
			wantValue: []interface{}{"active", "pending"},
			wantOp:    OpIn,
		},
		{
			name: "is null",
			op: IsNull{
				Property: "deletedAt",
			},
			wantProp:  "deletedAt",
			wantValue: nil,
			wantOp:    OpIsNull,
		},
		{
			name: "s_intersects",
			op: SIntersects{
				Property: "location",
				Geometry: poly,
			},
			wantProp:  "location",
			wantValue: poly,
			wantOp:    OpSIntersects,
		},
		{
			name: "t_intersects",
			op: TIntersects{
				Property: "timeRange",
				Interval: interval,
			},
			wantProp:  "timeRange",
			wantValue: interval,
			wantOp:    OpTIntersects,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.op.GetProperty(); got != tt.wantProp {
				t.Errorf("GetProperty() = %v, want %v", got, tt.wantProp)
			}

			gotValue := tt.op.GetValue()
			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("GetValue() = %v, want %v", gotValue, tt.wantValue)
			}

			if got := tt.op.Type(); got != tt.wantOp {
				t.Errorf("GetOp() = %v, want %v", got, tt.wantOp)
			}
		})
	}
}

func TestGroupByOperator(t *testing.T) {
	expr := NewBuilder().
		Equal("name", "test").
		Equal("email", "test@example.com"). // Same operator
		Between("age", 18, 65).
		Like("bio", "%engineer%").
		Like("title", "%senior%"). // Same operator
		Build()

	ops, err := ExtractTerminalOps(expr)
	if err != nil {
		t.Fatalf("ExtractTerminalOps() error = %v", err)
	}

	grouped := GroupByOperator(ops)

	expectedCounts := map[Operator]int{
		OpEqual:   2,
		OpBetween: 1,
		OpLike:    2,
	}

	for op, expectedCount := range expectedCounts {
		if got := len(grouped[op]); got != expectedCount {
			t.Errorf("GroupByOperator() operator %q got %d ops, want %d", op, got, expectedCount)
		}
	}
}

// Rest of the existing tests...
func TestExtractTerminalOps(t *testing.T) {
	poly := geom.NewPolygonFlat(geom.XY, []float64{0, 0, 1, 0, 1, 1, 0, 1, 0, 0}, []int{10})

	tests := []struct {
		name          string
		expr          Expression
		wantOpsCount  int
		wantErr       bool
		wantErrString string
	}{
		{
			name: "simple comparison",
			expr: NewBuilder().
				Equal("name", "test").
				Build(),
			wantOpsCount: 1,
		},
		{
			name: "multiple AND conditions",
			expr: NewBuilder().
				Equal("name", "test").
				GreaterThan("age", 18).
				Like("email", "%@example.com").
				Build(),
			wantOpsCount: 3,
		},
		{
			name: "nested AND conditions",
			expr: NewBuilder().
				Equal("name", "test").
				And(
					NewBuilder().
						GreaterThan("age", 18).
						Like("email", "%@example.com").
						Build(),
				).
				Build(),
			wantOpsCount: 3,
		},
		{
			name: "OR not supported",
			expr: NewBuilder().
				Or(
					NewBuilder().Equal("name", "test1").Build(),
					NewBuilder().Equal("name", "test2").Build(),
				).
				Build(),
			wantErr:       true,
			wantErrString: "only AND operations are supported",
		},
		{
			name: "mix of operations",
			expr: NewBuilder().
				Equal("type", "user").
				Between("age", 18, 65).
				In("status", []interface{}{"active", "pending"}).
				IsNull("deletedAt").
				Build(),
			wantOpsCount: 4,
		},
		{
			name: "spatial operations",
			expr: NewBuilder().
				Equal("collection", "landsat").
				SIntersects("geometry", poly).
				Build(),
			wantOpsCount: 2,
		},
		{
			name: "complex mix with spatial",
			expr: NewBuilder().
				Equal("collection", "landsat").
				Between("cloudCover", 0, 20).
				SIntersects("geometry", poly).
				And(
					NewBuilder().
						GreaterThan("resolution", 10).
						Build(),
				).
				Build(),
			wantOpsCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops, err := ExtractTerminalOps(tt.expr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExtractTerminalOps() error = nil, want error containing %q", tt.wantErrString)
					return
				}
				if tt.wantErrString != "" && !contains(err.Error(), tt.wantErrString) {
					t.Errorf("ExtractTerminalOps() error = %v, want error containing %q", err, tt.wantErrString)
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractTerminalOps() unexpected error = %v", err)
				return
			}

			if len(ops) != tt.wantOpsCount {
				t.Errorf("ExtractTerminalOps() got %d ops, want %d", len(ops), tt.wantOpsCount)
			}
		})
	}
}

func TestGroupByProperty(t *testing.T) {
	expr := NewBuilder().
		Equal("name", "test").
		Equal("name", "test2"). // Duplicate property
		GreaterThan("age", 18).
		LessThan("age", 65). // Duplicate property
		Like("email", "%@example.com").
		Build()

	ops, err := ExtractTerminalOps(expr)
	if err != nil {
		t.Fatalf("ExtractTerminalOps() error = %v", err)
	}

	grouped := GroupByProperty(ops)

	expectedCounts := map[string]int{
		"name":  2,
		"age":   2,
		"email": 1,
	}

	for prop, expectedCount := range expectedCounts {
		if got := len(grouped[prop]); got != expectedCount {
			t.Errorf("GroupByProperty() property %q got %d ops, want %d", prop, got, expectedCount)
		}
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}

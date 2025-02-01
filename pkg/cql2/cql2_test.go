// File: cql2_test.go
package cql2

import (
	"fmt"
	"strings"
	"testing"
)

// TestVisitor collects operator calls during parsing.
// It implements the Visitor interface.
type TestVisitor struct {
	t             *testing.T
	actualCalls   []string
	expectedCalls []string
}

func (v *TestVisitor) OnEquals(prop string, value interface{}) error {
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("Equals:%s:%v", prop, value))
	return nil
}
func (v *TestVisitor) OnNotEquals(prop string, value interface{}) error {
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("NotEquals:%s:%v", prop, value))
	return nil
}
func (v *TestVisitor) OnLessThan(prop string, value interface{}) error {
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("LessThan:%s:%v", prop, value))
	return nil
}
func (v *TestVisitor) OnGreaterThan(prop string, value interface{}) error {
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("GreaterThan:%s:%v", prop, value))
	return nil
}
func (v *TestVisitor) OnLessThanOrEquals(prop string, value interface{}) error {
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("LessThanOrEquals:%s:%v", prop, value))
	return nil
}
func (v *TestVisitor) OnGreaterThanOrEquals(prop string, value interface{}) error {
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("GreaterThanOrEquals:%s:%v", prop, value))
	return nil
}
func (v *TestVisitor) OnSIntersects(prop string, geom interface{}) error {
	m := geom.(map[string]interface{})
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("SIntersects:%s:%s", prop, strings.ToLower(m["type"].(string))))
	return nil
}
func (v *TestVisitor) OnSContains(prop string, geom interface{}) error {
	m := geom.(map[string]interface{})
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("SContains:%s:%s", prop, strings.ToLower(m["type"].(string))))
	return nil
}
func (v *TestVisitor) OnSWithin(prop string, geom interface{}) error {
	m := geom.(map[string]interface{})
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("SWithin:%s:%s", prop, strings.ToLower(m["type"].(string))))
	return nil
}
func (v *TestVisitor) OnAnd(args []interface{}) error {
	v.actualCalls = append(v.actualCalls, "And")
	return nil
}
func (v *TestVisitor) OnOr(args []interface{}) error {
	v.actualCalls = append(v.actualCalls, "Or")
	return nil
}
func (v *TestVisitor) OnNot(arg interface{}) error {
	v.actualCalls = append(v.actualCalls, "Not")
	return nil
}
func (v *TestVisitor) Verify() {
	if len(v.actualCalls) != len(v.expectedCalls) {
		v.t.Fatalf("expected %d calls, got %d:\nexpected: %v\nactual:   %v",
			len(v.expectedCalls), len(v.actualCalls), v.expectedCalls, v.actualCalls)
	}
	for i, exp := range v.expectedCalls {
		if exp != v.actualCalls[i] {
			v.t.Errorf("call %d: expected %q, got %q", i, exp, v.actualCalls[i])
		}
	}
}

// TestAllOperators exercises all comparison, spatial, and logical operators
// using a top-level "and" query.
func TestAllOperators(t *testing.T) {
	query := `{
		"op": "and",
		"args": [
			{"op": "=","args": [{"property": "a"}, "val"]},
			{"op": "!=","args": [{"property": "b"}, "val"]},
			{"op": "<","args": [{"property": "c"}, 1]},
			{"op": ">","args": [{"property": "d"}, 2]},
			{"op": "<=","args": [{"property": "e"}, 3]},
			{"op": ">=","args": [{"property": "f"}, 4]},
			{"op": "s_intersects","args": [{"property": "g"}, {"type": "Polygon", "coordinates": [[[0,0],[1,0],[1,1],[0,1],[0,0]]]}]},
			{"op": "s_contains","args": [{"property": "h"}, {"type": "Polygon", "coordinates": [[[0,0],[1,0],[1,1],[0,1],[0,0]]]}]},
			{"op": "s_within","args": [{"property": "i"}, {"type": "Polygon", "coordinates": [[[0,0],[1,0],[1,1],[0,1],[0,0]]]}]}
		]
	}`
	expected := []string{
		"And",
		"Equals:a:val",
		"NotEquals:b:val",
		"LessThan:c:1",
		"GreaterThan:d:2",
		"LessThanOrEquals:e:3",
		"GreaterThanOrEquals:f:4",
		"SIntersects:g:polygon",
		"SContains:h:polygon",
		"SWithin:i:polygon",
	}
	visitor := &TestVisitor{t: t, expectedCalls: expected}
	parser := NewParser(NewAdapter(visitor))
	if err := parser.Parse(query); err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	visitor.Verify()
}

// TestLogicalOperators verifies handling of "or" and "not" operators.
func TestLogicalOperators(t *testing.T) {
	query := `{
		"op": "or",
		"args": [
			{"op": "=","args": [{"property": "x"}, 10]},
			{"op": "not", "args": [
				{"op": "=","args": [{"property": "y"}, 20]}
			]}
		]
	}`
	expected := []string{"Or", "Equals:x:10", "Not"}
	visitor := &TestVisitor{t: t, expectedCalls: expected}
	parser := NewParser(NewAdapter(visitor))
	if err := parser.Parse(query); err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	visitor.Verify()
}

// TestTimestampTransformation checks that a timestamp argument is processed correctly.
func TestTimestampTransformation(t *testing.T) {
	query := `{
		"op": ">=",
		"args": [
			{"property": "datetime"},
			{"timestamp": "2021-04-08T04:39:23Z"}
		]
	}`
	expected := []string{"GreaterThanOrEquals:datetime:2021-04-08T04:39:23Z"}
	visitor := &TestVisitor{t: t, expectedCalls: expected}
	parser := NewParser(NewAdapter(visitor))
	if err := parser.Parse(query); err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	visitor.Verify()
}

// TestErrorCases uses table-driven tests to exercise error conditions in the parser.
func TestErrorCases(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		errSubstring string
	}{
		{
			name:         "MissingArgs",
			query:        `{"op": "=","args": [{"property": "x"}]}`,
			errSubstring: "comparison operator requires exactly two arguments",
		},
		{
			name:         "UnknownOperator",
			query:        `{"op": "foobar","args": [{"property": "x"}, "val"]}`,
			errSubstring: "unknown operator",
		},
		{
			name:         "NotOperatorWrongArgCount",
			query:        `{"op": "not","args": [{"property": "x"}, "val"]}`,
			errSubstring: "not operator requires exactly one argument",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visitor := &TestVisitor{t: t}
			parser := NewParser(NewAdapter(visitor))
			err := parser.Parse(tt.query)
			if err == nil {
				t.Errorf("Expected error for query: %s", tt.query)
			} else if !strings.Contains(err.Error(), tt.errSubstring) {
				t.Errorf("Expected error to contain %q, got %q", tt.errSubstring, err.Error())
			}
		})
	}
}

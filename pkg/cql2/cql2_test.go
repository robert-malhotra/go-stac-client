package cql2

import (
	"fmt"
	"strings"
	"testing"
)

type TestVisitor struct {
	t             *testing.T
	actualCalls   []string
	expectedCalls []string
}

func NewTestVisitor(t *testing.T) *TestVisitor {
	return &TestVisitor{
		t: t,
		expectedCalls: []string{
			"And",
			"Equals:collection:landsat8_l1tp",
			"LessThanOrEquals:eo:cloud_cover:10",
			"GreaterThanOrEquals:datetime:2021-04-08T04:39:23Z",
			"SIntersects:geometry:polygon",
		},
	}
}

func (v *TestVisitor) OnEquals(property string, value interface{}) error {
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("Equals:%s:%v", property, value))
	return nil
}

func (v *TestVisitor) OnLessThan(property string, value interface{}) error {
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("LessThan:%s:%v", property, value))
	return nil
}

func (v *TestVisitor) OnGreaterThan(property string, value interface{}) error {
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("GreaterThan:%s:%v", property, value))
	return nil
}

func (v *TestVisitor) OnLessThanOrEquals(property string, value interface{}) error {
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("LessThanOrEquals:%s:%v", property, value))
	return nil
}

func (v *TestVisitor) OnGreaterThanOrEquals(property string, value interface{}) error {
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("GreaterThanOrEquals:%s:%v", property, value))
	return nil
}

func (v *TestVisitor) OnNotEquals(property string, value interface{}) error {
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("NotEquals:%s:%v", property, value))
	return nil
}

func (v *TestVisitor) OnSIntersects(property string, geometry interface{}) error {
	geom := geometry.(map[string]interface{})
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("SIntersects:%s:%s", property, strings.ToLower(geom["type"].(string))))
	return nil
}

func (v *TestVisitor) OnSContains(property string, geometry interface{}) error {
	geom := geometry.(map[string]interface{})
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("SContains:%s:%s", property, strings.ToLower(geom["type"].(string))))
	return nil
}

func (v *TestVisitor) OnSWithin(property string, geometry interface{}) error {
	geom := geometry.(map[string]interface{})
	v.actualCalls = append(v.actualCalls, fmt.Sprintf("SWithin:%s:%s", property, strings.ToLower(geom["type"].(string))))
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
	if len(v.expectedCalls) != len(v.actualCalls) {
		v.t.Errorf("Expected %d calls but got %d", len(v.expectedCalls), len(v.actualCalls))
		v.t.Errorf("Expected: %v", v.expectedCalls)
		v.t.Errorf("Actual: %v", v.actualCalls)
		return
	}

	for i, expected := range v.expectedCalls {
		if expected != v.actualCalls[i] {
			v.t.Errorf("Call %d: expected %s but got %s", i, expected, v.actualCalls[i])
		}
	}
}

func TestComplexQuery(t *testing.T) {
	query := `{
		"op": "and",
		"args": [
			{
				"op": "=",
				"args": [
					{ "property": "collection" },
					"landsat8_l1tp"
				]
			},
			{
				"op": "<=",
				"args": [
					{ "property": "eo:cloud_cover" },
					10
				]
			},
			{
				"op": ">=",
				"args": [
					{ "property": "datetime" },
					{ "timestamp": "2021-04-08T04:39:23Z" }
				]
			},
			{
				"op": "s_intersects",
				"args": [
					{ "property": "geometry" },
					{
						"type": "Polygon",
						"coordinates": [[
							[43.5845, -79.5442],
							[43.6079, -79.4893],
							[43.5677, -79.4632],
							[43.6129, -79.3925],
							[43.6223, -79.3238],
							[43.6576, -79.3163],
							[43.7945, -79.1178],
							[43.8144, -79.1542],
							[43.8555, -79.1714],
							[43.7509, -79.6390],
							[43.5845, -79.5442]
						]]
					}
				]
			}
		]
	}`

	visitor := NewTestVisitor(t)
	parser := NewParser(NewAdapter(visitor))

	err := parser.Parse(query)
	if err != nil {
		t.Fatalf("Failed to parse query: %v", err)
	}

	visitor.Verify()
}

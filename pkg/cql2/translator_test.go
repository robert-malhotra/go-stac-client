package cql2

import (
	"testing"
)

func TestODataTranslation(t *testing.T) {
	comparison := NewComparisonNode("age", OpGt, 30)
	spatial := NewSpatialNode(SpatialIntersects, map[string]interface{}{
		"type":        "Polygon",
		"coordinates": [][][]float64{{{102.0, 0.0}, {103.0, 0.0}, {103.0, 1.0}, {102.0, 1.0}, {102.0, 0.0}}},
	})
	logical := NewLogicalNode(LogicalAnd, comparison, spatial)

	translator := &ODataTranslator{}
	query, err := translator.Translate(logical)
	if err != nil {
		t.Fatalf("Error translating to OData: %v", err)
	}

	expected := `(age gt 30 and INTERSECTS({"coordinates":[[[102,0],[103,0],[103,1],[102,1],[102,0]]],"type":"Polygon"}))`
	if query != expected {
		t.Errorf("Expected: %s, got: %s", expected, query)
	}
}

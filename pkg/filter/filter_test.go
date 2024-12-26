// pkg/filter/filter_test.go

package filter

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
)

// TestParseExpression tests parsing of complex expressions involving logical and temporal operators.
func TestParseExpression(t *testing.T) {
	jsonStr := `{
        "op": "and",
        "args": [
            {
                "op": "=",
                "args": [ { "property": "assetType" }, "image" ]
            },
            {
                "op": "s_intersects",
                "args": [
                    { "property": "geometry" },
                    {
                        "type": "Polygon",
                        "coordinates": [[
                            [-77.0824, 38.7886], [-77.0189, 38.7886],
                            [-77.0189, 38.8351], [-77.0824, 38.8351],
                            [-77.0824, 38.7886]
                        ]]
                    }
                ]
            },
            {
                "op": "t_intersects",
                "args": [
                    { "property": "datetime" },
                    { "interval" : [ "2023-01-01T00:00:00Z", "2023-12-31T23:59:59Z" ] }
                ]
            }
        ]
    }`

	expr, err := ParseExpression([]byte(jsonStr))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	andExpr, ok := expr.(And)
	if !ok {
		t.Fatalf("Expected And expression")
	}

	if len(andExpr.Children) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(andExpr.Children))
	}

	// Further assertions can be added here to verify each child expression
}

// TestSerializeExpression tests serialization of the TIntersects operator.
func TestSerializeExpression(t *testing.T) {
	tIntersects := TIntersects{
		Property: "timestamp",
		Interval: TimeInterval{
			Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
		},
	}

	jsonData, err := SerializeExpression(tIntersects)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	expected := `{
      "op": "t_intersects",
      "args": [
        {
          "property": "timestamp"
        },
        {
          "interval": [
            "2023-01-01T00:00:00Z",
            "2023-12-31T23:59:59Z"
          ]
        }
      ]
    }`

	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(expected), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(jsonData, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}

// TestPropertyPropertyComparison tests parsing and serialization of property-property comparisons.
func TestPropertyPropertyComparison(t *testing.T) {
	jsonStr := `{
        "op": "=",
        "args": [
            { "property": "prop1" },
            { "property": "prop2" }
        ]
    }`

	expr, err := ParseExpression([]byte(jsonStr))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	ppc, ok := expr.(PropertyPropertyComparison)
	if !ok {
		t.Fatalf("Expected PropertyPropertyComparison expression")
	}

	if ppc.Property1 != "prop1" || ppc.Property2 != "prop2" || ppc.Operator != "=" {
		t.Fatalf("PropertyPropertyComparison fields mismatch")
	}

	serialized, err := SerializeExpression(ppc)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	expected := `{
      "op": "=",
      "args": [
        {
          "property": "prop1"
        },
        {
          "property": "prop2"
        }
      ]
    }`

	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(expected), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}

// TestFunctionExpression tests parsing and serialization of function expressions like 'casei'.
func TestFunctionExpression(t *testing.T) {
	jsonStr := `{
        "op": "casei",
        "args": [
            { "property": "provider" },
            "coolsat"
        ]
    }`

	expr, err := ParseExpression([]byte(jsonStr))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	fn, ok := expr.(Function)
	if !ok {
		t.Fatalf("Expected Function expression")
	}

	if fn.Name != "casei" {
		t.Fatalf("Expected function name 'casei', got '%s'", fn.Name)
	}

	if len(fn.Args) != 2 {
		t.Fatalf("Expected 2 arguments for function 'casei'")
	}

	// Further assertions can be added here to verify function arguments
}

// TestSerializeAndParseExpression tests the serialization and parsing cycle for complex expressions.
func TestSerializeAndParseExpression(t *testing.T) {
	// Create a complex expression
	builder := NewBuilder()
	expr := builder.
		PropertyIsEqualTo("assetType", "image").
		SIntersects("geometry", GeoJSONGeometry{
			Type: "Polygon",
			Coordinates: [][][]float64{
				{
					{-77.0824, 38.7886},
					{-77.0189, 38.7886},
					{-77.0189, 38.8351},
					{-77.0824, 38.8351},
					{-77.0824, 38.7886},
				},
			},
		}).
		TIntersects("datetime", TimeInterval{
			Start: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
		}).
		Build()

	// Serialize
	jsonData, err := SerializeExpression(expr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Parse back
	parsedExpr, err := ParseExpression(jsonData)
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	// Serialize again
	jsonData2, err := SerializeExpression(parsedExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare both JSON outputs using reflect.DeepEqual
	var original, parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &original); err != nil {
		t.Fatalf("Failed to unmarshal original JSON: %v", err)
	}
	if err := json.Unmarshal(jsonData2, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal parsed JSON: %v", err)
	}

	if !reflect.DeepEqual(original, parsed) {
		t.Errorf("Serialized JSON does not match after parsing.\nOriginal: %v\nParsed: %v", original, parsed)
	}
}

// TestAndOperator tests parsing and serialization of the And logical operator with multiple child expressions.
func TestAndOperator(t *testing.T) {
	jsonStr := `{
		"op": "and",
		"args": [
			{
				"op": ">",
				"args": [ { "property": "age" }, 30 ]
			},
			{
				"op": "like",
				"args": [ { "property": "name" }, "J%" ]
			}
		]
	}`

	expr, err := ParseExpression([]byte(jsonStr))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	andExpr, ok := expr.(And)
	if !ok {
		t.Fatalf("Expected And expression")
	}

	if len(andExpr.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(andExpr.Children))
	}

	// Serialize back to JSON
	serialized, err := SerializeExpression(andExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare using reflect.DeepEqual
	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}

// TestOrOperator tests parsing and serialization of the Or logical operator with multiple child expressions.
func TestOrOperator(t *testing.T) {
	jsonStr := `{
		"op": "or",
		"args": [
			{
				"op": "<=",
				"args": [ { "property": "score" }, 50 ]
			},
			{
				"op": "isNull",
				"args": [ { "property": "nickname" } ]
			}
		]
	}`

	expr, err := ParseExpression([]byte(jsonStr))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	orExpr, ok := expr.(Or)
	if !ok {
		t.Fatalf("Expected Or expression")
	}

	if len(orExpr.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(orExpr.Children))
	}

	// Serialize back to JSON
	serialized, err := SerializeExpression(orExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare using reflect.DeepEqual
	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}

// TestNotOperator tests parsing and serialization of the Not logical operator with a single child expression.
func TestNotOperator(t *testing.T) {
	jsonStr := `{
		"op": "not",
		"args": [
			{
				"op": "in",
				"args": [ { "property": "status" }, ["active", "pending"] ]
			}
		]
	}`

	expr, err := ParseExpression([]byte(jsonStr))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	notExpr, ok := expr.(Not)
	if !ok {
		t.Fatalf("Expected Not expression")
	}

	// Serialize back to JSON
	serialized, err := SerializeExpression(notExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare using reflect.DeepEqual
	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}

// TestPropertyIsNotEqualTo tests parsing and serialization of the PropertyIsNotEqualTo operator.
func TestPropertyIsNotEqualTo(t *testing.T) {
	jsonStr := `{
		"op": "<>",
		"args": [
			{ "property": "category" },
			"electronics"
		]
	}`

	expr, err := ParseExpression([]byte(jsonStr))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	notEqualExpr, ok := expr.(PropertyIsNotEqualTo)
	if !ok {
		t.Fatalf("Expected PropertyIsNotEqualTo expression")
	}

	if notEqualExpr.Property != "category" || notEqualExpr.Value != "electronics" {
		t.Fatalf("PropertyIsNotEqualTo fields mismatch")
	}

	// Serialize back to JSON
	serialized, err := SerializeExpression(notEqualExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare using reflect.DeepEqual
	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}

// TestPropertyIsLessThan tests parsing and serialization of the PropertyIsLessThan operator.
func TestPropertyIsLessThan(t *testing.T) {
	jsonStr := `{
		"op": "<",
		"args": [
			{ "property": "price" },
			100
		]
	}`

	expr, err := ParseExpression([]byte(jsonStr))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	lessThanExpr, ok := expr.(PropertyIsLessThan)
	if !ok {
		t.Fatalf("Expected PropertyIsLessThan expression")
	}

	if lessThanExpr.Property != "price" || lessThanExpr.Value != 100.0 {
		t.Fatalf("PropertyIsLessThan fields mismatch")
	}

	// Serialize back to JSON
	serialized, err := SerializeExpression(lessThanExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare using reflect.DeepEqual
	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}

// TestPropertyIsGreaterThanOrEqualTo tests parsing and serialization of the PropertyIsGreaterThanOrEqualTo operator.
func TestPropertyIsGreaterThanOrEqualTo(t *testing.T) {
	jsonStr := `{
		"op": ">=",
		"args": [
			{ "property": "rating" },
			4.5
		]
	}`

	expr, err := ParseExpression([]byte(jsonStr))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	greaterThanOrEqualExpr, ok := expr.(PropertyIsGreaterThanOrEqualTo)
	if !ok {
		t.Fatalf("Expected PropertyIsGreaterThanOrEqualTo expression")
	}

	if greaterThanOrEqualExpr.Property != "rating" || greaterThanOrEqualExpr.Value != 4.5 {
		t.Fatalf("PropertyIsGreaterThanOrEqualTo fields mismatch")
	}

	// Serialize back to JSON
	serialized, err := SerializeExpression(greaterThanOrEqualExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare using reflect.DeepEqual
	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}

// TestBetweenOperator tests parsing and serialization of the Between operator.
func TestBetweenOperator(t *testing.T) {
	jsonStr := `{
		"op": "between",
		"args": [
			{ "property": "createdDate" },
			"2023-01-01T00:00:00Z",
			"2023-12-31T23:59:59Z"
		]
	}`

	expr, err := ParseExpression([]byte(jsonStr))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	betweenExpr, ok := expr.(Between)
	if !ok {
		t.Fatalf("Expected Between expression")
	}

	if betweenExpr.Property != "createdDate" ||
		betweenExpr.Lower != "2023-01-01T00:00:00Z" ||
		betweenExpr.Upper != "2023-12-31T23:59:59Z" {
		t.Fatalf("Between fields mismatch")
	}

	// Serialize back to JSON
	serialized, err := SerializeExpression(betweenExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare using reflect.DeepEqual
	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}

// TestLikeOperator tests parsing and serialization of the Like operator.
func TestLikeOperator(t *testing.T) {
	jsonStr := `{
		"op": "like",
		"args": [
			{ "property": "username" },
			"admin%"
		]
	}`

	expr, err := ParseExpression([]byte(jsonStr))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	likeExpr, ok := expr.(Like)
	if !ok {
		t.Fatalf("Expected Like expression")
	}

	if likeExpr.Property != "username" || likeExpr.Pattern != "admin%" {
		t.Fatalf("Like fields mismatch")
	}

	// Serialize back to JSON
	serialized, err := SerializeExpression(likeExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare using reflect.DeepEqual
	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}

// TestInOperator tests parsing and serialization of the In operator.
func TestInOperator(t *testing.T) {
	jsonStr := `{
		"op": "in",
		"args": [
			{ "property": "status" },
			["active", "inactive", "pending"]
		]
	}`

	expr, err := ParseExpression([]byte(jsonStr))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	inExpr, ok := expr.(In)
	if !ok {
		t.Fatalf("Expected In expression")
	}

	if inExpr.Property != "status" || len(inExpr.Values) != 3 {
		t.Fatalf("In fields mismatch")
	}

	// Serialize back to JSON
	serialized, err := SerializeExpression(inExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare using reflect.DeepEqual
	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}

// TestCaseiFunction tests parsing and serialization of the 'casei' function.
func TestCaseiFunction(t *testing.T) {
	jsonStr := `{
		"op": "casei",
		"args": [
			{ "property": "provider" },
			"coolsat"
		]
	}`

	expr, err := ParseExpression([]byte(jsonStr))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	fnExpr, ok := expr.(Function)
	if !ok {
		t.Fatalf("Expected Function expression")
	}

	if fnExpr.Name != "casei" {
		t.Fatalf("Expected function name 'casei', got '%s'", fnExpr.Name)
	}

	if len(fnExpr.Args) != 2 {
		t.Fatalf("Expected 2 arguments for function 'casei'")
	}

	// Serialize back to JSON
	serialized, err := SerializeExpression(fnExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare using reflect.DeepEqual
	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}

// TestSIntersectsOperatorWithDifferentGeometries tests the SIntersects operator with various GeoJSON geometries.
func TestSIntersectsOperatorWithDifferentGeometries(t *testing.T) {
	// Test with LineString
	jsonStrLineString := `{
		"op": "s_intersects",
		"args": [
			{ "property": "path" },
			{
				"type": "LineString",
				"coordinates": [
					[-77.0824, 38.7886],
					[-77.0189, 38.8351]
				]
			}
		]
	}`

	expr, err := ParseExpression([]byte(jsonStrLineString))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	sIntersectsExpr, ok := expr.(SIntersects)
	if !ok {
		t.Fatalf("Expected SIntersects expression")
	}

	if sIntersectsExpr.Property != "path" || sIntersectsExpr.Geometry.Type != "LineString" {
		t.Fatalf("SIntersects fields mismatch for LineString")
	}

	// Serialize back to JSON
	serialized, err := SerializeExpression(sIntersectsExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare using reflect.DeepEqual
	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStrLineString), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected for LineString.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}

	// Test with Point
	jsonStrPoint := `{
		"op": "s_intersects",
		"args": [
			{ "property": "location" },
			{
				"type": "Point",
				"coordinates": [-77.0365, 38.8977]
			}
		]
	}`

	expr, err = ParseExpression([]byte(jsonStrPoint))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	sIntersectsExpr, ok = expr.(SIntersects)
	if !ok {
		t.Fatalf("Expected SIntersects expression")
	}

	if sIntersectsExpr.Property != "location" || sIntersectsExpr.Geometry.Type != "Point" {
		t.Fatalf("SIntersects fields mismatch for Point")
	}

	// Serialize back to JSON
	serialized, err = SerializeExpression(sIntersectsExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare using reflect.DeepEqual
	if err := json.Unmarshal([]byte(jsonStrPoint), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected for Point.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}

// TestTIntersectsOperator tests the TIntersects operator with overlapping and non-overlapping intervals.
func TestTIntersectsOperator(t *testing.T) {
	// Overlapping interval
	jsonStrOverlapping := `{
		"op": "t_intersects",
		"args": [
			{ "property": "eventTime" },
			{ "interval": [ "2023-06-01T00:00:00Z", "2023-06-30T23:59:59Z" ] }
		]
	}`

	expr, err := ParseExpression([]byte(jsonStrOverlapping))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	tIntersectsExpr, ok := expr.(TIntersects)
	if !ok {
		t.Fatalf("Expected TIntersects expression")
	}

	if tIntersectsExpr.Property != "eventTime" {
		t.Fatalf("TIntersects property mismatch")
	}

	expectedStart := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2023, 6, 30, 23, 59, 59, 0, time.UTC)
	if !tIntersectsExpr.Interval.Start.Equal(expectedStart) || !tIntersectsExpr.Interval.End.Equal(expectedEnd) {
		t.Fatalf("TIntersects interval mismatch")
	}

	// Serialize back to JSON
	serialized, err := SerializeExpression(tIntersectsExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare using reflect.DeepEqual
	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStrOverlapping), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected for overlapping interval.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}

	// Non-overlapping interval
	jsonStrNonOverlapping := `{
		"op": "t_intersects",
		"args": [
			{ "property": "eventTime" },
			{ "interval": [ "2024-01-01T00:00:00Z", "2024-01-31T23:59:59Z" ] }
		]
	}`

	expr, err = ParseExpression([]byte(jsonStrNonOverlapping))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	tIntersectsExpr, ok = expr.(TIntersects)
	if !ok {
		t.Fatalf("Expected TIntersects expression")
	}

	if tIntersectsExpr.Property != "eventTime" {
		t.Fatalf("TIntersects property mismatch")
	}

	expectedStart = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd = time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
	if !tIntersectsExpr.Interval.Start.Equal(expectedStart) || !tIntersectsExpr.Interval.End.Equal(expectedEnd) {
		t.Fatalf("TIntersects interval mismatch")
	}

	// Serialize back to JSON
	serialized, err = SerializeExpression(tIntersectsExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare using reflect.DeepEqual
	if err := json.Unmarshal([]byte(jsonStrNonOverlapping), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected for non-overlapping interval.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}

// TestPropertyPropertyComparisonWithDifferentOperators tests property-property comparisons with various operators.
func TestPropertyPropertyComparisonWithDifferentOperators(t *testing.T) {
	operators := []string{"=", "<>", "<", "<=", ">", ">="}
	for _, op := range operators {
		t.Run(fmt.Sprintf("Operator_%s", op), func(t *testing.T) {
			jsonStr := fmt.Sprintf(`{
				"op": "%s",
				"args": [
					{ "property": "startDate" },
					{ "property": "endDate" }
				]
			}`, op)

			expr, err := ParseExpression([]byte(jsonStr))
			if err != nil {
				t.Fatalf("ParseExpression failed for operator '%s': %v", op, err)
			}

			ppc, ok := expr.(PropertyPropertyComparison)
			if !ok {
				t.Fatalf("Expected PropertyPropertyComparison expression for operator '%s'", op)
			}

			if ppc.Property1 != "startDate" || ppc.Property2 != "endDate" || ppc.Operator != op {
				t.Fatalf("PropertyPropertyComparison fields mismatch for operator '%s'", op)
			}

			// Serialize back to JSON
			serialized, err := SerializeExpression(ppc)
			if err != nil {
				t.Fatalf("SerializeExpression failed for operator '%s': %v", op, err)
			}

			// Compare using reflect.DeepEqual
			var expectedMap, actualMap map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &expectedMap); err != nil {
				t.Fatalf("Failed to unmarshal expected JSON for operator '%s': %v", op, err)
			}
			if err := json.Unmarshal(serialized, &actualMap); err != nil {
				t.Fatalf("Failed to unmarshal actual JSON for operator '%s': %v", op, err)
			}

			if !reflect.DeepEqual(expectedMap, actualMap) {
				t.Errorf("Serialized JSON does not match expected for operator '%s'.\nExpected: %v\nGot: %v", op, expectedMap, actualMap)
			}
		})
	}
}

// TestOperatorsTableDriven tests multiple operators using a table-driven approach for scalability.
func TestOperatorsTableDriven(t *testing.T) {
	tests := []struct {
		name         string
		jsonStr      string
		expectedExpr Expression
	}{
		{
			name: "PropertyIsEqualTo",
			jsonStr: `{
				"op": "=",
				"args": [
					{ "property": "status" },
					"active"
				]
			}`,
			expectedExpr: PropertyIsEqualTo{
				Property: "status",
				Value:    "active",
			},
		},
		{
			name: "PropertyIsNotEqualTo",
			jsonStr: `{
				"op": "<>",
				"args": [
					{ "property": "type" },
					"inactive"
				]
			}`,
			expectedExpr: PropertyIsNotEqualTo{
				Property: "type",
				Value:    "inactive",
			},
		},
		{
			name: "PropertyIsLessThan",
			jsonStr: `{
				"op": "<",
				"args": [
					{ "property": "age" },
					30
				]
			}`,
			expectedExpr: PropertyIsLessThan{
				Property: "age",
				Value:    30.0,
			},
		},
		{
			name: "PropertyIsLessThanOrEqualTo",
			jsonStr: `{
				"op": "<=",
				"args": [
					{ "property": "height" },
					180
				]
			}`,
			expectedExpr: PropertyIsLessThanOrEqualTo{
				Property: "height",
				Value:    180.0,
			},
		},
		{
			name: "PropertyIsGreaterThan",
			jsonStr: `{
				"op": ">",
				"args": [
					{ "property": "score" },
					75
				]
			}`,
			expectedExpr: PropertyIsGreaterThan{
				Property: "score",
				Value:    75.0,
			},
		},
		{
			name: "PropertyIsGreaterThanOrEqualTo",
			jsonStr: `{
				"op": ">=",
				"args": [
					{ "property": "rating" },
					4.5
				]
			}`,
			expectedExpr: PropertyIsGreaterThanOrEqualTo{
				Property: "rating",
				Value:    4.5,
			},
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseExpression([]byte(tt.jsonStr))
			if err != nil {
				t.Fatalf("ParseExpression failed: %v", err)
			}

			if !reflect.DeepEqual(expr, tt.expectedExpr) {
				t.Fatalf("Parsed expression does not match expected.\nExpected: %+v\nGot: %+v", tt.expectedExpr, expr)
			}

			// Serialize back to JSON
			serialized, err := SerializeExpression(expr)
			if err != nil {
				t.Fatalf("SerializeExpression failed: %v", err)
			}

			// Compare original and serialized JSON
			var expectedMap, actualMap map[string]interface{}
			if err := json.Unmarshal([]byte(tt.jsonStr), &expectedMap); err != nil {
				t.Fatalf("Failed to unmarshal expected JSON: %v", err)
			}
			if err := json.Unmarshal(serialized, &actualMap); err != nil {
				t.Fatalf("Failed to unmarshal serialized JSON: %v", err)
			}

			if !reflect.DeepEqual(expectedMap, actualMap) {
				t.Errorf("Serialized JSON does not match expected for operator '%s'.\nExpected: %v\nGot: %v", tt.name, expectedMap, actualMap)
			}
		})
	}
}

// TestBuilderChain tests the builder pattern for constructing complex expressions.
func TestBuilderChain(t *testing.T) {
	builder := NewBuilder()
	expr := builder.
		PropertyIsEqualTo("type", "satellite").
		PropertyIsGreaterThan("resolution", 0.5).
		And(
			PropertyIsLessThanOrEqualTo{"cloudCover", 20.0},
			Like{"name", "SAT-%"},
		).
		Build()

	// Expected JSON
	expectedJSON := `{
		"op": "and",
		"args": [
			{
				"op": "=",
				"args": [ { "property": "type" }, "satellite" ]
			},
			{
				"op": ">",
				"args": [ { "property": "resolution" }, 0.5 ]
			},
			{
				"op": "<=",
				"args": [ { "property": "cloudCover" }, 20 ]
			},
			{
				"op": "like",
				"args": [ { "property": "name" }, "SAT-%" ]
			}
		]
	}`

	// Serialize expr
	serialized, err := SerializeExpression(expr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	// Compare using reflect.DeepEqual
	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(expectedJSON), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected for builder chain.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}
}

// TestParseQueryables tests parsing and serialization of the queryables JSON schema.
func TestParseQueryables(t *testing.T) {
	queryablesJSON := `{
      "$schema" : "https://json-schema.org/draft/2019-09/schema",
      "$id" : "https://stac-api.example.com/queryables",
      "type" : "object",
      "title" : "Queryables for Example STAC API",
      "description" : "Queryable names for the example STAC API Item Search filter.",
      "properties" : {
        "id" : {
          "description" : "ID",
          "$ref": "https://schemas.stacspec.org/v1.0.0/item-spec/json-schema/item.json#/id"
        },
        "collection" : {
          "description" : "Collection",
          "$ref": "https://schemas.stacspec.org/v1.0.0/item-spec/json-schema/item.json#/collection"
        },
        "geometry" : {
          "description" : "Geometry",
          "$ref": "https://schemas.stacspec.org/v1.0.0/item-spec/json-schema/item.json#/geometry"
        },
        "datetime" : {
          "description" : "Datetime",
          "$ref": "https://schemas.stacspec.org/v1.0.0/item-spec/json-schema/datetime.json#/properties/datetime"
        },
        "eo:cloud_cover" : {
          "description" : "Cloud Cover",
          "$ref": "https://stac-extensions.github.io/eo/v1.0.0/schema.json#/properties/eo:cloud_cover"
        },
        "gsd" : {
          "description" : "Ground Sample Distance",
          "$ref": "https://schemas.stacspec.org/v1.0.0/item-spec/json-schema/instrument.json#/properties/gsd"
        },
        "assets_bands" : {
          "description" : "Asset eo:bands common names",
          "$ref": "https://stac-extensions.github.io/eo/v1.0.0/schema.json#/properties/eo:bands/common_name"
        }
      },
      "additionalProperties": true
    }`

	q, err := ParseQueryables([]byte(queryablesJSON))
	if err != nil {
		t.Fatalf("Failed to parse queryables: %v", err)
	}

	if err := ValidateQueryables(q); err != nil {
		t.Fatalf("Invalid queryables: %v", err)
	}

	// Serialize back to JSON
	serialized, err := SerializeQueryables(q)
	if err != nil {
		t.Fatalf("SerializeQueryables failed: %v", err)
	}

	// Compare with original (ignoring formatting)
	var originalMap, serializedMap map[string]interface{}
	if err := json.Unmarshal([]byte(queryablesJSON), &originalMap); err != nil {
		t.Fatalf("Failed to unmarshal original queryables JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &serializedMap); err != nil {
		t.Fatalf("Failed to unmarshal serialized queryables JSON: %v", err)
	}

	if !reflect.DeepEqual(originalMap, serializedMap) {
		t.Errorf("Serialized queryables JSON does not match original.\nOriginal: %v\nSerialized: %v", originalMap, serializedMap)
	}
}

// TestIsNullOperator tests parsing and serialization of the IsNull operator, including edge cases.
func TestIsNullOperator(t *testing.T) {
	// Valid IsNull expression
	jsonStrValid := `{
		"op": "isNull",
		"args": [
			{ "property": "middleName" }
		]
	}`

	expr, err := ParseExpression([]byte(jsonStrValid))
	if err != nil {
		t.Fatalf("ParseExpression failed: %v", err)
	}

	isNullExpr, ok := expr.(IsNull)
	if !ok {
		t.Fatalf("Expected IsNull expression")
	}

	if isNullExpr.Property != "middleName" {
		t.Fatalf("IsNull property mismatch")
	}

	// Serialize back to JSON
	serialized, err := SerializeExpression(isNullExpr)
	if err != nil {
		t.Fatalf("SerializeExpression failed: %v", err)
	}

	expected := `{
      "op": "isNull",
      "args": [
        {
          "property": "middleName"
        }
      ]
    }`

	var expectedMap, actualMap map[string]interface{}
	if err := json.Unmarshal([]byte(expected), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}
	if err := json.Unmarshal(serialized, &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if !reflect.DeepEqual(expectedMap, actualMap) {
		t.Errorf("Serialized JSON does not match expected.\nExpected: %v\nGot: %v", expectedMap, actualMap)
	}

	// Edge Case: Missing 'property' field
	jsonStrInvalid := `{
		"op": "isNull",
		"args": [
			{ "attribute": "middleName" }
		]
	}`

	_, err = ParseExpression([]byte(jsonStrInvalid))
	if err == nil {
		t.Fatalf("Expected error for invalid IsNull expression with missing 'property' field")
	}
}

// TestParseExpression_MissingOp tests that parsing fails when the 'op' field is missing.
func TestParseExpression_MissingOp(t *testing.T) {
	jsonStr := `{
		"args": [
			{ "property": "age" },
			30
		]
	}`

	_, err := ParseExpression([]byte(jsonStr))
	if err == nil {
		t.Fatalf("Expected error for missing 'op' field")
	}
}

// TestParseExpression_InvalidOperator tests that parsing fails with an unsupported operator.
func TestParseExpression_InvalidOperator(t *testing.T) {
	jsonStr := `{
		"op": "unsupported_op",
		"args": [
			{ "property": "age" },
			30
		]
	}`

	_, err := ParseExpression([]byte(jsonStr))
	if err == nil {
		t.Fatalf("Expected error for unsupported operator")
	}
}

// TestParseExpression_IncorrectArgTypes tests that parsing fails when argument types are incorrect.
func TestParseExpression_IncorrectArgTypes(t *testing.T) {
	// 'property' field is not a string
	jsonStr := `{
		"op": "=",
		"args": [
			{ "property": 123 },
			"active"
		]
	}`

	_, err := ParseExpression([]byte(jsonStr))
	if err == nil {
		t.Fatalf("Expected error for non-string 'property' field")
	}
}

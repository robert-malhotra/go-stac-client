// pkg/filter/parser.go

package filter

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ParseExpression parses a JSON byte slice into an Expression
func ParseExpression(data []byte) (Expression, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return parseExpressionMap(raw)
}

func parseExpressionMap(raw map[string]interface{}) (Expression, error) {
	exprType, ok := raw["op"].(string)
	if !ok {
		return nil, errors.New("missing or invalid 'op' field")
	}
	exprType = strings.ToLower(exprType)

	switch exprType {
	case "and":
		args, ok := raw["args"].([]interface{})
		if !ok {
			return nil, errors.New("'args' must be an array for 'and'")
		}
		var children []Expression
		for _, arg := range args {
			argMap, ok := arg.(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid child expression in 'and'")
			}
			expr, err := parseExpressionMap(argMap)
			if err != nil {
				return nil, err
			}
			children = append(children, expr)
		}
		return And{Children: children}, nil

	case "or":
		args, ok := raw["args"].([]interface{})
		if !ok {
			return nil, errors.New("'args' must be an array for 'or'")
		}
		var children []Expression
		for _, arg := range args {
			argMap, ok := arg.(map[string]interface{})
			if !ok {
				return nil, errors.New("invalid child expression in 'or'")
			}
			expr, err := parseExpressionMap(argMap)
			if err != nil {
				return nil, err
			}
			children = append(children, expr)
		}
		return Or{Children: children}, nil

	case "not":
		args, ok := raw["args"].([]interface{})
		if !ok || len(args) != 1 {
			return nil, errors.New("'args' must be an array with one element for 'not'")
		}
		childMap, ok := args[0].(map[string]interface{})
		if !ok {
			return nil, errors.New("invalid child expression in 'not'")
		}
		childExpr, err := parseExpressionMap(childMap)
		if err != nil {
			return nil, err
		}
		return Not{Child: childExpr}, nil

	case "=":
		return parsePropertyComparison(raw, "=")

	case "<>":
		return parsePropertyComparison(raw, "<>")

	case "<":
		return parsePropertyComparison(raw, "<")

	case "<=":
		return parsePropertyComparison(raw, "<=")

	case ">":
		return parsePropertyComparison(raw, ">")

	case ">=":
		return parsePropertyComparison(raw, ">=")

	case "between":
		return parseBetween(raw)

	case "like":
		return parseLike(raw)

	case "in":
		return parseIn(raw)

	case "s_intersects":
		return parseSIntersects(raw)

	case "t_intersects":
		return parseTIntersects(raw)

	case "isnull":
		return parseIsNull(raw)

	default:
		// Check if it's a function
		if isFunction(exprType) {
			return parseFunction(exprType, raw)
		}
		// Check for property-property comparison
		if isPropertyPropertyComparison(exprType) {
			return parsePropertyPropertyComparison(raw, exprType)
		}
		return nil, fmt.Errorf("unsupported or unknown operator: %s", exprType)
	}
}

func parsePropertyComparison(raw map[string]interface{}, operator string) (Expression, error) {
	args, ok := raw["args"].([]interface{})
	if !ok || len(args) != 2 {
		return nil, fmt.Errorf("operator '%s' requires exactly two arguments", operator)
	}

	propMap, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("first argument must be a property object")
	}
	property, ok := propMap["property"].(string)
	if !ok {
		return nil, errors.New("'property' field must be a string")
	}

	value := args[1]

	// Check if the second argument is also a property
	if valueMap, ok := value.(map[string]interface{}); ok {
		if prop2, exists := valueMap["property"]; exists {
			prop2Str, ok := prop2.(string)
			if !ok {
				return nil, errors.New("'property2' field must be a string")
			}
			return PropertyPropertyComparison{
				Property1: property,
				Operator:  operator,
				Property2: prop2Str,
			}, nil
		}
	}

	// Otherwise, return a standard comparison expression
	switch operator {
	case "=":
		return PropertyIsEqualTo{
			Property: property,
			Value:    value,
		}, nil
	case "<>":
		return PropertyIsNotEqualTo{
			Property: property,
			Value:    value,
		}, nil
	case "<":
		return PropertyIsLessThan{
			Property: property,
			Value:    value,
		}, nil
	case "<=":
		return PropertyIsLessThanOrEqualTo{
			Property: property,
			Value:    value,
		}, nil
	case ">":
		return PropertyIsGreaterThan{
			Property: property,
			Value:    value,
		}, nil
	case ">=":
		return PropertyIsGreaterThanOrEqualTo{
			Property: property,
			Value:    value,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported operator: %s", operator)
	}
}

func parseBetween(raw map[string]interface{}) (Expression, error) {
	args, ok := raw["args"].([]interface{})
	if !ok || len(args) != 3 {
		return nil, errors.New("operator 'between' requires exactly three arguments")
	}

	propMap, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("first argument must be a property object")
	}
	property, ok := propMap["property"].(string)
	if !ok {
		return nil, errors.New("'property' field must be a string")
	}

	lower := args[1]
	upper := args[2]

	return Between{
		Property: property,
		Lower:    lower,
		Upper:    upper,
	}, nil
}

func parseLike(raw map[string]interface{}) (Expression, error) {
	args, ok := raw["args"].([]interface{})
	if !ok || len(args) != 2 {
		return nil, errors.New("operator 'like' requires exactly two arguments")
	}

	propMap, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("first argument must be a property object")
	}
	property, ok := propMap["property"].(string)
	if !ok {
		return nil, errors.New("'property' field must be a string")
	}

	pattern, ok := args[1].(string)
	if !ok {
		return nil, errors.New("'pattern' must be a string")
	}

	return Like{
		Property: property,
		Pattern:  pattern,
	}, nil
}

func parseIn(raw map[string]interface{}) (Expression, error) {
	args, ok := raw["args"].([]interface{})
	if !ok || len(args) != 2 {
		return nil, errors.New("operator 'in' requires exactly two arguments")
	}

	propMap, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("first argument must be a property object")
	}
	property, ok := propMap["property"].(string)
	if !ok {
		return nil, errors.New("'property' field must be a string")
	}

	values, ok := args[1].([]interface{})
	if !ok {
		return nil, errors.New("'values' must be an array")
	}

	return In{
		Property: property,
		Values:   values,
	}, nil
}

func parseSIntersects(raw map[string]interface{}) (Expression, error) {
	args, ok := raw["args"].([]interface{})
	if !ok || len(args) != 2 {
		return nil, errors.New("operator 's_intersects' requires exactly two arguments")
	}

	propMap, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("first argument must be a property object")
	}
	property, ok := propMap["property"].(string)
	if !ok {
		return nil, errors.New("'property' field must be a string")
	}

	geometryMap, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, errors.New("second argument must be a geometry object")
	}

	geometry, err := parseGeoJSONGeometry(geometryMap)
	if err != nil {
		return nil, err
	}

	return SIntersects{
		Property: property,
		Geometry: geometry,
	}, nil
}

func parseTIntersects(raw map[string]interface{}) (Expression, error) {
	args, ok := raw["args"].([]interface{})
	if !ok || len(args) != 2 {
		return nil, errors.New("operator 't_intersects' requires exactly two arguments")
	}

	propMap, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("first argument must be a property object")
	}
	property, ok := propMap["property"].(string)
	if !ok {
		return nil, errors.New("'property' field must be a string")
	}

	intervalMap, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, errors.New("second argument must be an interval object")
	}

	interval, err := parseTimeInterval(intervalMap)
	if err != nil {
		return nil, err
	}

	return TIntersects{
		Property: property,
		Interval: interval,
	}, nil
}

func parseIsNull(raw map[string]interface{}) (Expression, error) {
	args, ok := raw["args"].([]interface{})
	if !ok || len(args) != 1 {
		return nil, errors.New("operator 'isNull' requires exactly one argument")
	}

	propMap, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("argument must be a property object")
	}
	property, ok := propMap["property"].(string)
	if !ok {
		return nil, errors.New("'property' field must be a string")
	}

	return IsNull{
		Property: property,
	}, nil
}

func parseFunction(name string, raw map[string]interface{}) (Expression, error) {
	args, ok := raw["args"].([]interface{})
	if !ok {
		return nil, errors.New("function 'args' must be an array")
	}
	return Function{
		Name: name,
		Args: args,
	}, nil
}

func parsePropertyPropertyComparison(raw map[string]interface{}, operator string) (Expression, error) {
	args, ok := raw["args"].([]interface{})
	if !ok || len(args) != 2 {
		return nil, fmt.Errorf("property-property comparison '%s' requires exactly two arguments", operator)
	}

	prop1Map, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, errors.New("first argument must be a property object")
	}
	property1, ok := prop1Map["property"].(string)
	if !ok {
		return nil, errors.New("'property1' field must be a string")
	}

	prop2Map, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, errors.New("second argument must be a property object")
	}
	property2, ok := prop2Map["property"].(string)
	if !ok {
		return nil, errors.New("'property2' field must be a string")
	}

	return PropertyPropertyComparison{
		Property1: property1,
		Operator:  operator,
		Property2: property2,
	}, nil
}

func parseGeoJSONGeometry(raw map[string]interface{}) (GeoJSONGeometry, error) {
	geoType, ok := raw["type"].(string)
	if !ok {
		return GeoJSONGeometry{}, errors.New("geometry must have a 'type' field")
	}

	coordinates, ok := raw["coordinates"]
	if !ok {
		return GeoJSONGeometry{}, errors.New("geometry must have 'coordinates'")
	}

	return GeoJSONGeometry{
		Type:        geoType,
		Coordinates: coordinates,
	}, nil
}

func parseTimeInterval(raw map[string]interface{}) (TimeInterval, error) {
	interval, ok := raw["interval"].([]interface{})
	if !ok || len(interval) != 2 {
		return TimeInterval{}, errors.New("interval must be an array of two elements")
	}

	startStr, ok := interval[0].(string)
	if !ok {
		return TimeInterval{}, errors.New("interval start must be a string")
	}

	endStr, ok := interval[1].(string)
	if !ok {
		return TimeInterval{}, errors.New("interval end must be a string")
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return TimeInterval{}, fmt.Errorf("invalid 'start' time format: %v", err)
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		return TimeInterval{}, fmt.Errorf("invalid 'end' time format: %v", err)
	}

	return TimeInterval{
		Start: start,
		End:   end,
	}, nil
}

func isFunction(op string) bool {
	functions := []string{"casei", "accenti"}
	for _, f := range functions {
		if op == f {
			return true
		}
	}
	return false
}

func isPropertyPropertyComparison(op string) bool {
	comparisons := []string{"=", "<>", "<", "<=", ">", ">="}
	for _, c := range comparisons {
		if op == c {
			return true
		}
	}
	return false
}

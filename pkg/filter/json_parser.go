// pkg/filter/parser.go

package filter

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

type expressionWrapper struct {
	Op   string            `json:"op"`
	Args []json.RawMessage `json:"args"`
}

func ParseExpression(data []byte) (Expression, error) {
	var wrapper expressionWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal expression: %w", err)
	}

	op := Operator(wrapper.Op)
	switch op {
	case OpAnd, OpOr, OpNot:
		return parseLogical(op, wrapper.Args)
	case OpEqual, OpNotEqual, OpLessThan, OpLessOrEqual, OpGreaterThan, OpGreaterOrEqual:
		return parseComparison(op, wrapper.Args)
	case OpBetween:
		return parseBetween(wrapper.Args)
	case OpLike:
		return parseLike(wrapper.Args)
	case OpIn:
		return parseIn(wrapper.Args)
	case OpSIntersects:
		return parseSIntersects(wrapper.Args)
	case OpTIntersects:
		return parseTIntersects(wrapper.Args)
	case OpIsNull:
		return parseIsNull(wrapper.Args)
	default:
		if isFunction(string(op)) {
			return parseFunction(string(op), wrapper.Args)
		}
		return nil, fmt.Errorf("unsupported operator: %s", op)
	}
}

func parseLogical(op Operator, args []json.RawMessage) (Expression, error) {
	children := make([]Expression, 0, len(args))
	for _, arg := range args {
		child, err := ParseExpression(arg)
		if err != nil {
			return nil, fmt.Errorf("failed to parse logical child: %w", err)
		}
		children = append(children, child)
	}

	return Logical{Op: op, Children: children}, nil
}

// func parseComparison(op Operator, args []json.RawMessage) (Expression, error) {
// 	if len(args) != 2 {
// 		return nil, fmt.Errorf("comparison requires exactly two arguments")
// 	}

// 	var prop property
// 	if err := json.Unmarshal(args[0], &prop); err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal property: %w", err)
// 	}

// 	var value interface{}
// 	if err := json.Unmarshal(args[1], &value); err != nil {
// 		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
// 	}

// 	return Comparison{Op: op, Property: prop.Property, Value: value}, nil
// }

func parseBetween(args []json.RawMessage) (Expression, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("between requires exactly three arguments")
	}

	var prop property
	if err := json.Unmarshal(args[0], &prop); err != nil {
		return nil, fmt.Errorf("failed to unmarshal property: %w", err)
	}

	var lower, upper interface{}
	if err := json.Unmarshal(args[1], &lower); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lower bound: %w", err)
	}
	if err := json.Unmarshal(args[2], &upper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal upper bound: %w", err)
	}

	return Between{Property: prop.Property, Lower: lower, Upper: upper}, nil
}

func parseLike(args []json.RawMessage) (Expression, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("like requires exactly two arguments")
	}

	var prop property
	if err := json.Unmarshal(args[0], &prop); err != nil {
		return nil, fmt.Errorf("failed to unmarshal property: %w", err)
	}

	var pattern string
	if err := json.Unmarshal(args[1], &pattern); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pattern: %w", err)
	}

	return Like{Property: prop.Property, Pattern: pattern}, nil
}

func parseIn(args []json.RawMessage) (Expression, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("in requires exactly two arguments")
	}

	var prop property
	if err := json.Unmarshal(args[0], &prop); err != nil {
		return nil, fmt.Errorf("failed to unmarshal property: %w", err)
	}

	var values []interface{}
	if err := json.Unmarshal(args[1], &values); err != nil {
		return nil, fmt.Errorf("failed to unmarshal values: %w", err)
	}

	return In{Property: prop.Property, Values: values}, nil
}

func parseSIntersects(args []json.RawMessage) (Expression, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("s_intersects requires exactly two arguments")
	}

	var prop property
	if err := json.Unmarshal(args[0], &prop); err != nil {
		return nil, fmt.Errorf("failed to unmarshal property: %w", err)
	}

	var geom geom.T
	if err := geojson.Unmarshal(args[1], &geom); err != nil {
		return nil, fmt.Errorf("failed to unmarshal geometry: %w", err)
	}

	return SIntersects{Property: prop.Property, Geometry: geom}, nil
}

func parseTIntersects(args []json.RawMessage) (Expression, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("t_intersects requires exactly two arguments")
	}

	var prop property
	if err := json.Unmarshal(args[0], &prop); err != nil {
		return nil, fmt.Errorf("failed to unmarshal property: %w", err)
	}

	var intervalWrapper struct {
		Interval []string `json:"interval"`
	}
	if err := json.Unmarshal(args[1], &intervalWrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal interval: %w", err)
	}

	if len(intervalWrapper.Interval) != 2 {
		return nil, fmt.Errorf("interval must contain exactly two timestamps")
	}

	start, err := time.Parse(time.RFC3339, intervalWrapper.Interval[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time: %w", err)
	}

	end, err := time.Parse(time.RFC3339, intervalWrapper.Interval[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse end time: %w", err)
	}

	return TIntersects{
		Property: prop.Property,
		Interval: TimeInterval{Start: start, End: end},
	}, nil
}

func parseIsNull(args []json.RawMessage) (Expression, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("isNull requires exactly one argument")
	}

	var prop property
	if err := json.Unmarshal(args[0], &prop); err != nil {
		return nil, fmt.Errorf("failed to unmarshal property: %w", err)
	}

	return IsNull{Property: prop.Property}, nil
}

func parseFunction(name string, args []json.RawMessage) (Expression, error) {
	var parsedArgs []interface{}
	for _, arg := range args {
		var value interface{}
		if err := json.Unmarshal(arg, &value); err != nil {
			return nil, fmt.Errorf("failed to unmarshal function argument: %w", err)
		}
		parsedArgs = append(parsedArgs, value)
	}

	return Function{Name: name, Args: parsedArgs}, nil
}

type property struct {
	Property string `json:"property"`
}

func parseProperty(data json.RawMessage) (string, error) {
	var prop property
	if err := json.Unmarshal(data, &prop); err != nil {
		return "", fmt.Errorf("failed to unmarshal property: %w", err)
	}
	if prop.Property == "" {
		return "", fmt.Errorf("failed to unmarshal property: missing 'property' field")
	}
	return prop.Property, nil
}

// Update parseComparisonExpression to use parseProperty
func parseComparison(op Operator, args []json.RawMessage) (Expression, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("comparison requires exactly two arguments")
	}

	prop, err := parseProperty(args[0])
	if err != nil {
		return nil, err
	}

	var value interface{}
	if err := json.Unmarshal(args[1], &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return Comparison{Op: op, Property: prop, Value: value}, nil
}

func isFunction(op string) bool {
	functions := map[string]bool{
		"casei":   true,
		"accenti": true,
	}
	return functions[op]
}

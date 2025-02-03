package cql2

import (
	"encoding/json"
	"fmt"
)

func (c Comparison) MarshalJSON() ([]byte, error) {
	// When serializing the left operand (a property) wrap it as an object.
	left := c.Left
	if s, ok := c.Left.(string); ok {
		left = map[string]string{"property": s}
	}
	return json.Marshal(struct {
		OP   string        `json:"op"`
		Args []interface{} `json:"args"`
	}{
		OP:   string(c.Operator),
		Args: []interface{}{left, c.Right},
	})
}

func (lo LogicalOperator) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		OP   string        `json:"op"`
		Args []interface{} `json:"args"`
	}{
		OP:   string(lo.Operator),
		Args: []interface{}{lo.Left, lo.Right},
	})
}

func (n Not) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		OP   string        `json:"op"`
		Args []interface{} `json:"args"`
	}{
		OP:   "NOT",
		Args: []interface{}{n.Expression},
	})
}

// SerializeJSON serializes an expression to JSON.
func SerializeJSON(expr Expression) ([]byte, error) {
	if expr == nil {
		return nil, fmt.Errorf("cannot serialize nil expression")
	}
	return json.Marshal(expr)
}

type cqlJSON struct {
	OP   string            `json:"op"`
	Args []json.RawMessage `json:"args"`
}

func DeserializeJSON(data []byte) (Expression, error) {
	var raw cqlJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	switch raw.OP {
	case "NOT":
		return parseNot(raw.Args)
	case "AND", "OR":
		return parseLogical(raw.OP, raw.Args)
	default:
		return parseComparison(raw.OP, raw.Args)
	}
}

func parseNot(args []json.RawMessage) (Expression, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("NOT requires 1 argument, got %d", len(args))
	}
	expr, err := ParseJSON(args[0])
	if err != nil {
		return nil, err
	}
	return Not{Expression: expr}, nil
}

func parseLogical(op string, args []json.RawMessage) (Expression, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("%s requires 2 arguments, got %d", op, len(args))
	}

	left, err := ParseJSON(args[0])
	if err != nil {
		return nil, err
	}
	right, err := ParseJSON(args[1])
	if err != nil {
		return nil, err
	}

	return LogicalOperator{
		Operator: Operator(op),
		Left:     left,
		Right:    right,
	}, nil
}

func parseComparison(op string, args []json.RawMessage) (Expression, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("comparison requires 2 arguments, got %d", len(args))
	}

	left, err := parseArg(args[0])
	if err != nil {
		return nil, err
	}
	right, err := parseArg(args[1])
	if err != nil {
		return nil, err
	}

	return Comparison{
		Operator: Operator(op),
		Left:     left,
		Right:    right,
	}, nil
}

func parseArg(data json.RawMessage) (interface{}, error) {
	// Try property first.
	var prop struct {
		Property string `json:"property"`
	}
	if err := json.Unmarshal(data, &prop); err == nil && prop.Property != "" {
		return prop.Property, nil
	}

	// Try literal value.
	var literal interface{}
	if err := json.Unmarshal(data, &literal); err == nil {
		return literal, nil
	}

	return nil, fmt.Errorf("invalid argument format")
}

package cql2

import (
	"encoding/json"
	"errors"
	"fmt"
)

func ParseJSON(input []byte) (Expression, error) {
	var raw json.RawMessage
	if err := json.Unmarshal(input, &raw); err != nil {
		return nil, err
	}
	return parseJSONExpr(raw)
}

func parseJSONExpr(data json.RawMessage) (Expression, error) {
	// Try logical operator first
	var logical struct {
		Op   string            `json:"op"`
		Args []json.RawMessage `json:"args"`
	}
	if err := json.Unmarshal(data, &logical); err == nil {
		switch logical.Op {
		case "AND", "OR":
			if len(logical.Args) != 2 {
				return nil, fmt.Errorf("%s requires 2 arguments", logical.Op)
			}
			left, err := parseJSONExpr(logical.Args[0])
			if err != nil {
				return nil, err
			}
			right, err := parseJSONExpr(logical.Args[1])
			if err != nil {
				return nil, err
			}
			return &LogicalOperator{
				Operator: Operator(logical.Op),
				Left:     left,
				Right:    right,
			}, nil
		case "NOT":
			if len(logical.Args) != 1 {
				return nil, errors.New("NOT requires 1 argument")
			}
			expr, err := parseJSONExpr(logical.Args[0])
			if err != nil {
				return nil, err
			}
			return &Not{Expression: expr}, nil
		}
	}

	// Try comparison
	var comp struct {
		Op   string            `json:"op"`
		Args []json.RawMessage `json:"args"`
	}
	if err := json.Unmarshal(data, &comp); err == nil {
		if len(comp.Args) != 2 {
			return nil, errors.New("comparison requires exactly 2 arguments")
		}

		left, err := parseJSONArg(comp.Args[0])
		if err != nil {
			return nil, err
		}
		right, err := parseJSONArg(comp.Args[1])
		if err != nil {
			return nil, err
		}

		return &Comparison{
			Operator: Operator(comp.Op),
			Left:     left,
			Right:    right,
		}, nil
	}

	return nil, errors.New("invalid expression format")
}

func parseJSONArg(data json.RawMessage) (Expression, error) {
	// Try to parse as property
	var prop struct {
		Property string `json:"property"`
	}
	if err := json.Unmarshal(data, &prop); err == nil && prop.Property != "" {
		return Property{Name: prop.Property}, nil
	}

	// Try to parse as literal value
	var literal interface{}
	if err := json.Unmarshal(data, &literal); err == nil {
		return Literal{Value: literal}, nil
	}

	return nil, errors.New("invalid argument format")
}

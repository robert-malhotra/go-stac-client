package cql2

import (
	"encoding/json"
	"fmt"
)

type Visitor interface {
	VisitComparison(op string, left interface{}, right interface{}) error
	VisitLogical(op string, args []interface{}) error
	VisitFunction(name string, args []interface{}) error
	VisitProperty(name string) error
	VisitLiteral(value interface{}) error
}

type Parser struct {
	visitor Visitor
}

func NewParser(v Visitor) *Parser {
	return &Parser{visitor: v}
}

func (p *Parser) Parse(jsonStr string) error {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return p.visit(data)
}

func (p *Parser) visit(node interface{}) error {
	switch v := node.(type) {
	case map[string]interface{}:
		if op, ok := v["op"].(string); ok {
			args, ok := v["args"].([]interface{})
			if !ok {
				return fmt.Errorf("invalid args for operator %s", op)
			}

			operator, ok := GetOperator(op)
			if !ok {
				return fmt.Errorf("unknown operator: %s", op)
			}

			switch operator {
			case OpEquals, OpNotEquals, OpLessThan, OpGreaterThan, OpLessThanEquals, OpGreaterThanEquals:
				if len(args) != 2 {
					return fmt.Errorf("comparison operator requires exactly two arguments")
				}

				// Handle the right operand being a timestamp
				right := args[1]
				if m, ok := right.(map[string]interface{}); ok {
					if ts, ok := m["timestamp"].(string); ok {
						right = ts
					}
				}

				return p.visitor.VisitComparison(op, args[0], right)

			case OpAnd, OpOr:
				if err := p.visitor.VisitLogical(op, args); err != nil {
					return err
				}
				for _, arg := range args {
					if err := p.visit(arg); err != nil {
						return err
					}
				}
				return nil

			case OpNot:
				if len(args) != 1 {
					return fmt.Errorf("not operator requires exactly one argument")
				}
				return p.visitor.VisitLogical(op, args)

			case OpSIntersects, OpSContains, OpSWithin:
				if len(args) != 2 {
					return fmt.Errorf("spatial operator requires exactly two arguments")
				}
				return p.visitor.VisitFunction(op, args)
			}
		}

		if property, ok := v["property"].(string); ok {
			return p.visitor.VisitProperty(property)
		}

		return p.visitor.VisitLiteral(v)

	case []interface{}:
		for _, item := range v {
			if err := p.visit(item); err != nil {
				return err
			}
		}
		return nil

	default:
		return p.visitor.VisitLiteral(v)
	}
}

func GetOperator(op string) (Operator, bool) {
	if _, ok := map[Operator]bool{
		OpEquals:            true,
		OpNotEquals:         true,
		OpLessThan:          true,
		OpGreaterThan:       true,
		OpLessThanEquals:    true,
		OpGreaterThanEquals: true,
		OpSIntersects:       true,
		OpSContains:         true,
		OpSWithin:           true,
		OpAnd:               true,
		OpOr:                true,
		OpNot:               true,
	}[Operator(op)]; ok {
		return Operator(op), true
	}
	return "", false
}

type CQL2Visitor interface {
	OnEquals(property string, value interface{}) error
	OnLessThan(property string, value interface{}) error
	OnGreaterThan(property string, value interface{}) error
	OnLessThanOrEquals(property string, value interface{}) error
	OnGreaterThanOrEquals(property string, value interface{}) error
	OnNotEquals(property string, value interface{}) error
	OnSIntersects(property string, geometry interface{}) error
	OnSContains(property string, geometry interface{}) error
	OnSWithin(property string, geometry interface{}) error
	OnAnd(args []interface{}) error
	OnOr(args []interface{}) error
	OnNot(arg interface{}) error
}

type Adapter struct {
	cql2 CQL2Visitor
}

func NewAdapter(v CQL2Visitor) *Adapter {
	return &Adapter{cql2: v}
}

func (a *Adapter) VisitComparison(op string, left, right interface{}) error {
	prop, ok := left.(map[string]interface{})
	if !ok {
		return fmt.Errorf("left operand must be property")
	}

	propName := prop["property"].(string)
	operator, ok := GetOperator(op)
	if !ok {
		return fmt.Errorf("unknown comparison operator: %s", op)
	}

	switch operator {
	case OpEquals:
		return a.cql2.OnEquals(propName, right)
	case OpLessThan:
		return a.cql2.OnLessThan(propName, right)
	case OpGreaterThan:
		return a.cql2.OnGreaterThan(propName, right)
	case OpLessThanEquals:
		return a.cql2.OnLessThanOrEquals(propName, right)
	case OpGreaterThanEquals:
		return a.cql2.OnGreaterThanOrEquals(propName, right)
	case OpNotEquals:
		return a.cql2.OnNotEquals(propName, right)
	default:
		return fmt.Errorf("operator %s is not a comparison operator", operator)
	}
}

func (a *Adapter) VisitFunction(name string, args []interface{}) error {
	if len(args) != 2 {
		return fmt.Errorf("spatial functions require exactly two arguments")
	}

	prop, ok := args[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("first argument must be property")
	}

	propName := prop["property"].(string)
	operator, ok := GetOperator(name)
	if !ok {
		return fmt.Errorf("unknown function: %s", name)
	}

	switch operator {
	case OpSIntersects:
		return a.cql2.OnSIntersects(propName, args[1])
	case OpSContains:
		return a.cql2.OnSContains(propName, args[1])
	case OpSWithin:
		return a.cql2.OnSWithin(propName, args[1])
	default:
		return fmt.Errorf("operator %s is not a spatial operator", operator)
	}
}

func (a *Adapter) VisitLogical(op string, args []interface{}) error {
	operator, ok := GetOperator(op)
	if !ok {
		return fmt.Errorf("unknown logical operator: %s", op)
	}

	switch operator {
	case OpAnd:
		return a.cql2.OnAnd(args)
	case OpOr:
		return a.cql2.OnOr(args)
	case OpNot:
		if len(args) != 1 {
			return fmt.Errorf("not operator requires exactly one argument")
		}
		return a.cql2.OnNot(args[0])
	default:
		return fmt.Errorf("operator %s is not a logical operator", operator)
	}
}

func (a *Adapter) VisitProperty(name string) error {
	return nil
}

func (a *Adapter) VisitLiteral(value interface{}) error {
	return nil
}

// pkg/filter/terminal.go

package filter

import (
	"fmt"
)

// TerminalOperation represents any terminal (leaf) operation in a filter
type TerminalOperation interface {
	// GetProperty returns the property name this operation acts on
	GetProperty() string
	// GetValue returns the value of the operation
	GetValue() interface{}
	// GetOp returns the operator of the operation
	Type() Operator
}

// Ensure all terminal operations implement the interface
var (
	_ TerminalOperation = Comparison{}
	_ TerminalOperation = Between{}
	_ TerminalOperation = Like{}
	_ TerminalOperation = In{}
	_ TerminalOperation = IsNull{}
	_ TerminalOperation = SIntersects{}
	_ TerminalOperation = TIntersects{}
)

// Implement interface methods for all terminal operations

// Comparison
func (c Comparison) GetProperty() string   { return c.Property }
func (c Comparison) GetValue() interface{} { return c.Value }

// Between
func (b Between) GetProperty() string { return b.Property }
func (b Between) GetValue() interface{} {
	return map[string]interface{}{
		"lower": b.Lower,
		"upper": b.Upper,
	}
}

// Like
func (l Like) GetProperty() string   { return l.Property }
func (l Like) GetValue() interface{} { return l.Pattern }

// In
func (i In) GetProperty() string   { return i.Property }
func (i In) GetValue() interface{} { return i.Values }

// IsNull
func (n IsNull) GetProperty() string   { return n.Property }
func (n IsNull) GetValue() interface{} { return nil }

// SIntersects
func (s SIntersects) GetProperty() string   { return s.Property }
func (s SIntersects) GetValue() interface{} { return s.Geometry }

// TIntersects
func (t TIntersects) GetProperty() string   { return t.Property }
func (t TIntersects) GetValue() interface{} { return t.Interval }

// ExtractTerminalOps extracts all terminal operations from an expression
// It only supports AND operations and terminal operations
func ExtractTerminalOps(expr Expression) ([]TerminalOperation, error) {
	if expr == nil {
		return nil, nil
	}

	var ops []TerminalOperation

	switch e := expr.(type) {
	case Logical:
		if e.Op != OpAnd {
			return nil, fmt.Errorf("only AND operations are supported, got: %s", e.Op)
		}
		for _, child := range e.Children {
			childOps, err := ExtractTerminalOps(child)
			if err != nil {
				return nil, err
			}
			ops = append(ops, childOps...)
		}

	case Comparison:
		ops = append(ops, e)
	case Between:
		ops = append(ops, e)
	case Like:
		ops = append(ops, e)
	case In:
		ops = append(ops, e)
	case IsNull:
		ops = append(ops, e)
	case SIntersects:
		ops = append(ops, e)
	case TIntersects:
		ops = append(ops, e)
	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}

	return ops, nil
}

// GroupByProperty groups terminal operations by their property name
func GroupByProperty(ops []TerminalOperation) map[string][]TerminalOperation {
	result := make(map[string][]TerminalOperation)
	for _, op := range ops {
		prop := op.GetProperty()
		result[prop] = append(result[prop], op)
	}
	return result
}

// GroupByOperator groups terminal operations by their operator type
func GroupByOperator(ops []TerminalOperation) map[Operator][]TerminalOperation {
	result := make(map[Operator][]TerminalOperation)
	for _, op := range ops {
		opType := op.Type()
		result[opType] = append(result[opType], op)
	}
	return result
}

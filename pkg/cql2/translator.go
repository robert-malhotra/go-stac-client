package cql2

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Translator is a generic interface for translating nodes.
type Translator[T any] interface {
	Translate(node Node) (T, error)
}

// ODataTranslator translates nodes to OData query strings.
type ODataTranslator struct{}

func (t *ODataTranslator) Translate(node Node) (string, error) {
	switch n := node.(type) {
	case *ComparisonNode:
		return fmt.Sprintf("%s %s %v", n.Property, odataOperator(n.Operator), n.Value), nil
	case *LogicalNode:
		parts := []string{}
		for _, child := range n.Children {
			childStr, err := t.Translate(child)
			if err != nil {
				return "", err
			}
			parts = append(parts, childStr)
		}
		joiner := " and "
		if n.Operator == LogicalOr {
			joiner = " or "
		}
		return fmt.Sprintf("(%s)", strings.Join(parts, joiner)), nil
	case *SpatialNode:
		geometryJSON, err := json.Marshal(n.Geometry)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s(%s)", n.Operator, geometryJSON), nil
	default:
		return "", fmt.Errorf("unsupported node type: %s", node.Type())
	}
}

func odataOperator(op ComparisonOperator) string {
	switch op {
	case OpEq:
		return "eq"
	case OpNe:
		return "ne"
	case OpLt:
		return "lt"
	case OpLe:
		return "le"
	case OpGt:
		return "gt"
	case OpGe:
		return "ge"
	default:
		return string(op)
	}
}

// SQLTranslator translates nodes to SQL query strings.
type SQLTranslator struct{}

func (t *SQLTranslator) Translate(node Node) (string, error) {
	switch n := node.(type) {
	case *ComparisonNode:
		return fmt.Sprintf("%s %s '%v'", n.Property, sqlOperator(n.Operator), n.Value), nil
	case *LogicalNode:
		parts := []string{}
		for _, child := range n.Children {
			childStr, err := t.Translate(child)
			if err != nil {
				return "", err
			}
			parts = append(parts, childStr)
		}
		joiner := " AND "
		if n.Operator == LogicalOr {
			joiner = " OR "
		}
		return fmt.Sprintf("(%s)", strings.Join(parts, joiner)), nil
	case *SpatialNode:
		return fmt.Sprintf("%s(%v)", n.Operator, n.Geometry), nil
	default:
		return "", fmt.Errorf("unsupported node type: %s", node.Type())
	}
}

func sqlOperator(op ComparisonOperator) string {
	switch op {
	case OpEq:
		return "="
	case OpNe:
		return "<>"
	case OpLt:
		return "<"
	case OpLe:
		return "<="
	case OpGt:
		return ">"
	case OpGe:
		return ">="
	default:
		return string(op)
	}
}

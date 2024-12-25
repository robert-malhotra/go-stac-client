package cql2

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Node represents a CQL2 expression node.
type Node interface {
	Type() string
	Render() string
}

// BaseNode provides a common structure for all nodes.
type BaseNode struct {
	NodeType string `json:"type"`
}

func (n *BaseNode) Type() string {
	return n.NodeType
}

// ComparisonNode represents a comparison operation.
type ComparisonNode struct {
	BaseNode
	Property string             `json:"property"`
	Operator ComparisonOperator `json:"operator"`
	Value    interface{}        `json:"value"`
}

func NewComparisonNode(property string, operator ComparisonOperator, value interface{}) *ComparisonNode {
	return &ComparisonNode{
		BaseNode: BaseNode{NodeType: "comparison"},
		Property: property,
		Operator: operator,
		Value:    value,
	}
}

func (n *ComparisonNode) Render() string {
	return fmt.Sprintf("%s %s %v", n.Property, n.Operator, n.Value)
}

// LogicalNode represents a logical operation (AND/OR).
type LogicalNode struct {
	BaseNode
	Operator LogicalOperator `json:"operator"`
	Children []Node          `json:"children"`
}

func NewLogicalNode(operator LogicalOperator, children ...Node) *LogicalNode {
	return &LogicalNode{
		BaseNode: BaseNode{NodeType: "logical"},
		Operator: operator,
		Children: children,
	}
}

func (n *LogicalNode) Render() string {
	subExpressions := []string{}
	for _, child := range n.Children {
		subExpressions = append(subExpressions, child.Render())
	}
	return fmt.Sprintf("(%s)", strings.Join(subExpressions, fmt.Sprintf(" %s ", n.Operator)))
}

// SpatialNode represents a spatial filter.
type SpatialNode struct {
	BaseNode
	Operator SpatialOperator `json:"operator"`
	Geometry interface{}     `json:"geometry"`
}

func NewSpatialNode(operator SpatialOperator, geometry interface{}) *SpatialNode {
	return &SpatialNode{
		BaseNode: BaseNode{NodeType: "spatial"},
		Operator: operator,
		Geometry: geometry,
	}
}

func (n *SpatialNode) Render() string {
	geometryJSON, _ := json.Marshal(n.Geometry)
	return fmt.Sprintf("%s(%s)", n.Operator, geometryJSON)
}

// ComparisonOperator defines comparison operators.
type ComparisonOperator string

const (
	OpEq ComparisonOperator = "="
	OpNe ComparisonOperator = "!="
	OpLt ComparisonOperator = "<"
	OpLe ComparisonOperator = "<="
	OpGt ComparisonOperator = ">"
	OpGe ComparisonOperator = ">="
)

// LogicalOperator defines logical operators (AND/OR).
type LogicalOperator string

const (
	LogicalAnd LogicalOperator = "AND"
	LogicalOr  LogicalOperator = "OR"
)

// SpatialOperator defines spatial relationships.
type SpatialOperator string

const (
	SpatialIntersects SpatialOperator = "INTERSECTS"
	SpatialContains   SpatialOperator = "CONTAINS"
	SpatialWithin     SpatialOperator = "WITHIN"
)

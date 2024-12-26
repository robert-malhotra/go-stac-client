// pkg/filter/models.go

package filter

import (
	"time"
)

// Base interface for all CQL2 expressions
type Expression interface {
	ExpressionType() string
}

// Logical Operators

type And struct {
	Children []Expression `json:"children"`
}

func (a And) ExpressionType() string { return "and" }

type Or struct {
	Children []Expression `json:"children"`
}

func (o Or) ExpressionType() string { return "or" }

type Not struct {
	Child Expression `json:"child"`
}

func (n Not) ExpressionType() string { return "not" }

// Comparison Operators

type PropertyIsEqualTo struct {
	Property string      `json:"property"`
	Value    interface{} `json:"value"`
}

func (p PropertyIsEqualTo) ExpressionType() string { return "=" }

type PropertyIsNotEqualTo struct {
	Property string      `json:"property"`
	Value    interface{} `json:"value"`
}

func (p PropertyIsNotEqualTo) ExpressionType() string { return "<>" }

type PropertyIsLessThan struct {
	Property string      `json:"property"`
	Value    interface{} `json:"value"`
}

func (p PropertyIsLessThan) ExpressionType() string { return "<" }

type PropertyIsLessThanOrEqualTo struct {
	Property string      `json:"property"`
	Value    interface{} `json:"value"`
}

func (p PropertyIsLessThanOrEqualTo) ExpressionType() string { return "<=" }

type PropertyIsGreaterThan struct {
	Property string      `json:"property"`
	Value    interface{} `json:"value"`
}

func (p PropertyIsGreaterThan) ExpressionType() string { return ">" }

type PropertyIsGreaterThanOrEqualTo struct {
	Property string      `json:"property"`
	Value    interface{} `json:"value"`
}

func (p PropertyIsGreaterThanOrEqualTo) ExpressionType() string { return ">=" }

// Advanced Comparison Operators

type Between struct {
	Property string      `json:"property"`
	Lower    interface{} `json:"lower"`
	Upper    interface{} `json:"upper"`
}

func (b Between) ExpressionType() string { return "between" }

type Like struct {
	Property string `json:"property"`
	Pattern  string `json:"pattern"`
}

func (l Like) ExpressionType() string { return "like" }

type In struct {
	Property string        `json:"property"`
	Values   []interface{} `json:"values"`
}

func (i In) ExpressionType() string { return "in" }

// Functions

type Function struct {
	Name string        `json:"function"`
	Args []interface{} `json:"args"`
}

func (f Function) ExpressionType() string { return f.Name }

// Spatial Operators

type SIntersects struct {
	Property string          `json:"property"`
	Geometry GeoJSONGeometry `json:"geometry"`
}

func (s SIntersects) ExpressionType() string { return "s_intersects" }

// Temporal Operators

type TIntersects struct {
	Property string       `json:"property"`
	Interval TimeInterval `json:"interval"`
}

func (t TIntersects) ExpressionType() string { return "t_intersects" }

// Property-Property Comparison

type PropertyPropertyComparison struct {
	Property1 string `json:"property1"`
	Operator  string `json:"operator"`
	Property2 string `json:"property2"`
}

func (ppc PropertyPropertyComparison) ExpressionType() string { return ppc.Operator }

// IS NULL Operator

type IsNull struct {
	Property string `json:"property"`
}

func (i IsNull) ExpressionType() string { return "isNull" }

// Utility Structures

type GeoJSONGeometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

type TimeInterval struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

package filter

import (
	"time"

	"github.com/twpayne/go-geom"
)

// Operator represents any filter operator
type Operator string

// Define all operators in one place
const (
	// Logical operators
	OpAnd Operator = "and"
	OpOr  Operator = "or"
	OpNot Operator = "not"

	// Comparison operators
	OpEqual          Operator = "="
	OpNotEqual       Operator = "<>"
	OpLessThan       Operator = "<"
	OpLessOrEqual    Operator = "<="
	OpGreaterThan    Operator = ">"
	OpGreaterOrEqual Operator = ">="
	OpBetween        Operator = "between"
	OpLike           Operator = "like"
	OpIn             Operator = "in"
	OpIsNull         Operator = "isNull"

	// Spatial and temporal operators
	OpSIntersects Operator = "s_intersects"
	OpTIntersects Operator = "t_intersects"
)

// Expression interface represents any filter expression
type Expression interface {
	Type() Operator
}

// Standard expression types
type (
	Logical struct {
		Op       Operator
		Children []Expression
	}

	Comparison struct {
		Op       Operator
		Property string
		Value    interface{}
	}

	Between struct {
		Property string
		Lower    interface{}
		Upper    interface{}
	}

	Like struct {
		Property string
		Pattern  string
	}

	In struct {
		Property string
		Values   []interface{}
	}

	IsNull struct {
		Property string
	}

	Function struct {
		Name string
		Args []interface{}
	}

	SIntersects struct {
		Property string
		Geometry geom.T
	}

	TIntersects struct {
		Property string
		Interval TimeInterval
	}
)

// TimeInterval represents a time range
type TimeInterval struct {
	Start time.Time
	End   time.Time
}

// Type implementations
func (e Logical) Type() Operator     { return e.Op }
func (e Comparison) Type() Operator  { return e.Op }
func (e Between) Type() Operator     { return OpBetween }
func (e Like) Type() Operator        { return OpLike }
func (e In) Type() Operator          { return OpIn }
func (e IsNull) Type() Operator      { return OpIsNull }
func (e Function) Type() Operator    { return Operator(e.Name) }
func (e SIntersects) Type() Operator { return OpSIntersects }
func (e TIntersects) Type() Operator { return OpTIntersects }

// Builder provides a fluent interface for constructing expressions
type Builder struct {
	expr Expression
}

// NewBuilder creates a new filter builder
func NewBuilder() *Builder {
	return &Builder{}
}

// Build returns the final expression
func (b *Builder) Build() Expression {
	return b.expr
}

// Helper method to combine expressions with AND
func (b *Builder) addWithAnd(expr Expression) *Builder {
	if b.expr == nil {
		b.expr = expr
		return b
	}

	if curr, ok := b.expr.(Logical); ok && curr.Op == OpAnd {
		curr.Children = append(curr.Children, expr)
		b.expr = curr
	} else {
		b.expr = Logical{Op: OpAnd, Children: []Expression{b.expr, expr}}
	}
	return b
}

// Builder methods for each expression type
func (b *Builder) And(exprs ...Expression) *Builder {
	if len(exprs) == 0 {
		return b
	}
	return b.addWithAnd(Logical{Op: OpAnd, Children: exprs})
}

func (b *Builder) Or(exprs ...Expression) *Builder {
	if len(exprs) == 0 {
		return b
	}
	orExpr := Logical{Op: OpOr, Children: exprs}
	return b.addWithAnd(orExpr)
}

func (b *Builder) Not(expr Expression) *Builder {
	return b.addWithAnd(Logical{Op: OpNot, Children: []Expression{expr}})
}

func (b *Builder) Equal(property string, value interface{}) *Builder {
	return b.addWithAnd(Comparison{Op: OpEqual, Property: property, Value: value})
}

func (b *Builder) NotEqual(property string, value interface{}) *Builder {
	return b.addWithAnd(Comparison{Op: OpNotEqual, Property: property, Value: value})
}

func (b *Builder) LessThan(property string, value interface{}) *Builder {
	return b.addWithAnd(Comparison{Op: OpLessThan, Property: property, Value: value})
}

func (b *Builder) GreaterThan(property string, value interface{}) *Builder {
	return b.addWithAnd(Comparison{Op: OpGreaterThan, Property: property, Value: value})
}

func (b *Builder) LessThanEqual(property string, value interface{}) *Builder {
	return b.addWithAnd(Comparison{Op: OpLessOrEqual, Property: property, Value: value})
}

func (b *Builder) GreaterThanEqual(property string, value interface{}) *Builder {
	return b.addWithAnd(Comparison{Op: OpGreaterOrEqual, Property: property, Value: value})
}

func (b *Builder) Between(property string, lower, upper interface{}) *Builder {
	return b.addWithAnd(Between{Property: property, Lower: lower, Upper: upper})
}

func (b *Builder) Like(property, pattern string) *Builder {
	return b.addWithAnd(Like{Property: property, Pattern: pattern})
}

func (b *Builder) In(property string, values []interface{}) *Builder {
	return b.addWithAnd(In{Property: property, Values: values})
}

func (b *Builder) IsNull(property string) *Builder {
	return b.addWithAnd(IsNull{Property: property})
}

func (b *Builder) Function(name string, args ...interface{}) *Builder {
	return b.addWithAnd(Function{Name: name, Args: args})
}

func (b *Builder) SIntersects(property string, geometry geom.T) *Builder {
	return b.addWithAnd(SIntersects{Property: property, Geometry: geometry})
}

func (b *Builder) TIntersects(property string, interval TimeInterval) *Builder {
	return b.addWithAnd(TIntersects{Property: property, Interval: interval})
}

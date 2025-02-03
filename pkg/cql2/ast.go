package cql2

// Expression is now simply an interface implemented by composite nodes.
// Leaf values (properties and literal values) are represented by plain Go types.
type Expression interface {
	isExpr()
}

type Comparison struct {
	Operator Operator    // e.g. "=", ">", etc.
	Left     string      // left operand (property name as string)
	Right    interface{} // right operand (literal value, geometry, etc.)
}

func (Comparison) isExpr() {}

type LogicalOperator struct {
	Operator Operator   // e.g. "and", "or"
	Left     Expression // left sub-expression
	Right    Expression // right sub-expression
}

func (LogicalOperator) isExpr() {}

type Not struct {
	Expression Expression // the expression being negated
}

func (Not) isExpr() {}

type Operator string

const (
	// Comparison operators
	OpEquals            Operator = "="
	OpNotEquals         Operator = "!="
	OpLessThan          Operator = "<"
	OpGreaterThan       Operator = ">"
	OpLessThanEquals    Operator = "<="
	OpGreaterThanEquals Operator = ">="

	// Spatial operators
	OpSIntersects Operator = "s_intersects"
	OpSContains   Operator = "s_contains"
	OpSWithin     Operator = "s_within"

	// Logical operators
	OpAnd Operator = "AND"
	OpOr  Operator = "OR"
	OpNot Operator = "NOT"
)

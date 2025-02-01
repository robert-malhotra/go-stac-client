package cql2

type Expression interface {
	isExpr()
}

type Property struct {
	Name string
}

func (Property) isExpr() {}

type Literal struct {
	Value interface{}
}

func (Literal) isExpr() {}

type Comparison struct {
	Operator Operator
	Left     Expression
	Right    Expression
}

func (Comparison) isExpr() {}

type LogicalOperator struct {
	Operator Operator
	Left     Expression
	Right    Expression
}

func (LogicalOperator) isExpr() {}

type Not struct {
	Expression Expression
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
	OpAnd Operator = "and"
	OpOr  Operator = "or"
	OpNot Operator = "not"
)

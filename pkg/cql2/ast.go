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
	Operator string
	Left     Expression
	Right    Expression
}

func (Comparison) isExpr() {}

type LogicalOperator struct {
	Operator string
	Left     Expression
	Right    Expression
}

func (LogicalOperator) isExpr() {}

type Not struct {
	Expression Expression
}

func (Not) isExpr() {}

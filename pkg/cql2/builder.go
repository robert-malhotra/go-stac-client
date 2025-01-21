package cql2

import "encoding/json"

type QueryBuilder struct {
	current Expression
	stack   []Expression
	negate  bool
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

func (qb *QueryBuilder) Where(property string) *ComparisonBuilder {
	return &ComparisonBuilder{
		qb:   qb,
		left: Property{Name: property},
	}
}

type ComparisonBuilder struct {
	qb    *QueryBuilder
	left  Expression
	op    string
	right Expression
}

func (cb *ComparisonBuilder) Eq(value interface{}) *QueryBuilder {
	return cb.completeComparison("=", value)
}

func (cb *ComparisonBuilder) Neq(value interface{}) *QueryBuilder {
	return cb.completeComparison("<>", value)
}

func (cb *ComparisonBuilder) Gt(value interface{}) *QueryBuilder {
	return cb.completeComparison(">", value)
}

func (cb *ComparisonBuilder) Gte(value interface{}) *QueryBuilder {
	return cb.completeComparison(">=", value)
}

func (cb *ComparisonBuilder) Lt(value interface{}) *QueryBuilder {
	return cb.completeComparison("<", value)
}

func (cb *ComparisonBuilder) Lte(value interface{}) *QueryBuilder {
	return cb.completeComparison("<=", value)
}

func (cb *ComparisonBuilder) completeComparison(op string, value interface{}) *QueryBuilder {
	comparison := Comparison{
		Operator: op,
		Left:     cb.left,
		Right:    Literal{Value: value},
	}

	if cb.qb.negate {
		cb.qb.current = Not{Expression: comparison}
		cb.qb.negate = false
	} else {
		cb.qb.current = comparison
	}

	return cb.qb
}

func (qb *QueryBuilder) And() *QueryBuilder {
	return qb.logicalOperator("AND")
}

func (qb *QueryBuilder) Or() *QueryBuilder {
	return qb.logicalOperator("OR")
}

func (qb *QueryBuilder) Not() *QueryBuilder {
	qb.negate = true
	return qb
}

func (qb *QueryBuilder) logicalOperator(op string) *QueryBuilder {
	if qb.current == nil {
		return qb
	}

	qb.stack = append(qb.stack, LogicalOperator{
		Operator: op,
		Left:     qb.current,
	})
	qb.current = nil
	return qb
}

func (qb *QueryBuilder) Build() Expression {
	var result Expression

	for _, expr := range qb.stack {
		if lo, ok := expr.(LogicalOperator); ok {
			if result == nil {
				result = lo
			} else {
				result = LogicalOperator{
					Operator: lo.Operator,
					Left:     result,
					Right:    lo.Left,
				}
			}
		}
	}

	if qb.current != nil {
		if result == nil {
			result = qb.current
		} else {
			result = LogicalOperator{
				Operator: "AND",
				Left:     result,
				Right:    qb.current,
			}
		}
	}

	return result
}

func (qb *QueryBuilder) ToJSON() ([]byte, error) {
	expr := qb.Build()
	return json.Marshal(expr)
}

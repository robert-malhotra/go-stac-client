package cql2

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	cqlLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"Keyword", `(?i)\b(AND|OR|NOT)\b`},
		{"Ident", `[a-zA-Z_][a-zA-Z0-9_]*`},
		{"Number", `[-+]?\d*\.?\d+([eE][-+]?\d+)?`},
		{"String", `"(\\"|[^"])*"`},
		{"Boolean", `(?i)\b(true|false)\b`},
		{"Operator", `<>|!=|<=|>=|[-+*/%=<>()]`},
		{"Paren", `[()]`},
		{"whitespace", `\s+`},
	})

	parser = participle.MustBuild[TextExpression](
		participle.Lexer(cqlLexer),
		participle.Unquote("String"),
		participle.CaseInsensitive("Keyword", "Boolean"),
	)
)

type TextExpression struct {
	Or []*And `parser:"@@ ( ('OR' | 'or') @@ )*"`
}

func (e *TextExpression) ToAST() Expression {
	if len(e.Or) == 0 {
		return nil
	}
	result := e.Or[0].ToAST()
	for _, next := range e.Or[1:] {
		result = &LogicalOperator{
			Operator: "OR",
			Left:     result,
			Right:    next.ToAST(),
		}
	}
	return result
}

type And struct {
	Terms []*Term `parser:"@@ ( ('AND' | 'and') @@ )*"`
}

func (a *And) ToAST() Expression {
	if len(a.Terms) == 0 {
		return nil
	}
	result := a.Terms[0].ToAST()
	for _, next := range a.Terms[1:] {
		result = &LogicalOperator{
			Operator: "AND",
			Left:     result,
			Right:    next.ToAST(),
		}
	}
	return result
}

type Term struct {
	Not        *Term           `parser:"( ('NOT' | 'not') @@ )"`
	Group      *TextExpression `parser:"| '(' @@ ')' "`
	Comparison *TextComparison `parser:"| @@"`
}

func (t *Term) ToAST() Expression {
	if t.Not != nil {
		return &Not{Expression: t.Not.ToAST()}
	}
	if t.Group != nil {
		return t.Group.ToAST()
	}
	return t.Comparison.ToAST()
}

type TextComparison struct {
	Left  *Operand `parser:"@@"`
	Op    string   `parser:"@Operator"`
	Right *Operand `parser:"@@"`
}

func (c *TextComparison) ToAST() Expression {
	return &Comparison{
		Operator: c.Op,
		Left:     c.Left.ToExpr(),
		Right:    c.Right.ToExpr(),
	}
}

type Operand struct {
	Property *string      `parser:"@Ident"`
	Literal  *TextLiteral `parser:"| @@"`
}

type TextLiteral struct {
	Number  *float64 `parser:"  @Number"`
	String  *string  `parser:"| @String"`
	Boolean *bool    `parser:"| @Boolean"`
}

func (o *Operand) ToExpr() Expression {
	if o.Property != nil {
		return Property{Name: *o.Property}
	}
	return o.Literal.ToExpr()
}

func (l *TextLiteral) ToExpr() Expression {
	switch {
	case l.Number != nil:
		return Literal{Value: *l.Number}
	case l.String != nil:
		return Literal{Value: *l.String}
	case l.Boolean != nil:
		return Literal{Value: *l.Boolean}
	}
	return nil
}

func ParseText(input string) (Expression, error) {
	expr, err := parser.ParseString("", input)
	if err != nil {
		return nil, err
	}
	return expr.ToAST(), nil
}

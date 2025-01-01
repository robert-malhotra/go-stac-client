package filter

import (
	"fmt"
	"strings"
	"time"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/twpayne/go-geom"
)

// TextParser implements text format parsing using Participle
type TextParser struct {
	parser *participle.Parser[textExpr] // Specify the type parameter
}

// NewTextParser creates a new TextParser instance
func NewTextParser() (*TextParser, error) {
	// Define lexer rules
	textLexer := lexer.MustSimple([]lexer.SimpleRule{
		{Name: "whitespace", Pattern: `\s+`},
		{Name: "String", Pattern: `"(?:\\.|[^"])*"|'(?:\\.|[^'])*'`},
		{Name: "Number", Pattern: `[-+]?\d*\.?\d+([eE][-+]?\d+)?`},
		// Change the order and split operators for better matching
		{Name: "CompOp", Pattern: `<>|>=|<=|[=<>]`}, // Put compound operators first
		{Name: "Operator", Pattern: `(?i)(?:AND|OR|NOT|LIKE|IN|IS|NULL|BETWEEN|S_INTERSECTS|T_INTERSECTS)`},
		{Name: "Boolean", Pattern: `(?i)true|false`},
		{Name: "Null", Pattern: `(?i)null`},
		{Name: "Punct", Pattern: `[,()[\]/]`},
		{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
	})

	// Create parser with type parameter
	parser, err := participle.Build[textExpr](
		participle.Lexer(textLexer),
		participle.Unquote("String"),
		participle.CaseInsensitive("Operator", "Boolean", "Null"),
		participle.UseLookahead(2),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build parser: %w", err)
	}

	return &TextParser{parser: parser}, nil
}

// Text parsing AST structures
type textExpr struct {
	Expression *textExpression `@@`
}

type textExpression struct {
	LogicalExpr *textLogicalExpr `  @@`
	SimpleExpr  *textSimpleExpr  `| @@`
}

type textLogicalExpr struct {
	Operator string           `@Operator`
	Left     *textExpression  `"(" @@`
	Right    []textExpression `("," @@ )* ")"`
}

type textSimpleExpr struct {
	ComparisonExpr *textComparisonExpr `  @@`
	BetweenExpr    *textBetweenExpr    `| @@`
	LikeExpr       *textLikeExpr       `| @@`
	InExpr         *textInExpr         `| @@`
	NullExpr       *textNullExpr       `| @@`
	SpatialExpr    *textSpatialExpr    `| @@`
	TemporalExpr   *textTemporalExpr   `| @@`
}

type textComparisonExpr struct {
	Property string     `@Ident`
	Operator string     `@CompOp`
	Value    *textValue `@@`
}

type textBetweenExpr struct {
	Property string     `@Ident`
	Lower    *textValue `"BETWEEN" @@`
	Upper    *textValue `"AND" @@`
}

type textLikeExpr struct {
	Property string `@Ident`
	Pattern  string `"LIKE" @String`
}

type textInExpr struct {
	Property string       `@Ident`
	Values   []*textValue `"IN" "(" @@ ( "," @@ )* ")"`
}

type textNullExpr struct {
	Property string `@Ident "IS" "NULL"`
}

type textSpatialExpr struct {
	Property string     `@Ident`
	Point    *textPoint `"S_INTERSECTS" "POINT" @@`
}

type textPoint struct {
	X float64 `"(" @Number`
	Y float64 `@Number ")"`
}

type textTemporalExpr struct {
	Property string        `@Ident`
	Interval *textInterval `"T_INTERSECTS" "[" @@ "]"`
}

type textInterval struct {
	Start string `@String`
	End   string `"/" @String`
}

type textValue struct {
	String  *string  `  @String`
	Number  *float64 `| @Number`
	Boolean *bool    `| @Boolean`
	Null    bool     `| @Null`
}

// Parse implements the Parser interface for text format
func (p *TextParser) Parse(input string) (Expression, error) {
	ast, err := p.parser.ParseString("", input)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	return p.convertExpression(ast.Expression)
}

// Conversion methods
func (p *TextParser) convertExpression(expr *textExpression) (Expression, error) {
	if expr.LogicalExpr != nil {
		return p.convertLogical(expr.LogicalExpr)
	}
	if expr.SimpleExpr != nil {
		return p.convertSimple(expr.SimpleExpr)
	}
	return nil, fmt.Errorf("empty expression")
}

func (p *TextParser) convertLogical(expr *textLogicalExpr) (Expression, error) {
	op := Operator(strings.ToLower(expr.Operator))
	if !isLogicalOperator(op) {
		return nil, fmt.Errorf("invalid logical operator: %s", expr.Operator)
	}

	left, err := p.convertExpression(expr.Left)
	if err != nil {
		return nil, err
	}

	children := []Expression{left}
	for _, right := range expr.Right {
		rightExpr, err := p.convertExpression(&right)
		if err != nil {
			return nil, err
		}
		children = append(children, rightExpr)
	}

	return Logical{Op: op, Children: children}, nil
}

func (p *TextParser) convertSimple(expr *textSimpleExpr) (Expression, error) {
	switch {
	case expr.ComparisonExpr != nil:
		return p.convertComparison(expr.ComparisonExpr)
	case expr.BetweenExpr != nil:
		return p.convertBetween(expr.BetweenExpr)
	case expr.LikeExpr != nil:
		return p.convertLike(expr.LikeExpr)
	case expr.InExpr != nil:
		return p.convertIn(expr.InExpr)
	case expr.NullExpr != nil:
		return p.convertNull(expr.NullExpr)
	case expr.SpatialExpr != nil:
		return p.convertSpatial(expr.SpatialExpr)
	case expr.TemporalExpr != nil:
		return p.convertTemporal(expr.TemporalExpr)
	default:
		return nil, fmt.Errorf("empty simple expression")
	}
}

func (p *TextParser) convertComparison(expr *textComparisonExpr) (Expression, error) {
	value, err := p.convertValue(expr.Value)
	if err != nil {
		return nil, err
	}
	return Comparison{
		Op:       Operator(expr.Operator),
		Property: expr.Property,
		Value:    value,
	}, nil
}

func (p *TextParser) convertBetween(expr *textBetweenExpr) (Expression, error) {
	lower, err := p.convertValue(expr.Lower)
	if err != nil {
		return nil, err
	}

	upper, err := p.convertValue(expr.Upper)
	if err != nil {
		return nil, err
	}

	return Between{
		Property: expr.Property,
		Lower:    lower,
		Upper:    upper,
	}, nil
}

func (p *TextParser) convertLike(expr *textLikeExpr) (Expression, error) {
	return Like{
		Property: expr.Property,
		Pattern:  expr.Pattern,
	}, nil
}

func (p *TextParser) convertIn(expr *textInExpr) (Expression, error) {
	values := make([]interface{}, len(expr.Values))
	for i, v := range expr.Values {
		val, err := p.convertValue(v)
		if err != nil {
			return nil, err
		}
		values[i] = val
	}
	return In{
		Property: expr.Property,
		Values:   values,
	}, nil
}

func (p *TextParser) convertNull(expr *textNullExpr) (Expression, error) {
	return IsNull{Property: expr.Property}, nil
}

func (p *TextParser) convertSpatial(expr *textSpatialExpr) (Expression, error) {
	point := geom.NewPointFlat(geom.XY, []float64{expr.Point.X, expr.Point.Y})
	return SIntersects{
		Property: expr.Property,
		Geometry: point,
	}, nil
}

func (p *TextParser) convertTemporal(expr *textTemporalExpr) (Expression, error) {
	start, err := time.Parse(time.RFC3339, expr.Interval.Start)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %w", err)
	}

	end, err := time.Parse(time.RFC3339, expr.Interval.End)
	if err != nil {
		return nil, fmt.Errorf("invalid end time: %w", err)
	}

	return TIntersects{
		Property: expr.Property,
		Interval: TimeInterval{Start: start, End: end},
	}, nil
}

func (p *TextParser) convertValue(v *textValue) (interface{}, error) {
	if v.String != nil {
		return *v.String, nil
	}
	if v.Number != nil {
		return *v.Number, nil
	}
	if v.Boolean != nil {
		return *v.Boolean, nil
	}
	if v.Null {
		return nil, nil
	}
	return nil, fmt.Errorf("invalid value")
}

// Helper functions
func isLogicalOperator(op Operator) bool {
	switch op {
	case OpAnd, OpOr, OpNot:
		return true
	default:
		return false
	}
}

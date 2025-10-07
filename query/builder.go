package query

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	ogcfilter "github.com/planetlabs/go-ogc/filter"
)

// Builder accumulates filter expressions in a fluent manner.
type Builder struct {
	expr ogcfilter.BooleanExpression
}

// NewBuilder returns an empty Builder instance.
func NewBuilder() *Builder {
	return &Builder{}
}

// Where sets the expression if none exists or ANDs it with the current expression.
func (b *Builder) Where(expr ogcfilter.Expression) *Builder {
	be := toBooleanExpression(expr)
	if be == nil {
		return b
	}
	if b.expr == nil {
		b.expr = be
		return b
	}
	b.expr = &ogcfilter.And{Args: []ogcfilter.BooleanExpression{b.expr, be}}
	return b
}

// And adds multiple expressions combined with logical AND.
func (b *Builder) And(exprs ...ogcfilter.Expression) *Builder {
	args := make([]ogcfilter.BooleanExpression, 0, len(exprs)+1)
	if b.expr != nil {
		args = append(args, b.expr)
	}
	for _, expr := range exprs {
		if be := toBooleanExpression(expr); be != nil {
			args = append(args, be)
		}
	}
	if len(args) == 0 {
		return b
	}
	b.expr = &ogcfilter.And{Args: args}
	return b
}

// Or combines the current expression with the provided ones using logical OR.
func (b *Builder) Or(exprs ...ogcfilter.Expression) *Builder {
	if b.expr == nil && len(exprs) == 0 {
		return b
	}
	args := make([]ogcfilter.BooleanExpression, 0, len(exprs)+1)
	if b.expr != nil {
		args = append(args, b.expr)
	}
	for _, expr := range exprs {
		if be := toBooleanExpression(expr); be != nil {
			args = append(args, be)
		}
	}
	if len(args) == 0 {
		return b
	}
	b.expr = &ogcfilter.Or{Args: args}
	return b
}

// Not negates the current expression.
func (b *Builder) Not() *Builder {
	if b.expr == nil {
		return b
	}
	b.expr = &ogcfilter.Not{Arg: b.expr}
	return b
}

// Filter returns the built expression.
func (b *Builder) Filter() ogcfilter.BooleanExpression {
	return b.expr
}

// Must returns the built expression or panics if it is empty.
func (b *Builder) Must() ogcfilter.BooleanExpression {
	if b.expr == nil {
		panic("query builder: expression is empty")
	}
	return b.expr
}

// Property constructs a property expression builder.
func Property(name string) PropertyExpression {
	return PropertyExpression{property: &ogcfilter.Property{Name: name}}
}

// PropertyExpression exposes fluent helpers for comparisons.
type PropertyExpression struct {
	property *ogcfilter.Property
}

// Eq creates an equality predicate. Nil values generate an isNull expression.
func (p PropertyExpression) Eq(value any) ogcfilter.BooleanExpression {
	if value == nil {
		return &ogcfilter.IsNull{Value: p.property}
	}
	return &ogcfilter.Comparison{
		Name:  ogcfilter.Equals,
		Left:  p.property,
		Right: toScalarExpression(value),
	}
}

// Neq creates an inequality predicate. Nil values generate a negated isNull expression.
func (p PropertyExpression) Neq(value any) ogcfilter.BooleanExpression {
	if value == nil {
		return &ogcfilter.Not{Arg: &ogcfilter.IsNull{Value: p.property}}
	}
	return &ogcfilter.Comparison{
		Name:  ogcfilter.NotEquals,
		Left:  p.property,
		Right: toScalarExpression(value),
	}
}

// Lt creates a less-than predicate.
func (p PropertyExpression) Lt(value any) ogcfilter.BooleanExpression {
	return &ogcfilter.Comparison{
		Name:  ogcfilter.LessThan,
		Left:  p.property,
		Right: toScalarExpression(value),
	}
}

// Lte creates a less-than-or-equal predicate.
func (p PropertyExpression) Lte(value any) ogcfilter.BooleanExpression {
	return &ogcfilter.Comparison{
		Name:  ogcfilter.LessThanOrEquals,
		Left:  p.property,
		Right: toScalarExpression(value),
	}
}

// Gt creates a greater-than predicate.
func (p PropertyExpression) Gt(value any) ogcfilter.BooleanExpression {
	return &ogcfilter.Comparison{
		Name:  ogcfilter.GreaterThan,
		Left:  p.property,
		Right: toScalarExpression(value),
	}
}

// Gte creates a greater-than-or-equal predicate.
func (p PropertyExpression) Gte(value any) ogcfilter.BooleanExpression {
	return &ogcfilter.Comparison{
		Name:  ogcfilter.GreaterThanOrEquals,
		Left:  p.property,
		Right: toScalarExpression(value),
	}
}

// Like creates a pattern match predicate.
func (p PropertyExpression) Like(pattern string) ogcfilter.BooleanExpression {
	return &ogcfilter.Like{
		Value:   p.property,
		Pattern: &ogcfilter.String{Value: pattern},
	}
}

// In creates a set membership predicate.
func (p PropertyExpression) In(values ...any) ogcfilter.BooleanExpression {
	list := make([]ogcfilter.ScalarExpression, 0, len(values))
	if len(values) == 1 {
		if slice, ok := maybeSlice(values[0]); ok {
			values = slice
		}
	}
	for _, v := range values {
		expr := toScalarExpression(v)
		if expr == nil {
			continue
		}
		list = append(list, expr)
	}
	return &ogcfilter.In{
		Item: p.property,
		List: ogcfilter.ScalarList(list),
	}
}

// Between constrains the property between the provided numeric bounds.
func (p PropertyExpression) Between(low, high any) ogcfilter.BooleanExpression {
	lowExpr := toNumericExpression(low)
	highExpr := toNumericExpression(high)
	return &ogcfilter.Between{
		Value: p.property,
		Low:   lowExpr,
		High:  highExpr,
	}
}

// IsNull creates an isNull predicate for the property.
func (p PropertyExpression) IsNull() ogcfilter.BooleanExpression {
	return &ogcfilter.IsNull{Value: p.property}
}

// IsNotNull creates a negated isNull predicate for the property.
func (p PropertyExpression) IsNotNull() ogcfilter.BooleanExpression {
	return &ogcfilter.Not{Arg: &ogcfilter.IsNull{Value: p.property}}
}

// BBox builds a spatial intersects expression for the geometry property.
func BBox(minLon, minLat, maxLon, maxLat float64) ogcfilter.BooleanExpression {
	return &ogcfilter.SpatialComparison{
		Name:  ogcfilter.GeometryIntersects,
		Left:  &ogcfilter.Property{Name: "geometry"},
		Right: &ogcfilter.BoundingBox{Extent: []float64{minLon, minLat, maxLon, maxLat}},
	}
}

// Datetime builds a temporal intersects expression on the datetime property.
func Datetime(start, end time.Time) ogcfilter.BooleanExpression {
	return Between("datetime", start, end)
}

// Between constrains a temporal property between the provided instants (inclusive).
func Between(property string, start, end time.Time) ogcfilter.BooleanExpression {
	start, end = normalizeTimes(start, end)
	return &ogcfilter.TemporalComparison{
		Name: ogcfilter.TimeIntersects,
		Left: &ogcfilter.Property{Name: property},
		Right: &ogcfilter.Interval{
			Start: &ogcfilter.Timestamp{Value: start},
			End:   &ogcfilter.Timestamp{Value: end},
		},
	}
}

// Raw wraps a pre-built structure as a boolean expression. It panics if the value
// cannot be encoded or decoded into a valid filter expression.
func Raw(value any) ogcfilter.BooleanExpression {
	data, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Errorf("query.Raw: %w", err))
	}
	var filter ogcfilter.Filter
	if err := json.Unmarshal(data, &filter); err != nil {
		panic(fmt.Errorf("query.Raw: %w", err))
	}
	return filter.Expression
}

func toBooleanExpression(expr ogcfilter.Expression) ogcfilter.BooleanExpression {
	if expr == nil {
		return nil
	}
	be, ok := expr.(ogcfilter.BooleanExpression)
	if !ok {
		panic("query builder: expression must be boolean")
	}
	return be
}

func toScalarExpression(value any) ogcfilter.ScalarExpression {
	switch v := value.(type) {
	case nil:
		return nil
	case ogcfilter.ScalarExpression:
		return v
	case PropertyExpression:
		return v.property
	case *ogcfilter.Property:
		return v
	case string:
		return &ogcfilter.String{Value: v}
	case fmt.Stringer:
		return &ogcfilter.String{Value: v.String()}
	case bool:
		return &ogcfilter.Boolean{Value: v}
	case int:
		return &ogcfilter.Number{Value: float64(v)}
	case int8:
		return &ogcfilter.Number{Value: float64(v)}
	case int16:
		return &ogcfilter.Number{Value: float64(v)}
	case int32:
		return &ogcfilter.Number{Value: float64(v)}
	case int64:
		return &ogcfilter.Number{Value: float64(v)}
	case uint:
		return &ogcfilter.Number{Value: float64(v)}
	case uint8:
		return &ogcfilter.Number{Value: float64(v)}
	case uint16:
		return &ogcfilter.Number{Value: float64(v)}
	case uint32:
		return &ogcfilter.Number{Value: float64(v)}
	case uint64:
		return &ogcfilter.Number{Value: float64(v)}
	case float32:
		return &ogcfilter.Number{Value: float64(v)}
	case float64:
		return &ogcfilter.Number{Value: v}
	case time.Time:
		return &ogcfilter.Timestamp{Value: v}
	case []any:
		list := make([]ogcfilter.ScalarExpression, 0, len(v))
		for _, item := range v {
			if expr := toScalarExpression(item); expr != nil {
				list = append(list, expr)
			}
		}
		return ogcfilter.ScalarList(list)
	case []string:
		list := make([]ogcfilter.ScalarExpression, len(v))
		for i, s := range v {
			list[i] = &ogcfilter.String{Value: s}
		}
		return ogcfilter.ScalarList(list)
	case []int:
		list := make([]ogcfilter.ScalarExpression, len(v))
		for i, n := range v {
			list[i] = &ogcfilter.Number{Value: float64(n)}
		}
		return ogcfilter.ScalarList(list)
	case []float64:
		list := make([]ogcfilter.ScalarExpression, len(v))
		for i, n := range v {
			list[i] = &ogcfilter.Number{Value: n}
		}
		return ogcfilter.ScalarList(list)
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice {
			length := rv.Len()
			list := make([]ogcfilter.ScalarExpression, 0, length)
			for i := 0; i < length; i++ {
				if expr := toScalarExpression(rv.Index(i).Interface()); expr != nil {
					list = append(list, expr)
				}
			}
			return ogcfilter.ScalarList(list)
		}
		return &ogcfilter.String{Value: fmt.Sprint(value)}
	}
}

func toNumericExpression(value any) ogcfilter.NumericExpression {
	expr := toScalarExpression(value)
	if expr == nil {
		return nil
	}
	numeric, ok := expr.(ogcfilter.NumericExpression)
	if !ok {
		panic("query builder: expected numeric value")
	}
	return numeric
}

func normalizeTimes(start, end time.Time) (time.Time, time.Time) {
	if end.IsZero() {
		end = start
	}
	if start.IsZero() {
		start = end
	}
	if end.Before(start) {
		start, end = end, start
	}
	return start.UTC(), end.UTC()
}

func maybeSlice(value any) ([]any, bool) {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return nil, false
	}
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}
	length := rv.Len()
	out := make([]any, 0, length)
	for i := 0; i < length; i++ {
		out = append(out, rv.Index(i).Interface())
	}
	return out, true
}

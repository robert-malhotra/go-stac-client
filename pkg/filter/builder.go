// pkg/filter/builder.go

package filter

// Builder helps in constructing CQL2 expressions
type Builder struct {
	expr Expression
}

// NewBuilder initializes a new Builder
func NewBuilder() *Builder {
	return &Builder{}
}

// And adds an And expression
func (b *Builder) And(expressions ...Expression) *Builder {
	if b.expr == nil {
		b.expr = And{Children: expressions}
	} else {
		if _, ok := b.expr.(And); !ok {
			b.expr = And{Children: []Expression{b.expr}}
		}
		b.expr = And{Children: append(getChildren(b.expr), expressions...)}
	}
	return b
}

// Or adds an Or expression
func (b *Builder) Or(expressions ...Expression) *Builder {
	if b.expr == nil {
		b.expr = Or{Children: expressions}
	} else {
		if _, ok := b.expr.(Or); !ok {
			b.expr = Or{Children: []Expression{b.expr}}
		}
		b.expr = Or{Children: append(getChildren(b.expr), expressions...)}
	}
	return b
}

// Not adds a Not expression
func (b *Builder) Not(expression Expression) *Builder {
	b.expr = Not{Child: expression}
	return b
}

// PropertyIsEqualTo adds a PropertyIsEqualTo expression
func (b *Builder) PropertyIsEqualTo(property string, value interface{}) *Builder {
	expr := PropertyIsEqualTo{
		Property: property,
		Value:    value,
	}
	if b.expr == nil {
		b.expr = expr
	} else {
		b.And(expr)
	}
	return b
}

// PropertyIsNotEqualTo adds a PropertyIsNotEqualTo expression
func (b *Builder) PropertyIsNotEqualTo(property string, value interface{}) *Builder {
	expr := PropertyIsNotEqualTo{
		Property: property,
		Value:    value,
	}
	if b.expr == nil {
		b.expr = expr
	} else {
		b.And(expr)
	}
	return b
}

// PropertyIsLessThan adds a PropertyIsLessThan expression
func (b *Builder) PropertyIsLessThan(property string, value interface{}) *Builder {
	expr := PropertyIsLessThan{
		Property: property,
		Value:    value,
	}
	if b.expr == nil {
		b.expr = expr
	} else {
		b.And(expr)
	}
	return b
}

// PropertyIsLessThanOrEqualTo adds a PropertyIsLessThanOrEqualTo expression
func (b *Builder) PropertyIsLessThanOrEqualTo(property string, value interface{}) *Builder {
	expr := PropertyIsLessThanOrEqualTo{
		Property: property,
		Value:    value,
	}
	if b.expr == nil {
		b.expr = expr
	} else {
		b.And(expr)
	}
	return b
}

// PropertyIsGreaterThan adds a PropertyIsGreaterThan expression
func (b *Builder) PropertyIsGreaterThan(property string, value interface{}) *Builder {
	expr := PropertyIsGreaterThan{
		Property: property,
		Value:    value,
	}
	if b.expr == nil {
		b.expr = expr
	} else {
		b.And(expr)
	}
	return b
}

// PropertyIsGreaterThanOrEqualTo adds a PropertyIsGreaterThanOrEqualTo expression
func (b *Builder) PropertyIsGreaterThanOrEqualTo(property string, value interface{}) *Builder {
	expr := PropertyIsGreaterThanOrEqualTo{
		Property: property,
		Value:    value,
	}
	if b.expr == nil {
		b.expr = expr
	} else {
		b.And(expr)
	}
	return b
}

// Between adds a Between expression
func (b *Builder) Between(property string, lower, upper interface{}) *Builder {
	expr := Between{
		Property: property,
		Lower:    lower,
		Upper:    upper,
	}
	if b.expr == nil {
		b.expr = expr
	} else {
		b.And(expr)
	}
	return b
}

// Like adds a Like expression
func (b *Builder) Like(property, pattern string) *Builder {
	expr := Like{
		Property: property,
		Pattern:  pattern,
	}
	if b.expr == nil {
		b.expr = expr
	} else {
		b.And(expr)
	}
	return b
}

// In adds an In expression
func (b *Builder) In(property string, values []interface{}) *Builder {
	expr := In{
		Property: property,
		Values:   values,
	}
	if b.expr == nil {
		b.expr = expr
	} else {
		b.And(expr)
	}
	return b
}

// Function adds a Function expression
func (b *Builder) Function(name string, args ...interface{}) *Builder {
	expr := Function{
		Name: name,
		Args: args,
	}
	if b.expr == nil {
		b.expr = expr
	} else {
		b.And(expr)
	}
	return b
}

// SIntersects adds an SIntersects expression
func (b *Builder) SIntersects(property string, geometry GeoJSONGeometry) *Builder {
	expr := SIntersects{
		Property: property,
		Geometry: geometry,
	}
	if b.expr == nil {
		b.expr = expr
	} else {
		b.And(expr)
	}
	return b
}

// TIntersects adds a TIntersects expression
func (b *Builder) TIntersects(property string, interval TimeInterval) *Builder {
	expr := TIntersects{
		Property: property,
		Interval: interval,
	}
	if b.expr == nil {
		b.expr = expr
	} else {
		b.And(expr)
	}
	return b
}

// PropertyPropertyComparison adds a PropertyPropertyComparison expression
func (b *Builder) PropertyPropertyComparison(property1, operator, property2 string) *Builder {
	expr := PropertyPropertyComparison{
		Property1: property1,
		Operator:  operator,
		Property2: property2,
	}
	if b.expr == nil {
		b.expr = expr
	} else {
		b.And(expr)
	}
	return b
}

// IsNull adds an IsNull expression
func (b *Builder) IsNull(property string) *Builder {
	expr := IsNull{
		Property: property,
	}
	if b.expr == nil {
		b.expr = expr
	} else {
		b.And(expr)
	}
	return b
}

// Build returns the constructed Expression
func (b *Builder) Build() Expression {
	return b.expr
}

// Helper to extract children from And or Or expressions
func getChildren(expr Expression) []Expression {
	switch e := expr.(type) {
	case And:
		return e.Children
	case Or:
		return e.Children
	default:
		return []Expression{expr}
	}
}

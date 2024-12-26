// pkg/filter/serializer.go

package filter

import (
	"encoding/json"
	"fmt"
	"time"
)

// SerializeExpression serializes an Expression into JSON
func SerializeExpression(expr Expression) ([]byte, error) {
	exprMap, err := expressionToMap(expr)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(exprMap, "", "  ")
}

func expressionToMap(expr Expression) (map[string]interface{}, error) {
	switch e := expr.(type) {
	case And:
		args := []interface{}{}
		for _, child := range e.Children {
			childMap, err := expressionToMap(child)
			if err != nil {
				return nil, err
			}
			args = append(args, childMap)
		}
		return map[string]interface{}{
			"op":   e.ExpressionType(),
			"args": args,
		}, nil

	case Or:
		args := []interface{}{}
		for _, child := range e.Children {
			childMap, err := expressionToMap(child)
			if err != nil {
				return nil, err
			}
			args = append(args, childMap)
		}
		return map[string]interface{}{
			"op":   e.ExpressionType(),
			"args": args,
		}, nil

	case Not:
		childMap, err := expressionToMap(e.Child)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"op":   e.ExpressionType(),
			"args": []interface{}{childMap},
		}, nil

	case PropertyIsEqualTo:
		return map[string]interface{}{
			"op":   e.ExpressionType(),
			"args": []interface{}{map[string]interface{}{"property": e.Property}, e.Value},
		}, nil

	case PropertyIsNotEqualTo:
		return map[string]interface{}{
			"op":   e.ExpressionType(),
			"args": []interface{}{map[string]interface{}{"property": e.Property}, e.Value},
		}, nil

	case PropertyIsLessThan:
		return map[string]interface{}{
			"op":   e.ExpressionType(),
			"args": []interface{}{map[string]interface{}{"property": e.Property}, e.Value},
		}, nil

	case PropertyIsLessThanOrEqualTo:
		return map[string]interface{}{
			"op":   e.ExpressionType(),
			"args": []interface{}{map[string]interface{}{"property": e.Property}, e.Value},
		}, nil

	case PropertyIsGreaterThan:
		return map[string]interface{}{
			"op":   e.ExpressionType(),
			"args": []interface{}{map[string]interface{}{"property": e.Property}, e.Value},
		}, nil

	case PropertyIsGreaterThanOrEqualTo:
		return map[string]interface{}{
			"op":   e.ExpressionType(),
			"args": []interface{}{map[string]interface{}{"property": e.Property}, e.Value},
		}, nil

	case Between:
		return map[string]interface{}{
			"op":   e.ExpressionType(),
			"args": []interface{}{map[string]interface{}{"property": e.Property}, e.Lower, e.Upper},
		}, nil

	case Like:
		return map[string]interface{}{
			"op":   e.ExpressionType(),
			"args": []interface{}{map[string]interface{}{"property": e.Property}, e.Pattern},
		}, nil

	case In:
		return map[string]interface{}{
			"op":   e.ExpressionType(),
			"args": []interface{}{map[string]interface{}{"property": e.Property}, e.Values},
		}, nil

	case Function:
		args := []interface{}{}
		for _, arg := range e.Args {
			switch a := arg.(type) {
			case Expression:
				argMap, err := expressionToMap(a)
				if err != nil {
					return nil, err
				}
				args = append(args, argMap)
			default:
				args = append(args, a)
			}
		}
		return map[string]interface{}{
			"op":   e.Name,
			"args": args,
		}, nil

	case SIntersects:
		geometryBytes, err := json.Marshal(e.Geometry)
		if err != nil {
			return nil, err
		}
		var geometry interface{}
		if err := json.Unmarshal(geometryBytes, &geometry); err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"op":   e.ExpressionType(),
			"args": []interface{}{map[string]interface{}{"property": e.Property}, geometry},
		}, nil

	case TIntersects:
		interval := []string{
			e.Interval.Start.Format(time.RFC3339),
			e.Interval.End.Format(time.RFC3339),
		}
		return map[string]interface{}{
			"op":   e.ExpressionType(),
			"args": []interface{}{map[string]interface{}{"property": e.Property}, map[string]interface{}{"interval": interval}},
		}, nil

	case PropertyPropertyComparison:
		return map[string]interface{}{
			"op":   e.Operator,
			"args": []interface{}{map[string]interface{}{"property": e.Property1}, map[string]interface{}{"property": e.Property2}},
		}, nil

	case IsNull:
		return map[string]interface{}{
			"op":   e.ExpressionType(),
			"args": []interface{}{map[string]interface{}{"property": e.Property}},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

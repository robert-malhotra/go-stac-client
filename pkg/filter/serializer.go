// pkg/filter/serializer.go

package filter

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/twpayne/go-geom/encoding/geojson"
)

func SerializeExpression(expr Expression) ([]byte, error) {
	wrapper, err := expressionToWrapper(expr)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(wrapper, "", "  ")
}

func expressionToWrapper(expr Expression) (map[string]interface{}, error) {
	switch e := expr.(type) {
	case Logical:
		args := make([]interface{}, len(e.Children))
		for i, child := range e.Children {
			childWrapper, err := expressionToWrapper(child)
			if err != nil {
				return nil, err
			}
			args[i] = childWrapper
		}
		return map[string]interface{}{
			"op":   string(e.Op),
			"args": args,
		}, nil

	case Comparison:
		return map[string]interface{}{
			"op": string(e.Op),
			"args": []interface{}{
				map[string]interface{}{"property": e.Property},
				e.Value,
			},
		}, nil

	case Between:
		return map[string]interface{}{
			"op": "between",
			"args": []interface{}{
				map[string]interface{}{"property": e.Property},
				e.Lower,
				e.Upper,
			},
		}, nil

	case Like:
		return map[string]interface{}{
			"op": "like",
			"args": []interface{}{
				map[string]interface{}{"property": e.Property},
				e.Pattern,
			},
		}, nil

	case In:
		return map[string]interface{}{
			"op": "in",
			"args": []interface{}{
				map[string]interface{}{"property": e.Property},
				e.Values,
			},
		}, nil

	case IsNull:
		return map[string]interface{}{
			"op": "isNull",
			"args": []interface{}{
				map[string]interface{}{"property": e.Property},
			},
		}, nil

	// case Function:
	// 	args := make([]interface{}, len(e.Args))
	// 	for i, arg := range e.Args {
	// 		switch a := arg.(type) {
	// 		case Expression:
	// 			wrapper, err := expressionToWrapper(a)
	// 			if err != nil {
	// 				return nil, err
	// 			}
	// 			args[i] = wrapper
	// 		default:
	// 			args[i] = a
	// 		}
	// 	}
	// 	return map[string]interface{}{
	// 		"op":   e.Name,
	// 		"args": args,
	// 	}, nil

	case SIntersects:
		geometryBytes, err := geojson.Marshal(e.Geometry)
		if err != nil {
			return nil, err
		}
		var geometry interface{}
		if err := json.Unmarshal(geometryBytes, &geometry); err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"op": "s_intersects",
			"args": []interface{}{
				map[string]interface{}{"property": e.Property},
				geometry,
			},
		}, nil

	case TIntersects:
		return map[string]interface{}{
			"op": "t_intersects",
			"args": []interface{}{
				map[string]interface{}{"property": e.Property},
				map[string]interface{}{
					"interval": []string{
						e.Interval.Start.Format(time.RFC3339),
						e.Interval.End.Format(time.RFC3339),
					},
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

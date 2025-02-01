package cql2

import (
	"fmt"
	"strconv"
	"strings"
)

func SerializeText(expr Expression) (string, error) {
	if expr == nil {
		return "", fmt.Errorf("cannot serialize nil expression")
	}
	return serialize(expr, 0)
}

func serialize(expr Expression, parentPrecedence int) (string, error) {
	switch e := expr.(type) {
	case Property:
		return e.Name, nil

	case Literal:
		return serializeLiteral(e.Value)

	case Comparison:
		left, err := serialize(e.Left, 0)
		if err != nil {
			return "", err
		}
		right, err := serialize(e.Right, 0)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s %s %s", left, e.Operator, right), nil

	case LogicalOperator:
		currentPrecedence := getPrecedence(e.Operator)
		left, err := serialize(e.Left, currentPrecedence)
		if err != nil {
			return "", err
		}
		right, err := serialize(e.Right, currentPrecedence)
		if err != nil {
			return "", err
		}

		result := fmt.Sprintf("%s %s %s", left, strings.ToUpper(string(e.Operator)), right)

		// Add parentheses if nested within another operator
		if parentPrecedence > 0 {
			result = fmt.Sprintf("(%s)", result)
		}
		return result, nil

	case Not:
		inner, err := serialize(e.Expression, 3) // NOT has highest precedence
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("NOT %s", inner), nil

	default:
		return "", fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func serializeLiteral(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, v), nil
	case bool:
		return strings.ToUpper(strconv.FormatBool(v)), nil
	case int, int32, int64, uint, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%g", v), nil
	default:
		return "", fmt.Errorf("unsupported literal type: %T", value)
	}
}

func getPrecedence(op Operator) int {
	switch op {
	case "NOT":
		return 3
	case "AND":
		return 2
	case "OR":
		return 1
	default:
		return 0
	}
}

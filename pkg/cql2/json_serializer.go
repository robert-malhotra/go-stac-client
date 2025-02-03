package cql2

import (
	"encoding/json"
	"fmt"
)

type jsonFormat struct {
	OP   Operator      `json:"op"`
	Args []interface{} `json:"args"`
}

func (c Comparison) MarshalJSON() ([]byte, error) {
	// When serializing the left operand (a property) wrap it as an object.
	return json.Marshal(
		jsonFormat{
			OP: c.Operator,
			Args: []interface{}{
				struct {
					Property string `json:"property"`
				}{Property: c.Left},
				c.Right,
			},
		})
}

func (lo LogicalOperator) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		jsonFormat{
			OP:   lo.Operator,
			Args: []interface{}{lo.Left, lo.Right},
		})
}

func (n Not) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		jsonFormat{
			OP:   OpNot,
			Args: []interface{}{n.Expression},
		})
}

// SerializeJSON serializes an expression to JSON.
func SerializeJSON(expr Expression) ([]byte, error) {
	if expr == nil {
		return nil, fmt.Errorf("cannot serialize nil expression")
	}
	return json.Marshal(expr)
}

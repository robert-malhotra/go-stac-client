package formatting

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func FormatGeometry(geometry interface{}) string {
	if geometry == nil {
		return ""
	}

	raw, err := json.Marshal(geometry)
	if err != nil {
		return fmt.Sprintf("%v", geometry)
	}

	var geo map[string]interface{}
	if err := json.Unmarshal(raw, &geo); err != nil {
		return string(raw)
	}

	var sections []string

	if typ, ok := geo["type"].(string); ok {
		sections = append(sections, typ)

		if typ == "GeometryCollection" {
			if geoms, ok := geo["geometries"].([]interface{}); ok {
				var summaries []string
				for _, g := range geoms {
					if s := FormatGeometry(g); s != "" {
						summaries = append(summaries, s)
					}
				}
				if len(summaries) > 0 {
					sections = append(sections, strings.Join(summaries, " | "))
				}
			}
		}
	}

	if coords, ok := geo["coordinates"]; ok {
		if coordStr := formatCoordinateValue(coords, 0); coordStr != "" {
			sections = append(sections, wrapCoordinateString(coordStr, 70))
		}
	}

	if bbox, ok := geo["bbox"]; ok {
		if bboxStr := formatCoordinateValue(bbox, 0); bboxStr != "" {
			sections = append(sections, "bbox "+wrapCoordinateString(bboxStr, 70))
		}
	}

	if len(sections) == 0 {
		return string(raw)
	}

	return strings.Join(sections, "\n")
}

func formatCoordinateValue(value interface{}, depth int) string {
	switch v := value.(type) {
	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		parts := make([]string, len(v))
		for i, elem := range v {
			parts[i] = formatCoordinateValue(elem, depth+1)
		}
		sep := ", "
		if depth >= 2 {
			sep = " "
		}
		return "[" + strings.Join(parts, sep) + "]"
	case float64:
		return strconv.FormatFloat(v, 'f', 5, 64)
	case json.Number:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func wrapCoordinateString(s string, width int) string {
	if len(s) <= width || width <= 0 {
		return s
	}

	var out strings.Builder
	lineLen := 0

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if lineLen >= width && (ch == ',' || ch == ']' || ch == ' ') {
			out.WriteByte('\n')
			lineLen = 0
			if ch == ' ' {
				continue
			}
		}

		out.WriteByte(ch)
		lineLen++
	}

	return out.String()
}

package formatting

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	stac "github.com/planetlabs/go-stac"
)

func FormatItemSummary(item *stac.Item) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("[yellow]ID: [white]%s\n", item.Id))
	if dt, ok := item.Properties["datetime"].(string); ok {
		builder.WriteString(fmt.Sprintf("[yellow]Datetime: [white]%s\n", dt))
	}
	if p, ok := item.Properties["platform"].(string); ok {
		builder.WriteString(fmt.Sprintf("[yellow]Platform: [white]%s\n", p))
	}
	if c, ok := item.Properties["constellation"].(string); ok {
		builder.WriteString(fmt.Sprintf("[yellow]Constellation: [white]%s\n", c))
	}
	if geomText := FormatGeometry(item.Geometry); geomText != "" {
		builder.WriteString("[yellow]Geometry:[white]\n")
		builder.WriteString(geomText)
		if !strings.HasSuffix(geomText, "\n") {
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

func FormatProperties(properties map[string]interface{}, indent int) string {
	var builder strings.Builder
	keys := make([]string, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		val := properties[key]
		indentedKey := fmt.Sprintf("%s%s:", strings.Repeat("  ", indent), key)
		builder.WriteString(fmt.Sprintf("[yellow]%-30s[white]", indentedKey))

		jsonBytes, err := json.MarshalIndent(val, "", "  ")
		if err != nil {
			builder.WriteString(" Error marshalling value\n")
		} else {
			builder.WriteString(fmt.Sprintf(" %s\n", string(jsonBytes)))
		}
	}
	return builder.String()
}

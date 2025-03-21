package agent

import "strings"

// ExtractJSONFromResponse finds and extracts JSON content from a string by removing the markup
func ExtractJSONFromResponse(s string) string {
	var final string

	final = strings.ReplaceAll(s, "\n", "")
	final = strings.ReplaceAll(final, "```json", "")
	final = strings.ReplaceAll(final, "```", "")

	return final
}

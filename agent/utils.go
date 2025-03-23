package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

// ParseJSONFromResponse parses a JSON response from the LLM. The response will be
// stripped of any newline characters and any lines that start with "
func ParseJSONFromResponse[T any](s string) (T, error) {
	var extractedJSON string

	extractedJSON = strings.ReplaceAll(s, "\n", "")
	extractedJSON = strings.ReplaceAll(extractedJSON, "```json", "")
	extractedJSON = strings.ReplaceAll(extractedJSON, "```", "")

	var parsedResponse T
	err := json.Unmarshal([]byte(extractedJSON), &parsedResponse)
	if err != nil {
		log.Debug().Msgf("Raw JSON: %s", extractedJSON)
		return parsedResponse, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return parsedResponse, nil
}

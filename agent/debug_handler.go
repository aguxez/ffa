package agent

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// DebugHandler implements the callbacks.Handler interface for debugging
type DebugHandler struct {
	Verbose bool
}

// NewDebugHandler creates a new debugging handler
func NewDebugHandler(verbose bool) *DebugHandler {
	return &DebugHandler{
		Verbose: verbose,
	}
}

// logMessage logs a message to stdout and/or file
func (h *DebugHandler) logMessage(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	// Always log to console
	log.Printf("[AGENT-DEBUG] %s", msg)
}

// HandleText handles debugging text
func (h *DebugHandler) HandleText(ctx context.Context, text string) {
	if h.Verbose {
		h.logMessage("Text: %s", text)
	}
}

// HandleLLMStart logs when LLM starts processing
func (h *DebugHandler) HandleLLMStart(ctx context.Context, prompts []string) {
	if h.Verbose {
		h.logMessage("LLM Start - Prompts: %d", len(prompts))
		for i, prompt := range prompts {
			if len(prompt) > 100 {
				h.logMessage("Prompt %d (truncated): %s...", i, prompt[:100])
			} else {
				h.logMessage("Prompt %d: %s", i, prompt)
			}
		}
	}
}

// HandleLLMError logs LLM errors
func (h *DebugHandler) HandleLLMError(ctx context.Context, err error) {
	h.logMessage("LLM Error: %v", err)
}

// HandleChainStart logs when a chain starts
func (h *DebugHandler) HandleChainStart(ctx context.Context, inputs map[string]any) {
	if h.Verbose {
		h.logMessage("Chain Start - Inputs: %v", summarizeMap(inputs))
	}
}

// HandleChainEnd logs when a chain ends
func (h *DebugHandler) HandleChainEnd(ctx context.Context, outputs map[string]any) {
	h.logMessage("Chain End - Outputs: %v", summarizeMap(outputs))
}

// HandleChainError logs chain errors
func (h *DebugHandler) HandleChainError(ctx context.Context, err error) {
	h.logMessage("Chain Error: %v", err)
}

// HandleToolStart logs when a tool starts
func (h *DebugHandler) HandleToolStart(ctx context.Context, input string) {
	h.logMessage("Tool Start - Input: %s", truncateString(input, 100))
}

// HandleToolEnd logs when a tool ends
func (h *DebugHandler) HandleToolEnd(ctx context.Context, output string) {
	h.logMessage("Tool End - Output: %s", truncateString(output, 100))
}

// HandleToolError logs tool errors
func (h *DebugHandler) HandleToolError(ctx context.Context, err error) {
	h.logMessage("Tool Error: %v", err)
}

// HandleAgentAction logs agent actions - CRITICAL for debugging agents
func (h *DebugHandler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {
	h.logMessage("Agent Action - Tool: %s, Input: %s",
		action.Tool, truncateString(action.ToolInput, 100))
}

// HandleAgentFinish logs when an agent finishes
func (h *DebugHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {
	h.logMessage("Agent Finish - Output: %s",
		truncateString(finish.ReturnValues["output"].(string), 200))
}

// Implement the rest of the interface methods
func (h *DebugHandler) HandleRetrieverStart(ctx context.Context, query string) {}

func (h *DebugHandler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
}
func (h *DebugHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {}

// HandleLLMGenerateContentStart logs when LLM content generation starts
func (h *DebugHandler) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	// Implementation
}

// HandleLLMGenerateContentEnd logs when LLM content generation ends
func (h *DebugHandler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	// Implementation
}

// Helper functions
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func summarizeMap(m map[string]any) string {
	parts := []string{}
	for k, v := range m {
		var valueStr string
		switch val := v.(type) {
		case string:
			valueStr = truncateString(val, 50)
		default:
			valueStr = fmt.Sprintf("%v", v)
		}
		parts = append(parts, fmt.Sprintf("%s: %s", k, valueStr))
	}
	return strings.Join(parts, ", ")
}

package agent

type LLMModel string

const (
	LLMModelClaude LLMModel = "anthropic/claude-3.7-sonnet"
	LLMModelGemini LLMModel = "google/gemini-2.0-flash-lite-001"
)

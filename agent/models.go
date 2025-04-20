package agent

type LLMModel string

const (
	LLMModelClaude             LLMModel = "anthropic/claude-3.7-sonnet"
	LLMModelGeminiTwoFlashLite LLMModel = "google/gemini-2.0-flash-lite-001"
	LLMModelGeminiTwoFlash     LLMModel = "google/gemini-2.0-flash-001"
)

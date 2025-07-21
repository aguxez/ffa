package agent

type LLMModel string

const (
	LLMModelClaude             LLMModel = "anthropic/claude-sonnet-4"
	LLMModelGeminiTwoFlashLite LLMModel = "google/gemini-2.0-flash-lite-001"
	LLMModelGeminiTwoFlash     LLMModel = "google/gemini-2.5-flash"
)

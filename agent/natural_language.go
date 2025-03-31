package agent

import (
	"context"
	"time"

	"github.com/aguxez/ffa/database"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

// NaturalLanguageAgent represents the main agent with memory capabilities
type NaturalLanguageAgent struct {
	llm *openai.LLM
}

// NewNaturalLanguageAgent creates a new NaturalLanguageAgent with the provided LLM.
//
// This function initializes a NaturalLanguageAgent, which is capable of
// processing natural language using the specified LLM (Large Language Model).
//
// Parameters:
// - llm: A pointer to an openai.LLM instance to be used for language processing.
//
// Returns:
// A pointer to a newly created NaturalLanguageAgent instance.
func NewNaturalLanguageAgent(llm *openai.LLM) *NaturalLanguageAgent {
	return &NaturalLanguageAgent{
		llm: llm,
	}
}

// ParseMessage is a generic function to process different types of natural language messages
func ParseMessage[T any](
	ctx context.Context,
	a *NaturalLanguageAgent,
	message string,
	templateFunc func() string,
) (T, error) {
	var result T

	// Call the LLM to process the message
	promptTemplate := prompts.NewPromptTemplate(templateFunc(), []string{"Input"})
	promptInput := map[string]any{
		"Input": message,
	}

	prompt, err := promptTemplate.Format(promptInput)
	if err != nil {
		return result, err
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	output, err := a.llm.Call(ctxWithTimeout, prompt)
	if err != nil {
		return result, err
	}

	return ParseJSONFromResponse[T](output)
}

// ParseMacroTargetsMessage processes a message containing macro nutrient targets
func (a *NaturalLanguageAgent) ParseMacroTargetsMessage(
	ctx context.Context,
	message string,
) (database.UserMacrosPayload, error) {
	return ParseMessage[database.UserMacrosPayload](
		ctx,
		a,
		message,
		NewMacroTargetsTemplate,
	)
}

// ParseDietaryRestrictionsMessage processes a message containing dietary restrictions
func (a *NaturalLanguageAgent) ParseDietaryRestrictionsMessage(
	ctx context.Context,
	message string,
) (database.UserDietaryRestrictionsPayload, error) {
	return ParseMessage[database.UserDietaryRestrictionsPayload](
		ctx,
		a,
		message,
		func() string { return NewFoodParsingTemplate("dietary restrictions") },
	)
}

// ParseUserFoodsMessage processes a message containing foods that the user likes.
func (a *NaturalLanguageAgent) ParseUserFoodsMessage(
	ctx context.Context,
	message string,
) (database.UserFoodsPayload, error) {
	return ParseMessage[database.UserFoodsPayload](
		ctx,
		a,
		message,
		func() string { return NewFoodParsingTemplate("foods") },
	)
}

package agent

// NewPromptTemplate returns a string template for generating meal plans using a conversational agent.
// The template provides a structured format for the agent to follow when interacting with users
// as a personal nutritionist. It leverages historical meal plans and tool access to generate
// comprehensive meal plans that meet user-defined macro targets, include diverse food options,
// and are suitable for meal prep. The template specifies JSON format for outputting meal plans
// with details such as food name, weight, macros, and explanations. The meal plan should consist
// of five meals per day, including a protein shake, and provide weekday portions with daily macro
// breakdowns. The response must be terminal-displayable, with explanations and preparations that
// incorporate color compatibility.
func NewPromptTemplate() string {
	return `
	You are a personal nutritionist with access to historical meal plans and their outcomes.

	When generating a meal plan, follow these strict rules:
	1. Always format your final answer as a valid JSON object
	2. The context you are given is to give you ideas of what foods the person likes to consume.
			You are free to use this to give recommendations rather than strictly recipes with the foods
			you are given.
	3. Your response MUST respect the JSON format you will be provided

	Context:
	{{.CombinedInput}}

	Your Final Answer must be in this specific JSON format:
	{
			"plan": [
					{
							"food": "Example Food",
							"weight": "100g",
							"macros": "P: 20g, C: 30g, F: 10g, Cal: 290",
							"foodExplanation": "Short explanation",
							"foodCategory": "breakfast"
					}
			],
			"planExplanation": "Explanation in markdown",
			"planPreparation": {
				"breakfast": "Prep instructions for breakfast",
				"lunch": "Prep instructions for lunch",
				"dinner": "Prep instructions for dinner"
			}
	}

	Where "foodCategory" must be one of: breakfast, lunch, dinner, or snack.`
}

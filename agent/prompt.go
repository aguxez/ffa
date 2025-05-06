package agent

import "fmt"

const (
	BeginnerNumberOfAcceptedPlans     = 11
	IntermediateNumberOfAcceptedPlans = 12
	AdvancedNumberOfAcceptedPlans     = 20
)

// NewMealPlanPromptTemplate returns a string template for generating meal plans using a conversational agent.
// The template provides a structured format for the agent to follow when interacting with users
// as a personal nutritionist. It leverages historical meal plans and tool access to generate
// comprehensive meal plans that meet user-defined macro targets, include diverse food options,
// and are suitable for meal prep. The template specifies JSON format for outputting meal plans
// with details such as food name, weight, macros, and explanations. The meal plan should consist
// of five meals per day, including a protein shake, and provide weekday portions with daily macro
// breakdowns. The response must be terminal-displayable, with explanations and preparations that
// incorporate color compatibility.
func NewMealPlanPromptTemplate() string {
	return fmt.Sprintf(`
You are a personal nutritionist AI assistant with expertise in meal planning and nutritional analysis, informed by simulated historical meal plan data and outcomes.

Your goal is to generate a personalized meal plan based on user context and preferences.

When generating the meal plan, follow these strict rules AT ALL TIMES:
1.  **JSON Format Absolute:** Your *entire* response MUST be a single, valid JSON object conforming precisely to the provided SCHEMA. No introductory text, no explanations outside the JSON structure, no markdown formatting, no code blocks.
2.  **Context Interpretation:** Use the user preferences, dietary needs, macros, liked foods as inspiration and guidance. Prioritize suggestions aligning with these inputs, but you are encouraged to propose variations or complementary dishes rather than only listing foods strictly mentioned.
3.  **Error Handling:** If you encounter any insurmountable error or cannot generate a valid plan based on the inputs, return an empty JSON object with the specified keys. This ensures downstream parsing safety.
4.  **No Extraneous Output:** ABSOLUTELY NO text before or after the JSON object. Failure to adhere to this will break parsing and cause operational KKTs/issues.
5.  **Dietary Compliance:** Strictly adhere to all dietary restrictions mentioned in the context. Only suggest foods compliant with these restrictions.
6.  **Suggestion Limit:** Generate a maximum of ONE food suggestions per foodCategory.
7.  **Variety Boost:** Analyze the previous plan. Actively avoid suggesting the *exact same meals* listed in the previous plan to ensure week-to-week variety.
8.  **Macro Awareness:** If specific macro targets (calories, protein, carbs, fat) are provided in the context, strive to suggest meals that align reasonably well with those targets. The macros field in your output should represent estimated macronutrients for a standard serving of the suggested food.
9.  **Language Adherence:** All user-facing string values within the JSON (food, foodExplanation, foodCategory) MUST be generated *in* the language specified by the user.
10. **Food Categories:** Must be at least one of the following: breakfast, lunch, dinner, and snack

**Plan Type Guidance:**
This is a {{.PlanType}} plan. Adapt your suggestions accordingly:
*   **Weekday Plans:** Focus on meals suitable for advance preparation (e.g., batch cooking, easy reheating). Prioritize balanced, sustainable options for a routine.
*   **Weekend Plans:** Be more adventurous. Suggest meals that might involve more cooking time on the day, explore different cuisines or flavor profiles, or offer slightly more indulgent options, providing a distinct break from the weekday pattern.

**Difficulty Guidance:**
Consider the fact this user has completed {{.NumberOfAcceptedPlans}} meal plans.
We want to ensure the user improves their cooking skills by considering different
levels of cooking based on the number of accepted plans which translates to cooked meals in a week.

*   **Beginner:** Focus on simple, quick, and easy meals. Until %d accepted plans
*   **Intermediate:** Focus on more complex, longer, and more involved meals. At least %d accepted plans
*   **Advanced:** Focus on more complex, longer, and more involved meals. At least %d accepted plans

**User Context:**
{{.CombinedInput}}

**User Macros (empty if not provided):**
{{.UserMacros}}

**Last Accepted Meal Plan (null if none):**
{{.PreviousPlanInput}}

**Output Language:**
{{.UserLanguage}}

**Required JSON SCHEMA:**
{
    "plan": [
        {
            "food": "string",
            "macros": "string",
            "foodExplanation": "string",
            "foodCategory": "string"
        }
    ]
}

**EXAMPLE (Illustrative, follow SCHEMA strictly):**
{
    "plan": [
        {
            "food": "Overnight Oats with Berries and Nuts",
            "macros": "P: 15g, C: 50g, F: 12g, Cal: 370",
            "foodExplanation": "Easy to prepare the night before, providing sustained energy for the morning. Customizable with preferred toppings.",
            "foodCategory": "breakfast"
        },
        {
            "food": "Quinoa Salad with Roasted Vegetables and Chickpeas",
            "macros": "P: 18g, C: 60g, F: 15g, Cal: 450",
            "foodExplanation": "A hearty and nutritious lunch that holds up well when prepped ahead. Provides complex carbs and plant-based protein.",
            "foodCategory": "lunch"
        }
    ]
}

**JSON Response:**
	`, BeginnerNumberOfAcceptedPlans, IntermediateNumberOfAcceptedPlans, AdvancedNumberOfAcceptedPlans)
}

// NewMealPlanExplanationTemplate returns a string representing a template for a prompt
// given to the LLM for generating a comprehensive explanation of a meal plan.
// The template provides a structured format for the agent to follow when generating
// meal plan explanations, which should cover nutritional balance, user preferences, and overall benefits.
func NewMealPlanExplanationTemplate() string {
	return `
You are a personal nutritionist. Based on the following meal plan items, create a comprehensive explanation of the overall nutrition strategy.

	When generating a meal plan, follow these strict rules:
	1. Always format your final answer as a valid JSON object
	2. Your response MUST respect the JSON format you will be provided. If you encounter an error,
			return an empty JSON object with the provided keys so it can be parsed.
	3. Do NOT use markdown or any other type of markup. No asterisks or other symbols are allowed for your string output.
		DO NOT add any other string other than the JSON, otherwise you wil mess up the parsing of the response and this can
		incur in monetary loss.
	4. Once again. DO NOT ADD Markdown to the text. No asterisks or other symbols are allowed for your string output.
	5. Explain the nutritional balance, how it addresses the user's preferences from context, and the overall benefits.
	6. Do not be verbose in your response. This needs to be easily digestible by the reader, so, do a summary.

This meal plan is for a {{.PlanType}} plan.

Context:
{{.PlanFromChain1}}

Consider the user language for the output generation:
{{.UserLanguage}}

SCHEMA:
	{
		"planExplanation": "(string, NO MARKDOWN)",
	}

JSON Response:
`
}

// NewMealPlanPreparationTemplate returns a string representing a template for a prompt
// given to the LLM for generating meal preparation instructions from a list of foods.
func NewMealPlanPreparationTemplate() string {
	return fmt.Sprintf(`
You are a personal nutritionist specializing in meal prep. Based on the following meal plan items, create practical preparation instructions organized by meal type.

	When generating a meal plan, follow these strict rules:
	1. Always format your final answer as a valid JSON object
	2. Your response MUST respect the JSON format you will be provided. If you encounter an error,
			return an empty JSON object with the provided keys so it can be parsed.
	3. Do NOT use markdown or any other type of markup. No asterisks or other symbols are allowed for your string output.
		DO NOT add any other string other than the JSON, otherwise you wil mess up the parsing.

	This is for a {{.PlanType}} plan. For weekdays, focus on efficient batch cooking methods,
	storage recommendations, and how to maintain freshness throughout the week. For weekends,
	keep it simple and easy to prepare, no need to focus on storage since we're not cooking
	food in advance.

**Difficulty Guidance:**
Consider the fact this user has completed {{.NumberOfAcceptedPlans}} meal plans.
We want to ensure the user improves their cooking skills by considering different
levels of cooking based on the number of accepted plans which translates to cooked meals in a week.

*   **Beginner:** Until %d	 accepted plans
*   **Intermediate:** At least %d accepted plans
*   **Advanced:** At least %d accepted plans

Ensure preparation steps align with the user's cooking experience level defined here.

**Meal Plan Shopping List::**
Here is the shopping list we have given the user, so, ensure your preparation steps
only focus on the items in the shopping list:
{{.ShoppingList}}

Meal Plan Items:
{{.PlanFromChain1}}

Consider the user language for the output generation:
{{.UserLanguage}}

Your output should include only the content for the "planPreparation" object in the final JSON with these three keys:
- breakfast: Step by step prep instructions for breakfast items
- lunch: Step by step prep instructions for lunch items
- dinner: Step by step prep instructions for dinner items

SCHEMA:
{
		"planPreparation": {
			"breakfast": "string",
			"lunch": "string",
			"dinner": "string"
		}
}

EXAMPLE:
{
	"planPreparation": {
		"breakfast": "Protein-Packed Chicken and Vegetable Breakfast Skillet: (Intermediate)\n1. Dice chicken breast into bite-sized pieces.\n2. Chop bell peppers and onions.\n3. Sauté chicken in a pan until cooked through.\n4. Add bell peppers and onions, cook until softened.\n5. Stir in spinach until wilted.\n6. Portion into containers for the week.\n7. Reheat in the microwave or on the stovetop each morning.",
		"lunch": "Mediterranean Chicken Wrap with Hummus: (Beginner) \n1. Grill or cook chicken breast. \n2. Prepare hummus. \n3. Wash and chop cucumber and tomato. \n4. Assemble wraps each morning by spreading hummus on a whole-grain wrap, adding mixed greens, chicken, cucumber, and tomato. \n5. Wrap tightly.",
		"dinner": "Herb-Crusted Baked Chicken with Sweet Potato and Broccoli: (Intermediate) \n1. Preheat oven to 400°F (200°C). \n2. Toss sweet potato cubes with olive oil, salt, and pepper. Roast for 20 minutes. \n3. Prepare chicken by coating with herb and panko mixture. \n4. Place chicken on a baking sheet with sweet potatoes. \n5. Bake chicken and sweet potatoes for 20-25 minutes, or until chicken is cooked through and sweet potatoes are tender. \n6. Steam broccoli. \n7. Portion chicken, sweet potatoes, and broccoli into containers for the week."
	}
}

JSON Response:
`, BeginnerNumberOfAcceptedPlans, IntermediateNumberOfAcceptedPlans, AdvancedNumberOfAcceptedPlans)
}

// NewMacroTargetsTemplate returns a string representing a template for a prompt
// given to the LLM for parsing user-provided macronutrient targets.
func NewMacroTargetsTemplate() string {
	return `
	Take this text and parse the macronutrients as given to you.

	When generating a meal plan, follow these strict rules:
	1. Always format your final answer as a valid JSON object
	2. Your response MUST respect the JSON format you will be provided. If you encounter an error,
			return an empty JSON object with the provided keys so it can be parsed.
	3. Do NOT use markdown or any other type of markup. No asterisks or other symbols are allowed for your string output.
		DO NOT add any other string other than the JSON, otherwise you wil mess up the parsing of the response and this can
		incur in monetary loss.

	Context:
	{{.Input}}

	EXAMPLE:
	{
		"calories": 2000,
		"protein": 100,
		"fat": 50,
		"carbs": 300
	}

	SCHEMA:
	{
		"calories": number,
		"protein": number,
		"fat": number,
		"carbs": number
	}

	JSON Response:
	`
}

// NewFoodParsingTemplate returns a string representing a template for a prompt
// given to the LLM for parsing user-provided dietary restrictions.
func NewFoodParsingTemplate(foodType string) string {
	return fmt.Sprintf(`
	Take this text and parse the %s as given to you.

	When generating a meal plan, follow these strict rules:
	1. Always format your final answer as a valid JSON object
	2. Your response MUST respect the JSON format you will be provided. If you encounter an error,
			return an empty JSON object with the provided keys so it can be parsed.
	3. Do NOT use markdown or any other type of markup. No asterisks or other symbols are allowed for your string output.
		DO NOT add any other string other than the JSON, otherwise you wil mess up the parsing of the response and this can
		incur in monetary loss.
	4. The JSON is simply an array of strings. Do not add objects or other keys.

	Context:
	{{.Input}}

	[string]
	`, foodType)
}

// NewMealPlanShoppingListTemplate returns a string representing a template for a prompt
// given to the LLM for generating a shopping list from a list of foods.
func NewMealPlanShoppingListTemplate() string {
	return fmt.Sprintf(`
	Here you have some foods, read them all and create a shopping list for all of them.
	You are given context for a meal plan that was already created.

	When generating a meal plan, follow these strict rules:
	1. Always format your final answer as a valid JSON object
	2. Your response MUST respect the JSON format you will be provided. If you encounter an error,
			return an empty JSON object with the provided keys so it can be parsed.
	3. Do NOT use markdown or any other type of markup. No asterisks or other symbols are allowed for your string output.
		DO NOT add any other string other than the JSON, otherwise you wil mess up the parsing of the response and this can
		incur in monetary loss.
	4. Take the user macros into consideration (if they're available in the input) when suggesting food portions
	5. Try as much as possible to add the portions to the foods so the user can have better perspective on how
		to buy.


	**Difficulty Guidance:**
	Consider the fact this user has completed {{.NumberOfAcceptedPlans}} meal plans.
	We want to ensure the user improves their cooking skills by considering different
	levels of cooking based on the number of accepted plans which translates to cooked meals in a week.

	*   **Beginner:** Focus on simple, quick, and easy meals. Until %d accepted plans
	*   **Intermediate:** Focus on more complex, longer, and more involved meals. At least %d accepted plans
	*   **Advanced:** Focus on more complex, longer, and more involved meals. At least %d accepted plans

	FOODS FROM GIVEN PLAN:
	{{.PlanFoods}}

	MACROS FROM GIVEN PLAN:
	{{.PlanMacros}}

	USER MACROS (empty if not provided):
	{{.UserMacros}}

	USER LANGUAGE FOR OUTPUT:
	{{.UserLanguage}}

	SCHEMA:
	[
		{
			"foodGroup": string,
			"foods": [string]
		}
	]

	EXAMPLE:
	[
		{
			"foodGroup": "Protein",
			"foods": ["Lean Meat (100g)", "Chicken Breast (200g)", "Salmon (420g)"]
		},
		{
			"foodGroup": "Carbs",
			"foods": ["Brown Rice (200g)", "Quinoa (200g)", "Oatmeal (200g)"]
		},
		{
			"foodGroup": "Fruits",
			"foods": ["Bananas (100g)", "Oranges (100g)", "Apples (100g)"]
		}
	]

	JSON Response:
	`, BeginnerNumberOfAcceptedPlans, IntermediateNumberOfAcceptedPlans, AdvancedNumberOfAcceptedPlans)
}

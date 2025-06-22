package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cbrgm/go-mcp-server/mcp"
)

const (
	teaTypeGreen  = "Green Tea"
	teaTypeBlack  = "Black Tea"
	teaTypeOolong = "Oolong Tea"
	teaTypeWhite  = "White Tea"

	caffeineLevelVeryLow = "Very Low"
	caffeineLevelLow     = "Low"
	caffeineLevelMedium  = "Medium"
	caffeineLevelHigh    = "High"

	toolGetTeaNames   = "getTeaNames"
	toolGetTeaInfo    = "getTeaInfo"
	toolGetTeasByType = "getTeasByType"
)

type TeaHandler struct{}

type Tea struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Origin      string  `json:"origin"`
	Caffeine    string  `json:"caffeine"`
	Flavor      string  `json:"flavor"`
	Temperature int     `json:"temperature"`
	SteepTime   string  `json:"steepTime"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

var teaMenu = map[string]Tea{
	"dragonwell": {
		Name:        "Dragonwell",
		Type:        teaTypeGreen,
		Origin:      "China",
		Caffeine:    caffeineLevelMedium,
		Flavor:      "Delicate, sweet, nutty",
		Temperature: 175,
		SteepTime:   "2-3 minutes",
		Description: "A classic Chinese green tea with a smooth, mellow flavor and beautiful flat leaves.",
		Price:       8.50,
	},
	"earl-grey": {
		Name:        "Earl Grey",
		Type:        teaTypeBlack,
		Origin:      "England",
		Caffeine:    caffeineLevelHigh,
		Flavor:      "Citrusy, bergamot, bold",
		Temperature: 212,
		SteepTime:   "3-5 minutes",
		Description: "A traditional English black tea infused with bergamot oil for a distinctive citrus aroma.",
		Price:       7.00,
	},
	"da-hong-pao": {
		Name:        "Da Hong Pao",
		Type:        "Oolong Tea",
		Origin:      "China",
		Caffeine:    "Medium",
		Flavor:      "Complex, roasted, fruity",
		Temperature: 200,
		SteepTime:   "1-2 minutes",
		Description: "A legendary Chinese oolong with a rich, complex flavor and beautiful amber liquor.",
		Price:       15.00,
	},
	"white-peony": {
		Name:        "White Peony",
		Type:        "White Tea",
		Origin:      "China",
		Caffeine:    "Low",
		Flavor:      "Subtle, floral, sweet",
		Temperature: 185,
		SteepTime:   "4-6 minutes",
		Description: "A delicate white tea with silvery buds and a light, refreshing taste.",
		Price:       12.00,
	},
	"gyokuro": {
		Name:        "Gyokuro",
		Type:        "Green Tea",
		Origin:      "Japan",
		Caffeine:    "High",
		Flavor:      "Umami, sweet, vegetal",
		Temperature: 140,
		SteepTime:   "1-2 minutes",
		Description: "Premium Japanese green tea grown in shade, producing a rich umami flavor.",
		Price:       18.00,
	},
	"assam": {
		Name:        "Assam",
		Type:        "Black Tea",
		Origin:      "India",
		Caffeine:    "High",
		Flavor:      "Malty, robust, brisk",
		Temperature: 212,
		SteepTime:   "3-5 minutes",
		Description: "A full-bodied Indian black tea perfect for breakfast and pairs well with milk.",
		Price:       6.50,
	},
	"tie-guan-yin": {
		Name:        "Tie Guan Yin",
		Type:        "Oolong Tea",
		Origin:      "China",
		Caffeine:    "Medium",
		Flavor:      "Floral, orchid-like, smooth",
		Temperature: 195,
		SteepTime:   "1-3 minutes",
		Description: "Iron Goddess of Mercy - a premium Chinese oolong with floral notes and lasting sweetness.",
		Price:       13.50,
	},
	"silver-needle": {
		Name:        "Silver Needle",
		Type:        "White Tea",
		Origin:      "China",
		Caffeine:    "Very Low",
		Flavor:      "Delicate, honey, fresh",
		Temperature: 175,
		SteepTime:   "5-7 minutes",
		Description: "The most prized white tea made from young buds, offering exceptional delicacy and sweetness.",
		Price:       22.00,
	},
}

func (h *TeaHandler) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	return []mcp.Tool{
		{
			Name:        toolGetTeaNames,
			Description: "Get a list of all available tea names in our collection",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]interface{}{},
			},
		},
		{
			Name:        toolGetTeaInfo,
			Description: "Get detailed information about a specific tea including brewing instructions",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "The name of the tea (e.g., 'dragonwell', 'earl-grey')",
					},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        toolGetTeasByType,
			Description: "Get all teas of a specific type (Green Tea, Black Tea, Oolong Tea, White Tea)",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"type": map[string]interface{}{
						"type":        "string",
						"description": "The tea type (e.g., 'Green Tea', 'Black Tea', 'Oolong Tea', 'White Tea')",
					},
				},
				Required: []string{"type"},
			},
		},
	}, nil
}

func (h *TeaHandler) CallTool(ctx context.Context, params mcp.ToolCallParams) (mcp.ToolResponse, error) {
	switch params.Name {
	case "getTeaNames":
		var names []string
		for key := range teaMenu {
			names = append(names, key)
		}

		result, err := json.Marshal(names)
		if err != nil {
			return mcp.ToolResponse{}, fmt.Errorf("failed to marshal tea names: %w", err)
		}

		return mcp.ToolResponse{
			Content: []mcp.ContentItem{
				{
					Type: "text",
					Text: string(result),
				},
			},
		}, nil

	case "getTeaInfo":
		nameInterface, ok := params.Arguments["name"]
		if !ok {
			return mcp.ToolResponse{}, fmt.Errorf("name parameter is required")
		}

		name, ok := nameInterface.(string)
		if !ok {
			return mcp.ToolResponse{}, fmt.Errorf("name parameter must be a string")
		}

		tea, exists := teaMenu[name]
		if !exists {
			return mcp.ToolResponse{
				Content: []mcp.ContentItem{
					{
						Type: "text",
						Text: fmt.Sprintf("Tea '%s' not found in our collection", name),
					},
				},
			}, nil
		}

		result, err := json.Marshal(tea)
		if err != nil {
			return mcp.ToolResponse{}, fmt.Errorf("failed to marshal tea info: %w", err)
		}

		return mcp.ToolResponse{
			Content: []mcp.ContentItem{
				{
					Type: "text",
					Text: string(result),
				},
			},
		}, nil

	case "getTeasByType":
		typeInterface, ok := params.Arguments["type"]
		if !ok {
			return mcp.ToolResponse{}, fmt.Errorf("type parameter is required")
		}

		teaType, ok := typeInterface.(string)
		if !ok {
			return mcp.ToolResponse{}, fmt.Errorf("type parameter must be a string")
		}

		var matchingTeas []Tea
		for _, tea := range teaMenu {
			if tea.Type == teaType {
				matchingTeas = append(matchingTeas, tea)
			}
		}

		if len(matchingTeas) == 0 {
			return mcp.ToolResponse{
				Content: []mcp.ContentItem{
					{
						Type: "text",
						Text: fmt.Sprintf("No teas found of type '%s'", teaType),
					},
				},
			}, nil
		}

		result, err := json.Marshal(matchingTeas)
		if err != nil {
			return mcp.ToolResponse{}, fmt.Errorf("failed to marshal teas by type: %w", err)
		}

		return mcp.ToolResponse{
			Content: []mcp.ContentItem{
				{
					Type: "text",
					Text: string(result),
				},
			},
		}, nil

	default:
		return mcp.ToolResponse{}, fmt.Errorf("unknown tool: %s", params.Name)
	}
}

func (h *TeaHandler) ListResources(ctx context.Context) ([]mcp.Resource, error) {
	return []mcp.Resource{
		{
			URI:  "menu://tea",
			Name: "Tea Menu",
		},
	}, nil
}

func (h *TeaHandler) ReadResource(ctx context.Context, params mcp.ResourceParams) (mcp.ResourceResponse, error) {
	switch params.URI {
	case "menu://tea":
		menuData, err := json.MarshalIndent(teaMenu, "", "  ")
		if err != nil {
			return mcp.ResourceResponse{}, fmt.Errorf("failed to marshal tea menu: %w", err)
		}

		return mcp.ResourceResponse{
			Contents: []mcp.ResourceContent{
				{
					URI:  params.URI,
					Text: string(menuData),
				},
			},
		}, nil
	default:
		return mcp.ResourceResponse{}, fmt.Errorf("unknown resource URI: %s", params.URI)
	}
}

func (h *TeaHandler) ListResourceTemplates(ctx context.Context) ([]mcp.ResourceTemplate, error) {
	return []mcp.ResourceTemplate{}, nil
}

func (h *TeaHandler) ListPrompts(ctx context.Context) ([]mcp.Prompt, error) {
	return []mcp.Prompt{
		{
			Name:        "tea_recommendation",
			Description: "Get personalized tea recommendations based on preferences",
			Arguments: []mcp.PromptArgument{
				{
					Name:        "mood",
					Description: "Current mood or desired effect (e.g., 'energizing', 'relaxing', 'focus')",
					Required:    false,
				},
				{
					Name:        "caffeine_preference",
					Description: "Caffeine level preference (e.g., 'high', 'medium', 'low', 'none')",
					Required:    false,
				},
				{
					Name:        "flavor_profile",
					Description: "Preferred flavor profile (e.g., 'floral', 'robust', 'delicate', 'complex')",
					Required:    false,
				},
			},
		},
		{
			Name:        "brewing_guide",
			Description: "Get detailed brewing instructions for a specific tea",
			Arguments: []mcp.PromptArgument{
				{
					Name:        "tea_name",
					Description: "Name of the tea to get brewing instructions for",
					Required:    true,
				},
			},
		},
		{
			Name:        "tea_pairing",
			Description: "Get food pairing suggestions for a specific tea",
			Arguments: []mcp.PromptArgument{
				{
					Name:        "tea_name",
					Description: "Name of the tea to get pairing suggestions for",
					Required:    true,
				},
			},
		},
	}, nil
}

func (h *TeaHandler) GetPrompt(ctx context.Context, params mcp.PromptParams) (mcp.PromptResponse, error) {
	arguments := h.convertArguments(params.Arguments)

	switch params.Name {
	case "tea_recommendation":
		return h.generateTeaRecommendation(arguments)
	case "brewing_guide":
		return h.generateBrewingGuide(arguments)
	case "tea_pairing":
		return h.generateTeaPairing(arguments)
	default:
		return mcp.PromptResponse{}, fmt.Errorf("unknown prompt: %s", params.Name)
	}
}

func (h *TeaHandler) convertArguments(args map[string]any) map[string]string {
	arguments := make(map[string]string)
	for k, v := range args {
		if str, ok := v.(string); ok {
			arguments[k] = str
		}
	}
	return arguments
}

func (h *TeaHandler) generateTeaRecommendation(arguments map[string]string) (mcp.PromptResponse, error) {
	mood := arguments["mood"]
	caffeinePreference := arguments["caffeine_preference"]
	flavorProfile := arguments["flavor_profile"]

	prompt := "Based on our tea collection, here are some recommendations:\n\n"

	if mood != "" {
		prompt += h.getMoodRecommendations(mood)
	}

	if caffeinePreference != "" {
		prompt += h.getCaffeineRecommendations(caffeinePreference)
	}

	if flavorProfile != "" {
		prompt += h.getFlavorRecommendations(flavorProfile)
	}

	return h.createPromptResponse(prompt), nil
}

func (h *TeaHandler) getMoodRecommendations(mood string) string {
	prompt := fmt.Sprintf("For a %s mood:\n", mood)
	switch mood {
	case "energizing":
		prompt += "- Gyokuro (high caffeine, umami flavor)\n- Assam (robust, perfect morning tea)\n"
	case "relaxing":
		prompt += "- White Peony (low caffeine, delicate)\n- Silver Needle (very low caffeine, honey notes)\n"
	case "focus":
		prompt += "- Earl Grey (bergamot aids concentration)\n- Da Hong Pao (complex flavors for mindful drinking)\n"
	}
	return prompt + "\n"
}

func (h *TeaHandler) getCaffeineRecommendations(caffeinePreference string) string {
	prompt := fmt.Sprintf("For %s caffeine preference:\n", caffeinePreference)
	switch caffeinePreference {
	case "high":
		prompt += "- Gyokuro, Earl Grey, Assam\n"
	case "medium":
		prompt += "- Dragonwell, Da Hong Pao, Tie Guan Yin\n"
	case "low":
		prompt += "- White Peony\n"
	case "none", "very low":
		prompt += "- Silver Needle\n"
	}
	return prompt + "\n"
}

func (h *TeaHandler) getFlavorRecommendations(flavorProfile string) string {
	prompt := fmt.Sprintf("For %s flavor profile:\n", flavorProfile)
	switch flavorProfile {
	case "floral":
		prompt += "- Tie Guan Yin (orchid-like), White Peony (subtle floral)\n"
	case "robust":
		prompt += "- Assam (malty), Earl Grey (bold bergamot)\n"
	case "delicate":
		prompt += "- Silver Needle (honey sweetness), Dragonwell (gentle nuttiness)\n"
	case "complex":
		prompt += "- Da Hong Pao (roasted, fruity), Gyokuro (umami depth)\n"
	}
	return prompt
}

func (h *TeaHandler) generateBrewingGuide(arguments map[string]string) (mcp.PromptResponse, error) {
	teaName := arguments["tea_name"]
	if teaName == "" {
		return mcp.PromptResponse{}, fmt.Errorf("tea_name is required for brewing guide")
	}

	tea, exists := teaMenu[teaName]
	if !exists {
		return mcp.PromptResponse{}, fmt.Errorf("tea '%s' not found in our collection", teaName)
	}

	prompt := fmt.Sprintf(`# Brewing Guide for %s

## Tea Information
- **Type**: %s
- **Origin**: %s
- **Caffeine Level**: %s

## Brewing Instructions
- **Water Temperature**: %d°F
- **Steeping Time**: %s
- **Flavor Profile**: %s

## Tips
%s

Enjoy your perfectly brewed %s!`,
		tea.Name, tea.Type, tea.Origin, tea.Caffeine,
		tea.Temperature, tea.SteepTime, tea.Flavor,
		tea.Description, tea.Name)

	return h.createPromptResponse(prompt), nil
}

func (h *TeaHandler) generateTeaPairing(arguments map[string]string) (mcp.PromptResponse, error) {
	teaName := arguments["tea_name"]
	if teaName == "" {
		return mcp.PromptResponse{}, fmt.Errorf("tea_name is required for pairing suggestions")
	}

	tea, exists := teaMenu[teaName]
	if !exists {
		return mcp.PromptResponse{}, fmt.Errorf("tea '%s' not found in our collection", teaName)
	}

	pairings := h.getTeaPairings(tea.Type)

	prompt := fmt.Sprintf(`# Food Pairings for %s

## Tea Profile
- **Type**: %s
- **Flavor**: %s
- **Origin**: %s

## Recommended Pairings
%s

## Why These Pairings Work
The %s characteristics of %s complement these foods perfectly, creating a harmonious tasting experience.

Price: $%.2f`,
		tea.Name, tea.Type, tea.Flavor, tea.Origin,
		pairings, tea.Flavor, tea.Name, tea.Price)

	return h.createPromptResponse(prompt), nil
}

func (h *TeaHandler) getTeaPairings(teaType string) string {
	switch teaType {
	case "Green Tea":
		return "Light appetizers, sushi, steamed vegetables, mild cheeses, fruit tarts"
	case "Black Tea":
		return "Breakfast pastries, chocolate desserts, hearty sandwiches, aged cheeses, spiced foods"
	case "Oolong Tea":
		return "Roasted nuts, grilled seafood, dim sum, stone fruits, semi-hard cheeses"
	case "White Tea":
		return "Fresh fruits, light salads, delicate pastries, soft cheeses, cucumber sandwiches"
	default:
		return "Light snacks and mild flavors that won't overpower the tea"
	}
}

func (h *TeaHandler) createPromptResponse(text string) mcp.PromptResponse {
	return mcp.PromptResponse{
		Messages: []mcp.PromptMessage{
			{
				Role: "user",
				Content: mcp.MessageContent{
					Type: "text",
					Text: text,
				},
			},
		},
	}
}

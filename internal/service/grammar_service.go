package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"languagepapi/internal/bridge"
	"languagepapi/internal/models"
	"languagepapi/internal/repository"
)

// GrammarService handles grammar explanation generation
type GrammarService struct{}

// NewGrammarService creates a new grammar service
func NewGrammarService() *GrammarService {
	return &GrammarService{}
}

// GetGrammarForCard generates or retrieves grammar explanation for a card
func (s *GrammarService) GetGrammarForCard(ctx context.Context, card *models.Card) (*models.GrammarTip, error) {
	// Check if we already have grammar linked to this card
	existing, err := repository.GetGrammarForCard(card.ID)
	if err == nil && existing != nil {
		examples := parseGrammarExamples(existing.Examples)
		return &models.GrammarTip{
			RuleKey:     existing.RuleKey,
			Title:       existing.Title,
			ShortExplan: truncate(existing.Explanation, 200),
			Examples:    examples,
		}, nil
	}

	// Generate new grammar explanation
	return s.GenerateGrammarForCard(ctx, card)
}

// GenerateGrammarForCard generates grammar explanation using AI
func (s *GrammarService) GenerateGrammarForCard(ctx context.Context, card *models.Card) (*models.GrammarTip, error) {
	gemini, err := bridge.NewGeminiService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI service: %w", err)
	}
	defer gemini.Close()

	prompt := fmt.Sprintf(`Analyze the Spanish word and explain the relevant grammar concept.

Word: "%s"
Translation: "%s"
Example: "%s"

Identify the most important grammar concept for this word (e.g., verb conjugation, gender agreement, tense, mood).

Return ONLY valid JSON:
{
  "rule_key": "present_tense_ar",
  "title": "Present Tense: -AR Verbs",
  "explanation": "Spanish -AR verbs are conjugated by removing -ar and adding endings: -o, -as, -a, -amos, -áis, -an.",
  "examples": [
    {"spanish": "Yo hablo español.", "english": "I speak Spanish."},
    {"spanish": "Ella habla rápido.", "english": "She speaks fast."}
  ],
  "difficulty": 1
}

Keep explanation under 150 words. Include 2-3 practical examples.`,
		card.Term, card.Translation, card.ExampleSentence)

	response, err := gemini.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate grammar: %w", err)
	}

	// Parse response
	text := strings.TrimSpace(response)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var result struct {
		RuleKey     string                   `json:"rule_key"`
		Title       string                   `json:"title"`
		Explanation string                   `json:"explanation"`
		Examples    []models.GrammarExample  `json:"examples"`
		Difficulty  int                      `json:"difficulty"`
	}

	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("failed to parse grammar response: %w", err)
	}

	// Save to database
	examplesJSON, _ := json.Marshal(result.Examples)
	ruleID, err := repository.SaveGrammarRule(
		result.RuleKey,
		result.Title,
		result.Explanation,
		string(examplesJSON),
		result.Difficulty,
	)
	if err != nil {
		log.Printf("Failed to save grammar rule: %v", err)
	} else {
		// Link card to grammar rule
		repository.LinkCardToGrammar(card.ID, ruleID)
	}

	return &models.GrammarTip{
		RuleKey:     result.RuleKey,
		Title:       result.Title,
		ShortExplan: truncate(result.Explanation, 200),
		Examples:    result.Examples,
	}, nil
}

// GetPreLessonTips analyzes cards in a lesson and returns relevant grammar tips
func (s *GrammarService) GetPreLessonTips(ctx context.Context, cards []models.LessonCard) ([]models.GrammarTip, error) {
	var tips []models.GrammarTip
	seenRules := make(map[string]bool)

	// Check for existing grammar rules linked to these cards
	for _, card := range cards {
		if len(tips) >= 2 { // Max 2 tips per lesson
			break
		}

		rule, err := repository.GetGrammarForCard(card.ID)
		if err == nil && rule != nil {
			if seenRules[rule.RuleKey] {
				continue
			}
			seenRules[rule.RuleKey] = true

			examples := parseGrammarExamples(rule.Examples)
			tips = append(tips, models.GrammarTip{
				RuleKey:     rule.RuleKey,
				Title:       rule.Title,
				ShortExplan: truncate(rule.Explanation, 150),
				Examples:    examples[:min(2, len(examples))],
			})
		}
	}

	// If we have tips from existing rules, return them
	if len(tips) > 0 {
		return tips, nil
	}

	// Otherwise, try to generate a tip based on the first few cards
	if len(cards) == 0 {
		return tips, nil
	}

	// Generate tip for first new card
	for _, card := range cards {
		if card.IsNew {
			tip, err := s.GenerateGrammarForCard(ctx, &card.Card)
			if err == nil && tip != nil {
				tips = append(tips, *tip)
			}
			break
		}
	}

	return tips, nil
}

// GetAllGrammarRules returns all grammar rules organized by difficulty
func (s *GrammarService) GetAllGrammarRules() (map[int][]models.GrammarRule, error) {
	rules, err := repository.GetAllGrammarRules()
	if err != nil {
		return nil, err
	}

	// Group by difficulty
	byDifficulty := make(map[int][]models.GrammarRule)
	for _, rule := range rules {
		byDifficulty[rule.DifficultyLevel] = append(byDifficulty[rule.DifficultyLevel], rule)
	}

	return byDifficulty, nil
}

// GetGrammarRule gets a specific grammar rule by key
func (s *GrammarService) GetGrammarRule(ruleKey string) (*models.GrammarRule, []models.GrammarExample, error) {
	rule, err := repository.GetGrammarRuleByKey(ruleKey)
	if err != nil {
		return nil, nil, err
	}

	examples := parseGrammarExamples(rule.Examples)
	return rule, examples, nil
}

// parseGrammarExamples parses JSON examples string into slice
func parseGrammarExamples(examplesJSON string) []models.GrammarExample {
	var examples []models.GrammarExample
	if err := json.Unmarshal([]byte(examplesJSON), &examples); err != nil {
		return nil
	}
	return examples
}

// truncate truncates a string to max length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// min returns the smaller of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

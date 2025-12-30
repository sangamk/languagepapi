package handlers

import (
	"net/http"
	"strconv"

	"languagepapi/components"
	"languagepapi/internal/models"
	"languagepapi/internal/repository"
	"languagepapi/internal/service"
)

var grammarService = service.NewGrammarService()

// HandleGrammar renders the grammar index page
func HandleGrammar(w http.ResponseWriter, r *http.Request) {
	rulesByDifficulty, err := grammarService.GetAllGrammarRules()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Flatten the map to a slice for the component
	var rules []models.GrammarRule
	for _, levelRules := range rulesByDifficulty {
		rules = append(rules, levelRules...)
	}

	components.GrammarPage(rules).Render(r.Context(), w)
}

// HandleGrammarDetail renders a specific grammar rule
func HandleGrammarDetail(w http.ResponseWriter, r *http.Request) {
	ruleKey := r.PathValue("rule_key")
	if ruleKey == "" {
		http.Redirect(w, r, "/grammar", http.StatusSeeOther)
		return
	}

	rule, examples, err := grammarService.GetGrammarRule(ruleKey)
	if err != nil {
		http.Error(w, "Grammar rule not found", http.StatusNotFound)
		return
	}

	// Get related cards
	cards, _ := repository.GetCardsForGrammarRule(rule.ID)

	components.GrammarDetail(rule, examples, cards).Render(r.Context(), w)
}

// HandleGrammarForCard returns grammar explanation for a specific card (HTMX)
func HandleGrammarForCard(w http.ResponseWriter, r *http.Request) {
	cardIDStr := r.PathValue("card_id")
	if cardIDStr == "" {
		http.Error(w, "card_id required", http.StatusBadRequest)
		return
	}

	cardID, err := strconv.ParseInt(cardIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid card_id", http.StatusBadRequest)
		return
	}

	card, err := repository.GetCard(cardID)
	if err != nil {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	tip, err := grammarService.GetGrammarForCard(r.Context(), card)
	if err != nil {
		// Return empty tip if generation fails
		tip = &models.GrammarTip{
			Title:       "Grammar",
			ShortExplan: "Grammar information not available for this card.",
			Examples:    nil,
		}
	}

	components.GrammarTipComponent(tip).Render(r.Context(), w)
}

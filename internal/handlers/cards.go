package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"languagepapi/components"
	"languagepapi/internal/bridge"
	"languagepapi/internal/models"
	"languagepapi/internal/repository"
)

const cardsPerPage = 50

// HandleWords renders the words list page
func HandleWords(w http.ResponseWriter, r *http.Request) {
	// Parse query params
	pageStr := r.URL.Query().Get("page")
	islandStr := r.URL.Query().Get("island")
	searchQuery := strings.TrimSpace(r.URL.Query().Get("q"))

	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	filterIsland := int64(0)
	if i, err := strconv.ParseInt(islandStr, 10, 64); err == nil {
		filterIsland = i
	}

	// Get islands for filter dropdown
	islands, _ := repository.GetAllIslands()

	// Get cards
	var cards []models.Card
	var totalCards int
	var err error

	if searchQuery != "" {
		// Search mode
		cards, err = repository.SearchCards(searchQuery, filterIsland)
		totalCards = len(cards)
	} else if filterIsland > 0 {
		cards, err = repository.GetCardsByIsland(filterIsland)
		totalCards = len(cards)
		// Apply pagination manually for filtered results
		start := (page - 1) * cardsPerPage
		end := start + cardsPerPage
		if start > len(cards) {
			cards = nil
		} else if end > len(cards) {
			cards = cards[start:]
		} else {
			cards = cards[start:end]
		}
	} else {
		totalCards, _ = repository.CountCards()
		cards, err = repository.GetAllCards(cardsPerPage, (page-1)*cardsPerPage)
	}

	if err != nil {
		http.Error(w, "Failed to load words", http.StatusInternalServerError)
		return
	}

	totalPages := (totalCards + cardsPerPage - 1) / cardsPerPage
	if totalPages == 0 {
		totalPages = 1
	}

	components.WordsList(cards, islands, page, totalPages, totalCards, filterIsland, searchQuery).Render(r.Context(), w)
}

// HandleAddCard renders the add card form
func HandleAddCard(w http.ResponseWriter, r *http.Request) {
	islands, _ := repository.GetAllIslands()
	components.AddCard(islands, "", false).Render(r.Context(), w)
}

// HandleCreateCard processes the add card form
func HandleCreateCard(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	term := r.FormValue("term")
	translation := r.FormValue("translation")
	islandIDStr := r.FormValue("island_id")
	example := r.FormValue("example")
	bridgeHindi := r.FormValue("bridge_hindi")
	bridgeDutch := r.FormValue("bridge_dutch")
	bridgeEnglish := r.FormValue("bridge_english")
	generateBridges := r.FormValue("generate_bridges") == "on"

	if term == "" || translation == "" {
		islands, _ := repository.GetAllIslands()
		components.AddCardPartial(islands, "Term and translation are required", false).Render(r.Context(), w)
		return
	}

	islandID := int64(1) // Default to first island
	if id, err := strconv.ParseInt(islandIDStr, 10, 64); err == nil {
		islandID = id
	}

	// Create the card
	card := &models.Card{
		IslandID:        sql.NullInt64{Int64: islandID, Valid: true},
		Term:            term,
		Translation:     translation,
		ExampleSentence: example,
	}

	if err := repository.CreateCard(card); err != nil {
		islands, _ := repository.GetAllIslands()
		components.AddCardPartial(islands, "Failed to create card: "+err.Error(), false).Render(r.Context(), w)
		return
	}

	// Create bridges if provided manually
	hasBridges := false
	if bridgeHindi != "" {
		repository.CreateBridge(&models.Bridge{
			CardID:        card.ID,
			BridgeType:    models.BridgeHindiPhonetic,
			BridgeContent: bridgeHindi,
		})
		hasBridges = true
	}
	if bridgeDutch != "" {
		repository.CreateBridge(&models.Bridge{
			CardID:        card.ID,
			BridgeType:    models.BridgeDutchSyntax,
			BridgeContent: bridgeDutch,
		})
		hasBridges = true
	}
	if bridgeEnglish != "" {
		repository.CreateBridge(&models.Bridge{
			CardID:        card.ID,
			BridgeType:    models.BridgeEnglishCognate,
			BridgeContent: bridgeEnglish,
		})
		hasBridges = true
	}

	// Generate AI bridges if requested and no manual bridges provided
	if generateBridges && !hasBridges {
		go func(cardID int64) {
			ctx := context.Background()
			gemini, err := bridge.NewGeminiService(ctx)
			if err != nil {
				log.Printf("Failed to create Gemini service: %v", err)
				return
			}
			defer gemini.Close()

			if err := gemini.GenerateAndSaveBridges(ctx, cardID); err != nil {
				log.Printf("Failed to generate bridges for card %d: %v", cardID, err)
			} else {
				log.Printf("Generated bridges for card %d", cardID)
			}
		}(card.ID)
	}

	islands, _ := repository.GetAllIslands()
	components.AddCardPartial(islands, "Word added successfully!", true).Render(r.Context(), w)
}

// HandleEditCard renders the edit card form
func HandleEditCard(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	card, err := repository.GetCardWithBridges(id)
	if err != nil {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	islands, _ := repository.GetAllIslands()
	components.EditCard(card, islands, "", false).Render(r.Context(), w)
}

// HandleUpdateCard processes the edit card form
func HandleUpdateCard(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	term := r.FormValue("term")
	translation := r.FormValue("translation")
	islandIDStr := r.FormValue("island_id")
	example := r.FormValue("example")
	bridgeHindi := r.FormValue("bridge_hindi")
	bridgeDutch := r.FormValue("bridge_dutch")
	bridgeEnglish := r.FormValue("bridge_english")

	if term == "" || translation == "" {
		card, _ := repository.GetCardWithBridges(id)
		islands, _ := repository.GetAllIslands()
		components.EditCard(card, islands, "Term and translation are required", false).Render(r.Context(), w)
		return
	}

	islandID := int64(1)
	if iid, err := strconv.ParseInt(islandIDStr, 10, 64); err == nil {
		islandID = iid
	}

	card := &models.Card{
		ID:              id,
		IslandID:        sql.NullInt64{Int64: islandID, Valid: true},
		Term:            term,
		Translation:     translation,
		ExampleSentence: example,
	}

	if err := repository.UpdateCard(card); err != nil {
		card, _ := repository.GetCardWithBridges(id)
		islands, _ := repository.GetAllIslands()
		components.EditCard(card, islands, "Failed to update card: "+err.Error(), false).Render(r.Context(), w)
		return
	}

	// Update bridges - delete old ones first
	repository.DeleteBridgesForCard(id)

	if bridgeHindi != "" {
		repository.CreateBridge(&models.Bridge{
			CardID:        id,
			BridgeType:    models.BridgeHindiPhonetic,
			BridgeContent: bridgeHindi,
		})
	}
	if bridgeDutch != "" {
		repository.CreateBridge(&models.Bridge{
			CardID:        id,
			BridgeType:    models.BridgeDutchSyntax,
			BridgeContent: bridgeDutch,
		})
	}
	if bridgeEnglish != "" {
		repository.CreateBridge(&models.Bridge{
			CardID:        id,
			BridgeType:    models.BridgeEnglishCognate,
			BridgeContent: bridgeEnglish,
		})
	}

	// Redirect to words list
	w.Header().Set("HX-Redirect", "/words")
	w.WriteHeader(http.StatusOK)
}

// HandleSearchCards searches for cards
func HandleSearchCards(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	islandStr := r.URL.Query().Get("island")

	filterIsland := int64(0)
	if i, err := strconv.ParseInt(islandStr, 10, 64); err == nil {
		filterIsland = i
	}

	islands, _ := repository.GetAllIslands()

	if query == "" {
		// No search query, show regular list
		cards, _ := repository.GetAllCards(cardsPerPage, 0)
		totalCards, _ := repository.CountCards()
		totalPages := (totalCards + cardsPerPage - 1) / cardsPerPage
		if totalPages == 0 {
			totalPages = 1
		}
		components.WordsList(cards, islands, 1, totalPages, totalCards, filterIsland, query).Render(r.Context(), w)
		return
	}

	// Search cards
	cards, err := repository.SearchCards(query, filterIsland)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	components.WordsList(cards, islands, 1, 1, len(cards), filterIsland, query).Render(r.Context(), w)
}

// HandleGenerateBridges generates AI bridges for an existing card
func HandleGenerateBridges(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	gemini, err := bridge.NewGeminiService(ctx)
	if err != nil {
		http.Error(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer gemini.Close()

	// Delete existing bridges first
	repository.DeleteBridgesForCard(id)

	if err := gemini.GenerateAndSaveBridges(ctx, id); err != nil {
		http.Error(w, "Failed to generate bridges", http.StatusInternalServerError)
		return
	}

	// Return updated card bridges
	card, err := repository.GetCardWithBridges(id)
	if err != nil {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	components.CardBridges(card.Bridges).Render(r.Context(), w)
}

// HandleGenerateExample generates an AI example sentence for a card
func HandleGenerateExample(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	card, err := repository.GetCard(id)
	if err != nil {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	gemini, err := bridge.NewGeminiService(ctx)
	if err != nil {
		http.Error(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer gemini.Close()

	example, err := gemini.GenerateExampleSentence(ctx, card.Term, card.Translation)
	if err != nil {
		http.Error(w, "Failed to generate example", http.StatusInternalServerError)
		return
	}

	// Update the card with the example
	card.ExampleSentence = example
	if err := repository.UpdateCard(card); err != nil {
		http.Error(w, "Failed to save example", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(example))
}

// HandleDeleteCard deletes a card
func HandleDeleteCard(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := repository.DeleteCard(id); err != nil {
		http.Error(w, "Failed to delete card", http.StatusInternalServerError)
		return
	}

	// Return empty response to remove the element
	components.WordDeleted().Render(r.Context(), w)
}

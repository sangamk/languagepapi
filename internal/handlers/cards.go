package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"languagepapi/components"
	"languagepapi/internal/models"
	"languagepapi/internal/repository"
)

const cardsPerPage = 50

// HandleWords renders the words list page
func HandleWords(w http.ResponseWriter, r *http.Request) {
	// Parse query params
	pageStr := r.URL.Query().Get("page")
	islandStr := r.URL.Query().Get("island")

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

	if filterIsland > 0 {
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

	components.WordsList(cards, islands, page, totalPages, totalCards, filterIsland).Render(r.Context(), w)
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

	// Create bridges if provided
	if bridgeHindi != "" {
		repository.CreateBridge(&models.Bridge{
			CardID:        card.ID,
			BridgeType:    models.BridgeHindiPhonetic,
			BridgeContent: bridgeHindi,
		})
	}
	if bridgeDutch != "" {
		repository.CreateBridge(&models.Bridge{
			CardID:        card.ID,
			BridgeType:    models.BridgeDutchSyntax,
			BridgeContent: bridgeDutch,
		})
	}
	if bridgeEnglish != "" {
		repository.CreateBridge(&models.Bridge{
			CardID:        card.ID,
			BridgeType:    models.BridgeEnglishCognate,
			BridgeContent: bridgeEnglish,
		})
	}

	islands, _ := repository.GetAllIslands()
	components.AddCardPartial(islands, "Word added successfully!", true).Render(r.Context(), w)
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

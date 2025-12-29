package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"languagepapi/internal/db"
	"languagepapi/internal/models"
)

// GeminiService handles bridge generation using Gemini API
type GeminiService struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// BridgeResponse represents the JSON response from Gemini
type BridgeResponse struct {
	Hindi   *string `json:"hindi"`
	Dutch   *string `json:"dutch"`
	English *string `json:"english"`
}

// NewGeminiService creates a new Gemini service
func NewGeminiService(ctx context.Context) (*GeminiService, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-1.5-flash")
	model.SetTemperature(0.7)

	return &GeminiService{
		client: client,
		model:  model,
	}, nil
}

// Close closes the Gemini client
func (s *GeminiService) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

// GenerateBridges generates bridges for a single word
func (s *GeminiService) GenerateBridges(ctx context.Context, term, translation string) (*BridgeResponse, error) {
	prompt := fmt.Sprintf(`You are helping a polyglot learn Spanish. They speak English, Dutch, and Hindi.

Given the Spanish word "%s" meaning "%s":

Generate memory bridges to help remember this word:

1. Hindi Phonetic Bridge: Find phonetic similarity to Hindi words/sounds (use Devanagari if helpful). Only if genuinely useful.

2. Dutch Syntax Bridge: Identify grammatical patterns similar to Dutch (V2 rule, word order, gender, cognates). Only if genuinely useful.

3. English Cognate: Find English words with same Latin/Romance root or similar spelling/sound. Only if genuinely useful.

Respond ONLY with valid JSON in this exact format (use null for unhelpful bridges):
{"hindi": "...", "dutch": "...", "english": "..."}

Be concise - max 50 characters per bridge. Focus on the most memorable connection.`, term, translation)

	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	// Extract text response
	text := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if t, ok := part.(genai.Text); ok {
			text += string(t)
		}
	}

	// Clean up the response (remove markdown code blocks if present)
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	// Parse JSON
	var bridges BridgeResponse
	if err := json.Unmarshal([]byte(text), &bridges); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (response: %s)", err, text)
	}

	return &bridges, nil
}

// GenerateAndSaveBridges generates bridges for a card and saves them to the database
func (s *GeminiService) GenerateAndSaveBridges(ctx context.Context, cardID int64) error {
	// Get the card
	card, err := repository.GetCard(cardID)
	if err != nil {
		return fmt.Errorf("failed to get card: %w", err)
	}

	// Generate bridges
	bridges, err := s.GenerateBridges(ctx, card.Term, card.Translation)
	if err != nil {
		return fmt.Errorf("failed to generate bridges: %w", err)
	}

	// Save bridges to database
	if bridges.Hindi != nil && *bridges.Hindi != "" {
		_, err = db.DB.Exec(`
			INSERT INTO bridges (card_id, bridge_type, bridge_content) VALUES (?, ?, ?)
		`, cardID, models.BridgeHindiPhonetic, *bridges.Hindi)
		if err != nil {
			log.Printf("Failed to save Hindi bridge: %v", err)
		}
	}

	if bridges.Dutch != nil && *bridges.Dutch != "" {
		_, err = db.DB.Exec(`
			INSERT INTO bridges (card_id, bridge_type, bridge_content) VALUES (?, ?, ?)
		`, cardID, models.BridgeDutchSyntax, *bridges.Dutch)
		if err != nil {
			log.Printf("Failed to save Dutch bridge: %v", err)
		}
	}

	if bridges.English != nil && *bridges.English != "" {
		_, err = db.DB.Exec(`
			INSERT INTO bridges (card_id, bridge_type, bridge_content) VALUES (?, ?, ?)
		`, cardID, models.BridgeEnglishCognate, *bridges.English)
		if err != nil {
			log.Printf("Failed to save English bridge: %v", err)
		}
	}

	return nil
}

// GenerateExampleSentence generates an example sentence for a word
func (s *GeminiService) GenerateExampleSentence(ctx context.Context, term, translation string) (string, error) {
	prompt := fmt.Sprintf(`Generate a simple, natural Spanish sentence using the word "%s" (meaning: %s).

Rules:
- Use everyday, conversational Spanish
- Keep it under 15 words
- Include the word in a natural context
- Return ONLY the Spanish sentence, nothing else`, term, translation)

	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	text := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if t, ok := part.(genai.Text); ok {
			text += string(t)
		}
	}

	return strings.TrimSpace(text), nil
}

// GenerateHint generates a hint for a card when the user gets it wrong
func (s *GeminiService) GenerateHint(ctx context.Context, term, translation string, bridges []models.Bridge) (string, error) {
	bridgeInfo := ""
	for _, b := range bridges {
		bridgeInfo += fmt.Sprintf("- %s: %s\n", b.BridgeType, b.BridgeContent)
	}

	prompt := fmt.Sprintf(`The user is trying to remember the Spanish word "%s" meaning "%s".

Existing memory bridges:
%s

Generate a SHORT (1-2 sentence) hint to help them remember. Be creative and memorable.
Return ONLY the hint text.`, term, translation, bridgeInfo)

	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	text := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if t, ok := part.(genai.Text); ok {
			text += string(t)
		}
	}

	return strings.TrimSpace(text), nil
}

// GenerateCards generates vocabulary cards for a topic
func (s *GeminiService) GenerateCards(ctx context.Context, topic string, count int) ([]CardSuggestion, error) {
	prompt := fmt.Sprintf(`Generate %d Spanish vocabulary words for the topic: "%s"

Return ONLY valid JSON array in this format:
[{"term": "Spanish word", "translation": "English translation", "example": "Example sentence in Spanish"}]

Focus on practical, commonly used words.`, count, topic)

	resp, err := s.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	text := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		if t, ok := part.(genai.Text); ok {
			text += string(t)
		}
	}

	// Clean up the response
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var cards []CardSuggestion
	if err := json.Unmarshal([]byte(text), &cards); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return cards, nil
}

// CardSuggestion represents an AI-generated card suggestion
type CardSuggestion struct {
	Term        string `json:"term"`
	Translation string `json:"translation"`
	Example     string `json:"example"`
}

// GenerateBridgesForBatch generates bridges for multiple cards with rate limiting
func GenerateBridgesForBatch(ctx context.Context, limit int) error {
	service, err := NewGeminiService(ctx)
	if err != nil {
		return err
	}
	defer service.Close()

	// Get cards without bridges
	rows, err := db.DB.Query(`
		SELECT c.id, c.term, c.translation
		FROM cards c
		LEFT JOIN bridges b ON b.card_id = c.id
		WHERE b.id IS NULL
		ORDER BY c.frequency_rank ASC
		LIMIT ?
	`, limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	type cardInfo struct {
		ID          int64
		Term        string
		Translation string
	}

	var cards []cardInfo
	for rows.Next() {
		var c cardInfo
		if err := rows.Scan(&c.ID, &c.Term, &c.Translation); err != nil {
			continue
		}
		cards = append(cards, c)
	}

	log.Printf("Generating bridges for %d cards...", len(cards))

	for i, card := range cards {
		log.Printf("[%d/%d] %s - %s", i+1, len(cards), card.Term, card.Translation)

		err := service.GenerateAndSaveBridges(ctx, card.ID)
		if err != nil {
			log.Printf("  Error: %v", err)
		} else {
			log.Printf("  Done")
		}

		// Rate limiting: 15 requests per minute for free tier
		time.Sleep(4 * time.Second)
	}

	return nil
}

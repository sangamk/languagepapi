package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"languagepapi/internal/bridge"
	"languagepapi/internal/models"
	"languagepapi/internal/repository"
)

// QuestionService handles question generation and retrieval
type QuestionService struct{}

// NewQuestionService creates a new question service
func NewQuestionService() *QuestionService {
	return &QuestionService{}
}

// GetOrGenerateMCQ gets an existing MCQ or generates a new one
func (s *QuestionService) GetOrGenerateMCQ(ctx context.Context, card *models.Card) (*models.MCQData, error) {
	// Check if we already have an MCQ for this card
	existing, err := repository.GetQuestionByType(card.ID, models.QuestionMCQ)
	if err == nil && existing != nil {
		var mcq models.MCQData
		if err := json.Unmarshal([]byte(existing.QuestionData), &mcq); err == nil {
			return &mcq, nil
		}
	}

	// Generate new MCQ
	mcq, err := s.GenerateMCQ(ctx, card)
	if err != nil {
		return nil, err
	}

	// Save to database
	data, _ := json.Marshal(mcq)
	repository.SaveQuestion(card.ID, models.QuestionMCQ, string(data))

	return mcq, nil
}

// GenerateMCQ generates a multiple choice question using AI
func (s *QuestionService) GenerateMCQ(ctx context.Context, card *models.Card) (*models.MCQData, error) {
	gemini, err := bridge.NewGeminiService(ctx)
	if err != nil {
		// Fallback to simple MCQ without AI
		return s.generateSimpleMCQ(card), nil
	}
	defer gemini.Close()

	prompt := fmt.Sprintf(`Generate a multiple choice question for the Spanish vocabulary word.

Word: "%s"
Translation: "%s"
Example: "%s"

Create a question that tests understanding. Include:
1. A clear question stem
2. 4 options (1 correct, 3 plausible distractors)
3. Brief explanation of the correct answer

Return ONLY valid JSON in this exact format:
{"stem": "What does 'hablar' mean?", "options": ["to speak", "to walk", "to eat", "to sleep"], "correct_index": 0, "explanation": "Hablar means 'to speak' in Spanish."}

Keep options concise (1-4 words each). Make distractors plausible but clearly wrong.`,
		card.Term, card.Translation, card.ExampleSentence)

	response, err := gemini.GenerateContent(ctx, prompt)
	if err != nil {
		return s.generateSimpleMCQ(card), nil
	}

	// Parse response
	text := strings.TrimSpace(response)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var mcq models.MCQData
	if err := json.Unmarshal([]byte(text), &mcq); err != nil {
		return s.generateSimpleMCQ(card), nil
	}

	// Validate
	if len(mcq.Options) != 4 || mcq.CorrectIndex < 0 || mcq.CorrectIndex > 3 {
		return s.generateSimpleMCQ(card), nil
	}

	return &mcq, nil
}

// generateSimpleMCQ creates a basic MCQ without AI
func (s *QuestionService) generateSimpleMCQ(card *models.Card) *models.MCQData {
	// Simple translation MCQ
	distractors := []string{
		"something else", "another word", "different meaning",
	}

	options := make([]string, 4)
	correctIdx := rand.Intn(4)

	distIdx := 0
	for i := 0; i < 4; i++ {
		if i == correctIdx {
			options[i] = card.Translation
		} else {
			options[i] = distractors[distIdx%len(distractors)]
			distIdx++
		}
	}

	return &models.MCQData{
		Stem:         fmt.Sprintf("What does '%s' mean?", card.Term),
		Options:      options,
		CorrectIndex: correctIdx,
		Explanation:  fmt.Sprintf("'%s' means '%s' in English.", card.Term, card.Translation),
	}
}

// GetOrGenerateFillBlank gets an existing fill-blank or generates a new one
func (s *QuestionService) GetOrGenerateFillBlank(ctx context.Context, card *models.Card) (*models.FillBlankData, error) {
	existing, err := repository.GetQuestionByType(card.ID, models.QuestionFillBlank)
	if err == nil && existing != nil {
		var fb models.FillBlankData
		if err := json.Unmarshal([]byte(existing.QuestionData), &fb); err == nil {
			return &fb, nil
		}
	}

	fb, err := s.GenerateFillBlank(ctx, card)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(fb)
	repository.SaveQuestion(card.ID, models.QuestionFillBlank, string(data))

	return fb, nil
}

// GenerateFillBlank generates a fill-in-the-blank question using AI
func (s *QuestionService) GenerateFillBlank(ctx context.Context, card *models.Card) (*models.FillBlankData, error) {
	gemini, err := bridge.NewGeminiService(ctx)
	if err != nil {
		return s.generateSimpleFillBlank(card), nil
	}
	defer gemini.Close()

	prompt := fmt.Sprintf(`Create a fill-in-the-blank exercise for the Spanish word.

Word: "%s"
Translation: "%s"

Requirements:
- Create a natural Spanish sentence using this word
- Replace the target word with ____ (four underscores)
- Keep the sentence simple (5-10 words)
- Provide a hint and English translation

Return ONLY valid JSON:
{"sentence": "Yo ____ español.", "blank_position": 1, "answer": "hablo", "hint": "to speak (yo form)", "context": "I speak Spanish."}`,
		card.Term, card.Translation)

	response, err := gemini.GenerateContent(ctx, prompt)
	if err != nil {
		return s.generateSimpleFillBlank(card), nil
	}

	text := strings.TrimSpace(response)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var fb models.FillBlankData
	if err := json.Unmarshal([]byte(text), &fb); err != nil {
		return s.generateSimpleFillBlank(card), nil
	}

	return &fb, nil
}

// generateSimpleFillBlank creates a basic fill-blank without AI
func (s *QuestionService) generateSimpleFillBlank(card *models.Card) *models.FillBlankData {
	sentence := card.ExampleSentence
	if sentence == "" {
		sentence = fmt.Sprintf("____ es una palabra importante.", card.Term)
	} else {
		sentence = strings.Replace(sentence, card.Term, "____", 1)
	}

	return &models.FillBlankData{
		Sentence:      sentence,
		BlankPosition: 0,
		Answer:        card.Term,
		Hint:          card.Translation,
		Context:       card.Translation,
	}
}

// GetOrGenerateSentenceBuild gets an existing sentence-build or generates a new one
func (s *QuestionService) GetOrGenerateSentenceBuild(ctx context.Context, card *models.Card) (*models.SentenceBuildData, error) {
	existing, err := repository.GetQuestionByType(card.ID, models.QuestionSentenceBuild)
	if err == nil && existing != nil {
		var sb models.SentenceBuildData
		if err := json.Unmarshal([]byte(existing.QuestionData), &sb); err == nil {
			return &sb, nil
		}
	}

	sb, err := s.GenerateSentenceBuild(ctx, card)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(sb)
	repository.SaveQuestion(card.ID, models.QuestionSentenceBuild, string(data))

	return sb, nil
}

// GenerateSentenceBuild generates a sentence building question using AI
func (s *QuestionService) GenerateSentenceBuild(ctx context.Context, card *models.Card) (*models.SentenceBuildData, error) {
	gemini, err := bridge.NewGeminiService(ctx)
	if err != nil {
		return s.generateSimpleSentenceBuild(card), nil
	}
	defer gemini.Close()

	prompt := fmt.Sprintf(`Create a sentence building exercise for the Spanish word.

Word: "%s"
Translation: "%s"

Requirements:
- Create a simple Spanish sentence (5-8 words) using this word
- Provide the shuffled word bank
- Include English translation
- Optional hint to help

Return ONLY valid JSON:
{"target_sentence": "El gato negro duerme.", "word_bank": ["duerme", "El", "negro", "gato"], "translation": "The black cat sleeps.", "hint": "Start with the article"}`,
		card.Term, card.Translation)

	response, err := gemini.GenerateContent(ctx, prompt)
	if err != nil {
		return s.generateSimpleSentenceBuild(card), nil
	}

	text := strings.TrimSpace(response)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var sb models.SentenceBuildData
	if err := json.Unmarshal([]byte(text), &sb); err != nil {
		return s.generateSimpleSentenceBuild(card), nil
	}

	return &sb, nil
}

// generateSimpleSentenceBuild creates a basic sentence-build without AI
func (s *QuestionService) generateSimpleSentenceBuild(card *models.Card) *models.SentenceBuildData {
	sentence := card.ExampleSentence
	if sentence == "" {
		sentence = card.Term + " es importante."
	}

	// Split into words and shuffle
	words := strings.Fields(sentence)
	shuffled := make([]string, len(words))
	copy(shuffled, words)

	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return &models.SentenceBuildData{
		TargetSentence: sentence,
		WordBank:       shuffled,
		Translation:    card.Translation,
		Hint:           "Arrange the words to form a sentence",
	}
}

// GenerateQuestionsForCard generates all question types for a card (background task)
func (s *QuestionService) GenerateQuestionsForCard(cardID int64) {
	go func() {
		ctx := context.Background()

		card, err := repository.GetCard(cardID)
		if err != nil {
			log.Printf("Failed to get card %d for question generation: %v", cardID, err)
			return
		}

		// Generate MCQ
		time.Sleep(2 * time.Second) // Rate limiting
		if _, err := s.GetOrGenerateMCQ(ctx, card); err != nil {
			log.Printf("Failed to generate MCQ for card %d: %v", cardID, err)
		}

		// Generate Fill-blank
		time.Sleep(2 * time.Second)
		if _, err := s.GetOrGenerateFillBlank(ctx, card); err != nil {
			log.Printf("Failed to generate fill-blank for card %d: %v", cardID, err)
		}

		// Generate Sentence-build
		time.Sleep(2 * time.Second)
		if _, err := s.GetOrGenerateSentenceBuild(ctx, card); err != nil {
			log.Printf("Failed to generate sentence-build for card %d: %v", cardID, err)
		}

		log.Printf("Generated questions for card %d", cardID)
	}()
}

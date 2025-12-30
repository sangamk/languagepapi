package repository

import (
	"languagepapi/internal/db"
	"languagepapi/internal/models"
)

// SaveQuestion saves a generated question to the database
func SaveQuestion(cardID int64, questionType models.QuestionType, questionData string) (int64, error) {
	result, err := db.DB.Exec(`
		INSERT INTO questions (card_id, question_type, question_data)
		VALUES (?, ?, ?)
	`, cardID, questionType, questionData)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetQuestionsForCard retrieves all questions for a specific card
func GetQuestionsForCard(cardID int64) ([]models.Question, error) {
	rows, err := db.DB.Query(`
		SELECT id, card_id, question_type, question_data, created_at
		FROM questions
		WHERE card_id = ?
		ORDER BY created_at DESC
	`, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []models.Question
	for rows.Next() {
		var q models.Question
		if err := rows.Scan(&q.ID, &q.CardID, &q.QuestionType, &q.QuestionData, &q.CreatedAt); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, nil
}

// GetQuestionByType retrieves a question of a specific type for a card
func GetQuestionByType(cardID int64, questionType models.QuestionType) (*models.Question, error) {
	var q models.Question
	err := db.DB.QueryRow(`
		SELECT id, card_id, question_type, question_data, created_at
		FROM questions
		WHERE card_id = ? AND question_type = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, cardID, questionType).Scan(&q.ID, &q.CardID, &q.QuestionType, &q.QuestionData, &q.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// DeleteQuestionsForCard removes all questions for a card
func DeleteQuestionsForCard(cardID int64) error {
	_, err := db.DB.Exec(`DELETE FROM questions WHERE card_id = ?`, cardID)
	return err
}

// GetRandomQuestionByType gets a random question of a specific type from the pool
func GetRandomQuestionByType(questionType models.QuestionType, excludeCardIDs []int64) (*models.Question, error) {
	// Build query with exclusion list
	query := `
		SELECT id, card_id, question_type, question_data, created_at
		FROM questions
		WHERE question_type = ?
	`
	args := []interface{}{questionType}

	if len(excludeCardIDs) > 0 {
		query += " AND card_id NOT IN ("
		for i, id := range excludeCardIDs {
			if i > 0 {
				query += ","
			}
			query += "?"
			args = append(args, id)
		}
		query += ")"
	}

	query += " ORDER BY RANDOM() LIMIT 1"

	var q models.Question
	err := db.DB.QueryRow(query, args...).Scan(&q.ID, &q.CardID, &q.QuestionType, &q.QuestionData, &q.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &q, nil
}

// CountQuestionsForCard returns the count of questions for a card
func CountQuestionsForCard(cardID int64) (int, error) {
	var count int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM questions WHERE card_id = ?`, cardID).Scan(&count)
	return count, err
}

// HasQuestionType checks if a card has a question of a specific type
func HasQuestionType(cardID int64, questionType models.QuestionType) (bool, error) {
	var count int
	err := db.DB.QueryRow(`
		SELECT COUNT(*) FROM questions
		WHERE card_id = ? AND question_type = ?
	`, cardID, questionType).Scan(&count)
	return count > 0, err
}

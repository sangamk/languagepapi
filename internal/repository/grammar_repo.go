package repository

import (
	"database/sql"

	"languagepapi/internal/db"
	"languagepapi/internal/models"
)

// SaveGrammarRule saves a grammar rule to the database
func SaveGrammarRule(ruleKey, title, explanation, examples string, difficulty int) (int64, error) {
	// Try to insert, or update if exists
	result, err := db.DB.Exec(`
		INSERT INTO grammar_rules (rule_key, title, explanation, examples, difficulty_level)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(rule_key) DO UPDATE SET
			title = excluded.title,
			explanation = excluded.explanation,
			examples = excluded.examples,
			difficulty_level = excluded.difficulty_level
	`, ruleKey, title, explanation, examples, difficulty)
	if err != nil {
		return 0, err
	}

	// Get the ID (either new or existing)
	id, err := result.LastInsertId()
	if err != nil || id == 0 {
		// If conflict occurred, get the existing ID
		var existingID int64
		err = db.DB.QueryRow(`SELECT id FROM grammar_rules WHERE rule_key = ?`, ruleKey).Scan(&existingID)
		if err != nil {
			return 0, err
		}
		return existingID, nil
	}
	return id, nil
}

// LinkCardToGrammar links a card to a grammar rule
func LinkCardToGrammar(cardID, grammarRuleID int64) error {
	_, err := db.DB.Exec(`
		INSERT OR IGNORE INTO card_grammar (card_id, grammar_rule_id)
		VALUES (?, ?)
	`, cardID, grammarRuleID)
	return err
}

// GetGrammarForCard gets the grammar rule linked to a card
func GetGrammarForCard(cardID int64) (*models.GrammarRule, error) {
	var rule models.GrammarRule
	err := db.DB.QueryRow(`
		SELECT gr.id, gr.rule_key, gr.title, gr.explanation, gr.examples,
		       COALESCE(gr.related_cards, '[]'), gr.difficulty_level, gr.created_at
		FROM grammar_rules gr
		JOIN card_grammar cg ON cg.grammar_rule_id = gr.id
		WHERE cg.card_id = ?
		LIMIT 1
	`, cardID).Scan(
		&rule.ID, &rule.RuleKey, &rule.Title, &rule.Explanation,
		&rule.Examples, &rule.RelatedCards, &rule.DifficultyLevel, &rule.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rule, nil
}

// GetGrammarRuleByKey gets a grammar rule by its key
func GetGrammarRuleByKey(ruleKey string) (*models.GrammarRule, error) {
	var rule models.GrammarRule
	err := db.DB.QueryRow(`
		SELECT id, rule_key, title, explanation, examples,
		       COALESCE(related_cards, '[]'), difficulty_level, created_at
		FROM grammar_rules
		WHERE rule_key = ?
	`, ruleKey).Scan(
		&rule.ID, &rule.RuleKey, &rule.Title, &rule.Explanation,
		&rule.Examples, &rule.RelatedCards, &rule.DifficultyLevel, &rule.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetGrammarRuleByID gets a grammar rule by ID
func GetGrammarRuleByID(id int64) (*models.GrammarRule, error) {
	var rule models.GrammarRule
	err := db.DB.QueryRow(`
		SELECT id, rule_key, title, explanation, examples,
		       COALESCE(related_cards, '[]'), difficulty_level, created_at
		FROM grammar_rules
		WHERE id = ?
	`, id).Scan(
		&rule.ID, &rule.RuleKey, &rule.Title, &rule.Explanation,
		&rule.Examples, &rule.RelatedCards, &rule.DifficultyLevel, &rule.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetAllGrammarRules gets all grammar rules
func GetAllGrammarRules() ([]models.GrammarRule, error) {
	rows, err := db.DB.Query(`
		SELECT id, rule_key, title, explanation, examples,
		       COALESCE(related_cards, '[]'), difficulty_level, created_at
		FROM grammar_rules
		ORDER BY difficulty_level ASC, title ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.GrammarRule
	for rows.Next() {
		var rule models.GrammarRule
		if err := rows.Scan(
			&rule.ID, &rule.RuleKey, &rule.Title, &rule.Explanation,
			&rule.Examples, &rule.RelatedCards, &rule.DifficultyLevel, &rule.CreatedAt,
		); err != nil {
			continue
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// GetGrammarRulesByDifficulty gets rules of a specific difficulty
func GetGrammarRulesByDifficulty(difficulty int) ([]models.GrammarRule, error) {
	rows, err := db.DB.Query(`
		SELECT id, rule_key, title, explanation, examples,
		       COALESCE(related_cards, '[]'), difficulty_level, created_at
		FROM grammar_rules
		WHERE difficulty_level = ?
		ORDER BY title ASC
	`, difficulty)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.GrammarRule
	for rows.Next() {
		var rule models.GrammarRule
		if err := rows.Scan(
			&rule.ID, &rule.RuleKey, &rule.Title, &rule.Explanation,
			&rule.Examples, &rule.RelatedCards, &rule.DifficultyLevel, &rule.CreatedAt,
		); err != nil {
			continue
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// GetCardsForGrammarRule gets all cards linked to a grammar rule
func GetCardsForGrammarRule(grammarRuleID int64) ([]models.Card, error) {
	rows, err := db.DB.Query(`
		SELECT c.id, c.island_id, c.term, c.translation, c.example_sentence,
		       c.notes, c.audio_url, c.frequency_rank, c.created_at
		FROM cards c
		JOIN card_grammar cg ON cg.card_id = c.id
		WHERE cg.grammar_rule_id = ?
		ORDER BY c.frequency_rank ASC
	`, grammarRuleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		if err := rows.Scan(
			&card.ID, &card.IslandID, &card.Term, &card.Translation,
			&card.ExampleSentence, &card.Notes, &card.AudioURL,
			&card.FrequencyRank, &card.CreatedAt,
		); err != nil {
			continue
		}
		cards = append(cards, card)
	}
	return cards, nil
}

// CountGrammarRules returns the total number of grammar rules
func CountGrammarRules() (int, error) {
	var count int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM grammar_rules`).Scan(&count)
	return count, err
}

package repository

import (
	"database/sql"

	"languagepapi/internal/db"
	"languagepapi/internal/models"
)

// GetCard retrieves a card by ID
func GetCard(id int64) (*models.Card, error) {
	card := &models.Card{}
	err := db.DB.QueryRow(`
		SELECT id, island_id, term, translation,
		       COALESCE(example_sentence, ''), COALESCE(notes, ''), COALESCE(audio_url, ''),
		       frequency_rank, created_at
		FROM cards WHERE id = ?
	`, id).Scan(
		&card.ID, &card.IslandID, &card.Term, &card.Translation,
		&card.ExampleSentence, &card.Notes, &card.AudioURL, &card.FrequencyRank, &card.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return card, nil
}

// GetCardWithBridges retrieves a card with its bridges
func GetCardWithBridges(id int64) (*models.Card, error) {
	card, err := GetCard(id)
	if err != nil {
		return nil, err
	}

	rows, err := db.DB.Query(`
		SELECT id, card_id, bridge_type, bridge_content, explanation
		FROM bridges WHERE card_id = ?
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b models.Bridge
		if err := rows.Scan(&b.ID, &b.CardID, &b.BridgeType, &b.BridgeContent, &b.Explanation); err != nil {
			return nil, err
		}
		card.Bridges = append(card.Bridges, b)
	}
	return card, rows.Err()
}

// GetCardsByIsland retrieves all cards for an island
func GetCardsByIsland(islandID int64) ([]models.Card, error) {
	rows, err := db.DB.Query(`
		SELECT id, island_id, term, translation,
		       COALESCE(example_sentence, ''), COALESCE(notes, ''), COALESCE(audio_url, ''),
		       frequency_rank, created_at
		FROM cards WHERE island_id = ?
		ORDER BY frequency_rank ASC, id ASC
	`, islandID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var c models.Card
		if err := rows.Scan(
			&c.ID, &c.IslandID, &c.Term, &c.Translation,
			&c.ExampleSentence, &c.Notes, &c.AudioURL, &c.FrequencyRank, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

// CreateCard inserts a new card
func CreateCard(card *models.Card) error {
	result, err := db.DB.Exec(`
		INSERT INTO cards (island_id, term, translation, example_sentence, notes, audio_url, frequency_rank)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, card.IslandID, card.Term, card.Translation, card.ExampleSentence, card.Notes, card.AudioURL, card.FrequencyRank)
	if err != nil {
		return err
	}
	card.ID, err = result.LastInsertId()
	return err
}

// CreateBridge inserts a new bridge for a card
func CreateBridge(bridge *models.Bridge) error {
	result, err := db.DB.Exec(`
		INSERT INTO bridges (card_id, bridge_type, bridge_content, explanation)
		VALUES (?, ?, ?, ?)
	`, bridge.CardID, bridge.BridgeType, bridge.BridgeContent, bridge.Explanation)
	if err != nil {
		return err
	}
	bridge.ID, err = result.LastInsertId()
	return err
}

// DeleteCard removes a card and its bridges (cascades)
func DeleteCard(id int64) error {
	_, err := db.DB.Exec(`DELETE FROM cards WHERE id = ?`, id)
	return err
}

// UpdateCard updates an existing card
func UpdateCard(card *models.Card) error {
	_, err := db.DB.Exec(`
		UPDATE cards SET
			island_id = ?,
			term = ?,
			translation = ?,
			example_sentence = ?,
			notes = ?,
			audio_url = ?
		WHERE id = ?
	`, card.IslandID, card.Term, card.Translation, card.ExampleSentence, card.Notes, card.AudioURL, card.ID)
	return err
}

// DeleteBridgesForCard removes all bridges for a card
func DeleteBridgesForCard(cardID int64) error {
	_, err := db.DB.Exec(`DELETE FROM bridges WHERE card_id = ?`, cardID)
	return err
}

// SearchCards searches for cards by term or translation
func SearchCards(query string, islandID int64) ([]models.Card, error) {
	searchPattern := "%" + query + "%"
	var rows *sql.Rows
	var err error

	if islandID > 0 {
		rows, err = db.DB.Query(`
			SELECT id, island_id, term, translation,
			       COALESCE(example_sentence, ''), COALESCE(notes, ''), COALESCE(audio_url, ''),
			       frequency_rank, created_at
			FROM cards
			WHERE island_id = ? AND (term LIKE ? OR translation LIKE ? OR example_sentence LIKE ?)
			ORDER BY frequency_rank ASC, id ASC
			LIMIT 100
		`, islandID, searchPattern, searchPattern, searchPattern)
	} else {
		rows, err = db.DB.Query(`
			SELECT id, island_id, term, translation,
			       COALESCE(example_sentence, ''), COALESCE(notes, ''), COALESCE(audio_url, ''),
			       frequency_rank, created_at
			FROM cards
			WHERE term LIKE ? OR translation LIKE ? OR example_sentence LIKE ?
			ORDER BY frequency_rank ASC, id ASC
			LIMIT 100
		`, searchPattern, searchPattern, searchPattern)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var c models.Card
		if err := rows.Scan(
			&c.ID, &c.IslandID, &c.Term, &c.Translation,
			&c.ExampleSentence, &c.Notes, &c.AudioURL, &c.FrequencyRank, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

// GetBridgesForCard retrieves all bridges for a card
func GetBridgesForCard(cardID int64) ([]models.Bridge, error) {
	rows, err := db.DB.Query(`
		SELECT id, card_id, bridge_type, bridge_content, COALESCE(explanation, '')
		FROM bridges WHERE card_id = ?
	`, cardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bridges []models.Bridge
	for rows.Next() {
		var b models.Bridge
		if err := rows.Scan(&b.ID, &b.CardID, &b.BridgeType, &b.BridgeContent, &b.Explanation); err != nil {
			return nil, err
		}
		bridges = append(bridges, b)
	}
	return bridges, rows.Err()
}

// CountCards returns total card count
func CountCards() (int, error) {
	var count int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM cards`).Scan(&count)
	return count, err
}

// GetAllCards retrieves all cards with pagination
func GetAllCards(limit, offset int) ([]models.Card, error) {
	rows, err := db.DB.Query(`
		SELECT id, island_id, term, translation,
		       COALESCE(example_sentence, ''), COALESCE(notes, ''), COALESCE(audio_url, ''),
		       frequency_rank, created_at
		FROM cards
		ORDER BY frequency_rank ASC, id ASC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var c models.Card
		if err := rows.Scan(
			&c.ID, &c.IslandID, &c.Term, &c.Translation,
			&c.ExampleSentence, &c.Notes, &c.AudioURL, &c.FrequencyRank, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

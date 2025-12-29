package repository

import (
	"languagepapi/internal/db"
	"languagepapi/internal/models"
)

// GetAllIslands retrieves all islands
func GetAllIslands() ([]models.Island, error) {
	rows, err := db.DB.Query(`
		SELECT id, name, description, icon, unlock_xp, sort_order
		FROM islands ORDER BY sort_order ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var islands []models.Island
	for rows.Next() {
		var i models.Island
		if err := rows.Scan(&i.ID, &i.Name, &i.Description, &i.Icon, &i.UnlockXP, &i.SortOrder); err != nil {
			return nil, err
		}
		islands = append(islands, i)
	}
	return islands, rows.Err()
}

// GetUnlockedIslands retrieves islands the user has unlocked based on XP
func GetUnlockedIslands(userID int64) ([]models.Island, error) {
	rows, err := db.DB.Query(`
		SELECT i.id, i.name, i.description, i.icon, i.unlock_xp, i.sort_order
		FROM islands i
		INNER JOIN users u ON u.id = ?
		WHERE i.unlock_xp <= u.total_xp
		ORDER BY i.sort_order ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var islands []models.Island
	for rows.Next() {
		var i models.Island
		if err := rows.Scan(&i.ID, &i.Name, &i.Description, &i.Icon, &i.UnlockXP, &i.SortOrder); err != nil {
			return nil, err
		}
		islands = append(islands, i)
	}
	return islands, rows.Err()
}

// GetIsland retrieves a single island by ID
func GetIsland(id int64) (*models.Island, error) {
	i := &models.Island{}
	err := db.DB.QueryRow(`
		SELECT id, name, description, icon, unlock_xp, sort_order
		FROM islands WHERE id = ?
	`, id).Scan(&i.ID, &i.Name, &i.Description, &i.Icon, &i.UnlockXP, &i.SortOrder)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// GetIslandStats retrieves progress stats for an island
func GetIslandStats(userID, islandID int64) (*models.IslandStats, error) {
	island, err := GetIsland(islandID)
	if err != nil {
		return nil, err
	}

	stats := &models.IslandStats{Island: *island}

	// Total cards in island
	db.DB.QueryRow(`
		SELECT COUNT(*) FROM cards WHERE island_id = ?
	`, islandID).Scan(&stats.TotalCards)

	// Learned cards (have progress, not new)
	db.DB.QueryRow(`
		SELECT COUNT(*)
		FROM card_progress p
		INNER JOIN cards c ON c.id = p.card_id
		WHERE c.island_id = ? AND p.user_id = ? AND p.state != 'new'
	`, islandID, userID).Scan(&stats.LearnedCards)

	// Due cards
	db.DB.QueryRow(`
		SELECT COUNT(*)
		FROM card_progress p
		INNER JOIN cards c ON c.id = p.card_id
		WHERE c.island_id = ? AND p.user_id = ? AND p.due <= datetime('now') AND p.state IN ('learning', 'review', 'relearning')
	`, islandID, userID).Scan(&stats.DueCards)

	// Mastered cards (high stability)
	db.DB.QueryRow(`
		SELECT COUNT(*)
		FROM card_progress p
		INNER JOIN cards c ON c.id = p.card_id
		WHERE c.island_id = ? AND p.user_id = ? AND p.stability > 30 AND p.reps >= 5
	`, islandID, userID).Scan(&stats.MasteredCards)

	return stats, nil
}

// GetAllIslandsWithStats retrieves all islands with their stats
func GetAllIslandsWithStats(userID int64) ([]models.IslandStats, error) {
	islands, err := GetAllIslands()
	if err != nil {
		return nil, err
	}

	var stats []models.IslandStats
	for _, island := range islands {
		s, err := GetIslandStats(userID, island.ID)
		if err != nil {
			return nil, err
		}
		stats = append(stats, *s)
	}
	return stats, nil
}

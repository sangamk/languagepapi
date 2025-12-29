package repository

import (
	"encoding/json"

	"languagepapi/internal/db"
)

// UserSettings represents user preferences
type UserSettings struct {
	DailyGoal         int    `json:"daily_goal"`
	EnableTTS         bool   `json:"enable_tts"`
	ShowBridges       bool   `json:"show_bridges"`
	DefaultMode       string `json:"default_mode"`
	NewCardsPerDay    int    `json:"new_cards_per_day"`
	ReviewsPerSession int    `json:"reviews_per_session"`
}

// GetUserSettings retrieves settings for a user
func GetUserSettings(userID int64) (*UserSettings, error) {
	var settingsJSON string
	err := db.DB.QueryRow(`
		SELECT settings FROM user_settings WHERE user_id = ?
	`, userID).Scan(&settingsJSON)
	if err != nil {
		return nil, err
	}

	var settings UserSettings
	if err := json.Unmarshal([]byte(settingsJSON), &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// SaveUserSettings saves settings for a user
func SaveUserSettings(userID int64, settings *UserSettings) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	_, err = db.DB.Exec(`
		INSERT INTO user_settings (user_id, settings)
		VALUES (?, ?)
		ON CONFLICT(user_id) DO UPDATE SET settings = excluded.settings
	`, userID, string(settingsJSON))
	return err
}

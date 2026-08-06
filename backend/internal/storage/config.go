package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	// Add other imports as needed
)

func CreateConfig(db *sql.DB, configKey string, configValue interface{}, author, message string) (int, error) {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var configID int
	err = tx.QueryRow("SELECT id FROM configs WHERE config_key = $1", configKey).Scan(&configID)

	if err == sql.ErrNoRows {
		err = tx.QueryRow("INSERT INTO configs (config_key) VALUES ($1) RETURNING id", configKey).Scan(&configID)
		if err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	configValueJSON, err := json.Marshal(configValue)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal config value: %w", err)
	}

	var newVID int
	err = tx.QueryRow(
		"INSERT INTO versions (config_id, config_value, author, message) VALUES ($1, $2, $3, $4) RETURNING id",
		configID, configValueJSON, author, message,
	).Scan(&newVID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert version: %w", err)
	}

	err = tx.QueryRow(
		`INSERT INTO active_pointers (config_id, active_vid, last_updated)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (config_id) DO UPDATE SET active_vid = $2, last_updated = CURRENT_TIMESTAMP`,
		configID, newVID,
	).Scan()
	if err != nil {
		return 0, fmt.Errorf("failed to update active pointer: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newVID, nil
}

package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	// Add other imports as needed
)

type ConfigStore interface {
	CreateConfig(configKey string, configValue interface{}, author, message string) (int, error)
	RollbackConfig(configKey string, targetVID int) (int, error)	
	GetConfig(configKey string) ([]byte, error)
	ListVersions(configKey string) ([]Version, error)
}

// Version struct represents a single version of a config
type Version struct {
	ID        int
	ConfigID  int
	Value     []byte
	Timestamp string
	Author    string
	Message   string
	IsActive  bool
}

// Main transactional function to create a new config or update an existing one
// Returns the new version ID and an error if any
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
		"INSERT INTO versions (config_id, config_val, author, message) VALUES ($1, $2, $3, $4) RETURNING id",
		configID, configValueJSON, author, message,
	).Scan(&newVID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert version: %w", err)
	}

	_, err = tx.Exec(
		`INSERT INTO active_pointers (config_id, active_vid, last_updated)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (config_id) DO UPDATE SET active_vid = $2, last_updated = CURRENT_TIMESTAMP`,
		configID, newVID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update active pointer: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newVID, nil
}

// Fetches the active configuration value for a given config key
// Returns the configuration value as a byte slice and an error if any
func GetConfig(db *sql.DB, configKey string) ([]byte, error) {
	var configID int
	err := db.QueryRow("SELECT id FROM configs where config_key = $1", configKey).Scan(&configID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("config key not found: %s", configKey)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query config id: %w", err)
	}

	var activeVID int
	err = db.QueryRow("SELECT active_vid FROM active_pointers WHERE config_id = $1", configID).Scan(&activeVID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active version: %w", err)
	}

	var configVal []byte
	err = db.QueryRow("SELECT config_val FROM versions WHERE id = $1", activeVID).Scan(&configVal)
	if err != nil {
		return nil, fmt.Errorf("failed to query config value: %w", err)
	}

	return configVal, nil
}

// Shows all versions of a given config key, including the active version
// Returns a slice of Version structs and an error if any
func ListVersions(db *sql.DB, configKey string) ([]Version, error) {
	var configID int
	err := db.QueryRow("SELECT id FROM configs WHERE config_key = $1", configKey).Scan(&configID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("config key not found: %s", configKey)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query config id: %w", err)
	}

	var activeVID int
	err = db.QueryRow("SELECT active_vid FROM active_pointers WHERE config_id = $1", configID).Scan(&activeVID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active version: %w", err)
	}

	rows, err := db.Query(
		`SELECT id, config_id, config_val, timestamp, author, message 
		FROM versions WHERE config_id = $1 ORDER BY timestamp DESC`, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to query versions: %w", err)
	}
	defer rows.Close()

	var allVersions []Version
	for rows.Next() {
		var cv Version
		err := rows.Scan(&cv.ID, &cv.ConfigID, &cv.Value, &cv.Timestamp, &cv.Author, &cv.Message)
		if err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		cv.IsActive = (cv.ID == activeVID)
		allVersions = append(allVersions, cv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over versions: %w", err)
	}

	return allVersions, nil
}

// Rolls back the active version of a given config key to a specified version ID
// Returns the new active version ID and an error if any
func RollbackConfig(db *sql.DB, configKey string, targetVID int) (int, error) {

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var configID int
	err = db.QueryRow("SELECT id FROM configs WHERE config_key = $1", configKey).Scan(&configID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("config key not found: %s", configKey)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to query config id: %w", err)
	}

	var vExists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM versions WHERE id = $1 AND config_id = $2)", targetVID, configID).Scan(&vExists)
	if err != nil {
		return 0, fmt.Errorf("failed to check version existence: %w", err)
	}
	if !vExists {
		return 0, fmt.Errorf("version ID %d does not exist for config key %s", targetVID, configKey)
	}

	_, err = tx.Exec(
		`UPDATE active_pointers SET active_vid = $1, last_updated = CURRENT_TIMESTAMP WHERE config_id = $2`,
		targetVID, configID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to update active pointer: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return targetVID, nil
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (p *PostgresStore) CreateConfig(configKey string, configValue interface{}, author, message string) (int, error) {
	return CreateConfig(p.db, configKey, configValue, author, message)
}

func (p *PostgresStore) RollbackConfig(configKey string, targetVID int) (int, error) {
	return RollbackConfig(p.db, configKey, targetVID)
}

func (p *PostgresStore) GetConfig(configKey string) ([]byte, error) {
	return GetConfig(p.db, configKey)
}

func (p *PostgresStore) ListVersions(configKey string) ([]Version, error) {
	return ListVersions(p.db, configKey)
}
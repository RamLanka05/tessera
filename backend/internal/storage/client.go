package storage

import (
	"database/sql"
	"fmt"
)

func ValidateClient(db *sql.DB, clientID, clientSecret string) (bool, error) {
	
	var storedSecret string
	err := db.QueryRow("SELECT client_secret FROM clients WHERE client_id = $1", clientID).Scan(&storedSecret)

	if err != nil {
		if err == sql.ErrNoRows {
			// Client not found - return false, nil (not an error per Option A)
			return false, nil
		}
		// Database error - return the error
		return false, fmt.Errorf("database error: %w", err)
	}
	
	return storedSecret == clientSecret, nil
}
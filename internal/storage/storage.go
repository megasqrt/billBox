package storage

import "billBox/internal/storage/postgres"

// New is a factory function that creates a new storage instance.
func New(databaseURI string) (Storage, error) {
	// Currently, it only supports PostgreSQL.
	// In the future, we could select the implementation based on a config value.
	return postgres.New(databaseURI)
}
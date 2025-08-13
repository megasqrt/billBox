package postgres

import (
	"billBox/internal/models"
	"context"
	"database/sql"
	"errors"
)

// CreateUser saves a new user to the database and populates the user struct with generated fields.
func (db *DB) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, created_at, current_balance`
	err := db.conn.QueryRowContext(ctx, query, user.Login, user.PasswordHash).Scan(&user.ID, &user.CreatedAt, &user.CurrentBalance)
	if err != nil {
		// A proper way to handle this would be to check for pq.ErrUniqueViolation,
		// but we avoid a direct dependency on the driver's error types.
		// The service layer should call FindByLogin first to ensure login is unique.
		return err
	}
	return nil
}

// FindByLogin finds a user by their login.
func (db *DB) FindByLogin(ctx context.Context, login string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, login, password_hash, current_balance, created_at FROM users WHERE login = $1`
	err := db.conn.QueryRowContext(ctx, query, login).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CurrentBalance, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found is not an error in this context
		}
		return nil, err
	}
	return user, nil
}
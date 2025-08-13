package postgres

import (
	"billBox/internal/models"
	"context"
)

// Get retrieves the current and withdrawn balance for a user.
func (db *DB) Get(ctx context.Context, userID int64) (*models.Balance, error) {
	balance := &models.Balance{}

	// Get current balance from users table
	queryCurrent := `SELECT current_balance FROM users WHERE id = $1`
	err := db.conn.QueryRowContext(ctx, queryCurrent, userID).Scan(&balance.Current)
	if err != nil {
		return nil, err
	}

	// Get total withdrawn amount from withdrawals table
	queryWithdrawn := `SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1`
	err = db.conn.QueryRowContext(ctx, queryWithdrawn, userID).Scan(&balance.Withdrawn)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

// Withdraw deducts points from the user's balance and records the withdrawal.
// This is executed within a transaction to ensure atomicity.
func (db *DB) Withdraw(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback is a no-op if the transaction is committed.

	var currentBalance float64
	err = tx.QueryRowContext(ctx, "SELECT current_balance FROM users WHERE id = $1 FOR UPDATE", userID).Scan(&currentBalance)
	if err != nil {
		return err
	}

	if currentBalance < sum {
		return models.ErrInsufficientFunds
	}

	_, err = tx.ExecContext(ctx, "UPDATE users SET current_balance = current_balance - $1 WHERE id = $2", sum, userID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3)", userID, orderNumber, sum)
	if err != nil {
		return err
	}

	return tx.Commit()
}
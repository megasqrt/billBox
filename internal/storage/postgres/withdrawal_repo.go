package postgres

import (
	"billBox/internal/models"
	"context"
)

// FindForUser (WithdrawalRepository) finds all withdrawals for a given user.
func (db *DB) WithdrawalFindForUser(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	query := `SELECT id, user_id, order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`
	rows, err := db.conn.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []models.Withdrawal
	for rows.Next() {
		var w models.Withdrawal
		if err := rows.Scan(&w.ID, &w.UserID, &w.OrderNumber, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}

	return withdrawals, rows.Err()
}
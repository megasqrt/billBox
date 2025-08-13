package postgres

import (
	"billBox/internal/models"
	"context"
	"database/sql"
	"errors"
)

// CreateOrder saves a new order to the database.
func (db *DB) CreateOrder(ctx context.Context, order *models.Order) error {
	query := `INSERT INTO orders (user_id, number, status, uploaded_at) VALUES ($1, $2, $3, $4) RETURNING id`
	return db.conn.QueryRowContext(ctx, query, order.UserID, order.Number, order.Status, order.UploadedAt).Scan(&order.ID)
}

// FindByNumber finds an order by its number.
func (db *DB) FindByNumber(ctx context.Context, number string) (*models.Order, error) {
	order := &models.Order{}
	query := `SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE number = $1`
	err := db.conn.QueryRowContext(ctx, query, number).Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found is not an error
		}
		return nil, err
	}
	return order, nil
}

// FindForUser (OrderRepository) finds all orders for a given user.
func (db *DB) OrderFindForUser(ctx context.Context, userID int64) ([]models.Order, error) {
	// Spec: "отсортированы по времени загрузки от самых новых к самым старым" -> DESC
	query := `SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`
	rows, err := db.conn.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Number, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
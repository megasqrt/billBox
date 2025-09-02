package postgres

import (
	"billBox/internal/models"
	"context"
	"database/sql"
	"errors"
)

// CreateOrder сохраняет новый заказ в БД.
func (db *DB) CreateOrder(ctx context.Context, order *models.Order) error {
	query := `INSERT INTO orders (user_id, number, status, uploaded_at) VALUES ($1, $2, $3, $4)`
	_, err := db.conn.ExecContext(ctx, query, order.UserID, order.Number, order.Status, order.UploadedAt)
	return err
}

// FindByNumber находит заказ по его номеру.
func (db *DB) FindByNumber(ctx context.Context, number string) (*models.Order, error) {
	order := &models.Order{}
	query := `SELECT id, user_id, number, status, accrual, uploaded_at FROM orders WHERE number = $1`
	var accrual sql.NullFloat64
	err := db.conn.QueryRowContext(ctx, query, number).Scan(&order.ID, &order.UserID, &order.Number, &order.Status, &accrual, &order.UploadedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found is not an error here, handler logic depends on nil
		}
		return nil, err
	}
	if accrual.Valid {
		accrualVal := accrual.Float64
		order.Accrual = &accrualVal
	}
	return order, nil
}

// OrderFindForUser находит все заказы для указанного пользователя.
func (db *DB) OrderFindForUser(ctx context.Context, userID int64) ([]models.Order, error) {
	query := `SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`
	rows, err := db.conn.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var accrual sql.NullFloat64
		if err := rows.Scan(&order.Number, &order.Status, &accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		if accrual.Valid {
			accrualVal := accrual.Float64
			order.Accrual = &accrualVal
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

// FindProcessableOrders находит все заказы со статусом NEW или PROCESSING.
func (db *DB) FindProcessableOrders(ctx context.Context) ([]models.Order, error) {
	query := `SELECT number, status FROM orders WHERE status = 'NEW' OR status = 'PROCESSING'`
	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		if err := rows.Scan(&order.Number, &order.Status); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

// UpdateOrderAccrual обновляет статус заказа и начисленные баллы.
// Если статус PROCESSED, также обновляет баланс пользователя в рамках транзакции.
func (db *DB) UpdateOrderAccrual(ctx context.Context, number string, status models.OrderStatus, accrual float64) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3`, status, accrual, number)
	if err != nil {
		return err
	}

	if status == models.OrderStatusProcessed && accrual > 0 {
		_, err = tx.ExecContext(ctx, `UPDATE users u SET current_balance = u.current_balance + $1 FROM orders o WHERE u.id = o.user_id AND o.number = $2`, accrual, number)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
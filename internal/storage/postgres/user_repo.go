package postgres

import (
	"billBox/internal/models"
	"context"
)

// CreateUser сохраняет нового пользователя в БД и заполняет структуру пользователя сгенерированными полями.
func (db *DB) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, created_at, current_balance`
	err := db.conn.QueryRowContext(ctx, query, user.Login, user.PasswordHash).Scan(&user.ID, &user.CreatedAt, &user.CurrentBalance)
	if err != nil {
		//TODO
		// Слой-обработчик (handler) должен проверять ошибку на нарушение
		// ограничения уникальности (например, по коду ошибки PostgreSQL "23505").
		return err
	}
	return nil
}

// FindByLogin находит пользователя по его логину.
func (db *DB) FindByLogin(ctx context.Context, login string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, login, password_hash, current_balance, created_at FROM users WHERE login = $1`
	err := db.conn.QueryRowContext(ctx, query, login).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CurrentBalance, &user.CreatedAt)
	if err != nil {
		
		// if errors.Is(err, sql.ErrNoRows) {
		// 	return nil, nil 
		// }
		// return nil, err

		return nil, err // Возвращаем ошибку как есть, включая sql.ErrNoRows
	}
	return user, nil
}
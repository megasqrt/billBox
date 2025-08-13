package storage

import (
	"billBox/internal/models"
	"context"
)

// UserRepository определяет методы для работы с пользователями в хранилище.
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	FindByLogin(ctx context.Context, login string) (*models.User, error)
}

// OrderRepository определяет методы для работы с заказами в хранилище.
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	FindByNumber(ctx context.Context, number string) (*models.Order, error)
	OrderFindForUser(ctx context.Context, userID int64) ([]models.Order, error)
}

// WithdrawalRepository определяет методы для работы со списаниями в хранилище.
type WithdrawalRepository interface {
	WithdrawalFindForUser(ctx context.Context, userID int64) ([]models.Withdrawal, error)
}

// BalanceRepository определяет методы для работы с балансом пользователя.
type BalanceRepository interface {
	Get(ctx context.Context, userID int64) (*models.Balance, error)
	Withdraw(ctx context.Context, userID int64, orderNumber string, sum float64) error
}

// Storage - это интерфейс, который должен быть реализован хранилищем данных.
// Он объединяет в себе все репозитории.
// В main.go `db` будет иметь тип, реализующий этот интерфейс.
type Storage interface {
	UserRepository
	OrderRepository
	WithdrawalRepository
	BalanceRepository
	// Метод Close() нужен для корректного закрытия соединения с БД.
	Close()
}
package storage

import (
	"billBox/internal/models"
	"context"
)

// Storage определяет интерфейс для взаимодействия с хранилищем данных.
type Storage interface {
	// Методы для пользователей
	CreateUser(ctx context.Context, user *models.User) error
	FindByLogin(ctx context.Context, login string) (*models.User, error)
	Get(ctx context.Context, userID int64) (*models.Balance, error)

	// Методы для заказов
	CreateOrder(ctx context.Context, order *models.Order) error
	FindByNumber(ctx context.Context, number string) (*models.Order, error)
	OrderFindForUser(ctx context.Context, userID int64) ([]models.Order, error)
	FindProcessableOrders(ctx context.Context) ([]models.Order, error)
	UpdateOrderAccrual(ctx context.Context, number string, status models.OrderStatus, accrual float64) error

	// Методы для списаний
	Withdraw(ctx context.Context, userID int64, orderNumber string, sum float64) error
	WithdrawalFindForUser(ctx context.Context, userID int64) ([]models.Withdrawal, error)

	Close()
}
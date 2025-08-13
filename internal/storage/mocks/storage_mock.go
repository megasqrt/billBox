package mocks

import (
	"billBox/internal/models"
	"context"

	"github.com/stretchr/testify/mock"
)

// Storage is a mock type for the storage.Storage interface
type Storage struct {
	mock.Mock
}

// CreateUser mocks the CreateUser method
func (m *Storage) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// FindByLogin mocks the FindByLogin method
func (m *Storage) FindByLogin(ctx context.Context, login string) (*models.User, error) {
	args := m.Called(ctx, login)
	var r0 *models.User
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.User)
	}
	return r0, args.Error(1)
}

// CreateOrder mocks the CreateOrder method
func (m *Storage) CreateOrder(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

// FindByNumber mocks the FindByNumber method
func (m *Storage) FindByNumber(ctx context.Context, number string) (*models.Order, error) {
	args := m.Called(ctx, number)
	var r0 *models.Order
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.Order)
	}
	return r0, args.Error(1)
}

// OrderFindForUser mocks the OrderFindForUser method
func (m *Storage) OrderFindForUser(ctx context.Context, userID int64) ([]models.Order, error) {
	args := m.Called(ctx, userID)
	var r0 []models.Order
	if args.Get(0) != nil {
		r0 = args.Get(0).([]models.Order)
	}
	return r0, args.Error(1)
}

// WithdrawalFindForUser mocks the WithdrawalFindForUser method
func (m *Storage) WithdrawalFindForUser(ctx context.Context, userID int64) ([]models.Withdrawal, error) {
	args := m.Called(ctx, userID)
	var r0 []models.Withdrawal
	if args.Get(0) != nil {
		r0 = args.Get(0).([]models.Withdrawal)
	}
	return r0, args.Error(1)
}

// Get mocks the Get method for balance
func (m *Storage) Get(ctx context.Context, userID int64) (*models.Balance, error) {
	args := m.Called(ctx, userID)
	var r0 *models.Balance
	if args.Get(0) != nil {
		r0 = args.Get(0).(*models.Balance)
	}
	return r0, args.Error(1)
}

// FindProcessableOrders mocks the FindProcessableOrders method
func (m *Storage) FindProcessableOrders(ctx context.Context) ([]models.Order, error) {
	args := m.Called(ctx)
	var r0 []models.Order
	if args.Get(0) != nil {
		r0 = args.Get(0).([]models.Order)
	}
	return r0, args.Error(1)
}

// UpdateOrderAccrual mocks the UpdateOrderAccrual method
func (m *Storage) UpdateOrderAccrual(ctx context.Context, number string, status models.OrderStatus, accrual float64) error {
	args := m.Called(ctx, number, status, accrual)
	return args.Error(0)
}

// Withdraw mocks the Withdraw method
func (m *Storage) Withdraw(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	args := m.Called(ctx, userID, orderNumber, sum)
	return args.Error(0)
}

// Close mocks the Close method
func (m *Storage) Close() {
	m.Called()
}

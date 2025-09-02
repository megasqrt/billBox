package models

import (
	"errors"
	"time"
)

// AuthRequest представляет JSON-тело для запросов регистрации и аутентификации.
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// User представляет пользователя в базе данных.
type User struct {
	ID           int64
	Login        string
	PasswordHash string
	CurrentBalance float64 //TODO decimal
	CreatedAt      time.Time
}

// OrderStatus представляет статус обработки заказа.
type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

// Order представляет заказ в базе данных.
type Order struct {
	ID         int64       `json:"-"`
	UserID     int64       `json:"-"`
	Number     string      `json:"number"`
	Status     OrderStatus `json:"status"`
	Accrual    *float64    `json:"accrual,omitempty"` // Указатель для обработки null-значений //TODO decimal
	UploadedAt time.Time   `json:"uploaded_at"`
}

// Balance представляет баланс пользователя.
type Balance struct {
	Current   float64 `json:"current"` //TODO decimal
	Withdrawn float64 `json:"withdrawn"` //TODO decimal
}

// WithdrawRequest представляет JSON-тело для запроса на списание средств.
type WithdrawRequest struct {
	OrderNumber string  `json:"order"` // Тег 'order' соответствует спецификации
	Sum         float64 `json:"sum"`   // Тег 'sum' соответствует спецификации //TODO decimal
}

// Withdrawal представляет запись о списании в базе данных.
type Withdrawal struct {
	ID          int64     `json:"-"`
	UserID      int64     `json:"-"`
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"` //TODO decimal
	ProcessedAt time.Time `json:"processed_at"`
}

// ErrInsufficientFunds - ошибка, возвращаемая при недостатке средств на счете.
var ErrInsufficientFunds = errors.New("insufficient funds")
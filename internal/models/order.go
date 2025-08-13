package models

import "time"

// OrderStatus представляет собой статус обработки заказа.
type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

// Order представляет заказ, загруженный пользователем.
type Order struct {
	ID         int64       `json:"-"`
	UserID     int64       `json:"-"`
	Number     string      `json:"number"`
	Status     OrderStatus `json:"status"`
	Accrual    *float64    `json:"accrual,omitempty"` // Указатель для обработки null-значений
	UploadedAt time.Time   `json:"uploaded_at"`
}
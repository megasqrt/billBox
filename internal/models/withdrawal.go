package models

import "time"

// Withdrawal представляет модель списания баллов.
// Поля с тегами `json` используются для сериализации в ответе API.
type Withdrawal struct {
	ID          int64
	UserID      int64
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

// WithdrawRequest используется для десериализации запроса на списание.
type WithdrawRequest struct {
	OrderNumber string  `json:"order"`
	Sum         float64 `json:"sum"`
}
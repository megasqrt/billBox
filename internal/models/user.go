package models

import "time"

// User представляет модель пользователя в системе.
type User struct {
	ID             int64
	Login          string
	PasswordHash   string
	CurrentBalance float64
	CreatedAt      time.Time
}

// Balance представляет баланс пользователя для ответа API.
type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

// AuthRequest используется для десериализации запроса на регистрацию/логин.
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
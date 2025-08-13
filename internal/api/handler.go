package api

import (
	"billBox/internal/storage"
	"log"
)

// Handler содержит зависимости, необходимые для обработчиков API.
type Handler struct {
	storage      storage.Storage
	logger       *log.Logger
	jwtSecretKey string
}

// NewHandler создает новый экземпляр Handler.
func NewHandler(s storage.Storage, l *log.Logger, jwtKey string) *Handler {
	return &Handler{
		storage:      s,
		logger:       l,
		jwtSecretKey: jwtKey,
	}
}
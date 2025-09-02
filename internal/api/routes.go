package api

import (
	"billBox/internal/config"
	"billBox/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

// NewRouter создаёт новый роутер и регистрирует все хендлеры.
func NewRouter(db storage.Storage, logger *log.Logger, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// Используем стандартные middleware для логирования, восстановления после паник и т.д.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Создаем экземпляр хендлера с зависимостями
	h := NewHandler(db, logger, cfg.JWTSecretKey)

	r.Route("/api/user", func(r chi.Router) {
		// Публичные роуты
		r.Post("/register", h.RegisterUser)
		r.Post("/login", h.LoginUser)

		// Роуты, требующие аутентификации
		r.Group(func(r chi.Router) {
			r.Use(h.AuthMiddleware) // Middleware для проверки аутентификации

			r.Post("/orders", h.UploadOrder)
			r.Get("/orders", h.GetUserOrders)
			r.Get("/balance", h.GetUserBalance)
			r.Post("/balance/withdraw", h.Withdraw)
			r.Get("/withdrawals", h.GetUserWithdrawals)
		})
	})

	return r
}
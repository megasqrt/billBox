package api

import (
	"billBox/internal/auth"
	"context"
	"net/http"
)

// Определяем ключ для контекста, чтобы избежать коллизий.
type contextKey string

const userContextKey = contextKey("userID")

// AuthMiddleware - это middleware для проверки аутентификации пользователя.
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Получить токен из cookie.
		cookie, err := r.Cookie("token")
		if err != nil {
			// Если cookie не установлен, пользователь не аутентифицирован.
			http.Error(w, "Unauthorized: no token cookie", http.StatusUnauthorized)
			return
		}

		tokenString := cookie.Value

		// 2. Проверить валидность токена и извлечь ID пользователя.
		userID, err := auth.GetUserID(tokenString, h.jwtSecretKey)
		if err != nil {
			// Если токен невалиден, пользователь не аутентифицирован.
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		// 3. Добавить ID пользователя в контекст запроса.
		ctx := context.WithValue(r.Context(), userContextKey, userID)

		// 4. Передать управление следующему обработчику с обновленным контекстом.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext извлекает ID пользователя из контекста запроса.
// Это вспомогательная функция для использования в других хендлерах.
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userContextKey).(int64)
	return userID, ok
}
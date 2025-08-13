package api

import (
	"billBox/internal/auth"
	"billBox/internal/models"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

// RegisterUser обрабатывает запрос на регистрацию нового пользователя.
// POST /api/user/register
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	// 1. Распарсить JSON из тела запроса в структуру {login, password}.
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// 2. Проверить, что логин и пароль не пустые.
	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	// 3. Проверить, что логин еще не занят (запрос к БД).
	// 4. Если занят, вернуть 409 Conflict. (Это обрабатывается ниже при ошибке создания)

	// 5. Если свободен, хешировать пароль.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Printf("Error hashing password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 6. Сохранить нового пользователя в БД.
	newUser := &models.User{
		Login:        req.Login,
		PasswordHash: string(hashedPassword),
	}

	if err := h.storage.CreateUser(r.Context(), newUser); err != nil {
		var pgErr *pgconn.PgError
		// Проверяем на ошибку уникальности (код 23505 для PostgreSQL)
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			http.Error(w, "Login already taken", http.StatusConflict)
			return
		}
		h.logger.Printf("Error creating user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 7. Сгенерировать JWT или другой токен аутентификации.
	tokenString, err := auth.BuildJWTString(newUser.ID, h.jwtSecretKey)
	if err != nil {
		h.logger.Printf("Error building JWT for user %d: %v", newUser.ID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 8. Установить токен в http-only cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  time.Now().Add(12 * time.Hour),
		HttpOnly: true,
	})

	// 9. Вернуть 200 OK.
	w.WriteHeader(http.StatusOK)
}

// LoginUser обрабатывает запрос на аутентификацию пользователя.
// POST /api/user/login
func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	// 1. Распарсить JSON из тела запроса.
	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// 2. Проверить, что логин и пароль не пустые.
	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	// 3. Найти пользователя в БД по логину.
	user, err := h.storage.FindByLogin(r.Context(), req.Login)
	if err != nil {
		// Если пользователь не найден, возвращаем 401.
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid login or password", http.StatusUnauthorized)
			return
		}
		// В случае другой ошибки БД - 500.
		h.logger.Printf("Error finding user by login %s: %v", req.Login, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 4. Сравнить хеш пароля из БД с паролем из запроса.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}

	// 5. Сгенерировать JWT и установить cookie.
	tokenString, err := auth.BuildJWTString(user.ID, h.jwtSecretKey)
	if err != nil {
		h.logger.Printf("Error building JWT for user %d: %v", user.ID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "token", Value: tokenString, Expires: time.Now().Add(12 * time.Hour), HttpOnly: true})
	w.WriteHeader(http.StatusOK)
}
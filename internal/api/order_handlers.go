package api

import (
	"billBox/internal/helper"
	"billBox/internal/models"
	"errors"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// UploadOrder обрабатывает загрузку номера заказа.
// POST /api/user/orders
func (h *Handler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	// 1. Получить ID пользователя из контекста (после AuthMiddleware).
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Прочитать номер заказа из тела запроса (Content-Type: text/plain).
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Printf("Error reading request body: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	orderNumber := string(body)
	if orderNumber == "" {
		http.Error(w, "Invalid request format: empty order number", http.StatusBadRequest)
		return
	}

	// 3. Проверить номер заказа с помощью алгоритма Луна.
	if !helper.IsValid(orderNumber) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	// 4 & 5. Проверить, не был ли этот номер заказа загружен ранее.
	existingOrder, err := h.storage.FindByNumber(r.Context(), orderNumber)
	if err != nil {
		h.logger.Printf("Error finding order by number %s: %v", orderNumber, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existingOrder != nil {
		if existingOrder.UserID == userID {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "Order number already uploaded by another user", http.StatusConflict)
		return
	}

	// 6. Сохранить новый заказ в БД со статусом 'NEW'.
	newOrder := models.Order{
		UserID:     userID,
		Number:     orderNumber,
		Status:     models.OrderStatusNew,
		UploadedAt: time.Now(),
	}

	if err := h.storage.CreateOrder(r.Context(), &newOrder); err != nil {
		h.logger.Printf("Error creating order %s for user %d: %v", orderNumber, userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 7. Вернуть 202 Accepted.
	w.WriteHeader(http.StatusAccepted)
}

// GetUserOrders получает список заказов пользователя.
// GET /api/user/orders
func (h *Handler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	// 1. Получить ID пользователя из контекста.
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Получить из БД все заказы этого пользователя.
	// Метод OrderFindForUser уже сортирует их по времени загрузки (от новых к старым).
	orders, err := h.storage.OrderFindForUser(r.Context(), userID)
	if err != nil {
		h.logger.Printf("Error fetching orders for user %d: %v", userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 3. Если заказов нет, вернуть 204 No Content.
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// 4 & 5. Сериализовать и отправить список заказов.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		h.logger.Printf("Error encoding orders to JSON for user %d: %v", userID, err)
	}
}

// GetUserBalance получает текущий баланс пользователя.
// GET /api/user/balance
func (h *Handler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	// 1. Получить ID пользователя из контекста.
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2 & 3. Получить из БД текущий баланс и сумму списаний.
	balance, err := h.storage.Get(r.Context(), userID)
	if err != nil {
		h.logger.Printf("Error getting balance for user %d: %v", userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 4 & 5. Сформировать и отправить JSON-ответ.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(balance); err != nil {
		h.logger.Printf("Error encoding balance to JSON for user %d: %v", userID, err)
	}
}

// Withdraw обрабатывает запрос на списание баллов.
// POST /api/user/balance/withdraw
func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	// 1. Получить ID пользователя из контекста.
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Распарсить JSON из тела запроса.
	var req models.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// 3. Проверить номер заказа по алгоритму Луна.
	if !helper.IsValid(req.OrderNumber) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	// 4 & 5. Вызвать метод хранилища для списания.
	err := h.storage.Withdraw(r.Context(), userID, req.OrderNumber, req.Sum)
	if err != nil {
		if errors.Is(err, models.ErrInsufficientFunds) {
			http.Error(w, "Insufficient funds on balance", http.StatusPaymentRequired)
			return
		}
		h.logger.Printf("Error withdrawing funds for user %d: %v", userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 6. Вернуть 200 OK.
	w.WriteHeader(http.StatusOK)
}

// GetUserWithdrawals получает историю списаний пользователя.
// GET /api/user/withdrawals
func (h *Handler) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	// 1. Получить ID пользователя из контекста.
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Получить из БД всю историю списаний для этого пользователя.
	withdrawals, err := h.storage.WithdrawalFindForUser(r.Context(), userID)
	if err != nil {
		h.logger.Printf("Error fetching withdrawals for user %d: %v", userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 3. Если списаний нет, вернуть 204 No Content.
	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// 4 & 5. Сериализовать и отправить список списаний.
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		h.logger.Printf("Error encoding withdrawals to JSON for user %d: %v", userID, err)
	}
}
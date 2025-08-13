package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// AccrualClient - это клиент для взаимодействия с внешней системой расчёта баллов.
type AccrualClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewAccrualClient создаёт новый экземпляр клиента для системы начислений.
func NewAccrualClient(baseURL string) *AccrualClient {
	return &AccrualClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
	}
}

// AccrualResponse представляет ответ от системы начислений.
type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"` // REGISTERED, INVALID, PROCESSING, PROCESSED
	Accrual float64 `json:"accrual,omitempty"`
}

// ErrTooManyRequests возвращается, когда система начислений ограничивает количество запросов.
type ErrTooManyRequests struct {
	RetryAfter time.Duration
}

func (e *ErrTooManyRequests) Error() string {
	return fmt.Sprintf("too many requests, retry after %v", e.RetryAfter)
}

// ErrOrderNotRegistered возвращается, когда заказ не найден в системе начислений.
var ErrOrderNotRegistered = fmt.Errorf("order not registered in accrual system")

// GetOrderAccrual запрашивает информацию о начислениях для указанного номера заказа.
func (c *AccrualClient) GetOrderAccrual(ctx context.Context, orderNumber string) (*AccrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			return nil, fmt.Errorf("failed to decode successful response: %w", err)
		}
		return &accrualResp, nil
	case http.StatusNoContent:
		return nil, ErrOrderNotRegistered
	case http.StatusTooManyRequests:
		retryAfterHeader := resp.Header.Get("Retry-After")
		retryAfterSeconds, _ := strconv.Atoi(retryAfterHeader) // Ошибка игнорируется, по умолчанию будет 0
		return nil, &ErrTooManyRequests{RetryAfter: time.Duration(retryAfterSeconds) * time.Second}
	default:
		return nil, fmt.Errorf("unexpected status code from accrual system: %d", resp.StatusCode)
	}
}

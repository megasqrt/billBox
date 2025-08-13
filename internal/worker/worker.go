package worker

import (
	"billBox/internal/external"
	"billBox/internal/models"
	"billBox/internal/storage"
	"context"
	"errors"
	"log"
	"time"
)

// Worker опрашивает статусы заказов во внешней системе начислений.
type Worker struct {
	storage       storage.Storage
	accrualClient *external.AccrualClient
	logger        *log.Logger
	pollInterval  time.Duration
}

// New создает новый экземпляр воркера.
func New(storage storage.Storage, accrualSystemAddress string, logger *log.Logger, pollInterval time.Duration) *Worker {
	return &Worker{
		storage:       storage,
		accrualClient: external.NewAccrualClient(accrualSystemAddress),
		logger:        logger,
		pollInterval:  pollInterval,
	}
}

// Start запускает основной цикл воркера.
// Он будет работать до тех пор, пока не будет отменен переданный контекст.
func (w *Worker) Start(ctx context.Context) {
	w.logger.Println("Starting accrual worker...")
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Println("Stopping accrual worker...")
			return
		case <-ticker.C:
			w.processOrders(ctx)
		}
	}
}

// processOrders получает заказы для обработки и обновляет их статусы.
func (w *Worker) processOrders(ctx context.Context) {
	orders, err := w.storage.FindProcessableOrders(ctx)
	if err != nil {
		w.logger.Printf("Error fetching processable orders: %v", err)
		return
	}

	if len(orders) == 0 {
		return
	}

	for _, order := range orders {
		accrualInfo, err := w.accrualClient.GetOrderAccrual(ctx, order.Number)
		if err != nil {
			var tooManyRequestsErr *external.ErrTooManyRequests
			if errors.As(err, &tooManyRequestsErr) {
				w.logger.Printf("Accrual system rate limit exceeded. Pausing for %v", tooManyRequestsErr.RetryAfter)
				time.Sleep(tooManyRequestsErr.RetryAfter)
				return // Прерываем текущий цикл, чтобы подождать перед следующей пачкой запросов
			}
			if !errors.Is(err, external.ErrOrderNotRegistered) {
				w.logger.Printf("Error getting accrual for order %s: %v", order.Number, err)
			}
			continue
		}

		if string(order.Status) != accrualInfo.Status {
			w.logger.Printf("Updating order %s: status -> %s, accrual -> %f", order.Number, accrualInfo.Status, accrualInfo.Accrual)
			if err := w.storage.UpdateOrderAccrual(ctx, order.Number, models.OrderStatus(accrualInfo.Status), accrualInfo.Accrual); err != nil {
				w.logger.Printf("Error updating order %s: %v", order.Number, err)
			}
		}
	}
}
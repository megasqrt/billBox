package worker

import (
	"billBox/internal/external"
	"billBox/internal/models"
	"billBox/internal/storage"
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

// Worker опрашивает статусы заказов во внешней системе начислений.
type Worker struct {
	storage       storage.Storage
	accrualClient *external.AccrualClient
	logger        *log.Logger
	pollInterval  time.Duration
	mu            sync.Mutex
	pausedUntil   time.Time
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
// Он будет работать до тех пор, пока не будет отменен переданный контекст
// или пока не произойдет критическая ошибка при обработке.
func (w *Worker) Start(ctx context.Context) {
	w.logger.Println("Starting accrual worker...")

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Println("Stopping accrual worker (context canceled)...")
			return
		case <-ticker.C:
			w.mu.Lock()
			isPaused := time.Now().Before(w.pausedUntil)
			w.mu.Unlock()

			if isPaused {
				w.logger.Println("Worker is paused due to rate limit.")
				continue
			}

			if err := w.processOrders(ctx); err != nil {
				w.logger.Printf("Critical error in processOrders, stopping worker: %v", err)
				return
			}
		}
	}
}

// processOrders получает заказы для обработки и обновляет их статусы.
// При ошибке 429 Too Many Requests, воркер "засыпает".
// При других критических ошибках возвращает ошибку, что приводит к остановке воркера.
func (w *Worker) processOrders(ctx context.Context) error {
	orders, err := w.storage.FindProcessableOrders(ctx)
	if err != nil {
		w.logger.Printf("Error fetching processable orders: %v", err)
		return nil // Не считаем ошибку БД критической для остановки всего воркера
	}

	if len(orders) == 0 {
		return nil
	}

	// Контекст для пакета задач. Отменяется при первой же ошибке.
	batchCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errChan := make(chan error, len(orders))
	var wg sync.WaitGroup

	for i := range orders {
		wg.Add(1)
		go func(order models.Order) {
			defer wg.Done()

			// Проверяем, не был ли уже отменен контекст пакета задач
			select {
			case <-batchCtx.Done():
				return
			default:
			}

			accrualInfo, err := w.accrualClient.GetOrderAccrual(batchCtx, order.Number)
			if err != nil {
				var tooManyRequestsErr *external.ErrTooManyRequests
				if errors.As(err, &tooManyRequestsErr) {
					// Ошибка 429: нужно "уснуть"
					w.logger.Printf("Accrual system rate limit exceeded for order %s. Pausing for %v", order.Number, tooManyRequestsErr.RetryAfter)

					w.mu.Lock()
					pauseTarget := time.Now().Add(tooManyRequestsErr.RetryAfter)
					if w.pausedUntil.Before(pauseTarget) {
						w.pausedUntil = pauseTarget
					}
					w.mu.Unlock()

					errChan <- err
					cancel() // Отменяем весь пакет задач
					return
				}

				// Игнорируем ошибку, если заказ не зарегистрирован в системе начислений
				if errors.Is(err, external.ErrOrderNotRegistered) {
					return
				}

				// Другая критическая ошибка
				w.logger.Printf("Error getting accrual for order %s: %v", order.Number, err)
				errChan <- err
				cancel() // Отменяем весь пакет задач
				return
			}

			if string(order.Status) != accrualInfo.Status {
				w.logger.Printf("Updating order %s: status -> %s, accrual -> %f", order.Number, accrualInfo.Status, accrualInfo.Accrual)
				if err := w.storage.UpdateOrderAccrual(batchCtx, order.Number, models.OrderStatus(accrualInfo.Status), accrualInfo.Accrual); err != nil {
					w.logger.Printf("Error updating order %s: %v", order.Number, err)
					errChan <- err
					cancel() // Отменяем пакет задач при ошибке обновления в БД
				}
			}
		}(orders[i])
	}

	wg.Wait()
	close(errChan)

	// Проверяем, были ли в пакете критические ошибки (не 429)
	var criticalError error
	for err := range errChan {
		if err != nil {
			var tooManyRequestsErr *external.ErrTooManyRequests
			// Если ошибка не 429, считаем ее критической
			if !errors.As(err, &tooManyRequestsErr) {
				criticalError = err
			}
		}
	}

	return criticalError
}

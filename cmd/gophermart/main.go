package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"billBox/internal/api"
	"billBox/internal/config"
	"billBox/internal/storage"
)

func main() {
	// 1. Инициализация конфигурации (флаги и переменные окружения)
	cfg := config.New()

	// 2. Инициализация логгера
	logger := log.New(os.Stdout, "gophermart ", log.LstdFlags)

	// 3. Инициализация хранилища (подключение к PostgreSQL)
	db, err := storage.New(cfg.DatabaseURI)
	if err != nil {
		logger.Fatalf("failed to initialize storage: %v", err)
	}
	defer db.Close()

	// // 4. Инициализация воркера для опроса системы начислений
	// accrualPoller := worker.New(db, cfg.AccrualSystemAddress, logger)
	// go accrualPoller.Start(context.Background())

	// 5. Инициализация HTTP-сервера и роутера
	router := api.NewRouter(db, logger, cfg)
	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: router,
	}

	// 6. Graceful shutdown
	go func() {
		logger.Println("Starting server on", cfg.RunAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}

	logger.Println("Server exiting")
}

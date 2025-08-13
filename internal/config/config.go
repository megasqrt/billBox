package config

import (
	"flag"
	"os"
)

// Config хранит конфигурацию для приложения.
// Значения считываются из переменных окружения и флагов командной строки.
type Config struct {
	RunAddress           string // Адрес и порт запуска сервиса
	DatabaseURI          string // Адрес подключения к базе данных
	AccrualSystemAddress string // Адрес системы расчёта начислений
	JWTSecretKey         string // Секретный ключ для JWT
}

// New инициализирует новый объект Config, считывая флаги и переменные окружения.
func New() *Config {
	cfg := &Config{}

	// Определяем флаги командной строки с описаниями и значениями по умолчанию.
	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "Адрес и порт запуска сервиса (env: RUN_ADDRESS)")
	flag.StringVar(&cfg.DatabaseURI, "d", "", "Адрес подключения к БД (env: DATABASE_URI)")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Адрес системы расчёта начислений (env: ACCRUAL_SYSTEM_ADDRESS)")
	flag.StringVar(&cfg.JWTSecretKey, "k", "supersecretkey", "Секретный ключ для JWT (env: JWT_SECRET_KEY)")

	flag.Parse()

	// Переопределяем значения из переменных окружения, если они заданы.
	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		cfg.RunAddress = envRunAddress
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		cfg.DatabaseURI = envDatabaseURI
	}
	if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
		cfg.AccrualSystemAddress = envAccrualSystemAddress
	}
	if envJWTSecretKey := os.Getenv("JWT_SECRET_KEY"); envJWTSecretKey != "" {
		cfg.JWTSecretKey = envJWTSecretKey
	}

	return cfg
}
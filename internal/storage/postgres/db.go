package postgres

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB implements the storage.Storage interface for PostgreSQL.
type DB struct {
	conn *sql.DB
}

// New creates a new PostgreSQL storage client, connects to the DB and runs migrations.
func New(databaseURI string) (*DB, error) {
	if databaseURI == "" {
		return nil, fmt.Errorf("database URI is required")
	}

	conn, err := sql.Open("pgx", databaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err = conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err = runMigrations(conn); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database connection and migrations successful")
	return &DB{conn: conn}, nil
}

func runMigrations(db *sql.DB) error {
	// Проверяем, что директория с миграциями действительно встроена в бинарный файл.
	// Это помогает выдать более понятную ошибку, если go:embed не сработал.
	if _, err := migrationsFS.ReadDir("migrations"); err != nil {
		return fmt.Errorf("не удалось прочитать встроенный каталог 'migrations': %w. Убедитесь, что он находится в 'internal/storage/postgres/migrations' относительно корня проекта", err)
	}

	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source driver: %w", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration db driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

func (db *DB) Close() {
	if db.conn != nil {
		db.conn.Close()
	}
}
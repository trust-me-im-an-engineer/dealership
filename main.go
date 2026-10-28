package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v11"
	"github.com/jackc/pgx/v5/pgxpool"
)

type config struct {
	LogLevel slog.Level `env:"LOG_LEVEL"`
}

func main() {
	if err := run(); err != nil {
		slog.Error(
			"application failed",
			slog.Any("error", err),
		)
		os.Exit(1)
	} else {
		slog.Info("application stopped")
	}
}

// run запускает приложение и возвращает ошибку в случае неудачи
func run() error {
	// Загрузка конфигурации из переменных окружения
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		return fmt.Errorf("failed to parse config from environment: %w", err)
	}

	// Настройка логирования в формате json с заданным уровнем логирования
	slog.SetDefault(
		slog.New(
			slog.NewJSONHandler(
				os.Stderr,
				&slog.HandlerOptions{Level: cfg.LogLevel},
			),
		),
	)

	// Подготовка контекста, отменяемого сигналами SIGINT (Ctrl+C) и SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Получение конфигурации postgres из стандартных переменных окружения:
	// PGHOST, PGDATABASE, PGUSER, PGPASSWORD и PGSSLMODE.
	config, err := pgxpool.ParseConfig("")
	if err != nil {
		return fmt.Errorf("failed to parse postgres config: %w", err)
	}

	// Подключение к postgres
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create pgx pool: %w", err)
	}
	defer func() {
		slog.Info("closing database pool...")
		pool.Close()
		slog.Info("db pool closed")
	}()

	// Проверка полученного подключения
	if err = pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	slog.Info("connected to postgres database")

	return nil
}

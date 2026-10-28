package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/caarlos0/env/v11"
	"github.com/jackc/pgx/v5/pgxpool"

	"dealership/internal/db"
	"dealership/internal/handlers"
	"dealership/internal/middleware"
	"dealership/internal/router"
	"dealership/internal/session"
)

// config содержит настройки приложения из переменных окружения.
type config struct {
	LogLevel    slog.Level `env:"LOG_LEVEL"`
	HTTPAddress string     `env:"HTTP_ADDRESS,required"`
}

// main выводит ошибку запуска и завершает процесс с ненулевым кодом.
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

	// Регистрация типа данных сессии и настройка её cookie
	gob.Register(session.User{})
	sessions := scs.New()
	sessions.Lifetime = 12 * time.Hour
	sessions.Cookie.SameSite = http.SameSiteLaxMode
	sessions.Cookie.Secure = true
	sessions.Cookie.Persist = true

	// Подключение обработчиков и middleware к HTTP-серверу
	queries := db.New(pool)
	h := handlers.New(sessions, queries)
	m := middleware.New(sessions)
	server := &http.Server{
		Addr:    cfg.HTTPAddress,
		Handler: router.Route(h, m, sessions),
	}
	serverErrors := make(chan error, 1)

	// Запуск в отдельной горутине позволяет одновременно ждать ошибки сервера и сигнала остановки.
	go func() {
		serverErrors <- server.ListenAndServe()
	}()
	slog.Info(
		"http server started",
		slog.Any("address", cfg.HTTPAddress),
	)

	// Остановка программы либо по ошибке сервера, либо по завершению контекста
	select {
	case err := <-serverErrors:
		return fmt.Errorf("http server failed: %w", err)

	case <-ctx.Done():
		shutdownCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()

		slog.Info("shutting server down...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("failed to shut down http server: %w", err)
		}

		slog.Info("server shutted down")
		return nil
	}
}

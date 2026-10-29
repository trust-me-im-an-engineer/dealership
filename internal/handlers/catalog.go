package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

// postgresError проверяет код ошибки PostgreSQL, например 23505 для UNIQUE или 23503 для FOREIGN KEY.
func postgresError(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}

// serverError записывает внутреннюю ошибку в журнал и скрывает её детали от пользователя.
func serverError(w http.ResponseWriter, msg string, err error) {
	slog.Error(
		msg,
		slog.Any("error", err),
	)
	http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
}

// После изменения данных htmx обновляет текущую страницу; обычная форма переходит на главную.
func mutationSuccess(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

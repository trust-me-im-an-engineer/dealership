package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

func postgresError(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}

func catalogServerError(w http.ResponseWriter, message string, err error) {
	slog.Error(
		message,
		slog.Any("error", err),
	)
	http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
}

// После изменения каталога htmx обновляет текущую страницу; обычная форма переходит на главную.
func catalogSuccess(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

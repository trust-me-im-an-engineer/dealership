package handlers

import (
	"bytes"
	"context"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"

	"dealership/internal/db"
)

// userRepository задаёт запросы, необходимые обработчикам для входа пользователя.
type userRepository interface {
	GetUserByUsername(context.Context, string) (db.User, error)
}

// Handlers хранит зависимости HTTP-обработчиков.
type Handlers struct {
	sessions       *scs.SessionManager
	userRepository userRepository
}

// New создаёт набор обработчиков с общим менеджером сессий и репозиторием пользователей.
func New(sessions *scs.SessionManager, userRepository userRepository) *Handlers {
	return &Handlers{
		sessions:       sessions,
		userRepository: userRepository,
	}
}

// render сначала собирает страницу в буфере, чтобы при ошибке шаблона не отправить часть ответа.
func (h *Handlers) render(w http.ResponseWriter, page *template.Template, data any, status int) {
	var buf bytes.Buffer
	if err := page.ExecuteTemplate(&buf, "base.html", data); err != nil {
		slog.Error(
			"failed to render page",
			slog.Any("error", err),
		)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

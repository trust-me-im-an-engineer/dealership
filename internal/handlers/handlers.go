package handlers

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"

	"dealership/internal/db"
)

// Handlers хранит зависимости HTTP-обработчиков.
type Handlers struct {
	sessions *scs.SessionManager
	queries  *db.Queries
}

// New создаёт набор обработчиков с общим менеджером сессий и запросами к базе данных.
func New(sessions *scs.SessionManager, queries *db.Queries) *Handlers {
	return &Handlers{
		sessions: sessions,
		queries:  queries,
	}
}

// render сначала собирает страницу в буфере, чтобы при ошибке шаблона не отправить часть ответа.
func (h *Handlers) render(w http.ResponseWriter, page *template.Template, data any, status int) {
	renderTemplate(w, page, "base.html", data, status)
}

// renderFragment отправляет только строку таблицы для обновления через htmx.
func (h *Handlers) renderFragment(w http.ResponseWriter, page *template.Template,
	name string, data any, status int,
) {
	renderTemplate(w, page, name, data, status)
}

func renderTemplate(w http.ResponseWriter, page *template.Template, name string, data any, status int) {
	var buf bytes.Buffer
	if err := page.ExecuteTemplate(&buf, name, data); err != nil {
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

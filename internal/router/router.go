package router

import (
	"net/http"

	"github.com/alexedwards/scs/v2"

	"dealership/internal/handlers"
	"dealership/internal/middleware"
)

// Route связывает пути с обработчиками и применяет проверку доступа к закрытым страницам.
func Route(h *handlers.Handlers, m *middleware.Middleware, sessions *scs.SessionManager) http.Handler {
	mux := http.NewServeMux()

	// Публичные пути
	mux.HandleFunc("GET /login", h.GetLogin)
	mux.HandleFunc("POST /login", h.PostLogin)

	// Пути, доступные только после входа
	mux.Handle("POST /logout", m.Auth(http.HandlerFunc(h.PostLogout)))
	mux.Handle("GET /{$}", m.Auth(http.HandlerFunc(h.Home)))

	// Загрузка и сохранение сессии выполняются для каждого HTTP-запроса.
	return sessions.LoadAndSave(mux)
}

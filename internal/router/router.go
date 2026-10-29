package router

import (
	"net/http"

	"github.com/alexedwards/scs/v2"

	"dealership/internal/handlers"
	"dealership/internal/middleware"
	"dealership/internal/templ"
)

// Route связывает пути с обработчиками и применяет проверку доступа к закрытым страницам.
func Route(h *handlers.Handlers, m *middleware.Middleware, sessions *scs.SessionManager) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServerFS(templ.Static()))

	// Публичные пути
	mux.HandleFunc("GET /login", h.GetLogin)
	mux.HandleFunc("POST /login", h.PostLogin)

	// Пути, доступные только после входа
	mux.Handle("POST /logout", m.Auth(http.HandlerFunc(h.PostLogout)))
	mux.Handle("GET /{$}", m.Auth(http.HandlerFunc(h.Home)))

	// Изменять каталог может только администратор.
	mux.Handle("GET /admin/models/edit", m.Admin(http.HandlerFunc(h.GetEditModel)))
	mux.Handle("GET /admin/models/row", m.Admin(http.HandlerFunc(h.GetModelRow)))
	mux.Handle("POST /admin/models", m.Admin(http.HandlerFunc(h.PostModel)))
	mux.Handle("POST /admin/models/update", m.Admin(http.HandlerFunc(h.PostUpdateModel)))
	mux.Handle("POST /admin/models/delete", m.Admin(http.HandlerFunc(h.PostDeleteModel)))
	mux.Handle("GET /admin/configurations/edit", m.Admin(http.HandlerFunc(h.GetEditConfiguration)))
	mux.Handle("GET /admin/configurations/row", m.Admin(http.HandlerFunc(h.GetConfigurationRow)))
	mux.Handle("POST /admin/configurations", m.Admin(http.HandlerFunc(h.PostConfiguration)))
	mux.Handle("POST /admin/configurations/update", m.Admin(http.HandlerFunc(h.PostUpdateConfiguration)))
	mux.Handle("POST /admin/configurations/delete", m.Admin(http.HandlerFunc(h.PostDeleteConfiguration)))

	// Загрузка и сохранение сессии выполняются для каждого HTTP-запроса.
	return sessions.LoadAndSave(mux)
}

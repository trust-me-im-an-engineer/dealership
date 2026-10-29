package middleware

import (
	"net/http"

	"github.com/alexedwards/scs/v2"

	"dealership/internal/session"
)

// Middleware хранит менеджер сессий для проверки доступа к защищённым путям.
type Middleware struct {
	sessions *scs.SessionManager
}

// New создаёт middleware с общим для приложения менеджером сессий.
func New(sessions *scs.SessionManager) *Middleware {
	return &Middleware{sessions: sessions}
}

// Auth пропускает запрос только при наличии данных вошедшего пользователя в сессии.
func (m *Middleware) Auth(next http.Handler) http.Handler {
	return m.checkAccess(next, "")
}

// Admin разрешает доступ только администратору.
func (m *Middleware) Admin(next http.Handler) http.Handler {
	return m.checkAccess(next, "admin")
}

// checkAccess требует сессию и, если роль задана, проверяет её перед вызовом обработчика.
func (m *Middleware) checkAccess(next http.Handler, role string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := m.sessions.Get(r.Context(), session.UserKey).(session.User)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if role != "" && user.Role != role {
			http.Error(w, "Доступ запрещён", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

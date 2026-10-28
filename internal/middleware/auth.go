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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := m.sessions.Get(r.Context(), session.UserKey).(session.User); !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

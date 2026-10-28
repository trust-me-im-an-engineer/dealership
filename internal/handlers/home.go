package handlers

import (
	"net/http"

	"dealership/internal/session"
	"dealership/internal/templ"
)

type homePage struct {
	Username string
	Role     string
}

// Шаблон главной страницы разбирается один раз при запуске приложения.
var tmplHome = templ.Parse("home.html")

// Home показывает данные пользователя, доступ к которым уже проверил middleware.
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	user := h.sessions.Get(r.Context(), session.UserKey).(session.User)
	data := homePage{Username: user.Username, Role: user.Role}
	h.render(w, tmplHome, data, http.StatusOK)
}

package handlers

import (
	"net/http"

	"dealership/internal/db"
	"dealership/internal/session"
)

// homePage содержит общие данные главных страниц администратора и продавца.
type homePage struct {
	Username       string
	RoleName       string
	Models         []db.ListCarModelsRow
	Configurations []db.ListCarConfigurationsRow
}

// Home направляет вошедшего пользователя на главную страницу его роли.
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	user := h.sessions.Get(r.Context(), session.UserKey).(session.User)
	switch user.Role {
	case "admin":
		h.adminHome(w, r, user)
	case "seller":
		h.sellerHome(w, r, user)
	default:
		http.Error(w, "Доступ запрещён", http.StatusForbidden)
	}
}

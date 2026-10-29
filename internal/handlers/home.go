package handlers

import (
	"context"
	"fmt"
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

// loadHomePage загружает общие данные главной страницы и указывает в ошибке неудавшийся запрос.
func (h *Handlers) loadHomePage(ctx context.Context, user session.User, roleName string) (homePage, error) {
	models, err := h.queries.ListCarModels(ctx)
	if err != nil {
		return homePage{}, fmt.Errorf("failed to list car models: %w", err)
	}

	configurations, err := h.queries.ListCarConfigurations(ctx)
	if err != nil {
		return homePage{}, fmt.Errorf("failed to list car configurations: %w", err)
	}

	return homePage{
		Username:       user.Username,
		RoleName:       roleName,
		Models:         models,
		Configurations: configurations,
	}, nil
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

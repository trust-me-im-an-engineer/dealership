package handlers

import (
	"log/slog"
	"net/http"

	"dealership/internal/db"
	"dealership/internal/session"
	"dealership/internal/templ"
)

type homePage struct {
	Username         string
	Role             string
	Models           []db.ListCarModelsRow
	Configurations   []db.ListCarConfigurationsRow
	NewModel         modelFormPage
	NewConfiguration configurationFormPage
}

// Шаблон главной страницы разбирается один раз при запуске приложения.
var tmplHome = templ.Parse(
	"home.html", "model-row.html", "model-editor-row.html",
	"configuration-row.html", "configuration-editor-row.html",
)

// Home показывает данные пользователя, доступ к которым уже проверил middleware.
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	user := h.sessions.Get(r.Context(), session.UserKey).(session.User)
	data := homePage{
		Username: user.Username,
		Role:     user.Role,
		NewModel: modelFormPage{Action: "/admin/models", FormID: "model-create"},
		NewConfiguration: configurationFormPage{
			Action: "/admin/configurations", FormID: "configuration-create",
		},
	}
	if user.Role == "admin" {
		var err error
		data.Models, err = h.queries.ListCarModels(r.Context())
		if err != nil {
			slog.Error(
				"failed to list car models",
				slog.Any("error", err),
			)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}
		data.Configurations, err = h.queries.ListCarConfigurations(r.Context())
		if err != nil {
			slog.Error(
				"failed to list car configurations",
				slog.Any("error", err),
			)
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}
		data.NewConfiguration.Models = data.Models
	}
	h.render(w, tmplHome, data, http.StatusOK)
}

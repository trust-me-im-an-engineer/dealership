package handlers

import (
	"log/slog"
	"net/http"

	"dealership/internal/db"
	"dealership/internal/session"
	"dealership/internal/templ"
)

// adminHomePage содержит данные трёх таблиц и пустых форм для страницы администратора.
type adminHomePage struct {
	homePage
	Orders           []db.ListOrdersRow
	NewModel         modelFormPage
	NewConfiguration configurationFormPage
}

// Шаблон страницы администратора разбирается один раз при запуске приложения.
var tmplAdminHome = templ.Parse(
	"home/admin.html",
	"home/header.html",
	"model-row.html",
	"model-editor-row.html",
	"configuration-row.html",
	"configuration-editor-row.html",
)

// adminHome показывает администратору каталог и все заказы.
func (h *Handlers) adminHome(w http.ResponseWriter, r *http.Request, user session.User) {
	data := adminHomePage{
		homePage: homePage{
			Username: user.Username,
			RoleName: "администратор",
		},
		NewModel: modelFormPage{
			catalogFormPage: catalogFormPage{
				Action: "/admin/models",
				FormID: "model-create",
			},
		},
		NewConfiguration: configurationFormPage{
			catalogFormPage: catalogFormPage{
				Action: "/admin/configurations",
				FormID: "configuration-create",
			},
		},
	}

	var err error

	data.Models, err = h.queries.ListCarModels(r.Context())
	if err != nil {
		slog.Error("failed to list car models", slog.Any("error", err))
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	data.Configurations, err = h.queries.ListCarConfigurations(r.Context())
	if err != nil {
		slog.Error("failed to list car configurations", slog.Any("error", err))
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	data.Orders, err = h.queries.ListOrders(r.Context())
	if err != nil {
		slog.Error("failed to list orders", slog.Any("error", err))
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	data.NewConfiguration.Models = data.Models
	h.render(w, tmplAdminHome, data, http.StatusOK)
}

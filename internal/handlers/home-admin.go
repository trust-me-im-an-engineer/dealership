package handlers

import (
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
	home, err := h.loadHomePage(r.Context(), user, "администратор")
	if err != nil {
		serverError(w, "failed to load admin home page", err)
		return
	}

	orders, err := h.queries.ListOrders(r.Context())
	if err != nil {
		serverError(w, "failed to list orders", err)
		return
	}

	data := adminHomePage{
		homePage: home,
		Orders:   orders,
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
			Models: home.Models,
		},
	}

	h.render(w, tmplAdminHome, data, http.StatusOK)
}

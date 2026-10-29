package handlers

import (
	"log/slog"
	"net/http"

	"dealership/internal/db"
	"dealership/internal/session"
	"dealership/internal/templ"
)

// sellerHomePage содержит заказы продавца и каталог для просмотра.
type sellerHomePage struct {
	homePage
	Orders []db.ListSellerOrdersRow
}

// Шаблон страницы продавца разбирается один раз при запуске приложения.
var tmplSellerHome = templ.Parse(
	"home/seller.html",
	"home/header.html",
	"seller-order-row.html",
)

// sellerHome показывает главную страницу продавца.
func (h *Handlers) sellerHome(w http.ResponseWriter, r *http.Request, user session.User) {
	data := sellerHomePage{
		homePage: homePage{
			Username: user.Username,
			RoleName: "продавец",
		},
	}

	var err error

	data.Orders, err = h.queries.ListSellerOrders(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to list seller orders", slog.Any("error", err))
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

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

	h.render(w, tmplSellerHome, data, http.StatusOK)
}

package handlers

import (
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
	"seller-order-create-row.html",
)

// sellerHome показывает главную страницу продавца.
func (h *Handlers) sellerHome(w http.ResponseWriter, r *http.Request, user session.User) {
	home, err := h.loadHomePage(r.Context(), user, "продавец")
	if err != nil {
		serverError(w, "failed to load seller home page", err)
		return
	}

	orders, err := h.queries.ListSellerOrders(r.Context(), user.ID)
	if err != nil {
		serverError(w, "failed to list seller orders", err)
		return
	}

	data := sellerHomePage{
		homePage: home,
		Orders:   orders,
	}

	render(w, tmplSellerHome, data, http.StatusOK)
}

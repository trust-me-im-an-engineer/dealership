package handlers

import (
	"net/http"

	"dealership/internal/session"
	"dealership/internal/templ"
)

type sellerHomePage struct {
	Username string
	RoleName string
}

// Шаблон страницы продавца разбирается один раз при запуске приложения.
var tmplSellerHome = templ.Parse(
	"home/seller.html",
	"home/header.html",
)

// sellerHome показывает главную страницу продавца.
func (h *Handlers) sellerHome(w http.ResponseWriter, user session.User) {
	data := sellerHomePage{
		Username: user.Username,
		RoleName: "продавец",
	}
	h.render(w, tmplSellerHome, data, http.StatusOK)
}

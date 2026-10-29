package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"dealership/internal/db"
	"dealership/internal/session"
	"dealership/internal/templ"
)

// sellerOrderEditorPage содержит заказ и варианты комплектаций для строки редактирования.
type sellerOrderEditorPage struct {
	db.GetSellerOrderRow
	Configurations []db.ListCarConfigurationsRow
}

// sellerOrderFormError сохраняет текст ошибки формы в исходном регистре.
type sellerOrderFormError string

// Error возвращает сообщение для пользователя.
func (e sellerOrderFormError) Error() string { return string(e) }

// Шаблоны строк заказа разбираются один раз при запуске приложения.
var (
	tmplSellerOrderRow       = templ.Parse("seller-order-row.html")
	tmplSellerOrderEditorRow = templ.Parse("seller-order-editor-row.html")
)

// sellerOrderID проверяет номер заказа из строки запроса.
func sellerOrderID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id < 1 {
		return 0, sellerOrderFormError("Некорректный номер заказа")
	}
	return id, nil
}

// GetEditSellerOrder возвращает продавцу строку редактирования его заказа.
func (h *Handlers) GetEditSellerOrder(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "true" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id, err := sellerOrderID(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := h.sessions.Get(r.Context(), session.UserKey).(session.User)
	params := db.GetSellerOrderParams{
		ID:       id,
		SellerID: user.ID,
	}
	order, err := h.queries.GetSellerOrder(r.Context(), params)
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, "failed to get seller order", err)
		return
	}

	configurations, err := h.queries.ListCarConfigurations(r.Context())
	if err != nil {
		serverError(w, "failed to list car configurations", err)
		return
	}

	page := sellerOrderEditorPage{
		GetSellerOrderRow: order,
		Configurations:    configurations,
	}
	h.renderFragment(w, tmplSellerOrderEditorRow, "seller-order-editor-row", page, http.StatusOK)
}

// GetSellerOrderRow возвращает обычную строку после отмены редактирования.
func (h *Handlers) GetSellerOrderRow(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "true" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id, err := sellerOrderID(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := h.sessions.Get(r.Context(), session.UserKey).(session.User)
	params := db.GetSellerOrderParams{
		ID:       id,
		SellerID: user.ID,
	}
	order, err := h.queries.GetSellerOrder(r.Context(), params)
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, "failed to get seller order", err)
		return
	}

	h.renderFragment(w, tmplSellerOrderRow, "seller-order-row", order, http.StatusOK)
}

// parseSellerOrderForm проверяет изменяемые поля заказа и возвращает параметры обновления.
func parseSellerOrderForm(r *http.Request) (db.UpdateSellerOrderParams, error) {
	if err := r.ParseForm(); err != nil {
		return db.UpdateSellerOrderParams{}, sellerOrderFormError("Некорректная форма")
	}

	configurationID, err := strconv.ParseInt(r.PostFormValue("car_configuration_id"), 10, 64)
	if err != nil || configurationID < 1 {
		return db.UpdateSellerOrderParams{}, sellerOrderFormError("Выберите комплектацию")
	}

	buyerName := strings.TrimSpace(r.PostFormValue("buyer_name"))
	buyerPhone := strings.TrimSpace(r.PostFormValue("buyer_phone"))
	if buyerName == "" || buyerPhone == "" {
		return db.UpdateSellerOrderParams{}, sellerOrderFormError("Укажите имя и телефон покупателя")
	}

	deliveryDate, err := time.Parse("2006-01-02", r.PostFormValue("expected_delivery_on"))
	if err != nil {
		return db.UpdateSellerOrderParams{}, sellerOrderFormError("Укажите корректную дату доставки")
	}

	status := db.OrderStatus(r.PostFormValue("status"))
	switch status {
	case db.OrderStatusPlaced, db.OrderStatusDelivered, db.OrderStatusCancelled:
	default:
		return db.UpdateSellerOrderParams{}, sellerOrderFormError("Выберите статус заказа")
	}

	var buyerEmail *string
	if email := strings.TrimSpace(r.PostFormValue("buyer_email")); email != "" {
		buyerEmail = &email
	}

	params := db.UpdateSellerOrderParams{
		CarConfigurationID: configurationID,
		BuyerName:          buyerName,
		BuyerPhone:         buyerPhone,
		BuyerEmail:         buyerEmail,
		ExpectedDeliveryOn: pgtype.Date{Time: deliveryDate, Valid: true},
		Status:             status,
	}
	return params, nil
}

// PostUpdateSellerOrder сохраняет изменения только в заказе текущего продавца.
func (h *Handlers) PostUpdateSellerOrder(w http.ResponseWriter, r *http.Request) {
	params, err := parseSellerOrderForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := sellerOrderID(r.PostFormValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := h.sessions.Get(r.Context(), session.UserKey).(session.User)
	params.ID = id
	params.SellerID = user.ID
	rows, err := h.queries.UpdateSellerOrder(r.Context(), params)
	if postgresError(err, "23503") {
		http.Error(w, "Выбранная комплектация не существует", http.StatusConflict)
		return
	}
	if err != nil {
		serverError(w, "failed to update seller order", err)
		return
	}
	if rows == 0 {
		http.NotFound(w, r)
		return
	}

	mutationSuccess(w, r)
}

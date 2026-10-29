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

// getSellerOrder находит заказ по номеру и продавцу из текущей сессии.
func (h *Handlers) getSellerOrder(r *http.Request) (db.GetSellerOrderRow, error) {
	id, err := sellerOrderID(r.URL.Query().Get("id"))
	if err != nil {
		return db.GetSellerOrderRow{}, err
	}

	user := h.sessions.Get(r.Context(), session.UserKey).(session.User)
	params := db.GetSellerOrderParams{
		ID:       id,
		SellerID: user.ID,
	}
	return h.queries.GetSellerOrder(r.Context(), params)
}

// respondSellerOrderError отправляет ответ по ошибке чтения заказа.
func respondSellerOrderError(w http.ResponseWriter, r *http.Request, err error) {
	var formErr sellerOrderFormError
	switch {
	case errors.As(err, &formErr):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, pgx.ErrNoRows):
		http.NotFound(w, r)
	default:
		serverError(w, "failed to get seller order", err)
	}
}

// GetEditSellerOrder возвращает продавцу строку редактирования его заказа.
func (h *Handlers) GetEditSellerOrder(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "true" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	order, err := h.getSellerOrder(r)
	if err != nil {
		respondSellerOrderError(w, r, err)
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
	renderFragment(w, tmplSellerOrderEditorRow, "seller-order-editor-row", page, http.StatusOK)
}

// GetSellerOrderRow возвращает обычную строку после отмены редактирования.
func (h *Handlers) GetSellerOrderRow(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "true" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	order, err := h.getSellerOrder(r)
	if err != nil {
		respondSellerOrderError(w, r, err)
		return
	}

	renderFragment(w, tmplSellerOrderRow, "seller-order-row", order, http.StatusOK)
}

// parseSellerOrderForm проверяет общие поля создания и редактирования заказа.
func parseSellerOrderForm(r *http.Request) (db.CreateSellerOrderParams, error) {
	if err := r.ParseForm(); err != nil {
		return db.CreateSellerOrderParams{}, sellerOrderFormError("Некорректная форма")
	}

	configurationID, err := strconv.ParseInt(r.PostFormValue("car_configuration_id"), 10, 64)
	if err != nil || configurationID < 1 {
		return db.CreateSellerOrderParams{}, sellerOrderFormError("Выберите комплектацию")
	}

	buyerName := strings.TrimSpace(r.PostFormValue("buyer_name"))
	buyerPhone := strings.TrimSpace(r.PostFormValue("buyer_phone"))
	if buyerName == "" || buyerPhone == "" {
		return db.CreateSellerOrderParams{}, sellerOrderFormError("Укажите имя и телефон покупателя")
	}

	var buyerEmail *string
	if email := strings.TrimSpace(r.PostFormValue("buyer_email")); email != "" {
		buyerEmail = &email
	}

	params := db.CreateSellerOrderParams{
		CarConfigurationID: configurationID,
		BuyerName:          buyerName,
		BuyerPhone:         buyerPhone,
		BuyerEmail:         buyerEmail,
	}
	return params, nil
}

// PostSellerOrder создаёт заказ от имени продавца из текущей сессии.
func (h *Handlers) PostSellerOrder(w http.ResponseWriter, r *http.Request) {
	params, err := parseSellerOrderForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := h.sessions.Get(r.Context(), session.UserKey).(session.User)
	params.SellerID = user.ID
	rows, err := h.queries.CreateSellerOrder(r.Context(), params)
	if err != nil {
		serverError(w, "failed to create seller order", err)
		return
	}
	if rows == 0 {
		http.Error(w, "Выбранная комплектация не существует", http.StatusConflict)
		return
	}

	mutationSuccess(w, r)
}

// PostUpdateSellerOrder сохраняет изменения только в заказе текущего продавца.
func (h *Handlers) PostUpdateSellerOrder(w http.ResponseWriter, r *http.Request) {
	input, err := parseSellerOrderForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	deliveryDate, err := time.Parse("2006-01-02", r.PostFormValue("expected_delivery_on"))
	if err != nil {
		http.Error(w, "Укажите корректную дату доставки", http.StatusBadRequest)
		return
	}

	status := db.OrderStatus(r.PostFormValue("status"))
	switch status {
	case db.OrderStatusPlaced, db.OrderStatusDelivered, db.OrderStatusCancelled:
	default:
		http.Error(w, "Выберите статус заказа", http.StatusBadRequest)
		return
	}

	id, err := sellerOrderID(r.PostFormValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := h.sessions.Get(r.Context(), session.UserKey).(session.User)
	params := db.UpdateSellerOrderParams{
		ID:                 id,
		SellerID:           user.ID,
		CarConfigurationID: input.CarConfigurationID,
		BuyerName:          input.BuyerName,
		BuyerPhone:         input.BuyerPhone,
		BuyerEmail:         input.BuyerEmail,
		ExpectedDeliveryOn: pgtype.Date{Time: deliveryDate, Valid: true},
		Status:             status,
	}
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

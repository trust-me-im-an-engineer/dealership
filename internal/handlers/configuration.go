package handlers

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"dealership/internal/db"
	"dealership/internal/templ"
)

// Цена ограничена точностью NUMERIC(12, 2) в базе данных.
var pricePattern = regexp.MustCompile(`^\d{1,10}(\.\d{1,2})?$`)

// configurationFormError содержит текст ошибки проверки для ответа пользователю.
type configurationFormError string

func (e configurationFormError) Error() string { return string(e) }

type configurationFormPage struct {
	Action       string
	FormID       string
	ID           int64
	Name         string
	CarModel     string
	Price        string
	DeliveryDays string
	Description  string
	OrderCount   int64
	IsEdit       bool
	Models       []db.ListCarModelsRow
}

var (
	tmplConfigurationRow       = templ.Parse("configuration-row.html")
	tmplConfigurationEditorRow = templ.Parse("configuration-editor-row.html")
)

// GetEditConfiguration возвращает строку редактирования комплектации для главной страницы.
func (h *Handlers) GetEditConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "true" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil || id < 1 {
		http.Error(w, "Некорректный номер комплектации", http.StatusBadRequest)
		return
	}

	configuration, err := h.queries.GetCarConfiguration(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		catalogServerError(w, "failed to get car configuration", err)
		return
	}

	page := configurationFormPage{
		Action:       "/admin/configurations/update",
		IsEdit:       true,
		FormID:       "configuration-edit-" + strconv.FormatInt(configuration.ID, 10),
		ID:           configuration.ID,
		Name:         configuration.Name,
		CarModel:     configuration.CarModel,
		Price:        configuration.Price,
		DeliveryDays: strconv.FormatInt(int64(configuration.DeliveryDays), 10),
		OrderCount:   configuration.OrderCount,
	}
	if configuration.Description != nil {
		page.Description = *configuration.Description
	}

	models, err := h.queries.ListCarModels(r.Context())
	if err != nil {
		catalogServerError(w, "failed to list car models", err)
		return
	}

	page.Models = models
	h.renderFragment(w, tmplConfigurationEditorRow, "configuration-editor-row", page, http.StatusOK)
}

// GetConfigurationRow возвращает обычную строку таблицы после отмены редактирования.
func (h *Handlers) GetConfigurationRow(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "true" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil || id < 1 {
		http.Error(w, "Некорректный номер комплектации", http.StatusBadRequest)
		return
	}

	configuration, err := h.queries.GetCarConfiguration(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		catalogServerError(w, "failed to get car configuration", err)
		return
	}

	h.renderFragment(w, tmplConfigurationRow, "configuration-row", configuration, http.StatusOK)
}

// parseConfigurationForm проверяет общие поля комплектации и возвращает параметры запроса.
func parseConfigurationForm(r *http.Request) (db.CreateCarConfigurationParams, error) {
	if err := r.ParseForm(); err != nil {
		return db.CreateCarConfigurationParams{}, configurationFormError("Некорректная форма")
	}

	name := strings.TrimSpace(r.PostFormValue("name"))
	carModel := r.PostFormValue("car_model")
	priceText := strings.TrimSpace(r.PostFormValue("price"))
	deliveryDays := r.PostFormValue("delivery_days")
	descriptionText := strings.TrimSpace(r.PostFormValue("description"))

	days, err := strconv.ParseInt(deliveryDays, 10, 32)
	if name == "" || carModel == "" || err != nil || days < 0 || !pricePattern.MatchString(priceText) {
		err := configurationFormError("Заполните название, модель, положительную цену и срок поставки")
		return db.CreateCarConfigurationParams{}, err
	}

	var price pgtype.Numeric
	if err := price.Scan(priceText); err != nil || price.Int == nil || price.Int.Sign() <= 0 {
		return db.CreateCarConfigurationParams{}, configurationFormError("Цена должна быть больше нуля")
	}

	var description *string
	if descriptionText != "" {
		description = &descriptionText
	}

	params := db.CreateCarConfigurationParams{
		Name: name, CarModel: carModel, Price: price,
		DeliveryDays: int32(days), Description: description,
	}
	return params, nil
}

// PostConfiguration создаёт комплектацию автомобиля.
func (h *Handlers) PostConfiguration(w http.ResponseWriter, r *http.Request) {
	input, err := parseConfigurationForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.queries.CreateCarConfiguration(r.Context(), input); err != nil {
		if postgresError(err, "23503") {
			http.Error(w, "Выбранная модель не существует", http.StatusConflict)
			return
		}
		catalogServerError(w, "failed to create car configuration", err)
		return
	}

	catalogSuccess(w, r)
}

// PostUpdateConfiguration сохраняет изменения комплектации.
func (h *Handlers) PostUpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	input, err := parseConfigurationForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(r.PostFormValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.Error(w, "Некорректный номер комплектации", http.StatusBadRequest)
		return
	}

	params := db.UpdateCarConfigurationParams{
		ID: id, Name: input.Name, CarModel: input.CarModel,
		Price: input.Price, DeliveryDays: input.DeliveryDays, Description: input.Description,
	}
	rows, err := h.queries.UpdateCarConfiguration(r.Context(), params)
	if postgresError(err, "23503") {
		http.Error(w, "Выбранная модель не существует", http.StatusConflict)
		return
	}
	if err != nil {
		catalogServerError(w, "failed to update car configuration", err)
		return
	}
	if rows == 0 {
		http.NotFound(w, r)
		return
	}

	catalogSuccess(w, r)
}

// PostDeleteConfiguration удаляет комплектацию, если для неё нет заказов.
func (h *Handlers) PostDeleteConfiguration(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Некорректная форма", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(r.PostFormValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.Error(w, "Некорректный номер комплектации", http.StatusBadRequest)
		return
	}

	rows, err := h.queries.DeleteCarConfiguration(r.Context(), id)
	if postgresError(err, "23503") {
		http.Error(w, "Нельзя удалить комплектацию, для которой есть заказы", http.StatusConflict)
		return
	}
	if err != nil {
		catalogServerError(w, "failed to delete car configuration", err)
		return
	}
	if rows == 0 {
		http.NotFound(w, r)
		return
	}

	catalogSuccess(w, r)
}

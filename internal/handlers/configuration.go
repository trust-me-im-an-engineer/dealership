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
	Error        string
	IsEdit       bool
	Models       []db.ListCarModelsRow
}

var tmplConfigurationRow = templ.Parse("configuration-row.html")
var tmplConfigurationEditorRow = templ.Parse("configuration-editor-row.html")

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
		Action: "/admin/configurations/update", IsEdit: true,
		FormID: "configuration-edit-" + strconv.FormatInt(configuration.ID, 10),
		ID:     configuration.ID, Name: configuration.Name, CarModel: configuration.CarModel,
		Price: configuration.Price, DeliveryDays: strconv.FormatInt(int64(configuration.DeliveryDays), 10),
		OrderCount: configuration.OrderCount,
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

func parseConfigurationForm(r *http.Request, page *configurationFormPage) (
	db.CreateCarConfigurationParams, bool,
) {
	if err := r.ParseForm(); err != nil {
		page.Error = "Некорректная форма"
		return db.CreateCarConfigurationParams{}, false
	}
	page.Name = strings.TrimSpace(r.PostFormValue("name"))
	page.CarModel = r.PostFormValue("car_model")
	page.Price = strings.TrimSpace(r.PostFormValue("price"))
	page.DeliveryDays = r.PostFormValue("delivery_days")
	page.Description = strings.TrimSpace(r.PostFormValue("description"))
	page.ID, _ = strconv.ParseInt(r.PostFormValue("id"), 10, 64)

	days, err := strconv.ParseInt(page.DeliveryDays, 10, 32)
	if page.Name == "" || page.CarModel == "" || err != nil || days < 0 ||
		!pricePattern.MatchString(page.Price) {
		page.Error = "Заполните название, модель, положительную цену и срок поставки"
		return db.CreateCarConfigurationParams{}, false
	}
	var price pgtype.Numeric
	if err := price.Scan(page.Price); err != nil || price.Int == nil || price.Int.Sign() <= 0 {
		page.Error = "Цена должна быть больше нуля"
		return db.CreateCarConfigurationParams{}, false
	}
	var description *string
	if page.Description != "" {
		description = &page.Description
	}
	return db.CreateCarConfigurationParams{
		Name: page.Name, CarModel: page.CarModel, Price: price,
		DeliveryDays: int32(days), Description: description,
	}, true
}

// PostConfiguration создаёт комплектацию автомобиля.
func (h *Handlers) PostConfiguration(w http.ResponseWriter, r *http.Request) {
	page := configurationFormPage{}
	input, ok := parseConfigurationForm(r, &page)
	if !ok {
		http.Error(w, page.Error, http.StatusBadRequest)
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
	page := configurationFormPage{}
	input, ok := parseConfigurationForm(r, &page)
	if !ok {
		http.Error(w, page.Error, http.StatusBadRequest)
		return
	}
	if page.ID < 1 {
		http.Error(w, "Некорректный номер комплектации", http.StatusBadRequest)
		return
	}
	rows, err := h.queries.UpdateCarConfiguration(r.Context(), db.UpdateCarConfigurationParams{
		ID: page.ID, Name: input.Name, CarModel: input.CarModel,
		Price: input.Price, DeliveryDays: input.DeliveryDays, Description: input.Description,
	})
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

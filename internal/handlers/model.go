package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"

	"dealership/internal/db"
	"dealership/internal/templ"
)

// modelFormPage хранит значения строки добавления или редактирования модели.
type modelFormPage struct {
	catalogFormPage
	OldName string
	Year    string
	Error   string
}

// modelFormError содержит текст ошибки формы в исходном регистре.
type modelFormError string

// Error возвращает сообщение для пользователя.
func (e modelFormError) Error() string { return string(e) }

// Шаблоны обычной и редактируемой строки модели разбираются при запуске.
var tmplModelRow = templ.Parse("model-row.html")
var tmplModelEditorRow = templ.Parse("model-editor-row.html")

// GetEditModel возвращает строку редактирования модели для главной страницы.
func (h *Handlers) GetEditModel(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "true" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Укажите модель", http.StatusBadRequest)
		return
	}

	model, err := h.queries.GetCarModel(r.Context(), name)
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, "failed to get car model", err)
		return
	}

	page := modelFormPage{
		catalogFormPage: catalogFormPage{
			Action:     "/admin/models/update",
			FormID:     "model-edit-" + url.QueryEscape(model.Name),
			Name:       model.Name,
			OrderCount: model.OrderCount,
			IsEdit:     true,
		},
		OldName: model.Name,
		Year:    strconv.FormatInt(int64(model.Year), 10),
	}
	h.renderFragment(w, tmplModelEditorRow, "model-editor-row", page, http.StatusOK)
}

// GetModelRow возвращает обычную строку таблицы после отмены редактирования.
func (h *Handlers) GetModelRow(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") != "true" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	model, err := h.queries.GetCarModel(r.Context(), r.URL.Query().Get("name"))
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, "failed to get car model", err)
		return
	}

	h.renderFragment(w, tmplModelRow, "model-row", model, http.StatusOK)
}

// parseModelForm проверяет поля модели и возвращает параметры запроса либо ошибку.
func parseModelForm(r *http.Request) (db.CreateCarModelParams, error) {
	if err := r.ParseForm(); err != nil {
		return db.CreateCarModelParams{}, modelFormError("Некорректная форма")
	}

	name := strings.TrimSpace(r.PostFormValue("name"))
	year, err := strconv.ParseInt(r.PostFormValue("year"), 10, 32)
	if name == "" || err != nil || year < 1 || year > 9999 {
		return db.CreateCarModelParams{}, modelFormError("Введите название и год от 1 до 9999")
	}

	params := db.CreateCarModelParams{
		Name: name,
		Year: int32(year),
	}
	return params, nil
}

// PostModel создаёт модель автомобиля.
func (h *Handlers) PostModel(w http.ResponseWriter, r *http.Request) {
	input, err := parseModelForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.queries.CreateCarModel(r.Context(), input); err != nil {
		if postgresError(err, "23505") {
			http.Error(w, "Модель с таким названием уже существует", http.StatusConflict)
			return
		}

		serverError(w, "failed to create car model", err)
		return
	}

	mutationSuccess(w, r)
}

// PostUpdateModel сохраняет новое название и год модели.
func (h *Handlers) PostUpdateModel(w http.ResponseWriter, r *http.Request) {
	input, err := parseModelForm(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	oldName := r.PostFormValue("old_name")
	if oldName == "" {
		http.Error(w, "Укажите исходную модель", http.StatusBadRequest)
		return
	}

	params := db.UpdateCarModelParams{
		OldName: oldName,
		NewName: input.Name,
		Year:    input.Year,
	}
	rows, err := h.queries.UpdateCarModel(r.Context(), params)
	if postgresError(err, "23505") {
		http.Error(w, "Модель с таким названием уже существует", http.StatusConflict)
		return
	}
	if err != nil {
		serverError(w, "failed to update car model", err)
		return
	}

	if rows == 0 {
		http.NotFound(w, r)
		return
	}

	mutationSuccess(w, r)
}

// PostDeleteModel удаляет модель и её комплектации, если для них нет заказов.
func (h *Handlers) PostDeleteModel(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Некорректная форма", http.StatusBadRequest)
		return
	}

	name := r.PostFormValue("name")
	if name == "" {
		http.Error(w, "Укажите модель", http.StatusBadRequest)
		return
	}

	rows, err := h.queries.DeleteCarModel(r.Context(), name)
	if postgresError(err, "23503") {
		http.Error(w, "Нельзя удалить модель: для её комплектаций есть заказы", http.StatusConflict)
		return
	}
	if err != nil {
		serverError(w, "failed to delete car model", err)
		return
	}

	if rows == 0 {
		http.NotFound(w, r)
		return
	}

	mutationSuccess(w, r)
}

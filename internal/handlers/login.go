package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"dealership/internal/session"
	"dealership/internal/templ"
)

type loginPage struct {
	Error string
}

// Шаблон страницы входа разбирается один раз при запуске приложения.
var tmplLogin = templ.Parse("login.html")

// GetLogin показывает форму входа или отправляет уже вошедшего пользователя на главную страницу.
func (h *Handlers) GetLogin(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.sessions.Get(r.Context(), session.UserKey).(session.User); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	h.render(w, tmplLogin, loginPage{}, http.StatusOK)
}

// PostLogin проверяет пароль и создаёт сессию пользователя.
func (h *Handlers) PostLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Некорректная форма", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.PostFormValue("username"))
	password := r.PostFormValue("password")

	// Для отсутствующего пользователя и неверного пароля возвращается одинаковая ошибка.
	user, err := h.queries.GetUserByUsername(r.Context(), username)
	if errors.Is(err, pgx.ErrNoRows) ||
		err == nil && bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		h.render(
			w,
			tmplLogin,
			loginPage{Error: "Неверное имя пользователя или пароль"},
			http.StatusUnauthorized,
		)
		return
	}
	if err != nil {
		slog.Error(
			"failed to get user by username",
			slog.Any("error", err),
			slog.Any("username", username),
		)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Новый токен не позволяет повторно использовать прежний идентификатор сессии после входа.
	if err := h.sessions.RenewToken(r.Context()); err != nil {
		slog.Error(
			"failed to renew session",
			slog.Any("error", err),
		)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	// В сессии хранятся только данные, нужные для авторизованных страниц.
	h.sessions.Put(r.Context(), session.UserKey, session.User{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// PostLogout удаляет сессию и возвращает пользователя на страницу входа.
func (h *Handlers) PostLogout(w http.ResponseWriter, r *http.Request) {
	if err := h.sessions.Destroy(r.Context()); err != nil {
		slog.Error(
			"failed to destroy session",
			slog.Any("error", err),
		)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

package session

// UserKey — ключ данных пользователя в сессии.
const UserKey = "user"

// User содержит данные, доступные в сессии после успешного входа.
type User struct {
	ID       int64
	Username string
	Role     string
}

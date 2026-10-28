-- Поиск учётной записи по имени для проверки пароля при входе.
-- name: GetUserByUsername :one
SELECT id, role, username, password_hash
FROM "user"
WHERE username = $1;

[linux]
set shell := ["bash", "-cu"]

[windows]
set shell := ["powershell.exe", "-NoProfile", "-Command"]

set dotenv-load

default:
    @just --list

# Создание базы данных
[linux]
db-init:
    runuser --preserve-environment -u postgres -- env -u PGHOST psql -X -v ON_ERROR_STOP=1 -U postgres -d postgres -f db/bootstrap.sql

# Создание базы данных
[windows]
db-init:
    psql -X -W -v ON_ERROR_STOP=1 -h localhost -U postgres -d postgres -f db/bootstrap.sql

# Сгенерировать миграции
create-migration name:
    migrate create -ext sql -dir db/migrations -seq -digits 3 {{ name }}

# Сгенерировать слой взаимодействия с базой данной
generate:
    sqlc generate

# Применить миграции вверх
migrate-up:
    migrate -path db/migrations -database 'postgres://@' up

# Применить одну миграцию вниз
migrate-down:
    migrate -path db/migrations -database 'postgres://@' down 1

# Заполнить базу тестовыми данными (один раз после миграций)
db-seed:
    psql -X -1 -v ON_ERROR_STOP=1 -f db/seed.sql

# Форматировать код
fmt:
    golangci-lint fmt ./...
    npx --no-install prettier --write internal/templ/templates

# Проверить стиль кода
lint:
    golangci-lint run ./...
    npx --no-install prettier --check internal/templ/templates

# Провести тесты кода
test:
    go test ./...

# Запустить приложение
run:
    go run .

# Собрать приложение
build:
    go build .
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
    runuser --preserve-environment -u postgres -- psql -X -v ON_ERROR_STOP=1 -d postgres -f db/bootstrap.sql

# Создание базы данных
[windows]
db-init:
    psql -X -v ON_ERROR_STOP=1 -h localhost -U postgres -d postgres -f db/bootstrap.sql

# Сгенерировать слой взаимодействия с базой данной
generate:
    sqlc generate

# Форматировать код
fmt:
    golangci-lint fmt ./...

# Проверить стиль кода
lint:
    golangci-lint run ./...

# Провести тесты кода
test:
    go test ./...

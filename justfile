[linux]
set shell := ["bash", "-cu"]

[windows]
set shell := ["powershell.exe", "-NoProfile", "-Command"]

set dotenv-load

default:
    @just --list

[linux]
db-init:
    runuser --preserve-environment -u postgres -- psql -X -v ON_ERROR_STOP=1 -d postgres -f db/bootstrap.sql

[windows]
db-init:
    psql -X -v ON_ERROR_STOP=1 -h localhost -U postgres -d postgres -f db/bootstrap.sql

generate:
    sqlc generate

fmt:
    golangci-lint fmt ./...

lint:
    golangci-lint run ./...

test:
    go test ./...

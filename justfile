default:
    @just --list

generate:
    sqlc generate

fmt:
    golangci-lint fmt ./...

lint:
    golangci-lint run ./...

test:
    go test ./...

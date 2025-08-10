
# AGENTS.md

This document provides guidelines for AI agents working on this codebase.

## Build, Lint, and Test

- **Build:** `go build -o ./tmp/main cmd/qc_api/main.go`
- **Run:** `./tmp/main`
- **Test All:** `go test ./...`
- **Test Package:** `go test ./internal/auth/` or `go test -v ./internal/auth/`
- **Test Single:** `go test -run TestRegister ./internal/auth/`
- **Lint:** `golangci-lint run` (install if needed)

## Code Style

- **Formatting:** Use `gofmt` for all Go code
- **Imports:** Group into blocks: standard library, third-party, internal (with blank lines between)
- **Naming:** PascalCase for exported types/functions, camelCase for variables/unexported
- **Database Models:** Inherit from `db.BaseModel` with GORM tags, use UUID primary keys
- **HTTP Handlers:** Use Echo framework, return JSON with proper status codes
- **Error Handling:** Explicit `if err != nil` checks, use `utils.ErrorResponse` for HTTP errors
- **Testing:** Use testify/assert, separate test packages (e.g., `package auth_test`)
- **Project Structure:** `cmd/` for main, `internal/` for business logic by feature

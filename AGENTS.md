# AGENTS.md

- Go module: `github.com/jzes/reqman`.
- App entrypoint is `main.go`; it loads requests from the directory argument or `.` and starts the Bubble Tea TUI in alt screen.
- Main packages: `internal/presentation` owns TUI state and key handling, `internal/adapter/requestfile` loads/saves `.curl` files, `internal/adapter/requesthttp` sends requests, `internal/domain/request` is the shared model.

## Commands

- Run all tests: `go test ./...`
- Run the TUI package tests: `go test ./internal/presentation`
- Run the HTTP adapter tests: `go test ./internal/adapter/requesthttp`
- Run the app locally: `go run . [requests-dir]`

## Repo-specific behavior

- Requests are file-based and only `.curl` files are parsed by default.
- `requestfile.Write` writes a single shell-quoted `curl` command and sorts headers before serializing.
- `requesthttp.Client` prepends `http://` when a request URL has no scheme.
- The `Do` panel is focused like an input, but it is read-only; `Enter` and `Space` still trigger the request when it is focused.

## Editing / testing

- If you change panel focus or key handling, update `internal/presentation/model_test.go` alongside the code.

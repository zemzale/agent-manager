AGENTS — repository conventions and quick commands

# Commands
- Build:    task build
- Format:   task fmt
- Vet/Lint: task vet
- Test all: task test
- Run a single package tests: go test ./internal/git -v
- Run a single test by name: go test ./internal/git -run "^TestClient$" -v

# Code style 
- Formatting: use go fmt / goimports; commit formatted code only.
- Imports: group/ordering handled by goimports (std lib, blank line, 3rd-party, blank line, local).
- Errors: return errors upward; wrap with fmt.Errorf("...: %w", err) when adding context; check and handle errors promptly.
- Context: accept context.Context on public functions that perform I/O or long-running work and pass it through.
- Tests: keep tests deterministic; use table-driven tests for multiple cases; use t.Run for subtests.

# Notes
- Keep changes minimal and idiomatic Go
- Run vet/lint/tests after a chnage
- Use `go doc` to review other package APIs

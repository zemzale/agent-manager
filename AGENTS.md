AGENTS — repository conventions and quick commands

Build / Lint / Test
- Build:    go build ./...
- Install deps: go mod download
- Format:   go fmt ./...  (or goimports -w . for import grouping)
- Vet/Lint: go vet ./...; golangci-lint run (if available)
- Test all: go test ./... -v
- Run a single package tests: go test ./internal/git -v
- Run a single test by name: go test ./internal/git -run "^TestClient$" -v

Code style (short)
- Formatting: use go fmt / goimports; commit formatted code only.
- Imports: group/ordering handled by goimports (std lib, blank line, 3rd-party, blank line, local).
- Naming: exported identifiers use PascalCase; unexported use camelCase; avoid stuttering (pkg.Type, not pkg.PkgType).
- Types: prefer concrete types unless an interface is required for mocking; keep interfaces small and name single-method interfaces with -er when sensible.
- Errors: return errors upward; wrap with fmt.Errorf("...: %w", err) when adding context; check and handle errors promptly.
- Context: accept context.Context on public functions that perform I/O or long-running work and pass it through.
- Tests: keep tests deterministic; use table-driven tests for multiple cases; use t.Run for subtests.

Cursor / Copilot rules
- No .cursor rules or GitHub Copilot instruction file detected in this repository. If added, include their rules here.

Agent note
- Keep changes minimal and idiomatic Go; run go vet and tests before committing.

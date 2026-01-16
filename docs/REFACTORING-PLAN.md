# Refactoring plan for wg-manage

## Goals

- Modernize the codebase for current Go tooling and idioms.
- Keep CLI behavior stable and validated by acceptance tests.
- Improve readability, error handling, and project structure.
- Replace obsolete CI and keep GitHub Actions as the only pipeline.

## Step-by-step plan

### 1. Baseline and documentation
- Ensure the acceptance test suite covers all CLI commands.
- Capture functional behavior in `docs/FUNC.md`.

### 2. Dependencies and Go version
- Upgrade `go.mod` to a modern Go version (latest stable).
- Update dependencies to their latest compatible versions.
- Run `go mod tidy` after dependency changes.

### 3. Project structure modernization
- Adopt `cmd/wg-manage` as the sole entrypoint and remove redundant root `main.go`/`command.go`.
- Move shared CLI wiring to a reusable package (e.g., `internal/cli`).
- Keep command packages under `cmd/<command>`.

### 4. Code readability and maintainability
- Replace deprecated APIs (e.g., `ioutil`) with modern equivalents.
- Use `os.MkdirAll` with readable permissions constants.
- Avoid shadowed or overly generic variable names.
- Introduce helper functions for repeated patterns in command handling.

### 5. Error handling and logging
- Normalize error handling to return errors where possible.
- Keep fatal exits at the command boundary (CLI layer).
- Add contextual error messages for file and parsing failures.

### 6. Acceptance tests and verification
- Ensure acceptance tests pass before and after each refactor step.
- Use `go test ./...` to validate changes.

### 7. CI updates
- Remove Azure Pipelines configuration (`azure-pipelines.yml`).
- Update GitHub Actions workflow to use the new Go version and run tests.

## Execution notes

- Refactor in small, verifiable steps.
- Run the acceptance tests after each structural change.
- Do not modify `snapcraft.yaml`.

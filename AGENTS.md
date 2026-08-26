# AGENTS.md - Development Guide for Agentic Coding

This file provides comprehensive guidelines for autonomous coding agents working on the wg-manage project.

## Build, Lint, and Test Commands

### Build
```bash
go build -v -o wg-manage .
```
Build for specific OS:
```bash
GOOS=windows go build -v -o wg-manage.exe .
GOOS=linux go build -v -o wg-manage .
GOOS=linux GOARCH=arm64 go build -v -o wg-manage-arm64 .
GOOS=darwin go build -v -o wg-manage .
```

### Linting & Formatting
```bash
go vet ./...
gofmt -w .
goimports -w .
```

### Testing
No existing test files in repository. When adding tests:
```bash
go test ./...
go test -v ./...
go test -run TestName ./...
```
Use table-driven tests with `t.Run` for multiple scenarios.

### Dependencies
```bash
go mod tidy
go mod download
```

## Code Style Guidelines

### Imports
- Use `goimports` to automatically organize imports
- Organize in groups: stdlib, third-party, local packages
- Example order:
  ```go
  import (
    "flag"
    "fmt"
    "os"
    
    "golang.zx2c4.com/wireguard/wgctrl"
    
    "github.com/ofcoursedude/wg-manage/models"
  )
  ```

### Formatting & Types
- Run `gofmt` on all files before committing
- Use `goimports` to manage imports
- Use lowercase package names (single word, no underscores)
- Use MixedCaps for exported symbols, mixedCaps for unexported
- Use pointers for optional fields in structs:
  ```go
  type Peer struct {
    Name string
    Description *string  // optional
    ListenPort *int      // optional
  }
  ```

### Naming Conventions
- **Packages**: lowercase, descriptive (e.g., `models`, `wg`, `utils`)
- **Functions**: mixedCaps for unexported, MixedCaps for exported
- **Interfaces**: -er suffix when possible (Reader, Writer, Formatter)
- **Constants**: MixedCaps for exported, mixedCaps for unexported
- **Variables**: descriptive names, avoid single letters except loop indices
- **Package declarations**: NEVER duplicate - each file has exactly ONE `package` line

### Error Handling
- Check errors immediately after function calls
- Use `fmt.Errorf` with `%w` verb to wrap errors with context
- Exported error helpers like `HandleError` in `utils/utils.go`
- Place error returns as the last return value
- Name error variables `err`
- Example from codebase:
  ```go
  func HandleError(err error, message string) {
    if err != nil {
      log.Fatal(message, err)
    }
  }
  ```

### Comments
- Avoid obvious/redundant comments - write self-documenting code
- Comment complex business logic and non-obvious algorithms
- Export public functions with comment starting with function name
- Use `//` for most comments, `/* */` sparingly
- **Avoid**: TODO comments, divider comments, outdated comments

### Code Patterns in this Project

#### Command Pattern
All commands implement the `Command` interface:
```go
type Command interface {
  PrintHelp()
  Run()
  ShortCommand() string
  LongCommand() string
}

// Commands registered in main.go availableCommands slice
```

#### Configuration Loading
- YAML files with `gopkg.in/yaml.v2` or `yaml.v3`
- Struct tags: `yaml:"fieldName,omitempty"`
- Use pointers for optional fields

#### Models
- Defined in `models/models.go`
- Key types: `Configuration`, `Peer`, `InterfaceSection`, `PeerSection`
- Include YAML tags on all struct fields
- Documentation in inline comments (see models.go for examples)

#### Error Handling
- Use `utils.HandleError(err, "message")` for fatal errors
- Log and exit on critical failures
- Return errors from functions, don't log at the source

### File Organization
```
.
├── cmd/              # Command implementations (add, bootstrap, etc.)
├── models/           # Configuration structs
├── utils/            # Utility functions (error handling, file ops)
├── wg/               # WireGuard-specific logic
├── oscustom/         # OS-specific code (windows.go, nonwindows.go)
├── go.mod / go.sum   # Dependencies
└── main.go           # Main entry point
```

## Best Practices

### Dependency Management
- Use Go modules exclusively
- Keep dependencies minimal - already using: barcode, wireguard/wgctrl, yaml
- Run `go mod tidy` after dependency changes

### Testing Approach
- Write tests alongside implementation in `*_test.go` files
- Use table-driven tests for multiple scenarios
- Test both success and error cases
- Use `t.Helper()` for test helper functions

### OS-Specific Code
- Place platform-specific code in `oscustom/` directory
- Use build tags: `//go:build windows` or `// +build windows`
- See `oscustom/windows.go` and `oscustom/nonwindows.go` for examples

## Repository Instructions (Required Reading)

This repository follows guidelines from:
- `.github/instructions/go.instructions.md` - Comprehensive Go best practices (374 lines)
- `.github/instructions/self-explanatory-code-commenting.instructions.md` - Comment guidelines
- `.github/instructions/task-implementation.instructions.md` - Task implementation workflow

Key requirements from Go instructions:
- Follow idiomatic Go (Effective Go, Go Code Review Comments)
- Implement `Command` interface for new CLI commands
- Use early return pattern to keep code left-aligned
- Minimize allocations in hot paths
- Handle errors immediately after function calls
- Never ignore errors without documentation
- Use `errors.Is` and `errors.As` for error checking

## Quick Reference

### Adding a New Command
1. Create `cmd/newcommand/newcommand.go`
2. Implement `Command` interface (ShortCommand, LongCommand, Run, PrintHelp)
3. Import in `main.go`
4. Add to `availableCommands` slice
5. Use `flag` package for CLI arguments (see `cmd/add/add.go`)

### Adding a Test
1. Create `path/to/file_test.go`
2. Use `Test_functionName_scenario` naming
3. Use table-driven tests: `for _, tt := range tests { t.Run(...) }`
4. Run: `go test ./...` or `go test -run TestName ./...`

### File Operations
- Use `utils.SaveToFile(name string, data []byte)`
- Use YAML for config serialization
- Handle file errors with `HandleError`

## Go Version
Project targets Go 1.16+ (defined in go.mod as 1.16)
- Use modern stdlib APIs available in 1.16
- If using features from Go 1.22+, add version checks

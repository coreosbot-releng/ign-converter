@AGENTS.md

## Claude Code Specific Workflows

This file imports the main repository specification from AGENTS.md and adds Claude Code-specific instructions.

### Task Management

When working on multi-step tasks, use `TodoWrite` to break down work and track progress.

### Agent Usage

- Use the **Explore** agent for broad codebase searches when simple Grep/Glob isn't enough
- Use **Plan** mode for complex changes affecting multiple files (e.g. adding a new translation module)
- Delegate independent research to subagents to keep main context clean

### Testing

Always run `go test ./...` before committing changes. The CI also runs `golangci-lint run -E=gofmt`, so ensure code is gofmt-formatted.

### Vendoring

After any `go.mod` changes, always run `go mod tidy && go mod vendor` before committing. The `vendor/` directory is committed to the repository.

# ign-converter

Library and CLI for converting Ignition configs between spec versions (v1/v2.x <-> v3.x, and v3.x downgrade translations).

## Tech Stack

- **Language**: Go 1.18+
- **Key Dependencies**: `github.com/coreos/ignition/v2` (spec types + translator), `github.com/coreos/ignition` (v1/v2 types), `github.com/stretchr/testify` (assertions)
- **Linting**: golangci-lint v1.51+ with `-E=gofmt`
- **CI**: GitHub Actions (`.github/workflows/go.yml`)

## Architecture

```
internal/           # CLI entrypoint (main package) - converts v2.4 <-> v3.1
translate/          # Translation packages (one per version pair)
  v23tov30/         # v2.3 -> v3.0
  v24tov31/         # v2.4 -> v3.1
  v30tov22/         # v3.0 -> v2.2
  v31tov22/         # v3.1 -> v2.2
  v31tov24/         # v3.1 -> v2.4
  v32tov22/         # v3.2 -> v2.2
  v32tov24/         # v3.2 -> v2.4
  v32tov31/         # v3.2 -> v3.1
  v33tov32/         # v3.3 -> v3.2
  v34tov33/         # v3.4 -> v3.3
util/               # Shared helpers (pointer utils, error types)
vendor/             # Vendored dependencies (committed)
translate_test.go   # All translation tests (root package)
```

## Build Commands

- `go test ./...` - Run all tests
- `go vet ./...` - Static analysis
- `go mod tidy && go mod vendor` - Update and vendor dependencies
- `golangci-lint run -E=gofmt` - Lint (enforces gofmt)

## Code Style

- gofmt enforced (no tabs vs spaces debate)
- Apache 2.0 license header on all `.go` files
- Each translation module is a single file: `translate/vXXtoYY/vXXtoYY.go`
- Import the newer spec types as `old_types` and target spec types as `types`
- Use `util.StrP()`, `util.IntP()`, `util.BoolP()` helpers for pointer creation in tests

## Testing

- **Framework**: Go `testing` + `testify/assert`
- **Test file**: `translate_test.go` (root package `ignconverter`)
- **Pattern**: Exhaustive config structs (`exhaustiveConfigX_Y`) covering all fields, round-trip and downgrade tests
- **Pre-commit**: Run `go test ./...` before committing

## Translation Module Pattern

Each `translate/vNEWtoOLD/` module follows this structure:
1. `translateIgnition()` - sets target version string
2. `translateConfig()` - registers custom translators, calls `translate.NewTranslator().Translate()`
3. `Translate()` (exported) - validates input, calls `recursiveIsOk()` to reject unsupported features, translates, validates output
4. `recursiveIsOk()` - walks config tree via reflection, returns errors for features not in the target spec

## Commit Conventions

**Format**: `component: description` (lowercase, imperative)

Examples from history:
- `translate: add translation from v34 to v33`
- `go.mod update to ignition 2.15.0`
- `workflows: add Go 1.20`
- `vendor: Bump to Ignition 2.11.0`

**Style**: Imperative mood, lowercase, no period. Component is the directory or area being changed.

## Important Rules

- `vendor/` is committed -- always run `go mod vendor` after dependency changes
- Do not modify vendored files directly
- New translation modules must check for features unsupported in the target spec via `recursiveIsOk()`
- Tests use exhaustive configs that exercise all spec fields
- License: Apache 2.0 (Red Hat copyright)

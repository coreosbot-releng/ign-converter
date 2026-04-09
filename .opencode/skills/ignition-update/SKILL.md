---
name: ignition-update
description: Update the Ignition library dependency to the latest release, re-vendor, and fix any breakage
---

# Ignition Library Update

## What it does

1. Discovers the current Ignition version from `go.mod` and the latest release from GitHub
2. Updates `go.mod` to the latest Ignition version
3. Runs `go mod tidy && go mod vendor`
4. Runs `go test ./...` to verify existing translations still work
5. Fixes any compilation or lint issues caused by the update

## Prerequisites

- Go toolchain installed
- `gh` CLI available (for querying latest release), or internet access for `go list -m -versions`

## Usage

```bash
# Fully automatic - discovers latest version from GitHub
/ignition-update
```

No arguments required.

## Workflow

### Step 1: Discover Versions

**1a. Read the current Ignition version from go.mod:**

```bash
grep 'coreos/ignition/v2' go.mod
```

Extract the current version (e.g., `v2.20.0`).

**1b. Find the latest Ignition release:**

```bash
gh release list --repo coreos/ignition --limit 10
```

Or:

```bash
go list -m -versions github.com/coreos/ignition/v2 | tr ' ' '\n' | tail -5
```

Pick the latest stable release (not a pre-release/RC).

**1c. Compare versions:**

If the current version is already the latest, inform the user and stop:
> "go.mod already has the latest Ignition version (vX.Y.Z). Nothing to update."

**1d. Confirm with user:**

> "Current: vX.Y.Z -> Latest: vA.B.C. Proceed with update?"

### Step 2: Update go.mod

Edit `go.mod` to change the `github.com/coreos/ignition/v2` version to the latest.

Only change the version in the `require` block, for example:

```
github.com/coreos/ignition/v2 v2.20.0
```

becomes:

```
github.com/coreos/ignition/v2 v2.26.0
```

Do NOT manually edit indirect dependencies - `go mod tidy` handles those.

### Step 3: Re-vendor

```bash
go mod tidy
go mod vendor
```

If `go mod tidy` fails, it usually means:
- The version doesn't exist yet (check the tag exists on GitHub)
- A transitive dependency conflict (read the error, may need to update the Go version in `go.mod`)

If the Go version in `go.mod` needs bumping (the new Ignition version requires a newer Go), update the `go` directive as well. Check what Go version the Ignition release requires:

```bash
# After vendoring, check what Ignition's go.mod requires
cat vendor/github.com/coreos/ignition/v2/go.mod | head -5
```

### Step 4: Run Tests

```bash
go test ./...
```

If tests pass, proceed to Step 6.

### Step 5: Fix Breakage (if tests fail)

Common causes of breakage after an Ignition update:

**5a. Type definition changes in existing specs**

Ignition sometimes adds new fields, validation, or changes types in existing spec versions. If a translation module references a field that changed:

- Read the compilation error to identify which file and type
- Compare the vendored type definition before and after
- Update the translation code to match the new type structure

**5b. New validation in existing specs**

Ignition may add stricter validation to existing spec versions. Test configs that were previously valid may now fail validation:

- Read the test failure to identify which config is rejected
- Update the test config to satisfy the new validation
- Or update the translation to handle the validation change

**5c. Transitive dependency API changes**

If a transitive dependency (e.g., `go-systemd`, `go-json`) changed its API:

- Read the compilation error
- Update the affected code to use the new API

After fixing, re-run:

```bash
go test ./...
```

Repeat until all tests pass.

### Step 6: Check Formatting

```bash
gofmt -l .
```

If any files are listed, format them:

```bash
gofmt -w .
```

### Step 7: Summarize Changes

After completion, summarize what changed for the user:

- Ignition version: old -> new
- Go version bumped (if applicable)
- New spec types introduced (if any new `v3_X` directories appeared in vendor)
- Transitive dependencies updated
- Any fixes applied to existing translations
- Test results

Suggest a commit message following the repo convention:

```
go.mod: update to Ignition vX.Y.Z
```

## Checklist Coverage

- [x] Latest version discovery from GitHub / Go module registry
- [x] Current version detection from go.mod
- [x] go.mod update
- [x] Re-vendoring with `go mod tidy && go mod vendor`
- [x] Test verification
- [x] Breakage diagnosis and fix
- [x] Format check

## What's NOT covered

- [ ] Adding a new spec translation (use `/add-spec-translation` for that)
- [ ] Creating a GitHub PR
- [ ] Updating CI workflow Go versions (if the new Ignition requires a newer Go than CI tests)

## Example Output

```
Discovering Ignition versions...
  Current (go.mod): v2.20.0
  Latest (GitHub):  v2.26.0

Update Ignition v2.20.0 -> v2.26.0? (y/n)

Updating go.mod...
Running go mod tidy && go mod vendor...

New spec types detected:
  - v3_6 (cex, clevis, config, device, directory, disk, ...)

Transitive dependencies updated:
  - aws-sdk-go -> aws-sdk-go-v2
  - go-systemd/v22: v22.5.0 -> v22.6.0
  - testify: v1.9.0 -> v1.11.1

Running go test ./...
PASS

gofmt check: clean

Suggested commit message:
  go.mod: update to Ignition v2.26.0
```

## References

- Design document: `.opencode/skills/ignition-update/DESIGN.md`
- Related skill: `.opencode/skills/add-spec-translation/SKILL.md`
- Example commits: `04f5d8a`, `fc3acd6`, `dc5d7d4`, `9b0adda`

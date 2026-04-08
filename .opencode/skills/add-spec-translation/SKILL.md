---
name: add-spec-translation
description: Add a new Ignition spec version downgrade translation module with dependency update, translation code, and tests
---

# Add Spec Version Translation

## What it does

1. Discovers the next spec version to translate by inspecting existing `translate/` directories and the latest Ignition release
2. Updates the Ignition dependency in `go.mod` to pull in the new spec types
3. Runs `go mod tidy && go mod vendor` to vendor the new types
4. Analyzes the diff between the new and old spec type definitions to identify what changed
5. Creates a new translation package `translate/vNEWtoOLD/vNEWtoOLD.go`
6. Adds comprehensive tests to `translate_test.go`
7. Validates everything with `go test ./...`

## Prerequisites

- Go toolchain installed (1.18+)
- The new Ignition spec version must be released and available in the Go module registry

## Usage

```bash
# Fully automatic - discovers versions from code and GitHub
/add-spec-translation
```

No arguments are required. The skill determines the versions automatically.

## Workflow

### Step 1: Discover the Next Translation Pair

Determine what already exists and what comes next.

**1a. Find the highest existing translation:**

```bash
ls translate/
```

Look at the directory names (e.g., `v34tov33`, `v35tov34`, `v36tov35`). Parse the highest NEW version from the directory named `vNEWtoOLD`. For example, if the latest is `v34tov33`, then `OLD_COMPACT = 34` and `NEW_COMPACT = 35`.

The next translation to create is `v{OLD_COMPACT+1}tov{OLD_COMPACT}` where `OLD_COMPACT` is the NEW from the latest existing directory.

**1b. Find the latest Ignition release version:**

```bash
gh release list --repo coreos/ignition --limit 5
```

Or check Go module versions:

```bash
go list -m -versions github.com/coreos/ignition/v2 | tr ' ' '\n' | tail -5
```

Pick the latest release that includes the new spec version's types.

**1c. Verify the new spec types exist in that release:**

The Ignition library publishes spec types at `github.com/coreos/ignition/v2/config/v{NEW_UNDERSCORE}/types/`. Confirm the new spec types package exists by checking the release notes or the module contents.

**1d. Derive all variables:**

From the discovered versions, compute:
- `NEW_MAJOR_MINOR` = e.g., `3.7` (with dot)
- `OLD_MAJOR_MINOR` = e.g., `3.6` (with dot)
- `NEW_UNDERSCORE` = e.g., `3_7` (for Go package paths)
- `OLD_UNDERSCORE` = e.g., `3_6` (for Go package paths)
- `NEW_COMPACT` = e.g., `37` (no dot, for directory names)
- `OLD_COMPACT` = e.g., `36` (no dot, for directory names)
- `PACKAGE_NAME` = e.g., `v37tov36`
- `IGNITION_VERSION` = e.g., `v2.28.0` (from the latest release)

**1e. Confirm with the user** before proceeding:

Present the discovered values and ask the user to confirm:
> "I'll create translation v{NEW_MAJOR_MINOR} -> v{OLD_MAJOR_MINOR} using Ignition {IGNITION_VERSION}. Proceed?"

### Step 2: Update Ignition Dependency

Read `go.mod` and update the Ignition version:

```
github.com/coreos/ignition/v2 {IGNITION_VERSION}
```

Then run:

```bash
go mod tidy
go mod vendor
```

Verify the new spec types are available:

```bash
ls vendor/github.com/coreos/ignition/v2/config/v{NEW_UNDERSCORE}/types/
```

If the types directory doesn't exist, the Ignition version may not include the new spec yet. Ask the user for guidance.

### Step 3: Analyze Spec Differences

Compare the type definitions between the new and old spec versions to understand what changed:

```bash
# List types in new spec
ls vendor/github.com/coreos/ignition/v2/config/v{NEW_UNDERSCORE}/types/

# List types in old spec
ls vendor/github.com/coreos/ignition/v2/config/v{OLD_UNDERSCORE}/types/
```

Read and compare key type files. Look for:
- New types (files in NEW that don't exist in OLD)
- New fields in existing types (diff struct definitions)
- Changed field types (e.g., `string` -> `*string`)

Also read the upstream translator to understand the translation direction:

```bash
cat vendor/github.com/coreos/ignition/v2/config/v{NEW_UNDERSCORE}/translate/translate.go
```

This file translates from OLD to NEW. We need to **reverse** it: translate from NEW to OLD.

### Step 4: Read the Most Recent Existing Translation

Read the most recent translation module to use as a structural template:

```bash
# Find the latest translation directory
ls translate/

# Read it
cat translate/{LATEST}/v{LATEST}.go
```

Also read `translate_test.go` to understand the test structure:

```bash
cat translate_test.go
```

### Step 5: Create the Translation File

Create `translate/v{NEW_COMPACT}tov{OLD_COMPACT}/v{NEW_COMPACT}tov{OLD_COMPACT}.go`

The file MUST follow this structure:

```go
// Copyright {YEAR} Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package {PACKAGE_NAME}

import (
    "fmt"
    "reflect"

    "github.com/coreos/ignition/v2/config/translate"
    "github.com/coreos/ignition/v2/config/v{OLD_UNDERSCORE}/types"
    old_types "github.com/coreos/ignition/v2/config/v{NEW_UNDERSCORE}/types"
    "github.com/coreos/ignition/v2/config/validate"
)
```

**Key rules for the translation code:**

1. **`translateIgnition()`** is ALWAYS required - it sets `ret.Version = types.MaxVersion.String()`

2. **Custom translate functions** are needed for any type where:
   - Fields were added in the new spec that don't exist in the old spec
   - Field types changed between versions
   - The function should copy only the fields that exist in BOTH versions

3. **`translateConfig()`** registers all custom translators and runs the full translation

4. **`Translate()`** is the public entry point:
   - Validates the input config
   - Calls `checkValue()` to reject unsupported features
   - Runs `translateConfig()`
   - Validates the output config (sanity check)

5. **`checkValue()`** uses reflection to walk the config and reject features not supported in the OLD spec:
   - Check for new types that don't exist in OLD
   - Check for new field values that aren't valid in OLD
   - Recursively walk structs, slices, and pointers
   - Return clear error messages: `fmt.Errorf("invalid input config: {feature} is not supported in spec v{OLD_MAJOR_MINOR}")`

6. **Mode bit masking**: If the OLD spec version is < 3.4, mask special mode bits in `translateFileEmbedded1()` and `translateDirectoryEmbedded1()` using `*old.Mode & ^07000`. If the OLD spec is >= 3.4, special mode bits are supported and no masking is needed.

### Step 6: Add Tests to translate_test.go

Add the following to `translate_test.go`:

1. **Import the new types package**:
   ```go
   types{NEW_UNDERSCORE} "github.com/coreos/ignition/v2/config/v{NEW_UNDERSCORE}/types"
   ```

2. **Import the new translate package**:
   ```go
   "github.com/coreos/ign-converter/translate/{PACKAGE_NAME}"
   ```

3. **Create `nonexhaustiveConfig{NEW_UNDERSCORE}`** - A config variable that exercises ALL undeprecated fields in the new spec version. Model it after the existing `nonexhaustiveConfig` variables, adding fields for any new features.

4. **Create `downtranslateConfig{OLD_UNDERSCORE}`** (if it doesn't already exist as a `nonexhaustiveConfig`) - The expected config after translation. This should be the same as `nonexhaustiveConfig{NEW_UNDERSCORE}` but:
   - Version string set to OLD version (e.g., `"3.6.0"`)
   - New-only fields removed
   - Mode bits masked if applicable

5. **Add test function**:
   ```go
   func TestV{NEW_COMPACT}toV{OLD_COMPACT}(t *testing.T) {
       // Test successful translation
       res, err := {PACKAGE_NAME}.Translate(nonexhaustiveConfig{NEW_UNDERSCORE})
       assert.NoError(t, err)
       assert.Equal(t, downtranslateConfig{OLD_UNDERSCORE}, res)

       // Test error cases for unsupported features
       // For each new-only feature, create a config that uses it
       // and verify Translate() returns an error
   }
   ```

### Step 7: Validate

Run the tests:

```bash
go test ./...
```

Check formatting:

```bash
gofmt -l .
```

If tests fail, analyze the error and fix. Common issues:
- Missing custom translator for a changed type (causes panic or wrong output)
- Incorrect field mapping in custom translator
- Missing error case in `checkValue()`
- Test config doesn't match expected output (version string, missing fields)

### Step 8: Verify CI Compatibility

The CI workflow (`.github/workflows/go.yml`) runs:
- `go test ./...` across multiple Go versions
- `golangci-lint` with `-E=gofmt`

Make sure the code passes both.

## Checklist Coverage

This skill automates:
- [x] Version discovery from existing code and latest Ignition release
- [x] Ignition dependency update (`go.mod`, vendor)
- [x] Translation file creation with correct structure
- [x] Custom translators for changed types
- [x] Feature rejection for unsupported features
- [x] Mode bit masking when applicable
- [x] Test creation with exhaustive configs
- [x] Error case testing
- [x] Validation with `go test`

## What's NOT covered

- [ ] Complex translation logic for fundamentally restructured types (rare, requires judgment)
- [ ] Updating `README.md` with new supported versions
- [ ] Creating a GitHub PR

## Example Output

When you run `/add-spec-translation`, you should see:

```
Discovering next translation pair...
  Existing translations: v32tov31, v33tov32, v34tov33
  Highest existing: v34 -> v33
  Next translation needed: v35 -> v34

Checking latest Ignition release...
  Latest: v2.20.0 (includes spec v3.5 types)

I'll create translation v3.5 -> v3.4 using Ignition v2.20.0. Proceed? (y/n)

Updating go.mod to Ignition v2.20.0...
Running go mod tidy && go mod vendor...

Analyzing spec differences between v3.5 and v3.4...
Found changes:
  - New type: Cex
  - Changed fields in Luks: added OpenOptions, Discard
  - Changed fields in Tang: added Advertisement

Creating translate/v35tov34/v35tov34.go...
  - translateIgnition()
  - translateLuks() (new fields: OpenOptions, Discard)
  - translateTang() (new field: Advertisement)
  - translateConfig()
  - Translate()
  - checkValue() (rejects: Cex type)

Updating translate_test.go...
  - Added nonexhaustiveConfig3_5
  - Added downtranslateConfig3_4
  - Added TestV35toV34()

Running go test ./...
PASS
ok   github.com/coreos/ign-converter  0.015s

All checks passed!
```

## References

- Design document: `.opencode/skills/add-spec-translation/DESIGN.md`
- Example - v36 to v35: `.opencode/skills/add-spec-translation/examples/v36tov35.md`
- Example - v34 to v33: `.opencode/skills/add-spec-translation/examples/v34tov33.md`
- Existing translations: `translate/v34tov33/`, `translate/v33tov32/`, `translate/v32tov31/`
- CI workflow: `.github/workflows/go.yml`
- Upstream Ignition: `vendor/github.com/coreos/ignition/v2/config/`

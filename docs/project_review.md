# Dyre Project Review

This review summarizes issues found via code inspection plus `go build`/`go test` runs, with precise file references and suggested fixes. All paths are repository‑relative.

## Test And Build Snapshot

- `go build ./...` succeeded.
- `go test ./...` results:
  - PASS: `endpoint`, `parser`, `transpiler` packages.
- `go vet ./...`: no issues reported.

---


## 2) Invalid `go` directive in `go.mod`

- File: `go.mod:3`
- Symptom: `go 1.24.4` is not valid. The `go` directive must be `major.minor`.
- Additional context: Tests use Go 1.22 features (e.g., `for i := range matchLen` in `endpoint/endpoint_parser_test.go:184`).
- Fix: Set to a valid version matching the language features, e.g.: `go 1.22`.
- Rationale: Ensures module toolchain works and matches features used in tests and code.

## 6) `datetime` builtin converts to `date`

- File: `transpiler/builtins.go:145–147`
- Symptom: Returns `objectType.DATETIME` but emits `CONVERT(date, %s, 127)`.
- Fix: Use a datetime target type consistent with the return type, e.g. `CONVERT(datetime2, %s, 127)`.
- Rationale: Aligns emitted SQL with the declared DyRe object type.

## 7) Endpoint JSON parsing: unsafe type assertions for required fields

- File: `endpoint/endpoint_parser.go:66–79`
- Symptom: Uses `m["name"].(string)` and `tableName.(string)` without checking types; malformed JSON can panic.
- Fix: Use the existing `parseString` helper for `name` and `tableName`:
  - `name, err := parseString(m, "name")`; `table, err := parseString(m, "tableName")` and handle errors.
- Rationale: Avoids panics and returns actionable parse errors when types don’t match.

## 8) String literal escaping in SQL emission

- File: `object/object.go:72–75`
- Symptom: `String.String()` returns `fmt.Sprintf("'%s'", s.Value)` without escaping embedded single quotes. Current lexer disallows quotes inside string tokens, but robust emitters should still escape.
- Fix: Escape `'` by doubling within `s.Value` before formatting; e.g., `strings.ReplaceAll(s.Value, "'", "''")`.
- Rationale: Safer SQL generation if/when inputs contain single quotes or if lexer rules expand.

## 9) Release script only tests root package

- File: `publish:7`
- Symptom: Uses `go test .` which skips tests in subpackages; currently misses failing `lexer` tests.
- Fix: Switch to `go test ./...` and consider adding `go vet ./...`.
- Rationale: Ensures all packages are validated before push/tag publish.

## 10) Formatting issues in `tools/` module

- Command: `gofmt -s -l .` flags unformatted files:
  - `tools/cmd/dyre-config/main.go`
  - `tools/cmd/dyre-lsp/main.go`
  - `tools/tools.go`
- Fix: Run `gofmt -s -w` on these files (inside `tools/`).
- Rationale: Keeps codebase consistent with `gofmt` expectations per repo guidelines.

## 11) Documentation issues in README

- File: `README.md`
- Typos / syntax:
  - `FASLE` at `README.md:144` → `FALSE`.
  - `getCustomersWithBiling` at `README.md:240,305` → `getCustomersWithBilling`.
  - Go examples use single quotes for strings and rune literals in Go snippets, e.g. `Re.Request('Customers', ...)` at `README.md:203,251,309`; use double quotes in Go.
- API alignment:
  - Example shows `Re, dyre_err = dyre.Init("./dyre.json")` at `README.md:185`, but current `dyre.Init` does not return an error; if you adopt the `Init(...)(Dyre, error)` change (see Issue 13), update docs accordingly.
- Rationale: Improves accuracy and reduces confusion for users adopting the library.

## 12) Potential non-deterministic join JSON order (low risk now)

- File: `endpoint/endpoint.go:64–70`
- Symptom: `Endpoint.JSON()` iterates a map (`e.Joins`) to build JSON; map iteration order is random. Current tests pass because each sample endpoint uses ≤ 1 join, so order doesn’t matter.
- Fix (optional): Emit joins using deterministic ordering (e.g., iterate `e.JoinNames`) to avoid future flakiness if multiple joins are added to tests/fixtures.

## 13) Library panics on init/parse errors

- File: `dyre.go:13–24, 27–41`
- Symptom: `dyre.Init` and helpers call `log.Panic` on file or parse errors.
- Fix: Prefer returning errors rather than panicking for a library API:
  - Change signature to `func Init(path string) (Dyre, error)` and propagate errors.
- Rationale: Avoids crashing applications embedding the library; lets callers handle configuration errors gracefully.

---

## Suggested Next Steps

1. Fix the `lexer` test expectations and the JOIN/brace bugs (Issues 1, 3, 4) — these are fast and reduce test breakage/confusion.
2. Update `go.mod` (Issue 2) to a valid version that matches language features in tests.
3. Harden parsing (Issue 7) and `CastType` (Issue 5) to prevent panics on malformed inputs.
4. Align `datetime` builtin emission (Issue 6) and add tests for LEFT/RIGHT/FULL joins.
5. Improve safety/formatting/docs (Issues 8–11); consider `dyre.Init` API change (Issue 13) if you want a non-panicking library surface.

If helpful, I can open a focused PR applying these changes with tests for LEFT/RIGHT/FULL joins and `datetime` conversion.


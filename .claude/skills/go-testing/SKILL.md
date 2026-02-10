---
name: go-testing
description: Write and review Go unit tests and fuzz tests following project conventions. Use when writing tests, adding test coverage, reviewing test quality, or when the user mentions testing, unit tests, fuzz tests, table-driven tests, or test coverage for this Go CLI project.
---

# Go Testing

Write tests that follow the project's established patterns and the best practices in `docs/test-best-practices.md`.

## When to use

- Writing new unit tests or fuzz tests
- Reviewing existing tests for quality
- Adding test coverage to untested code

## Instructions

### Before writing tests

1. Read the function under test to understand its behavior and edge cases
2. Check for existing tests in the same `_test.go` file
3. Read `docs/test-best-practices.md` for the full reference

### Test structure

Always use **table-driven tests** with subtests:

```go
func TestFuncName(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {name: "should do X when Y", input: "a", want: "A"},
        {name: "should fail on empty input", input: "", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            got, err := FuncName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
            }
            if !tt.wantErr && got != tt.want {
                t.Fatalf("got=%q want=%q", got, tt.want)
            }
        })
    }
}
```

### Naming conventions

- Test functions: `TestTypeName_MethodName` or `TestFuncName`
- Test case names: use `should X when Y` format (lowercase, descriptive intent)
- Fuzz functions: `FuzzFuncName`

### Required patterns

1. **Always `t.Parallel()`** at both the top-level test and inside `t.Run`
2. **Always capture loop variable**: `tt := tt` before `t.Run` (required for parallel subtests in Go < 1.22)
3. **Use `t.Fatalf`** for assertions, not `t.Errorf`, unless later assertions still make sense after failure
4. **Prefer stdlib**: use `testing`, `httptest`, `os.CreateTemp` — avoid third-party test libraries unless already in use
5. **Use `t.Helper()`** in any shared test helper function

### Edge cases to always cover

- Empty string / nil / zero value
- Boundary values
- Invalid or malformed input
- Error paths (not just happy path)
- Unicode and special characters where relevant

### Fuzz tests

Write fuzz tests for parsers, validators, encoders/decoders:

```go
func FuzzFuncName(f *testing.F) {
    f.Add([]byte(`valid input`))
    f.Add([]byte(`edge case`))
    f.Add([]byte{})

    f.Fuzz(func(t *testing.T, data []byte) {
        // Must never panic
        FuncName(data)
    })
}
```

Seed the corpus with values from existing unit test cases.

### File organization

- Tests go in `*_test.go` next to the source file
- Fuzz tests go in a separate `*_fuzz_test.go` file when there are multiple fuzz functions
- Test helpers stay in the same test file unless shared across packages

### What NOT to do

- Don't test private internals unless there's no public API to test through
- Don't use exact error string matching (check `err != nil` or use `errors.Is`)
- Don't add comments explaining obvious test code
- Don't mock internal collaborators — mock at boundaries (HTTP, DB, filesystem)
- Don't write slow tests — no real network, no real filesystem, no sleep

### Running tests

```bash
go test ./...                    # All tests
go test ./cmd/...                # Package tests
go test -race ./...              # With race detector
go test -fuzz=FuzzName ./pkg/    # Run specific fuzz test
go test -v -run TestName ./pkg/  # Run specific test
```

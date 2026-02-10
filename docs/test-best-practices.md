# Unit Testing Best Practices

## Core Principles (Language-Agnostic)

- **Test behavior, not implementation**
  - Assert outputs, state changes, returned errors, and observable side effects.
  - Avoid testing private/internal details unless unavoidable.

- **Keep tests small and focused**
  - One test = one reason to fail.
  - Prefer many small tests over large multi-purpose ones.

- **Use Arrange / Act / Assert**
  - Clearly separate setup, execution, and assertions.

- **Make tests deterministic**
  - No real time, randomness, network, filesystem, or shared globals.
  - Inject abstractions for anything non-deterministic.

- **Name tests by intent**
  - `GivenX_WhenY_ThenZ` or `ShouldX_WhenY`.

- **Test edge cases**
  - Empty, nil, zero, boundaries, invalid inputs, weird Unicode.
  - Always include happy + failure paths.

- **Assert the contract**
  - Don’t just check “no error” if output matters.
  - Assert what callers actually depend on.

- **Avoid brittle assertions**
  - Avoid exact error strings unless part of the API.
  - Avoid order-dependent checks unless order is guaranteed.

- **Mock at boundaries**
  - Fake DBs, HTTP clients, queues.
  - Avoid mocking internal collaborators.

- **Keep unit tests fast**
  - Slow tests get skipped.
  - Push real integrations to integration tests.

- **Use property-based / fuzz testing**
  - Great for parsers, validators, encoders/decoders.

- **Treat tests as production code**
  - Refactor duplication.
  - Keep helpers clean and readable.

---

## Go-Specific Best Practices

- **Use table-driven tests**
  - Cover many scenarios with minimal noise.
  - Always include a `name` field.

- **Use subtests**
  - `t.Run` for clear scenario grouping.

- **Parallelize carefully**
  - Use `t.Parallel()` for independent tests only.
  - Capture loop variables correctly.

- **Prefer stdlib testing tools**
  - `testing`, `httptest`, `go test -race`.
  - Add `testify/require` only if it improves clarity.

- **Use `t.Helper()`**
  - Makes failures point to the caller, not helpers.

- **Fake via interfaces**
  - Inject interfaces for clocks, stores, clients.
  - Don’t over-abstract prematurely.

- **Use in-memory deps**
  - `httptest.Server`, in-memory stores, temp dirs.

- **Golden files sparingly**
  - Good for large outputs (JSON/HTML).
  - Review diffs carefully to avoid blind updates.

- **Leverage Go fuzzing**
  - Ideal for parsers, validation, normalization logic.

---

## Unit Test Quality Checklist

- [ ] One clear failure reason
- [ ] Deterministic and repeatable
- [ ] No external dependencies
- [ ] Asserts real behavior
- [ ] Readable test name
- [ ] Passes with `-race` and repeated runs

---

## Idiomatic Go Table-Driven Example

```go
func TestThing_Do(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {name: "happy path", input: "a", want: "A"},
        {name: "empty input", input: "", wantErr: true},
    }

    for _, tc := range tests {
        tc := tc
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            got, err := Do(tc.input)
            if (err != nil) != tc.wantErr {
                t.Fatalf("err=%v wantErr=%v", err, tc.wantErr)
            }

            if !tc.wantErr && got != tc.want {
                t.Fatalf("got=%q want=%q", got, tc.want)
            }
        })
    }
}

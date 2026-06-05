# Error Handling Conventions

This document describes the error-handling conventions used in `go-passbolt-cli`.
It codifies the patterns already in use so that new code stays consistent.

## Wrapping

Always wrap a lower-level error with `%w` (never `%s` or `%v`) so the original
error stays inspectable with `errors.Is` / `errors.As` up the call stack:

```go
if err := client.DeleteResource(ctx, id); err != nil {
    return fmt.Errorf("deleting Resource: %w", err)
}
```

Use `%v` (or `%q`) only for non-error context values — IDs, flag names, raw
input. The combined `%w` + `%v` form is the idiomatic way to wrap a sentinel
while adding context:

```go
return fmt.Errorf("%w (valid: %s)", util.ErrNoColumns, strings.Join(valid, ", "))
```

## Message format

- Start with a lowercase first word and use **no trailing punctuation**
  (enforced by staticcheck `ST1005`).
- Domain nouns and initialisms keep their casing (e.g. `Resource`, `Folder`,
  `MFA`, `TOTP`, `UUID`, `URL`). This PascalCase-for-nouns style is the existing
  house convention — match it; do not lowercase those words.
- Lead with the operation being performed (`"deleting Resource: %w"`,
  `"parsing MFA Challenge"`), and include the resource ID/name when it helps the
  reader locate the problem.

## Sentinel errors

Recurring input-validation failures are defined as sentinel errors in
[`util/errors.go`](../util/errors.go) (e.g. `util.ErrNoID`, `util.ErrNoColumns`).
Return the sentinel directly, or wrap it with `%w` to add guidance. This lets
callers and tests match on the sentinel with `errors.Is` instead of comparing
message strings:

```go
if id == "" {
    return util.ErrNoID
}
```

## Surfacing to the user

Commands return their error from the cobra `RunE` handler; cobra prints it to
stderr and exits non-zero. Do not call `os.Exit` or `log.Fatal` from command
logic — return the error and let the root command handle it.

## Testing

Assert on behaviour with `errors.Is` / `errors.As` rather than on exact message
strings, which are brittle:

```go
if !errors.Is(err, util.ErrNoColumns) {
    t.Errorf("expected ErrNoColumns, got %v", err)
}
```

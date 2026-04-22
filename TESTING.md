# Testing

Project testing conventions:

- Always use external test packages, for example `package skills_test`.
- Use only the Go standard library `testing` package.
- Test exported behavior, not internal implementation details.
- Do not write tests for logging.
- Keep test files aligned with source files: if behavior belongs to `x.go`, put it in `x_test.go`.

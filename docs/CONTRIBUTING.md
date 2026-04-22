# Contribute to the documentation

The docs should stay close to the code. When the runtime changes, update the relevant guide or reference page in the same change set whenever possible.

## How to contribute

### Option 1: Edit directly on GitHub

1. Open the page you want to change.
2. Edit the corresponding `.mdx` or `.md` file.
3. Open a pull request with a short explanation of what behavior changed.

### Option 2: Local development

1. Clone the repository.
2. Install the Mintlify CLI with `npm i -g mint`.
3. Run `mint dev` from `docs/`.
4. Preview the site locally.
5. If you changed runtime behavior, run `go test ./...` from the repo root too.

## Writing guidelines

- Write for developers, not end users.
- Prefer practical examples over abstract descriptions.
- Document hardcoded defaults and limits explicitly.
- Keep environment variables, commands, and exported identifiers in backticks.
- Describe current behavior only.

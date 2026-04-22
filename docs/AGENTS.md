# hrns docs instructions

## About this project

- This is a Mintlify documentation site for the `hrns` Go project.
- The audience is developers who want to experiment with tool-using agents and embed `hrns` into their own codebases.
- Pages are MDX files with YAML frontmatter.
- Navigation lives in `docs.json`.

## Terminology

- Use `hrns` for the project name.
- Use "agent loop" for `loop.RunLoop`.
- Use "OpenAI-compatible endpoint" for provider configuration.
- Use "skill" specifically for `SKILL.md` files discovered by the `skills` package.
- Do not call skills "plugins".

## Style preferences

- Use active voice and second person ("you")
- Keep the tone direct and technical
- Prefer tutorial-style explanations before exhaustive reference detail
- State hardcoded defaults explicitly
- Use code formatting for commands, file names, environment variables, and exported identifiers

## Content boundaries

- Document current behavior only.
- Do not present `TODO.md` items as implemented.
- Do not invent configuration flags, commands, or environment variables.
- When behavior has a sharp edge or limitation, document it plainly.

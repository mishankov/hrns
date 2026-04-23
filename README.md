# hrns

`hrns` is a small Go harness for experimenting with tool-using AI agents and embedding them in your own code.

It currently gives you:

- an OpenAI-compatible chat completion client with streaming support
- a minimal agent loop that can execute tool calls and continue the conversation
- a small interactive TUI for testing prompts and tools
- skill discovery from `SKILL.md` files
- a handful of built-in tools for files, shell commands, and HTTP fetches

## Install

From a local checkout:

```bash
go install .
```

Directly from GitHub:

```bash
go install github.com/mishankov/hrns@latest
```

## Run

Start the app:

```bash
hrns
```

Or, from a checkout:

```bash
go run .
```

On first launch, the bundled TUI creates `~/.config/hrns/config.json` through an interactive onboarding flow. It asks for:

- provider name
- provider API URL
- provider API key
- default model
- whether to skip TLS verification

After onboarding, later runs reuse the saved config and print the active provider and model at startup.

Built-in commands:

- `/models` lists models exposed by the current provider's `/models` endpoint
- `/model <model>` updates the current provider's saved default model
- `/new` starts a fresh conversation
- `/help` shows the command list
- `/providers` lists saved providers
- `/provider <name>` switches the active provider, rebuilds the client, and saves it as current
- `/connect` adds another provider to the saved config

Note: `/connect` persists a new provider configuration and marks it as `currentProvider` in the config file, but it does not rebuild the active in-memory client. Use `/provider <name>` to switch immediately, or restart the app to pick up the saved current provider on startup.

## Configuration

The bundled binary stores provider settings in:

```text
~/.config/hrns/config.json
```

The config file contains:

- named providers with `url`, `key`, `model`, and `skipVerify`
- `currentProvider` to choose the default provider on startup

The bundled app also loads skills from:

- `~/.agents/skills`
- `./.agents/skills`

## Built-in tools

The default app registers these tools:

- `read_file`
- `list_files`
- `write_file`
- `run_command`
- `web_fetch`
- `load_skill`

## Package layout

- [`openai`](./openai): OpenAI-compatible client, request and response types, stream accumulator
- [`loop`](./loop): agent loop, chunk model, tool interface
- [`tools`](./tools): bundled tool implementations
- [`skills`](./skills): skill discovery, metadata loading, `load_skill`
- [`tui`](./tui): interactive terminal UI

## Tests

Run:

```bash
go test ./...
```

Testing conventions live in [TESTING.md](./TESTING.md).

## Taskfile

If you use [`task`](https://taskfile.dev), the repo includes a small [Taskfile.yml](./Taskfile.yml) with a few common commands:

- `task run`
- `task install`
- `task test`
- `task fmt`
- `task docs:dev`
- `task docs:check`

## Documentation

Project documentation lives in [`docs/`](./docs).

The docs cover:

- quickstart and provider setup
- the TUI workflow
- embedding `hrns` into your Go code
- adding tools and skills
- package-level references for the current implementation

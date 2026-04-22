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

Set provider configuration:

```bash
export HRNS_KEY="your-api-key"
export HRNS_BASE_URL="https://your-provider.example/v1"
```

Then start the app:

```bash
hrns
```

Or, from a checkout:

```bash
go run .
```

The bundled TUI starts with the hardcoded model `kimi-k2.6`. If your provider does not support that model, switch immediately:

```text
/model <your-model>
```

Other built-in commands:

- `/new`
- `/help`

## Configuration

The current binary reads these environment variables:

- `HRNS_KEY`: API key sent as `Authorization: Bearer ...`
- `HRNS_BASE_URL`: OpenAI-compatible base URL. Defaults to `https://api.openai.com/v1`
- `HRNS_SKIP_VERIFY`: set to `true` to disable TLS certificate verification

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

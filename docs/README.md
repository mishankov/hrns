# hrns documentation

This directory contains the Mintlify documentation site for `hrns`.

## Goals

- explain how to run the bundled TUI
- show developers how to embed `hrns` packages into their own Go code
- document the current behavior of the runtime and exported packages
- keep tutorial-style guides ahead of abstract reference pages

## Local preview

Install the Mintlify CLI and run the preview from this directory:

```bash
npm i -g mint
mint dev
```

## Structure

- `docs.json`: site configuration and navigation
- `index.mdx`: landing page
- `quickstart.mdx`: fastest way to run hrns
- `guides/`: practical workflows
- `internals/`: architecture-level explanations
- `reference/`: package reference pages
- `development.mdx`: contribution and maintenance guidance

## Writing rule

Describe current code, not intended future behavior.

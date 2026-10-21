*[日本語](README_ja.md) | **English***

# crube documentation

crube attaches descriptions to the files and directories of a project, keeps them in `.crube/`, and lets you reach the place you want quickly with `ls` / `tree` / `find` that show those descriptions. AI can write and update the descriptions, but reading them needs no AI.

```
$ cd src/auth
$ crube tree
  src/             Application source code
>   auth/          Authentication (login, sessions, token refresh)
      session.ts   Session storage and expiry
      *_test.go    Tests of the file of the same name
    main.go        Entry point
  data/            Data files
    *.csv          CSV data
```

## Contents

### design/ — Design

Why crube is shaped the way it is.

- [Overview and principles](design/overview.md) — purpose, scope, principles, terms
- [Architecture](design/architecture.md) — the core and the CLI / MCP around it, language and libraries
- [Data model](design/data-model.md) — the tree format, patterns, languages, symbolic links, length limits
- [Freshness and coverage](design/freshness.md) — what status reports, and why coverage is left to the user
- [Decisions](design/decisions.md) — what was decided, and why

### reference/ — Reference

The exact specification.

- [CLI](reference/cli.md) — commands, options, the `CRUBE_LANG` environment variable, exit codes
- [MCP](reference/mcp.md) — the tools and their arguments
- [File format](reference/file-format.md) — the grammar of `structure.txt`, files per language, `ignore`

### examples/ — Examples

- [Getting started](examples/getting-started.md) — from init to reading the descriptions
- [Updating descriptions with an AI agent](examples/ai-workflow.md) — updating with status --json, CI
- [A sample .crube/](examples/sample-crube/) — `structure.txt`, `structure_ja.txt`, `ignore`

## At a glance

| Item | Value |
| --- | --- |
| Language | Go (a single binary) |
| Data format | A text format of its own, written as a tree (no library needed) |
| Writes to | Only inside `.crube/` |
| Dependencies | Only `golang.org/x/text`, `golang.org/x/term` and the official MCP Go SDK (no Git needed) |
| License | MIT |

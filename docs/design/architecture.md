*[日本語](architecture_ja.md) | **English***

# Architecture

## The whole

All the logic is in a core library, and the CLI and the MCP server are thin wrappers of it. The core calls neither AI nor Git.

```
    User (a person)            AI agent
       │                    │        │
       ▼                    ▼        ▼
     ┌─────┐              ┌──────────────┐
     │ CLI │◀─────────────│ (via the CLI)│
     └──┬──┘              │  MCP server  │
        │                 └──────┬───────┘
        ▼                        ▼
  ┌──────────────────────────────────────────────────────────────┐
  │ Core library (does not call AI)                              │
  │ Store │ Language │ Matching │ Ignore │ Diff engine │ Search  │
  └───────┬──────────────────────────────┬───────────────────────┘
          │ reads and writes             │ only reads
          ▼                              ▼
     .crube/                       The project's files
   (structure, ignore)             (only walked, never changed)
```

## Modules of the core

| Module | Role |
| --- | --- |
| Store | Reads a structure file line by line and builds the tree. Writing outputs the whole file again in the canonical form |
| Language | Picks, from `CRUBE_LANG`, the one structure file to read and edit |
| Matching | Finds the entry that matches a real path (an exact match, or else the first `path.Match` from the top) |
| Ignore | Evaluates `.crube/ignore` |
| Diff engine | Compares the real directories with the structure, and finds untracked paths, orphans and coverage |
| Search | Searches paths, descs and details. It keeps no index: it reads the structure for every search |

## Implementation

- **Language**: Go, distributed as a single binary.
- **Structure files**: a text format of its own, read and written with the standard library only (`strings`, `path`).
- **License**: MIT. The license texts of the libraries it uses are shipped with it.

## What it writes (an invariant)

crube writes only inside `.crube/`. No command and no tool is an exception.

- The files and directories of the project are only walked and read. They are never created, changed, deleted or moved.
- `crube mv` / `crube rm` only move or remove descriptions (entries); they do not touch the real files.
- `crube init` creates only `.crube/` and what is in it.
- Right before writing, the core checks that the place written to is inside `.crube/`. One that leads outside through a symbolic link is refused too.

## Finding .crube/

A project has only one `.crube/`. Every command looks for `.crube/` from the current directory (or the directory given with `-C`) upwards, and uses the first one it finds. That directory is the project root.

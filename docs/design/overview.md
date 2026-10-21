*[日本語](overview_ja.md) | **English***

# Overview and principles

## What crube is

crube attaches descriptions to the files and directories of a project, keeps them in `.crube/`, and lets you reach the place you want quickly with `ls` / `tree` / `find` that show those descriptions.

AI can write and update the descriptions, but reading them needs no AI.

## The problem

When people or AI agents work on a project, they explore the directory structure and read source code again and again to find the file they need. The bigger and more complex the project, the more this search costs.

crube writes down "what is where" once and keeps it, so that anyone can look it up afterwards (without AI).

## Scope

| In scope | Out of scope |
| --- | --- |
| Descriptions of the project's structure (structure) | An order in which to read things for a purpose |
| The CLI commands to read and maintain them | Calling AI from the tool itself |
| An MCP server that offers the same | Taking in information from other tools |

How to work and what to know are left to the agent's own rules (`.claude/` and the like).

## Principles

- **Writing and reading are separate**: AI is used only to write and update descriptions. Reading (ls / tree / find / show / status) works on the saved data alone.
- **Paths are the anchor**: descriptions are tied to paths, so comparing them with the real directories is the freshness check.
- **Text people can read**: the data is text that writes the directory structure as it is, with indentation. People can read and write it directly, and diffs are easy to take.
- **Independent**: it depends on neither Git, nor a particular AI provider, nor any other tool, and works on its own.
- **Writes only inside `.crube/`**: crube only reads the files of the project, and never creates, changes or deletes anything outside `.crube/`.
- **Simplicity first**: a project has only one `.crube/`. crube does not guess or fix things by itself; it reports the facts.
- **A familiar interface**: the commands follow existing names (ls / tree / find), so there is little to learn.
- **The CLI and MCP are thin wrappers of the same core**: no difference in features or behavior between them.
- **A narrow role**: crube deals only with descriptions of "what is where". It does not take on many roles.
- **Decisions are left to the user**: the tool has no policy on "how many descriptions are enough" or "how long a description should be". It reports the facts, and the user decides what passes.

## Terms

| Term | Meaning |
| --- | --- |
| structure | All the descriptions of the project's structure. It is made of one file per language |
| entry | One line of a structure file: a name (of a file or directory, or a pattern) with its description (desc and detail) |
| project root | The directory where `.crube/` is |
| pattern | The name of an entry, such as `*.csv`, that matches several names directly in the same directory |
| target language | The language given by the `CRUBE_LANG` environment variable. Reading and editing use the file of this language |
| default language | English, written in `structure.txt`. It is used when there is no file for the target language |
| untracked path | A path that exists but has no description |
| orphan entry | An entry with a description whose path does not exist |
| coverage | How far the descriptions go (all match / dirs match) |

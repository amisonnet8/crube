*[日本語](ai-workflow_ja.md) | **English***

# Updating descriptions with an AI agent

crube itself does not call AI. An AI agent writes the descriptions through the CLI or MCP. The steps are the same whatever the AI.

## Steps

1. Get the places that need a description with `crube status --json`.
2. The AI reads those files and whatever sources it has (design documents, work records and so on).
3. The AI updates the structure with `crube set` / `mv` / `rm` (or the MCP writing tools).
4. Run `crube status` again to make sure no problem is left.
5. A person reviews the diff of `.crube/`.

## Example output of status --json

```json
{
  "structure": ".crube/structure.txt",
  "items": [
    { "kind": "untracked", "type": "dir",  "path": "src/db/" },
    { "kind": "untracked", "type": "file", "path": "src/db/conn.go" },
    { "kind": "orphan",    "path": "src/legacy/" }
  ],
  "coverage": {
    "all":  { "ok": false, "missing": 2 },
    "dirs": { "ok": false, "missing": 1 }
  }
}
```

For each field and the order, see the [CLI reference](../reference/cli.md#crube-status).

## Example instructions for agents

Instructions like these in the agent's rules (`.claude/` and the like) make it easy to run.

```markdown
## When looking for files
- First find the place with `crube find <keywords>` or `crube tree`.

## When finishing work
- Run `crube status --json`.
- If there are untracked / orphan items, read the files and
  update the descriptions with `crube set` / `crube mv` / `crube rm`.
- If you moved or renamed files, move their descriptions with `crube mv <old path> <new path>`.
- Write the desc in one line: what is there. Put the details in --detail.
- If there are files for several languages, switch `CRUBE_LANG` and do the same for each language.
```

## When to run it

- Run steps 1 to 4 when an AI agent finishes its work, or in a pre-commit hook.
- In CI, catch missing descriptions by the exit code of `crube status --require dirs` or the like. Which level to require is up to the project.

```yaml
# Example: CI steps
- uses: actions/checkout@v7
- uses: actions/setup-go@v7
  with:
    go-version: stable
- name: Install crube
  run: go install github.com/amisonnet8/crube/cmd/crube@latest
- name: Check crube descriptions
  run: crube status --require dirs
```

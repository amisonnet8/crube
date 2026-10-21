*[日本語](mcp_ja.md) | **English***

# MCP reference

`crube mcp` starts an MCP server on stdio. It calls the same core as the CLI, and returns results as JSON.

## Registering with a client

```json
{
  "mcpServers": {
    "crube": {
      "command": "crube",
      "args": ["mcp", "--allow-write"],
      "env": { "CRUBE_LANG": "ja" }
    }
  }
}
```

The server looks for `.crube/` from the directory where it started, upwards. The language is the `CRUBE_LANG` environment variable when the server starts (it can be set with `env` in the client's settings). If it is not in the form of a language tag, the server does not start and stops with exit code 2. The writing tools, too, write only inside `.crube/`. `remove_description` / `move_description` only remove or move descriptions; they do not touch files.

## Tools

The writing tools are offered only when the server is started with `crube mcp --allow-write`. To allow only reading, start it without `--allow-write`.

| Kind | Tool | Arguments | Returns |
| --- | --- | --- | --- |
| Read | `tree` | path, depth | The nested structure of the CLI's `crube tree --json`. No mark of the current place (`here` is `""` at the top, and `false` on each entry) |
| Read | `ls` | path | The same as the CLI's `crube ls --json` (what is directly in it, with descriptions) |
| Read | `find` | query, limit | The same as the CLI's `crube find --json` (the paths found with descriptions, and how many were left out) |
| Read | `show` | path | The same as the CLI's `crube show` (desc, detail, the path of the entry that matched, and the target of a symbolic link) |
| Maintain | `status` | kinds (a filter) | The same as `crube status --json` |
| Write | `set_description` | path, desc, detail | The entry after the update, `{ "path", "desc", "detail" }` |
| Write | `remove_description` | path | `{ "path", "removed" }` (`removed` is the number of entries removed, those below included) |
| Write | `move_description` | from, to | `{ "from", "to", "moved" }` (`moved` is the number of entries moved, those below included) |

- The arguments `path`, `depth`, `limit`, `kinds` and `detail` can be left out. `depth` and `limit` left out or 0 mean no limit; a negative number is an error. The `detail` of `set_description` stays as it is when left out, and `""` removes it.
- The `path` of `ls` and `show` is **relative to the project root** (whatever the place the server started in). MCP has no current place, and the paths it returns are relative to the root too, so a path in a result can be passed to the next call as it is.
- The `kinds` of `status` is an array of `untracked`, `orphan` and `invalid`. Only `items` are filtered; `coverage` is always that of the whole project. Other values are an error.
- The tools read `.crube/` again on every call. If `.crube/` is changed while the server runs, the next call sees it.

## Names of the tools

The reading and maintenance tools have the same names and give the same as the CLI's commands.

Only the writing tools have names other than the CLI's `set` / `rm` / `mv` (`set_description` / `remove_description` / `move_description`). Agents may judge what a tool is for from its name in the list, and `rm` or `mv` could be taken for tools that delete or move the files themselves. The names tell that they only work on descriptions.

## Notes on the design

- The descriptions of the tools say "before searching for files, use tree / find first", so that agents read these first.
- The output of tree is about as big as the structure file. In a big project, narrow it with depth or path.
- The writing tools check the same things as the CLI's set (length limits, no newlines).
- Writing, as in the CLI, outputs the whole file again in the canonical form.
- Errors are returned as tool errors. The text is the CLI's error message (what follows `crube: `).
- MCP is implemented with the official Go SDK (`github.com/modelcontextprotocol/go-sdk`).

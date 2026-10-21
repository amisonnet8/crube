*[日本語](cli_ja.md) | **English***

# CLI reference

Every command looks for `.crube/` from the current directory (or the directory given with `-C`) upwards, and works on it. Only `init` is different: it makes `.crube/` in the given directory.

## Commands

| Command | Kind | What it does |
| --- | --- | --- |
| [`crube init`](#crube-init) | Setup | Makes `.crube/` |
| [`crube ls`](#crube-ls) | Read | Lists the real files, with descriptions |
| [`crube tree`](#crube-tree) | Read | Shows the structure, marking the current place |
| [`crube find`](#crube-find) | Read | Searches paths and descriptions |
| [`crube show`](#crube-show) | Read | Everything about one path |
| [`crube status`](#crube-status) | Maintain | The difference between the real directories and the structure |
| [`crube set`](#crube-set) | Edit | Adds or changes a description |
| [`crube rm`](#crube-rm) | Edit | Removes an entry (with those below it) |
| [`crube mv`](#crube-mv) | Edit | Moves an entry (with those below it) |
| [`crube mcp`](#crube-mcp) | Integration | Runs as an MCP server |
| [`crube completion`](#crube-completion) | Integration | Prints a shell completion script |

## Common options

| Option | What it does |
| --- | --- |
| `--json` | Structured output (JSON) |
| `-C <dir>` | Works as if run in the given directory |
| `--no-ignore` | Turns the ignore settings off (`.crube/` stays left out) |

- Options can go anywhere before or after the positional arguments (`crube ls src --full`). An option with a value takes `--limit 3` or `--limit=3`.
- Everything after `--` is read as positional arguments (`crube set -- note.txt "-v flag help"`).
- `init` takes only `-C`. `set`, `rm` and `mv` take `--json` and `-C` (and their own options). `mcp` takes only `--allow-write`. `completion` takes no options.

## Usage and version

```
$ crube --help
usage: crube <command> [arguments]

commands:
  init        create .crube/
  ls          list a directory with descriptions
  tree        show the descriptions as a tree
  find        search paths and descriptions
  show        show the description of one path
  status      report untracked, orphan and invalid entries
  set         add or change a description
  rm          remove a description
  mv          move a description
  mcp         run as an MCP server over stdio
  completion  print a shell completion script

common options: --json  -C <dir>  --no-ignore

$ crube ls -h
usage: crube ls [path] [--full] [--json] [-C <dir>] [--no-ignore]

$ crube --version
crube v0.1.0
```

- `--help`, `-h` and `help` print to standard output, with exit code 0. `crube` with no arguments prints the same to standard error, with exit code 2.
- `-h` and `--help` of a command print its usage in one line.
- `--version` prints the version built into the release, or else the version Go recorded in the binary (`v0.1.0` after `go install …@v0.1.0`). A build with neither prints `crube (devel)`.

## Language

There is no settings file. The language is given only by the `CRUBE_LANG` environment variable (BCP 47: `ja`, `zh-CN` and so on).

Every command reads and edits the file of the `CRUBE_LANG` language (`.crube/structure_<language tag>.txt`). When `CRUBE_LANG` is unset or `en`, or there is no file for that language, `.crube/structure.txt` (English) is used. Only one file is ever used.

The value of `CRUBE_LANG` is used as it is (`ja-JP` is not brought to `ja`, and case is not changed). A value not in the form of a language tag (1 to 8 letters or digits, joined with `-`) is an error for every command (`crube: invalid CRUBE_LANG "ja_JP" (use a language tag such as ja or zh-CN)`).

## Entry paths

The path given to `tree` / `set` / `rm` / `mv` is the position of an entry in the structure file: the names from the root, joined with `/`. The last name can be a pattern. Wherever it is run, it is the position from the root.

```
data/*.csv        # the entry "*.csv" under data/
src/auth/         # a directory entry (ending in /)
```

- An entry path finds an entry by the exact name at each level. It is not matched as a pattern (`crube rm "src/*.go"` removes only the entry "`*.go`" under `src/`).
- A path with a name that a structure file cannot hold ([file-format.md](file-format.md#names-that-cannot-be-written), and `.`, `..`), or with an empty level (`a//b`), is an error.

The path given to `ls` / `show` is a real path. It is read relative to the current directory (with `-C`, that directory), and one that points outside the project is an error (`crube: ../: outside the project`).

On Windows, `\` in an entry path is read as `/` too (`src\auth\` is the same as `src/auth/`), since Windows names cannot have `\`. On other OSes, `\` is a character of a name. Entry paths in results and errors are separated by `/` on every OS.

## Exit codes

| Code | Meaning |
| --- | --- |
| 0 | Success |
| 1 | status found a problem (orphans, invalid values, or the coverage given with `--require` is not met) |
| 2 | An error (wrong arguments, no `.crube/` found, a syntax error in the structure file, a value rejected when writing, and so on) |

## Output conventions

- What crube itself prints (marks, headings, messages) is English only. `CRUBE_LANG` only picks the structure file.
- An error is one line `crube: <what>` on standard error, with exit code 2. The same with `--json`. A load error of the structure file comes with the file and the line number.

```
crube: .crube/structure.txt:12: indentation must be tabs
crube: no .crube/ found (run crube init)
crube: desc is 300 columns wide (limit 256)
crube: src/foo.go: no such file
crube: build/x: ignored by .crube/ignore (use --no-ignore)
crube: .crube: .crube/ is always left out, even with --no-ignore
```

Wrong arguments are one line with the command's name, `crube: <command>: <what>` (exit code 2). A wrong number of arguments comes with the usage of the command.

```
crube: unknown command "lss" (see crube --help)
crube: ls: unknown option --foo
crube: ls: -C needs a value
crube: find: invalid --limit "x" (use a number of 1 or more)
crube: ls: too many arguments (usage: crube ls [path] [--full] [--json] [-C <dir>] [--no-ignore])
crube: set: missing <desc> (usage: crube set <path> <desc> [--detail <text>] [--json] [-C <dir>])
```

### JSON

- Paths are separated by `/` and relative to the project root. Directories end in `/`.
- The `desc` and `detail` keys are always there, `""` when empty. `detail` is one string, with newlines as `\n`.
- `type` is `dir` or `file`. A symbolic link is `file`.
- Indented by 2 spaces. `<`, `>` and `&` are not escaped (descriptions are text for people).

---

## crube init

```
crube init [-C <dir>]
```

Makes `.crube/` and the structure file of the target language. It makes only `.crube/` and what is in it.

- If no `.crube/` is found upwards, it makes `.crube/` in the given directory, with the default `ignore` and an empty `structure.txt`.
- If `CRUBE_LANG` is not English, it also makes an empty file for that language, `structure_<language tag>.txt`.
- If `.crube/` is already there, it makes only the file of the target language, if that is missing. This is how to add a language file.
- If there is nothing to make, it is an error (exit code 2).

```
$ CRUBE_LANG=ja crube init
created .crube/structure_ja.txt
```

## crube ls

```
crube ls [path] [--full] [--json]
```

Lists what is directly in `path` (the current directory if left out), with descriptions.

```
$ crube ls src
auth/     Authentication (login, sessions, token refresh)
db/       ! no description
main.go   Entry point
```

- A path without a description shows `! no description`.
- Sorted by name (directories do not come first).
- The descs start at the same column: 3 spaces after the longest name (in display width).
- A desc is cut with `…` to fit the terminal width. `--full` shows it whole.
- Ignored paths are not listed. An ignored or missing path given as `path` is an error. So is a file (`crube: go.mod: not a directory (use crube show)`).
- Inside a directory that no entry matches, nothing has a description (matching goes down from the parent).
- With `--json`, `entry` is the path of the entry that matched (`""` for none).

```json
{
  "path": "src/",
  "items": [
    { "name": "auth/",   "type": "dir",  "entry": "src/auth/",   "desc": "Authentication (login, sessions, token refresh)" },
    { "name": "db/",     "type": "dir",  "entry": "",            "desc": "" },
    { "name": "main.go", "type": "file", "entry": "src/main.go", "desc": "Entry point" }
  ]
}
```

## crube tree

```
crube tree [path] [-d N] [--json]
```

Shows what is in the structure file, nested as in the file and easy to read, and marks the current place.

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

- **What it shows**: the entries of the structure file (patterns as they are). It does not walk the real file system. Use `ls` for the real files, and `status` for missing or extra descriptions.
- **Indentation**: 2 spaces per level (unlike the tabs of the file, so that the width is the same on any terminal).
- **Aligned descs**: the descs start at the same column across the whole view: 3 spaces after the name of the widest line (the mark column and the indentation included). The width of names is counted in display width (full width = 2).
- **detail**: not shown. Use `show` for the detail.
- **Lines without a desc**: only the name is shown.
- **Mark of the current place**: `>` in the column at the start of the line, on the entry that matches the current directory. Without such an entry, it goes on the line of the deepest ancestor that matches, and the current directory is added at the end of the line, as in `(here: src/db/)` (2 spaces after the desc; at the desc column if there is no desc). At the project root there is no mark. The mark column always takes 2 columns, mark or not.
- **Long descs**: when the output is a terminal, a desc is cut with `…` to fit its width. Only the desc is cut; `(here: …)` stays. Output to a pipe or a file is not cut.

```
$ cd src/db          # there is no entry for src/db/
$ crube tree src/ -d 2
> src/        Application source code  (here: src/db/)
    auth/     Authentication (login, sessions, token refresh)
    main.go   Entry point
```

- With `path` (an entry path), only the subtree of that entry is shown. Without such an entry, it is an error (`crube: no entry src/x/`). `-d` limits the number of levels shown (`-d 1` is the top level only; no limit if left out).
- With `--json`, the entries are output as nested JSON. The `here` at the top is the current directory (`""` at the project root), and the `here` of each entry tells whether it is the line with the mark. The example below is the output of `crube tree src/ -d 2 --json` run in `src/auth`.

```json
{
  "here": "src/auth/",
  "entries": [
    { "name": "src/", "path": "src/", "desc": "Application source code", "detail": "", "here": false,
      "children": [
        { "name": "auth/", "path": "src/auth/", "desc": "Authentication (login, sessions, token refresh)",
          "detail": "The entry point is index.ts.", "here": true, "children": [] },
        { "name": "main.go", "path": "src/main.go", "desc": "Entry point",
          "detail": "", "here": false, "children": [] }
      ] }
  ]
}
```

## crube find

```
crube find <query> [--limit N] [--json]
```

Searches the entries of the structure file by path, desc and detail. Results are shown as entry paths (a pattern entry as `data/*.csv`).

```
$ crube find session
src/auth/             Authentication (login, sessions, token refresh)
src/auth/session.ts   Session storage and expiry
```

- **Matching**: case-insensitive substring match. Several words separated by spaces match only what contains all of them (AND). Each word can be anywhere in the path, desc or detail.
- **Several arguments**: joined with spaces into one `query` (`crube find auth session` is the same as `crube find "auth session"`).
- **Order**: that of the structure file.
- **View**: two columns, path and desc. The descs start 3 spaces after the longest path. When the output is a terminal, descs are cut with `…`, as with `ls` and `tree`. With no match, nothing is printed (exit code 0).
- Ignored entries and those below them are not in the results. A query of only spaces is an error (`crube: empty query`).
- `--limit N` limits the number of results (no limit if left out). If some were left out, the count comes last, as in `… 3 more (use --limit)`.
- With `--json`, `more` is the number left out.

```json
{
  "query": "session",
  "more": 0,
  "items": [
    { "path": "src/auth/",           "desc": "Authentication (login, sessions, token refresh)", "detail": "The entry point is index.ts." },
    { "path": "src/auth/session.ts", "desc": "Session storage and expiry",
      "detail": "Sessions are kept in Redis; each access extends the expiry.\nThe expiry is set by session.ttl in config/auth.yaml." }
  ]
}
```

## crube show

```
crube show <path> [--json]
```

Shows, for one real path, the desc and detail of the entry that matches it. When a pattern matched, it shows the path of that entry (such as `data/*.csv`); for a symbolic link, also its target.

```
$ crube show data/2026.csv
data/2026.csv
  entry   data/*.csv
  desc    CSV data

$ crube show src/auth/session.ts
src/auth/session.ts
  desc    Session storage and expiry
  detail  Sessions are kept in Redis; each access extends the expiry.
          The expiry is set by session.ttl in config/auth.yaml.

$ crube show current
current
  link    releases/v2
  desc    ! no description
```

- The first line is the real path. Then the fields, one per line, with their labels.
- `entry` is shown only when the path of the entry that matched differs from the real path (when a pattern matched).
- `link` is shown only for a symbolic link.
- `detail` is not shown when there is none. Its second and later lines are indented to line up.
- Without a description, the desc is `! no description` (exit code 0).
- A missing path or an ignored path is an error. So is the project root itself (no entry matches the root).
- With `--json`, every key is always there (`entry` and `link` are `""` when there is none).

```json
{ "path": "data/2026.csv", "type": "file", "entry": "data/*.csv", "desc": "CSV data", "detail": "", "link": "" }
```

## crube status

```
crube status [--require all|dirs] [--json]
```

Compares the real directories with the structure, reports untracked paths, orphans and invalid values, and shows the coverage last. It only reports; it changes nothing. Wherever it runs, it checks the whole project.

| Shown | Meaning |
| --- | --- |
| `untracked` | Untracked (with the type, `dir` / `file`) |
| `orphan` | Orphan |
| `invalid` | A desc / detail over the limit |

```
$ crube status
untracked   dir   src/db/
untracked   file  src/db/conn.go
orphan      src/legacy/
invalid     src/x.go  desc 300/256

coverage    all: no (2)   dirs: no (1)
```

- **Layout**: the kind column is 12 columns wide, and the `dir` / `file` column of untracked is 6. An invalid line is "path, 2 spaces, `desc 300/256`", not aligned with the other lines.
- **The coverage line**: `yes` when met, `no (count)` when not. With nothing to report, only the coverage line is printed (no empty line either).

```
$ crube status
coverage    all: yes   dirs: yes
```

- **Order** (text and JSON): untracked dirs → untracked files → orphans → invalid.
  - Untracked paths are in the order of the walk. In each directory, files first, then subdirectories, both by name.
  - Orphans and invalid values are in the order of the structure file.
- An invalid line shows which field is how many columns wide (and the limit), as in `desc 300/256`.
- With `--json`, `structure` is the file checked.

```json
{
  "structure": ".crube/structure.txt",
  "items": [
    { "kind": "untracked", "type": "dir",  "path": "src/db/" },
    { "kind": "untracked", "type": "file", "path": "src/db/conn.go" },
    { "kind": "orphan",    "path": "src/legacy/" },
    { "kind": "invalid",   "path": "src/x.go", "field": "desc", "width": 300, "limit": 256 }
  ],
  "coverage": {
    "all":  { "ok": false, "missing": 2 },
    "dirs": { "ok": false, "missing": 1 }
  }
}
```

| Option | What it does |
| --- | --- |
| `--require all\|dirs` | Exit code 1 if the given coverage is not met. Without it, the coverage does not change the exit code |
| `--json` | Outputs the path and kind of each item as JSON |

For the details of what is reported, see [design/freshness.md](../design/freshness.md).

## crube set

```
crube set <path> <desc> [--detail <text>] [--json]
```

Adds or changes the description of an entry.

```
$ crube set data/ "Data files"
$ crube set "data/*.csv" "CSV data"
```

- It writes to the file used for the target language.
- A new entry goes after its siblings. Missing parents on the way are made as lines without a desc.
- Spaces around the desc are removed. A desc of `""` makes a line without a desc.
- Newlines in `--detail` become the lines of the detail as they are (CRLF and CR become LF, and spaces at the end of each line are removed). Without `--detail`, the detail stays as it is; `--detail ""` removes it.
- A desc over 256 columns or with a newline, or a detail over 4096 columns, is an error.
- It writes to a temporary file in `.crube/`, which then replaces the file. Readers never see a half-written file. Several crubes writing the same file at the same time are not kept from each other (the same for `rm` and `mv`).
- On success it prints nothing (the same for `rm` and `mv`). Only with `--json` does it print what it wrote, in the same shape as the MCP writing tools return.

```json
{ "path": "data/*.csv", "desc": "CSV data", "detail": "" }
```

## crube rm

```
crube rm <path> [--json]
```

Removes an entry from the file used for the target language. For a directory entry, every entry below it is removed too. The real files are not removed. Without such an entry, it is an error (`crube: no entry src/x.go`).

## crube mv

```
crube mv <from> <to> [--json]
```

Moves an entry to another position in the file used for the target language. For a directory entry, the entries below it move with it. The real files are not moved. Use it after moving files, to move their descriptions to the new paths.

- Missing parents of the new position are made as lines without a desc.
- Without an entry at `from`, it is an error. An entry already at `to` is an error too (also when `from` and `to` are the same).
- Moving between a directory entry (ending in `/`) and a file entry, or into itself, is an error.
- The entry moved goes after its new siblings.

## crube mcp

```
crube mcp [--allow-write]
```

Starts an MCP server on stdio. Only with `--allow-write` are the writing tools offered. See [mcp.md](mcp.md) for more.

## crube completion

```
crube completion <shell>
```

Prints a shell completion script to standard output. `<shell>` is one of `bash`, `zsh`, `fish` and `powershell`. Any other word is an error (`crube: completion: unknown shell "tcsh" (use bash, zsh, fish or powershell)`). Put the script where the shell looks for completions.

```
crube completion bash > ~/.local/share/bash-completion/completions/crube
crube completion zsh > "${fpath[1]}/_crube"
crube completion fish > ~/.config/fish/completions/crube.fish
crube completion powershell >> $PROFILE
```

The script is a short, fixed text with no list of commands. On every TAB it asks `crube __complete` for candidates. So the candidates never disagree with the crube installed, and the entry paths of the project in that place can be offered too.

What is completed:

| Where | Candidates |
| --- | --- |
| The first word | Commands (with a summary) |
| A word starting with `-` | The options the command takes that are not on the line yet, and `--help` (for the first word, `--help` and `--version`) |
| The value of `--require` | `all`, `dirs` |
| The path of `tree`, `set` and `rm`, and `from` / `to` of `mv` | Entry paths |
| After `completion` | The 4 shells |

- Entry paths are offered one level at a time: the entries directly under the entry up to the last `/` of the word being typed, as paths from the root, with their descs (`crube tree src/<TAB>` gives `src/auth/` and `src/main.go`). A directory entry ends in `/`, with no space after it. Ignored entries are not offered (as with `tree`).
- For the path of `set`, real paths without a description (directly in the same directory, with no entry of exactly that name) are offered too, besides the entries: as paths from the root, following the ignore settings. This is for adding new descriptions.
- A pattern entry (`data/*.csv`) is inserted in quotes so that the shell does not expand it (zsh and fish escape it instead). The quotes and backslashes of the word being typed are removed before looking for candidates.
- Some shells do not complete a value after `=` (`--require=`). Write `--require <TAB>`.
- The value of `-C` and the path of `ls` and `show` (real paths) are left to the shell's file completion. The values of `-d`, `--limit` and `--detail`, descs, and the words of `find` are not completed.
- `.crube/` is looked for from the directory of `-C <dir>` if it is on the line, or else from the current directory. Descs come from the file of the `CRUBE_LANG` language. `--no-ignore` on the line is followed.
- When there is no `.crube/`, the structure file cannot be read, and so on, entry paths are simply not offered, and nothing is said (a warning on every TAB would break the line being typed).

### __complete

```
crube __complete [--word=<word being typed>] -- <words>...
```

The command the scripts call. It is not for typing, so `crube --help` does not show it.

- The words after `--` are the line typed so far (without `crube` itself), as the shell split it. `--word` is the word being typed, empty when the cursor is at the start of a new word. It is not among the words because shells may drop an empty argument.
- The output is one candidate per line: the value, and if there is a description, a tab and the description (the summary of a command, or the desc of an entry cut at 60 columns of display width). Only those starting with `--word` are printed.
- Whatever the words, it ends with exit code 0 and prints nothing to standard error.

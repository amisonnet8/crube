*[日本語](decisions_ja.md) | **English***

# Decisions

What was decided while designing crube, and why.

| Item | Decision | Why |
| --- | --- | --- |
| Name | crube | — |
| Language | Go | A single binary is easy to distribute |
| License | MIT | — |
| Format of the structure | A text format of its own that writes the directory structure indented with tabs. One line = name + tab + desc; a line starting with `\|` is the detail | What you see is the tree, easy to read and write. No library needed |
| File names | `structure.txt`, and `structure_<language tag>.txt` per language | They open as text |
| Patterns | The name of an entry can be a `path.Match` pattern. It matches only names directly in the same directory. An exact match wins, then the order written | Matching needs only the standard library, and the rule of priority is simple |
| How it writes | It outputs the whole file again in the canonical form. The order is kept; alignment, empty lines and comments are not | Writing stays simple |
| Where it writes | Only inside `.crube/`. It never creates, changes or deletes the project's files | So that it is safe to bring in and to run |
| Git | Not depended on. Neither Git's information nor `.gitignore` is used | A simple specification, and the same result in every environment |
| Number of `.crube/` | One per project, not nested | No rules on what each one covers or on inheriting settings |
| Fields | Only desc and detail (both optional) | Entry points, other names and related paths can be written as text in the detail |
| Settings file | None. The language is the `CRUBE_LANG` environment variable; everything else is a command option | No machinery to manage settings |
| Languages | Just one independent file per language. The file of the `CRUBE_LANG` language is read and edited; without it, the English `structure.txt`. What matches what between languages is not managed | No tracking of translations, no fallback rules |
| Detecting moves | Not done. They are reported as orphans and untracked paths; the user moves the description with `mv` | The tool does not guess |
| Search index | None. Every search reads the structure | No cache to manage |
| Coverage | No default. status reports whether all / dirs match, and the user picks how far to require | The tool has no policy |
| Symbolic links | Not followed. One is a single entry like a file, and the description is attached to the link itself | No loops, duplicates or walking outside |
| Length of desc | Only limits (desc 256 columns with no newline, detail 4096 columns). What is recommended is up to the user | Keeps the one-line view. A recommendation is a policy, so the tool has none |
| Role | Only descriptions of "what is where". No order of reading for a purpose, no calling AI, no taking in from other tools | Not taking on many roles. Steps and knowledge are left to the agent's rules (`.claude/` and the like) |
| MCP | Provided, as a thin wrapper of the same core as the CLI. The writing tools are offered only when started with `crube mcp --allow-write` | AI without a shell can use it too, and gets structured data |
| Names of MCP tools | Reading and maintenance tools have the CLI's names (`show` and so on). Only the writing tools have other names, such as `set_description` | No drift in behavior. The writing tools get names that cannot be taken for file operations |
| Libraries | `golang.org/x/text/width` for display width, `golang.org/x/term` for the terminal width, the official Go SDK for MCP. Structure files and ignore are read with the standard library only | No tables or protocols of its own to keep up to date. Only official or near-official libraries |
| Ambiguous width | Counted as 1 column (○, ① and the like). Combining and format characters (U+0301, U+200D and the like) are 0 columns | The default of most terminals |
| Language of messages | crube itself prints only English. `CRUBE_LANG` only picks the structure file | No machinery for translations |
| Matching `CRUBE_LANG` | Only whether `structure_<the value of CRUBE_LANG>.txt` exists. `ja-JP` is not brought to `ja` | No fallback rules |
| Matching of find | A case-insensitive substring match. Words separated by spaces are ANDed. Results are in the order of the structure file | People and AI can look things up without minding how they are written |
| Format of ignore | All the rules of gitignore (`**`, `/` at the start or in the middle, `!`), implemented with the standard library only | A format people already know how to write |
| Shape of JSON | Paths are relative to the root, separated by `/`. The desc and detail keys are always there, `""` when empty | The same output on every OS, and users need not mind whether a key is there |
| How releases are built | goreleaser builds binaries for 3 OSes and attaches them to a draft Release | Building, checksums and attaching are taken care of |
| Arguments of the CLI | Options can go anywhere before or after the positional arguments, and everything after `--` is positional. Parsed with the standard library only. The arguments of find are one query | So that `crube ls src --full` works as the docs show, and a desc starting with `-` can be given |
| Aligning text | ls, tree and find align the descs 3 spaces after the widest name. Only on a terminal is a desc cut with `…`; tree's `(here: …)` is kept | The same look for the 3 commands, and nothing lost in output sent to a pipe |
| Result of writing | set returns the entry written; rm and mv return the number of entries, those below included. The same shape for the CLI's `--json` and MCP | So that the result can be checked, and the CLI and MCP do not drift |
| `\` on Windows | `\` in an entry path is read as `/`, on Windows only. Output is `/` on every OS | So that a Windows user typing `\` does not silently make an entry that matches nothing (`\` cannot be in a Windows name) |
| Paths of MCP | The paths of `ls` and `show` are relative to the project root. tree marks no current place | MCP has no current place, and a path it returned can be passed to the next call as it is |
| Paths for editing | The entry paths of `tree`, `set`, `rm` and `mv` look up each level by the exact name. They are not matched as patterns | So that `crube rm "src/*.go"` means that pattern entry, not the descriptions of the files it matches |
| Reach of orphans | Below an orphan, and below an ignored directory, nothing is reported as an orphan. An entry matching an ignored path or an empty directory is not an orphan | Moving one directory is reported in one line, and ignoring something does not make orphans |
| Leaving out `.crube/` | `.crube/` at the root is always left out; neither `!` nor `--no-ignore` brings it back | So that the structure files do not show as untracked |
| Default ignore | `init` writes only `.git/`, `.hg/`, `.svn/` and `node_modules/` | Things such as `dist/` and `vendor/`, which mean something in some projects, are not hidden on their own |
| Writing at the same time | No locking. Writing goes to a temporary file in `.crube/`, which then replaces the file | Readers never see a half-written file, and there are no locks to manage |
| Language of the docs | README and docs are pairs of English (`X.md`) and Japanese (`X_ja.md`). Neither is the master: one change fixes both. Examples are written with descriptions in each language, and tests check both against the real output | Readers of English and of Japanese both see the examples in their own language, and neither side goes stale |
| Shell completion | `crube completion <shell>` prints a fixed script for bash, zsh, fish or PowerShell, and the script asks the hidden command `crube __complete` for candidates on every TAB. Entry paths are offered one level at a time, and set also offers real paths without a description | The script holds no list of commands, so it never disagrees with the crube installed. Entry paths are positions from the root, which the shell's file completion from the current place cannot offer |

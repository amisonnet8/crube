*[日本語](file-format_ja.md) | **English***

# File format reference

## What is in .crube/

```
.crube/
├── structure.txt        # descriptions (English, the default)
├── structure_ja.txt     # descriptions (Japanese, optional). One file per language
└── ignore               # what to leave out
```

- There is only one `.crube/`, at the project root. They are not nested.
- Everything in `.crube/` is data that can be shared. Nothing such as a cache that has to be made again is kept there.
- The file of a language other than English is named `structure_<language tag>.txt`. The language tag is BCP 47 (`ja`, `zh-CN` and so on).
- `.crube/` itself is always left out of the walk.
- The encoding is UTF-8. Newlines are LF (CRLF can be read, but writing uses LF).

## structure.txt

Text written as a tree: the structure of the project indented with tabs. In the example below, the indentation and the separator between name and desc are all tabs.

```
src/	Application source code
	auth/	Authentication (login, sessions, token refresh)
		| The entry point is index.ts.
		| Settings are in config/auth.yaml.
		index.ts	Entry point of the module
		session.ts	Session storage and expiry
		*_test.go	Tests of the file of the same name
	my  file.txt	A name with spaces is written as it is
data/	Data files
	*.csv	CSV data
tmp/
	cache/	Temporary cache
```

### Grammar

```
line         = indentation body
indentation  = tab × depth
body         = detail-line | entry-line
detail-line  = "|" | "| " text
entry-line   = name [ tab desc ]
```

- **Indentation**: one tab per level. Spaces cannot be used to indent.
- **Entry line**: after the indentation, the name runs up to the first tab, and the desc is what follows (without spaces around it). Without a tab, the whole line is the name, with no desc. Spaces in a name are all part of the name.
- **Spaces at the end of a line**: removed when read.
- **Name**: ending in `/`, a directory; otherwise a file (symbolic links included). A name cannot have `/` in the middle.
- **Line without a desc**: an entry with no description, for structure only. It is used, for example, as the parent of a description deep down.
- **Detail line**: a line that is `|` alone, or starts with `| `. It is one line of the detail of the entry above (a line of `|` alone is an empty line). It is indented as deep as the children of that entry.
- **Empty lines**: ignored when read.

### Names that cannot be written

This grammar cannot express the names below. None of them is common.

- A name with a tab in it
- A name starting with a space (it cannot be told from indentation with spaces), or ending with one (spaces at the end of a line are removed)
- A name that is `|` alone, or starts with `| `

### Matching

Which entry matches a real path is decided, directory by directory, in this order:

1. If the same directory has an entry whose name is exactly the name, that one.
2. Otherwise, the first of the entries of the same directory, from the top, that matches with Go's `path.Match`.

- Only the names directly in that directory are matched.
- An entry ending in `/` matches only directories; the others only files. The trailing `/` is dropped when comparing.
- Below a directory matched by a pattern, matching goes on with the children of that entry.
- A name that is not a valid `path.Match` pattern (such as `[`) matches nothing as a pattern. It still matches as an exact name.

### Errors

In these cases the whole file fails to load (reported with the line number).

- The indentation has a space in it (including when an editor turned tabs into spaces)
- The indentation is two or more levels deeper than the parent (two or more tabs too many)
- A file entry (not ending in `/`) has children
- One directory has two entries of the same name
- A detail line with no entry before it
- A name has `/` in the middle

### Length limits

Lengths are counted in display width (full width = 2, half width = 1).

| Field | Limit |
| --- | --- |
| desc | 256 columns |
| detail | 4096 columns, all lines together |

A value over the limit is an error for set and for writing through MCP. In a file edited by hand, status reports it.

### How crube writes (the canonical form)

When crube writes, it outputs the whole file again in this form:

- The order of the entries is kept. A new entry goes after its siblings.
- Indentation is tabs. The name and the desc are separated by one tab. Nothing is aligned.
- A detail line is `| ` + text (an empty one is `|`).
- No empty lines, and no spaces at the end of lines.

Alignment and empty lines made by hand are gone once crube writes the file.

## structure_<language tag>.txt

The file of a language other than English. The format is exactly that of `structure.txt`.

- The file of each language is independent. It has no tie to `structure.txt`, and the entries need not be the same.
- Reading and editing use only one file: that of the `CRUBE_LANG` language. If it is not there (not a regular file), `structure.txt` is used.
- With `CRUBE_LANG` unset or `en`, `structure.txt` is used even if `structure_en.txt` exists.
- `crube init` can make an empty one (`CRUBE_LANG=ja crube init`).

## ignore

One pattern of paths to leave out per line. The format is that of gitignore. Only this file is read; `.gitignore` is not.

- A line starting with `#` is a comment, and empty lines are ignored. A line that is not a valid pattern is skipped, as git does.
- Spaces at the end of a line are removed (a space written as `\ ` stays).
- A trailing `/` matches only directories.
- A line starting with `!` takes back what an earlier line left out (later lines win). But nothing inside a directory left out can be taken back.
- A pattern with `/` at the start or in the middle is relative to the project root. Otherwise it matches a name at any depth.
- `*`, `?` and `[…]` do not cross `/`. `**` matches zero or more directories (`**/x`, `a/**`, `a/**/b`).
- `\` makes the next character a plain character (`\#`, `\!`).

```
# .crube/ignore
node_modules/
dist/
*.log
!dist/README.md
```

- `.crube/` directly in the project root is always left out; neither `!` nor `--no-ignore` brings it back. A `.crube` in a subdirectory is treated like any other name.
- Without `ignore`, nothing but `.crube/` is left out.
- `crube init` writes a default ignore with what almost every project wants to leave out (`.git/`, `.hg/`, `.svn/`, `node_modules/`). Things such as `dist/` and `vendor/`, which mean something in some projects, are not in it. Edit it as you like.
- A path left out is out of scope for ls, tree, find and status alike.
- Leaving a path out only hides it from views and checks; it does not remove descriptions from the structure.

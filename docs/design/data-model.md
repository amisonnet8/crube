*[日本語](data-model_ja.md) | **English***

# Data model

The exact definition of the file format is in [reference/file-format.md](../reference/file-format.md). This page explains why it has that shape.

## Text written as a tree

The structure is a text format of its own that writes the directory structure of the project as it is, indented with tabs. The depth is only the number of tabs at the start of a line, so there is no indentation width to count.

```
src/	Application source code
	auth/	Authentication (login, sessions, token refresh)
		| The entry point is index.ts. Settings are in config/auth.yaml.
		session.ts	Session storage and expiry
		*_test.go	Tests of the file of the same name
data/	Data files
	*.csv	CSV data
```

- One line is one entry: a name and a desc, separated by a tab.
- Tabs are used only to indent and to separate; every space is part of the name or the desc. So names with spaces are written as they are.
- Children are indented (one tab).
- A line starting with `|` is the detail (more about it) of the entry above.

Why:

- **What you see is the tree**: `crube tree` shows the same nesting as this file. The form you write is the form you read, so neither people nor AI get lost.
- **The reach of a pattern is clear**: a pattern matches only names directly in the directory where it is written (see below).
- **Moving is easy**: to move a directory, move its block as a whole. The patterns in it are relative, so they need no rewriting.
- **No dependencies**: a small line-by-line parser reads it. No library such as YAML is needed.

## Names and patterns

The name of an entry is the name of a real file or directory, or a pattern. They are not told apart. Matching goes in this order:

1. If the same directory has an entry whose name is exactly the name, that one is used.
2. Otherwise, the entries of the same directory are matched as patterns with Go's standard `path.Match`, from the top, and the first that matches is used.

- A pattern matches **only names directly in** the directory where it is written. `*.csv` under `data/` does not match `data/sub/x.csv`.
- An entry ending in `/` matches only directories; the others match only files (and symbolic links). `*/` matches every directory directly in it.
- Because of rule 1, a name with symbols of `path.Match`, such as `[id].tsx`, matches the real name when written as it is.
- Patterns take priority in the order written. Write the more specific ones above.

Matching needs only the standard library, and there is no syntax, such as `**`, that crosses levels.

## Only desc and detail

An entry has only a one-line description (desc) and an optional detail. Information such as the file to start from, other names, or related paths is written in the detail as text when needed.

A line without a desc can be written too. It means "no description": a line for structure only, put there as the parent of an entry deep down.

## Writing outputs the canonical form

When crube writes a structure file, it outputs the whole file again in the canonical form.

- The order of the entries is kept (the priority of patterns depends on it).
- Indentation is tabs, and the name and the desc are separated by one tab. Nothing is aligned.
- Empty lines and spaces at the end of lines are removed.

The file puts reading and writing by machine first, and is not aligned. The aligned view is the job of `crube tree`.

## Languages: one independent file per language

There is just one structure file per language. Nothing ties the files to each other.

| File | Language |
| --- | --- |
| `structure.txt` | English (the default) |
| `structure_<language tag>.txt` | That language (for example `structure_ja.txt`) |

- The **target language** is given by the `CRUBE_LANG` environment variable. Unset, or `en`, it is English.
- If there is a file for the target language, reading and editing use only that file.
- If there is no file for the target language (none has been made), reading and editing use `structure.txt`.
- The switch is per file. Languages never mix in one view.
- Each file is complete on its own.
- A language file is made with `crube init` (`CRUBE_LANG=ja crube init` makes `structure_ja.txt`).
- What matches what between languages is not managed. `mv` and `rm` work only on the file of the target language. What another language's file lacks or has too much shows as orphans and untracked paths when `status` runs in that language.

## Symbolic links

Symbolic links are not followed: one is a single entry, just like a regular file.

- No loops, no duplicates of the same thing, and no walking outside the project.
- The result is the same in every environment.
- Even a link to a directory gets no `/` at the end, since crube does not go into it.
- The description is attached to the link itself, not to its target. `show` shows the target for reference.

## Only one .crube/

There is only one `.crube/`, at the project root; they are not nested. The top level of a structure file is what is directly in the project root.

So there is no need for rules on the boundaries of what each one covers, on inheriting settings, or on choosing where to write, and the specification and the implementation stay simple.

## Length limits

Lengths are counted in display width (full width = 2, half width = 1). Counted in characters, the amount of information and the width on screen would differ a lot between Japanese and English.

crube has only limits; it sets no recommended length.

| Field | Limit |
| --- | --- |
| desc | 256 columns (about 128 Japanese characters) |
| detail | 4096 columns, all lines together |

The desc is shown on the same line as the name by ls / tree, so its limit is short. A long description goes in the detail (shown by `show`).

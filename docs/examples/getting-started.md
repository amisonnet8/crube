*[日本語](getting-started_ja.md) | **English***

# Getting started

How to bring crube into a small Go project, describe it, and look things up.

## 1. Initialize

```
$ cd myapp
$ crube init
created .crube/structure.txt
created .crube/ignore
```

crube creates only what is in `.crube/`. No command after this changes anything outside `.crube/`.

## 2. See what needs a description

```
$ crube status
untracked   dir   cmd/
untracked   dir   cmd/myapp/
untracked   dir   internal/
untracked   dir   internal/store/
untracked   file  go.mod
untracked   file  cmd/myapp/main.go
untracked   file  internal/store/store.go

coverage    all: no (7)   dirs: no (4)
```

## 3. Describe

```
$ crube set cmd/ "Executables"
$ crube set cmd/myapp/ "CLI entry point of myapp"
$ crube set internal/ "Private packages"
$ crube set internal/store/ "Persistence layer (SQLite)" --detail "Also called: db, database"
```

At this point `.crube/structure.txt` looks like this (the indentation, and the separator between name and description, are tabs).

```
cmd/	Executables
	myapp/	CLI entry point of myapp
internal/	Private packages
	store/	Persistence layer (SQLite)
		| Also called: db, database
```

You may edit this file directly. For example, where there are many files of the same kind, one line with a pattern is enough.

```
internal/	Private packages
	store/	Persistence layer (SQLite)
		*_test.go	Unit tests for the file of the same name
```

## 4. Look things up

`tree` shows the structure, nested in the same way, and puts the mark `>` on the line of the directory you are in. To see the real files with descriptions, use `ls`.

```
$ cd internal/store
$ crube tree
  cmd/        Executables
    myapp/    CLI entry point of myapp
  internal/   Private packages
>   store/    Persistence layer (SQLite)
$ cd ../..

$ crube find sqlite
internal/store/   Persistence layer (SQLite)
```

## (Optional) Make a Japanese file

Each language can have a file of its own. Set `CRUBE_LANG` and run `init`, and an empty file for that language is made.

```
$ export CRUBE_LANG=ja
$ crube init
created .crube/structure_ja.txt

$ crube set internal/store/ "永続化レイヤー(SQLite)"
$ crube tree
  internal/
    store/    永続化レイヤー(SQLite)
```

While `CRUBE_LANG=ja`, reading and editing use only `structure_ja.txt`. Descriptions of the English file never mix in. If `structure_ja.txt` has not been made, the English `structure.txt` is used even with `CRUBE_LANG=ja`.

The rest goes on without `CRUBE_LANG` (in English).

```
$ unset CRUBE_LANG
```

## 5. Check the coverage

```
$ crube status
untracked   file  go.mod
untracked   file  cmd/myapp/main.go
untracked   file  internal/store/store.go

coverage    all: no (3)   dirs: yes

$ crube status --require dirs; echo $?
...
0
```

Every directory now has a description, so dirs match is met. How far to require is up to your project.

## 6. Share

`.crube/` holds only text files, so if you keep them under version control or ship them with the project, other people and AI agents can use the same descriptions.

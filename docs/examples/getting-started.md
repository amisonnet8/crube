# はじめての crube

小さな Go プロジェクトに crube を導入し、説明を付けて引くまでの流れです。

## 1. 初期化

```
$ cd myapp
$ crube init
created .crube/structure.txt
created .crube/ignore
```

crube が作るのは `.crube/` の中身だけです。以後のどのコマンドも、`.crube/` の外は書き換えません。

## 2. 説明が必要な箇所を確認する

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

## 3. 説明を付ける

```
$ crube set cmd/ "Executables"
$ crube set cmd/myapp/ "CLI entry point of myapp"
$ crube set internal/ "Private packages"
$ crube set internal/store/ "Persistence layer (SQLite)" --detail "Also called: db, database"
```

この時点の `.crube/structure.txt` はこうなっています(インデントと、名前と説明の区切りはタブ)。

```
cmd/	Executables
	myapp/	CLI entry point of myapp
internal/	Private packages
	store/	Persistence layer (SQLite)
		| Also called: db, database
```

このファイルは直接編集してもかまいません。たとえば、同じ種類のファイルがたくさんある場所には、パターンで 1 行だけ書けます。

```
internal/	Private packages
	store/	Persistence layer (SQLite)
		*_test.go	Unit tests for the file of the same name
```

## 4. 参照する

`tree` は structure の内容を同じ入れ子の形で表示し、今いるディレクトリの行に `>` の印を付けます。実在するファイルを説明付きで見たいときは `ls` を使います。

```
$ cd internal/store
$ crube tree
  cmd/            Executables
    myapp/        CLI entry point of myapp
  internal/       Private packages
>   store/        Persistence layer (SQLite)
$ cd ../..

$ crube find sqlite
internal/store/   Persistence layer (SQLite)
```

## (任意)日本語のファイルを作る

言語ごとに独立したファイルを持てます。`CRUBE_LANG` を設定して `init` を実行すると、その言語の空ファイルができます。

```
$ export CRUBE_LANG=ja
$ crube init
created .crube/structure_ja.txt

$ crube set internal/store/ "永続化レイヤー(SQLite)"
$ crube tree
  internal/
    store/      永続化レイヤー(SQLite)
```

`CRUBE_LANG=ja` の間は、参照も編集も `structure_ja.txt` だけを使います。英語のファイルの説明が混ざることはありません。`structure_ja.txt` を作っていなければ、`CRUBE_LANG=ja` でも英語の `structure.txt` が使われます。

以下は `CRUBE_LANG` を外した(英語の)状態で続けます。

```
$ unset CRUBE_LANG
```

## 5. 充足度を確認する

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

ディレクトリには説明が揃ったので、dirs 一致を満たしています。どこまで求めるかはプロジェクトで決めてください。

## 6. 共有する

`.crube/` の中身はテキストファイルだけなので、プロジェクトと一緒にバージョン管理や配布をすれば、ほかの人や AI エージェントも同じ説明を使えます。

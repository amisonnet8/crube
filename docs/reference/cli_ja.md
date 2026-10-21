*[English](cli.md) | **日本語***

# CLI リファレンス

どのコマンドも、カレントディレクトリ（または `-C` で指定したディレクトリ）から上位へ `.crube/` を探して動作します。`init` だけは例外で、指定ディレクトリに `.crube/` を作ります。

## コマンド一覧

| コマンド | 種別 | 内容 |
| --- | --- | --- |
| [`crube init`](#crube-init) | 準備 | `.crube/` を作る |
| [`crube ls`](#crube-ls) | 参照 | 実在するファイルの一覧を、説明付きで表示 |
| [`crube tree`](#crube-tree) | 参照 | structure の内容を表示し、現在地に印を付ける |
| [`crube find`](#crube-find) | 参照 | パス・説明で検索 |
| [`crube show`](#crube-show) | 参照 | 1 パスの全情報 |
| [`crube status`](#crube-status) | 保守 | 実ディレクトリと structure の差分 |
| [`crube set`](#crube-set) | 編集 | 説明の追加・更新 |
| [`crube rm`](#crube-rm) | 編集 | エントリの削除（配下も含む） |
| [`crube mv`](#crube-mv) | 編集 | エントリの移動（配下も含む） |
| [`crube mcp`](#crube-mcp) | 連携 | MCP サーバーとして起動 |
| [`crube completion`](#crube-completion) | 連携 | シェルの補完スクリプトを出力 |

## 共通オプション

| オプション | 内容 |
| --- | --- |
| `--json` | 構造化出力（JSON） |
| `-C <dir>` | 指定したディレクトリで実行したものとして動作する |
| `--no-ignore` | 除外設定を無効にする（`.crube/` は除外したまま） |

- オプションは、位置引数の前後どこに書いてもよい（`crube ls src --full`）。値をとるオプションは `--limit 3` でも `--limit=3` でもよい。
- `--` より後ろは、すべて位置引数として読む（`crube set -- note.txt "-v flag help"`）。
- `init` が受け付けるのは `-C` だけ。`set`・`rm`・`mv` は `--json` と `-C`（と各自のオプション）。`mcp` は `--allow-write` だけ。`completion` はオプションを受け付けない。

## 使い方と版

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

- `--help`・`-h`・`help` は標準出力に出して終了コード 0。引数なしの `crube` は、同じものを標準エラー出力に出して終了コード 2。
- 各コマンドの `-h`・`--help` は、その書式を 1 行で出す。
- `--version` は配布物に埋めた版を出す。なければ、Go がバイナリに記録した版を出す（`go install …@v0.1.0` なら `v0.1.0`）。どちらもなければ `crube (devel)`。

## 言語

設定ファイルはありません。言語は環境変数 `CRUBE_LANG`（BCP 47。`ja`、`zh-CN` など）だけで指定します。

どのコマンドも、環境変数 `CRUBE_LANG` の言語のファイル（`.crube/structure_<言語タグ>.txt`）を参照・編集します。`CRUBE_LANG` が未設定か `en` の場合、またはその言語のファイルがない場合は `.crube/structure.txt`（英語）を使います。使うファイルは常に 1 つです。

`CRUBE_LANG` の値はそのまま使います（`ja-JP` を `ja` に寄せたり、大文字・小文字をそろえたりしない）。言語タグの形（英数字 1〜8 字を `-` でつなぐ）でない値は、どのコマンドでもエラーです（`crube: invalid CRUBE_LANG "ja_JP" (use a language tag such as ja or zh-CN)`）。

## エントリのパス

`tree` / `set` / `rm` / `mv` に渡すパスは、structure ファイル上のエントリの位置を、ルートからの名前を `/` でつないで表したものです。最後の名前はパターンでもかまいません。どこで実行しても、ルートからの位置です。

```
data/*.csv        # data/ の下の「*.csv」というエントリ
src/auth/         # ディレクトリのエントリ(末尾 /)
```

- エントリのパスは、各段を名前の完全一致で探してエントリを指す。パターンとして照合はしない（`crube rm "src/*.go"` が消すのは、`src/` の下の「`*.go`」というエントリだけ）。
- structure ファイルに書けない名前（[file-format_ja.md](file-format_ja.md#書けない名前)、と `.`・`..`）や、空の段（`a//b`）を含むパスはエラー。

`ls` / `show` に渡すパスは、実在するパスです。カレントディレクトリ（`-C` を付けたらそのディレクトリ）からの相対で読み、プロジェクトの外を指すとエラー（`crube: ../: outside the project`）です。

Windows では、エントリのパスの `\` も `/` として読みます（`src\auth\` は `src/auth/` と同じ）。Windows の名前には `\` を使えないためです。そのほかの OS では `\` は名前の文字です。結果やエラーに出るエントリのパスは、どの OS でも `/` 区切りです。

## 終了コード

| コード | 意味 |
| --- | --- |
| 0 | 正常 |
| 1 | status で問題あり（孤立・上限超え、または `--require` で指定した充足度に未達） |
| 2 | エラー（引数の誤り、`.crube/` が見つからない、structure ファイルの文法エラー、書き込みの検証エラーなど） |

## 出力の共通の決まり

- crube 自身が出す文字（印・見出し・メッセージ）は英語だけです。`CRUBE_LANG` は structure ファイルの選択にだけ効きます。
- エラーは標準エラー出力に `crube: <内容>` の 1 行で出し、終了コード 2 で終わります。`--json` を付けたときも同じです。structure ファイルの読み込みエラーには、ファイルと行番号を付けます。

```
crube: .crube/structure.txt:12: indentation must be tabs
crube: no .crube/ found (run crube init)
crube: desc is 300 columns wide (limit 256)
crube: src/foo.go: no such file
crube: build/x: ignored by .crube/ignore (use --no-ignore)
crube: .crube: .crube/ is always left out, even with --no-ignore
```

引数の誤りは、コマンド名を付けて `crube: <コマンド>: <内容>` の 1 行で出します（終了コード 2）。引数の数の誤りには、そのコマンドの書式を添えます。

```
crube: unknown command "lss" (see crube --help)
crube: ls: unknown option --foo
crube: ls: -C needs a value
crube: find: invalid --limit "x" (use a number of 1 or more)
crube: ls: too many arguments (usage: crube ls [path] [--full] [--json] [-C <dir>] [--no-ignore])
crube: set: missing <desc> (usage: crube set <path> <desc> [--detail <text>] [--json] [-C <dir>])
```

### JSON

- パスは `/` 区切りで、プロジェクトルートからの相対。ディレクトリは末尾 `/`。
- `desc`・`detail` のキーは常に出し、ないときは `""`。`detail` は改行を `\n` で含む 1 つの文字列。
- `type` は `dir` か `file`。シンボリックリンクは `file`。
- 2 字下げで出す。`<`・`>`・`&` はエスケープしない（説明は人が読む文字なので）。

---

## crube init

```
crube init [-C <dir>]
```

`.crube/` と、対象言語の structure ファイルを作ります。作成するのは `.crube/` とその中身だけです。

- 上位へ探しても `.crube/` がなければ、指定ディレクトリに `.crube/` を作り、既定の `ignore` と空の `structure.txt` を置く。
- `CRUBE_LANG` が英語以外なら、その言語の空ファイル `structure_<言語タグ>.txt` も作る。
- すでに `.crube/` がある場合は、対象言語のファイルがなければそれだけを作る。言語ファイルを追加するときはこの使い方をする。
- 作るものが何もなければエラー（終了コード 2）。

```
$ CRUBE_LANG=ja crube init
created .crube/structure_ja.txt
```

## crube ls

```
crube ls [path] [--full] [--json]
```

`path`（省略時はカレントディレクトリ）の直下を、説明付きで一覧表示します。

```
$ crube ls src
auth/     認証(ログイン、セッション、トークン更新)
db/       ! no description
main.go   エントリポイント
```

- 説明のないパスは `! no description` と表示する。
- 名前の順に並べる（ディレクトリを先にしない）。
- desc の開始位置を、いちばん長い名前の後ろに空白 3 つで揃える（幅は表示幅）。
- desc は端末幅に合わせて `…` で切り詰める。`--full` で全文を表示する。
- ignore されたパスは一覧に出さない。ignore されたパスや実在しないパスを `path` に指定するとエラー。ファイルを指定してもエラー（`crube: go.mod: not a directory (use crube show)`）。
- エントリが当たらないディレクトリの中は、どれも説明なしになる（照合は親から順にたどるため）。
- `--json` では、`entry` に当たったエントリのパス（なければ `""`）を出す。

```json
{
  "path": "src/",
  "items": [
    { "name": "auth/",   "type": "dir",  "entry": "src/auth/",   "desc": "認証(ログイン、セッション、トークン更新)" },
    { "name": "db/",     "type": "dir",  "entry": "",            "desc": "" },
    { "name": "main.go", "type": "file", "entry": "src/main.go", "desc": "エントリポイント" }
  ]
}
```

## crube tree

```
crube tree [path] [-d N] [--json]
```

structure ファイルの内容を、ファイルと同じ入れ子の形で見やすく表示し、現在地に印を付けます。

```
$ cd src/auth
$ crube tree
  src/             アプリケーションのソースコード
>   auth/          認証(ログイン、セッション、トークン更新)
      session.ts   セッションの保存と有効期限の管理
      *_test.go    同名ファイルのテスト
    main.go        エントリポイント
  data/            データ格納場所
    *.csv          CSVデータ
```

- **表示するもの**: structure ファイルのエントリ（パターンもそのまま）。実際のファイルシステムは走査しない。実在するファイルの一覧は `ls`、説明の過不足は `status` で確認する。
- **インデント**: 1 段あたりスペース 2 つで表示する（ファイルのタブとは違い、端末で幅が一定になるように）。
- **desc の桁揃え**: desc の開始位置を、表示全体で揃える。位置は、いちばん広い行の名前（印の欄と字下げを含む）の後ろに空白 3 つ。名前の幅は表示幅（全角 = 2）で数える。
- **detail**: 表示しない。detail は `show` で見る。
- **desc なしの行**: 名前だけを表示する。
- **現在地の印**: 行頭の欄に `>` を付ける。付けるのは、現在のディレクトリに当たるエントリ。そのエントリがなければ、当たる中でいちばん深い祖先の行に付け、行末に `(here: src/db/)` のように現在のディレクトリを添える（desc の後ろに空白 2 つ。desc がなければ desc の桁に）。プロジェクトルートにいるときは付けない。印の欄は、印がなくても常に 2 桁分空けておく。
- **長い desc**: 出力先が端末のときは、端末の幅に合わせて `…` で切り詰める。切るのは desc だけで、`(here: …)` は残す。パイプやファイルへの出力では切り詰めない。

```
$ cd src/db          # src/db/ のエントリがない
$ crube tree src/ -d 2
> src/        アプリケーションのソースコード  (here: src/db/)
    auth/     認証(ログイン、セッション、トークン更新)
    main.go   エントリポイント
```

- `path`（エントリのパス）を指定すると、そのエントリの部分木だけを表示する。エントリがなければエラー（`crube: no entry src/x/`）。`-d` で表示する段の数を制限する（`-d 1` は最上段だけ。省略時は制限なし）。
- `--json` では、エントリを入れ子の JSON で出力する。最上位の `here` は現在のディレクトリ（プロジェクトルートなら `""`）、各エントリの `here` は印を付ける行かどうか。次の例は、`src/auth` で `crube tree src/ -d 2 --json` を実行したときの出力。

```json
{
  "here": "src/auth/",
  "entries": [
    { "name": "src/", "path": "src/", "desc": "アプリケーションのソースコード", "detail": "", "here": false,
      "children": [
        { "name": "auth/", "path": "src/auth/", "desc": "認証(ログイン、セッション、トークン更新)",
          "detail": "入口は index.ts。", "here": true, "children": [] },
        { "name": "main.go", "path": "src/main.go", "desc": "エントリポイント",
          "detail": "", "here": false, "children": [] }
      ] }
  ]
}
```

## crube find

```
crube find <query> [--limit N] [--json]
```

structure ファイルのエントリを、パス・desc・detail で検索します。結果はエントリのパスで表示します（パターンのエントリは `data/*.csv` のように表示）。

```
$ crube find セッション
src/auth/             認証(ログイン、セッション、トークン更新)
src/auth/session.ts   セッションの保存と有効期限の管理
```

- **照合**: 大文字・小文字を区別しない部分一致。空白で区切った複数の語は、すべてを含むもの（AND）だけに当たる。語はパス・desc・detail のどこに含まれていてもよい。
- **複数の引数**: 空白でつないで 1 つの `query` にする（`crube find auth session` は `crube find "auth session"` と同じ）。
- **順序**: structure ファイルの順。
- **表示**: パスと desc の 2 列。desc の開始位置を、いちばん長いパスの後ろに空白 3 つで揃える。出力先が端末のときは、`ls`・`tree` と同じく desc を `…` で切り詰める。当たりがなければ何も出さない（終了コード 0）。
- ignore されたエントリとその下は、結果に出さない。空白だけの検索語はエラー（`crube: empty query`）。
- `--limit N` で件数を絞る（省略時は制限なし）。省いたものがあれば、最後に `… 3 more (use --limit)` のように件数を出す。
- `--json` では、`more` に省いた件数を出す。

```json
{
  "query": "セッション",
  "more": 0,
  "items": [
    { "path": "src/auth/",           "desc": "認証(ログイン、セッション、トークン更新)", "detail": "入口は index.ts。" },
    { "path": "src/auth/session.ts", "desc": "セッションの保存と有効期限の管理",
      "detail": "セッションは Redis に保存し、アクセスのたびに有効期限を延長する。\n有効期限は config/auth.yaml の session.ttl で設定する。" }
  ]
}
```

## crube show

```
crube show <path> [--json]
```

実在する 1 パスについて、当たったエントリの desc と detail を表示します。パターンで当たった場合はそのエントリのパス（`data/*.csv` など）を、シンボリックリンクならリンク先も表示します。

```
$ crube show data/2026.csv
data/2026.csv
  entry   data/*.csv
  desc    CSVデータ

$ crube show src/auth/session.ts
src/auth/session.ts
  desc    セッションの保存と有効期限の管理
  detail  セッションは Redis に保存し、アクセスのたびに有効期限を延長する。
          有効期限は config/auth.yaml の session.ttl で設定する。

$ crube show current
current
  link    releases/v2
  desc    ! no description
```

- 1 行目は実在のパス。続けて、見出しつきで縦に並べる。
- `entry` は、当たったエントリのパスが実在のパスと違うとき（パターンで当たったとき）だけ出す。
- `link` は、シンボリックリンクのときだけ出す。
- `detail` は、ないときは出さない。2 行目以降は字下げして揃える。
- 説明がなければ、desc を `! no description` とする（終了コード 0）。
- 実在しないパス、ignore されたパスはエラー。プロジェクトルートそのものもエラー（ルートに当たるエントリはない）。
- `--json` では、すべてのキーを常に出す（`entry`・`link` はないとき `""`）。

```json
{ "path": "data/2026.csv", "type": "file", "entry": "data/*.csv", "desc": "CSVデータ", "detail": "", "link": "" }
```

## crube status

```
crube status [--require all|dirs] [--json]
```

実ディレクトリと structure を突き合わせ、未説明・孤立・上限超えを報告し、最後に充足度を表示します。報告するだけで、何も書き換えません。どこで実行しても、プロジェクト全体を検査します。

| 表示 | 意味 |
| --- | --- |
| `untracked` | 未説明（`dir` / `file` の種別付き） |
| `orphan` | 孤立 |
| `invalid` | desc / detail が上限を超えている |

```
$ crube status
untracked   dir   src/db/
untracked   file  src/db/conn.go
orphan      src/legacy/
invalid     src/x.go  desc 300/256

coverage    all: no (2)   dirs: no (1)
```

- **表の形**: 種類の欄は 12 桁、untracked の `dir`・`file` の欄は 6 桁。invalid の行は「パス、空白 2 つ、`desc 300/256`」で、行どうしは揃えない。
- **coverage の行**: 満たしていれば `yes`、未達なら `no (件数)`。報告するものがなければ、coverage の行だけを出す（空行も出さない）。

```
$ crube status
coverage    all: yes   dirs: yes
```

- **並び順**（テキストも JSON も）: untracked の dir → untracked の file → orphan → invalid。
  - untracked は走査の順。各ディレクトリで、ファイルを先に、次にサブディレクトリを、どちらも名前の順にたどる。
  - orphan と invalid は structure ファイルの順。
- invalid の行は、どのフィールドが何桁か（上限）を `desc 300/256` の形で出す。
- `--json` では、`structure` に検査したファイルを出す。

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

| オプション | 内容 |
| --- | --- |
| `--require all\|dirs` | 指定した充足度を満たさなければ終了コード 1 にする。省略時は充足度は終了コードに影響しない |
| `--json` | 各項目のパス・種類を JSON で出力する |

判定の詳細は [design/freshness_ja.md](../design/freshness_ja.md) を参照してください。

## crube set

```
crube set <path> <desc> [--detail <text>] [--json]
```

エントリの説明を追加・更新します。

```
$ crube set data/ "データ格納場所"
$ crube set "data/*.csv" "CSVデータ"
```

- 書き込み先は、対象言語で使っているファイル。
- エントリがなければ、兄弟の末尾に追加する。途中の親がなければ、desc なしの行として作る。
- desc の前後の空白は取り除く。desc を `""` にすると、desc なしの行になる。
- `--detail` の改行は、そのまま detail の複数行になる（CRLF・CR は LF にし、各行の行末の空白は取り除く）。`--detail` を省くと detail はそのまま、`--detail ""` で detail を消す。
- desc が 256 桁を超える、改行を含む、または detail が 4096 桁を超える場合はエラー。
- 書き込みは、`.crube/` の中の一時ファイルに書いてから置き換える。読む側が書きかけのファイルを見ることはない。複数の crube が同時に同じファイルへ書き込む場合の排他はしない（`rm`・`mv` も同じ）。
- 成功したときは何も出力しない（`rm`・`mv` も同じ）。`--json` を付けたときだけ、書いた結果を出す。形は MCP の書き込みツールが返すものと同じ。

```json
{ "path": "data/*.csv", "desc": "CSVデータ", "detail": "" }
```

## crube rm

```
crube rm <path> [--json]
```

対象言語で使っているファイルから、エントリを削除します。ディレクトリのエントリなら、その下のエントリもすべて削除します。実際のファイルは削除しません。エントリがなければエラー（`crube: no entry src/x.go`）です。

## crube mv

```
crube mv <from> <to> [--json]
```

対象言語で使っているファイルの中で、エントリを別の位置へ移します。ディレクトリのエントリなら、その下のエントリも一緒に移ります。実際のファイルは移動しません。ファイルを移動したあとで、説明を新しいパスへ付け替えるために使います。

- 移動先の親がなければ、desc なしの行として作る。
- 移動元にエントリがなければエラー。移動先にすでにエントリがあればエラー（`from` と `to` が同じときも）。
- ディレクトリのエントリ（末尾 `/`）とファイルのエントリの間の移動、自分の下への移動はエラー。
- 移動したエントリは、新しい兄弟の末尾に置く。

## crube mcp

```
crube mcp [--allow-write]
```

stdio の MCP サーバーとして起動します。`--allow-write` を付けたときだけ、書き込み系ツールを公開します。詳しくは [mcp_ja.md](mcp_ja.md) を参照してください。

## crube completion

```
crube completion <shell>
```

シェルの補完スクリプトを標準出力に出します。`<shell>` は `bash`・`zsh`・`fish`・`powershell` のどれかです。ほかの語はエラー（`crube: completion: unknown shell "tcsh" (use bash, zsh, fish or powershell)`）です。スクリプトは、シェルが補完を探す場所に置きます。

```
crube completion bash > ~/.local/share/bash-completion/completions/crube
crube completion zsh > "${fpath[1]}/_crube"
crube completion fish > ~/.config/fish/completions/crube.fish
crube completion powershell >> $PROFILE
```

スクリプトはコマンドの一覧を持たない、短い固定のテキストです。TAB を押すたびに `crube __complete` に候補を聞きます。そのため、出る候補が入っている crube と食い違うことはありません。その場所のプロジェクトのエントリのパスも出せます。

補完するもの:

| 場所 | 候補 |
| --- | --- |
| 1 語目 | コマンド（要約付き） |
| `-` で始まる語 | そのコマンドが受け付けるオプションのうち、行にまだないものと `--help`（1 語目なら `--help`・`--version`） |
| `--require` の値 | `all`・`dirs` |
| `tree` のパス、`set` のパス、`rm` のパス、`mv` の `from`・`to` | エントリのパス |
| `completion` の次 | 4 つのシェル |

- エントリのパスは 1 段ずつ出す。打ちかけの語の最後の `/` までのエントリの直下にあるエントリを、ルートからのパスで、desc 付きで出す（`crube tree src/<TAB>` で `src/auth/`・`src/main.go`）。ディレクトリのエントリは末尾 `/` で、後ろに空白を入れない。除外されたエントリは出さない（`tree` と同じ）。
- `set` のパスには、エントリに加えて、説明のない実在のパス（同じディレクトリの直下で、名前が完全に一致するエントリがないもの）も出す。ルートからのパスで、除外設定に従う。新しく説明を付けるときのため。
- パターンのエントリ（`data/*.csv`）は、シェルに展開させないよう、引用符を付けて入れる（zsh と fish はエスケープする）。打ちかけの語の引用符とバックスラッシュは、取り除いてから候補を探す。
- 等号の形（`--require=`）の後ろの値は、補完しないシェルがある。`--require <TAB>` と書く。
- `-C` の値と、`ls`・`show` のパス（実在のパス）は、シェルのファイル補完に任せる。`-d`・`--limit`・`--detail` の値、desc、`find` の語は補完しない。
- 行に `-C <dir>` があればそのディレクトリから、なければカレントディレクトリから `.crube/` を探す。desc は `CRUBE_LANG` の言語のファイルから出す。行の `--no-ignore` に従う。
- `.crube/` がない、structure ファイルが読めないなどのときは、エントリのパスを出さないだけで、何も言わない（TAB のたびに警告が出ると、打っている行が崩れるため）。

### __complete

```
crube __complete [--word=<打ちかけの語>] -- <語>...
```

スクリプトが呼ぶコマンドです。打つためのものではないので、`crube --help` には出しません。

- `--` の後ろの語は、それまでに打った行（`crube` 自身は含まない）を、シェルが区切ったまま渡す。`--word` は今打っている語で、カーソルが新しい語の先頭にあるときは空。語の並びに入れないのは、シェルが空の引数を落とすことがあるため。
- 出力は 1 行に 1 件。値と、説明があれば、タブと説明（コマンドなら要約、エントリなら desc を表示幅 60 桁で切ったもの）。`--word` で始まるものだけを出す。
- どんな語でも、終了コード 0 で終わり、標準エラー出力に何も出さない。

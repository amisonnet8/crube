# CLI リファレンス

どのコマンドも、カレントディレクトリ(または `-C` で指定したディレクトリ)から上位へ `.crube/` を探して動作します。`init` だけは例外で、指定ディレクトリに `.crube/` を作ります。

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
| [`crube rm`](#crube-rm) | 編集 | エントリの削除(配下も含む) |
| [`crube mv`](#crube-mv) | 編集 | エントリの移動(配下も含む) |
| [`crube mcp`](#crube-mcp) | 連携 | MCP サーバーとして起動 |

## 共通オプション

| オプション | 内容 |
| --- | --- |
| `--json` | 構造化出力(JSON) |
| `-C <dir>` | 指定したディレクトリで実行したものとして動作する |
| `--no-ignore` | 除外設定を無効にする |

## 言語

設定ファイルはありません。言語は環境変数 `CRUBE_LANG`(BCP 47。`ja`、`zh-CN` など)だけで指定します。

どのコマンドも、環境変数 `CRUBE_LANG` の言語のファイル(`.crube/structure_<言語タグ>.txt`)を参照・編集します。`CRUBE_LANG` が未設定か `en` の場合、またはその言語のファイルがない場合は `.crube/structure.txt`(英語)を使います。使うファイルは常に 1 つです。

## エントリのパス

`tree` / `set` / `rm` / `mv` に渡すパスは、structure ファイル上のエントリの位置を、ルートからの名前を `/` でつないで表したものです。最後の名前はパターンでもかまいません。

```
data/*.csv        # data/ の下の「*.csv」というエントリ
src/auth/         # ディレクトリのエントリ(末尾 /)
```

`ls` / `show` に渡すパスは、実在するパスです。

## 終了コード

| コード | 意味 |
| --- | --- |
| 0 | 正常 |
| 1 | status で問題あり(孤立・上限超え、または `--require` で指定した充足度に未達) |
| 2 | エラー(引数の誤り、`.crube/` が見つからない、structure ファイルの文法エラー、書き込みの検証エラーなど) |

---

## crube init

```
crube init [-C <dir>]
```

`.crube/` と、対象言語の structure ファイルを作ります。作成するのは `.crube/` とその中身だけです。

- 上位へ探しても `.crube/` がなければ、指定ディレクトリに `.crube/` を作り、既定の `ignore` と空の `structure.txt` を置く。
- `CRUBE_LANG` が英語以外なら、その言語の空ファイル `structure_<言語タグ>.txt` も作る。
- すでに `.crube/` がある場合は、対象言語のファイルがなければそれだけを作る。言語ファイルを追加するときはこの使い方をする。
- 作るものが何もなければエラー(終了コード 2)。

```
$ CRUBE_LANG=ja crube init
created .crube/structure_ja.txt
```

## crube ls

```
crube ls [path] [--full] [--json]
```

`path`(省略時はカレントディレクトリ)の直下を、説明付きで一覧表示します。

```
$ crube ls src
auth/     認証(ログイン、セッション、トークン更新)
db/       ! 説明なし
main.go   エントリポイント
```

- 説明のないパスは `! 説明なし` と表示する。
- desc は端末幅に合わせて `…` で切り詰める。`--full` で全文を表示する。

## crube tree

```
crube tree [path] [-d N] [--json]
```

structure ファイルの内容を、ファイルと同じ入れ子の形で見やすく表示し、現在地に印を付けます。

```
$ cd src/auth
$ crube tree
  src/              アプリケーションのソースコード
>   auth/           認証(ログイン、セッション、トークン更新)
      session.ts    セッションの保存と有効期限の管理
      *_test.go     同名ファイルのテスト
  data/             データ格納場所
    *.csv           CSVデータ
```

- **表示するもの**: structure ファイルのエントリ(パターンもそのまま)。実際のファイルシステムは走査しない。実在するファイルの一覧は `ls`、説明の過不足は `status` で確認する。
- **インデント**: 1 段あたりスペース 2 つで表示する(ファイルのタブとは違い、端末で幅が一定になるように)。
- **desc の桁揃え**: desc の開始位置を、表示全体で揃える。名前の幅は表示幅(全角 = 2)で数える。
- **detail**: 表示しない。detail は `show` で見る。
- **desc なしの行**: 名前だけを表示する。
- **現在地の印**: 行頭の欄に `>` を付ける。付けるのは、現在のディレクトリに当たるエントリ。そのエントリがなければ、当たる中でいちばん深い祖先の行に付け、行末に `(here: src/db/)` のように現在のディレクトリを添える。プロジェクトルートにいるときは付けない。印の欄は、印がなくても常に 2 桁分空けておく。
- **長い desc**: 出力先が端末のときは、端末の幅に合わせて `…` で切り詰める。パイプやファイルへの出力では切り詰めない。
- `path` を指定すると、そのエントリの部分木だけを表示する。`-d` で深さを制限する(省略時は制限なし)。
- `--json` では、エントリを入れ子の JSON で出力する(desc、detail、現在地かどうかを含む)。

## crube find

```
crube find <query> [--limit N] [--json]
```

structure ファイルのエントリを、パス・desc・detail で検索します。結果はエントリのパスで表示します(パターンのエントリは `data/*.csv` のように表示)。

```
$ crube find セッション
src/auth/            認証(ログイン、セッション、トークン更新)
src/auth/session.ts  セッションの保存と有効期限の管理
```

## crube show

```
crube show <path> [--json]
```

実在する 1 パスについて、当たったエントリの desc と detail を表示します。パターンで当たった場合はそのエントリのパス(`data/*.csv` など)を、シンボリックリンクならリンク先も表示します。

## crube status

```
crube status [--require all|dirs] [--json]
```

実ディレクトリと structure を突き合わせ、未説明・孤立・不正な値を報告し、最後に充足度を表示します。報告するだけで、何も書き換えません。

| 表示 | 意味 |
| --- | --- |
| `untracked` | 未説明(`dir` / `file` の種別付き) |
| `orphan` | 孤立 |
| `invalid` | desc / detail が上限を超えている |

```
$ crube status
untracked   dir   src/db/
untracked   file  src/db/conn.go
orphan      src/legacy/

coverage    all: no (2)   dirs: no (1)
```

| オプション | 内容 |
| --- | --- |
| `--require all\|dirs` | 指定した充足度を満たさなければ終了コード 1 にする。省略時は充足度は終了コードに影響しない |
| `--json` | 各項目のパス・種類を JSON で出力する |

判定の詳細は [design/freshness.md](../design/freshness.md) を参照してください。

## crube set

```
crube set <path> <desc> [--detail <text>]
```

エントリの説明を追加・更新します。

```
$ crube set data/ "データ格納場所"
$ crube set "data/*.csv" "CSVデータ"
```

- 書き込み先は、対象言語で使っているファイル。
- エントリがなければ、兄弟の末尾に追加する。途中の親がなければ、desc なしの行として作る。
- `--detail` の改行は、そのまま detail の複数行になる。`--detail ""` で detail を消す。
- desc が 256 桁を超える、改行を含む、または detail が 4096 桁を超える場合はエラー。

## crube rm

```
crube rm <path>
```

対象言語で使っているファイルから、エントリを削除します。ディレクトリのエントリなら、その下のエントリもすべて削除します。実際のファイルは削除しません。

## crube mv

```
crube mv <from> <to>
```

対象言語で使っているファイルの中で、エントリを別の位置へ移します。ディレクトリのエントリなら、その下のエントリも一緒に移ります。実際のファイルは移動しません。ファイルを移動したあとで、説明を新しいパスへ付け替えるために使います。

- 移動先の親がなければ、desc なしの行として作る。
- 移動先にすでにエントリがあればエラー。
- 移動したエントリは、新しい兄弟の末尾に置く。

## crube mcp

```
crube mcp [--allow-write]
```

stdio の MCP サーバーとして起動します。`--allow-write` を付けたときだけ、書き込み系ツールを公開します。詳しくは [mcp.md](mcp.md) を参照してください。

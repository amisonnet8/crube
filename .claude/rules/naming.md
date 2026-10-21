# 命名規則

<!-- TODO: Goコードの命名（パッケージ名・型名など）と、用語の英語の表記（docs に無いもの）を、決まったら書く -->

## 製品名

- **crube**。文中でもコマンドでも、すべて小文字で書く
- データの置き場は`.crube/`、環境変数は`CRUBE_LANG`
- structure ファイルの名前は`structure.txt`（英語）と`structure_<言語タグ>.txt`（言語タグは BCP 47。例：`structure_ja.txt`）。除外の設定は`.crube/ignore`

## 用語

出所は`docs/design/overview_ja.md`「用語」と`docs/reference/`。**用語を変えたり足したりしたら、docs とこの表を同じ変更の中で直す。** CLI・JSON での表記は、docs にあるものだけを書いている。

| 用語 | CLI・JSON での表記 | 意味 |
|---|---|---|
| structure | — | プロジェクト構成の説明データ全体。言語ごとのファイルで構成される |
| エントリ | — | structure ファイルの1行。名前（ファイル名・ディレクトリ名、またはパターン）と、説明（desc と detail）の組 |
| エントリのパス | — | ルートからの名前を`/`でつないだ、structure ファイル上のエントリの位置（`data/*.csv`）。`tree`・`set`・`rm`・`mv`に渡す。`ls`・`show`に渡すのは実在するパス |
| プロジェクトルート | — | `.crube/`を置いたディレクトリ |
| パターン | — | `*.csv`のように、同じディレクトリ直下の複数の名前にまとめて当たるエントリの名前（`path.Match`） |
| 対象言語 | `CRUBE_LANG` | 参照・編集に使う言語 |
| デフォルト言語 | — | 英語。`structure.txt`に書く |
| 未説明（未説明パス） | `untracked`（`dir`／`file`） | 実在するが、説明が付いていないパス |
| 孤立（孤立エントリ） | `orphan` | 説明はあるが、対応するパスが実在しないエントリ |
| 上限超え | `invalid` | desc／detail が上限を超えている |
| 充足度 | `coverage`（`all`／`dirs`） | 説明がどこまで揃っているか |

## コマンド・ツールの名前

- CLI：`init`・`ls`・`tree`・`find`・`show`・`status`・`set`・`rm`・`mv`・`mcp`（`docs/reference/cli.md`）。既存のコマンド名（ls / tree / find）に合わせる方針（`docs/design/overview.md`）
- MCP：参照・保守は CLI と同じ名前（`tree`・`ls`・`find`・`show`・`status`）。書き込みだけ`set_description`・`remove_description`・`move_description`（ファイル操作と誤解されないように。`docs/reference/mcp.md`）
- フィールド：`desc`・`detail`だけ（`docs/reference/file-format.md`）
